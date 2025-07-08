package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tajimyradov/transcode/broker"
	"github.com/tajimyradov/transcode/database"
	"github.com/tajimyradov/transcode/models"
	"github.com/tajimyradov/transcode/repository"
	"github.com/tajimyradov/transcode/scripts"
	"github.com/tajimyradov/transcode/storage"
)

const ParentPath = "media"

func main() {
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

	fmt.Printf("Detected %d CPU cores.\n", numCPUs)
	fmt.Printf("GOMAXPROCS is set to %d.\n", runtime.GOMAXPROCS(-1))
	fmt.Println("-------------------------------")
	fmt.Println("Starting transcoding process...")

	file, err := os.Open("codes.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	langMap := make(map[string]models.Language)
	for _, row := range records {
		if len(row) < 2 {
			continue
		}
		langMap[row[0]] = models.Language{
			Alpha3: row[0],
			Alpha2: row[1],
			Name:   row[2],
		}
	}

	rand.NewSource(time.Now().UnixNano())

	appConfig, err := models.NewAppConfig("configs/local.yaml")
	if err != nil {
		log.Fatal("load config error: ", err)
	}

	db, err := database.NewPostgresDB(appConfig.FilmsDB)
	if err != nil {
		log.Fatal("postgres error: ", err)
	}

	consumer, err := broker.NewRabbitMQ(appConfig.RabbitMQ)
	if err != nil {
		log.Fatalf("Error initializing RabbitMQ: %v", err)
	}
	defer consumer.Connection.Close()
	defer consumer.Channel.Close()

	minio, err := storage.NewMinio(&appConfig.Minio)
	if err != nil {
		log.Fatal("minio error: ", err)
	}

	filmsRepo := repository.NewVideosRepository(db)

	if _, err := exec.LookPath("ffprobe"); err != nil {
		log.Fatal("ffprobe not found in PATH")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("ffmpeg not found in PATH")
	}

	go func() {
		for msg := range consumer.Messages {
			var m models.TranscodeRequstBody
			err := json.Unmarshal(msg.Body, &m)
			if err != nil {
				log.Printf("Failed to Unmarshal body: %v", err)
				_ = msg.Ack(false)
				continue
			}

			log.Println("Transcode start: ", m.Id)

			outputDir := ParentPath + "/" + strconv.Itoa(m.Id)

			err = os.RemoveAll(outputDir)
			if err != nil && !os.IsNotExist(err) {
				log.Printf(`Failed to clean working directory: %v`, err)
				_ = msg.Ack(false)
				continue
			}

			if err = os.MkdirAll(outputDir, 0777); err != nil {
				log.Printf("Failed to create output directory: %v", err)
				_ = msg.Ack(false)
				continue
			}

			logFile, err := os.OpenFile(outputDir+"/logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Failed to create log file: %v", err)
				_ = msg.Ack(false)
				continue
			}

			log.SetOutput(logFile)
			log.SetPrefix("Transcode: ")
			log.SetFlags(log.LstdFlags | log.Lshortfile)

			originalFilePath, err := filmsRepo.GetOriginalFileOfVideo(m.Id)
			if err != nil {
				log.Printf("Failed to get original file path from DB %v", err)
				_ = file.Close()
				_ = msg.Ack(false)
				continue
			}

			input := outputDir + "/" + strings.Split(originalFilePath, "/")[len(strings.Split(originalFilePath, "/"))-1]

			err = scripts.DownloadLargeObject(minio, appConfig.Minio.OrigianlFileBucket, originalFilePath, input)
			if err != nil {
				log.Printf("Failed to create output directory: %v", err)
				_ = file.Close()
				_ = msg.Ack(false)
				continue
			}

			cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", input)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("ffprobe failed: %v\n%s", err, string(out))
				_ = file.Close()
				_ = msg.Ack(false)
				continue
			}

			var result models.FFProbeOutput
			if err = json.Unmarshal(out, &result); err != nil {
				log.Printf("JSON parse error: %v", err)
				_ = file.Close()
				_ = msg.Ack(false)
				continue
			}

			fmt.Println(`Transcode started!!!`)
			for _, stream := range result.Streams {
				if stream.CodecType == "audio" {

					if err = scripts.TranscodeAudio(input, stream, outputDir, filmsRepo, logFile); err != nil {
						log.Printf("Transcoding audio failed: %v", err)
						_ = msg.Ack(false)
						break
					}

				}

				if stream.CodecType == "subtitle" {

					if err = scripts.TranscodeSubtitle(input, stream, outputDir, logFile); err != nil {
						log.Printf("Transcoding subtitle failed: %v", err)
						_ = msg.Ack(false)
						break
					}
				}

				if stream.CodecType == "video" {
					// Transcode to 480p

					if err = scripts.TranscodeVideoHLS(input, "480", outputDir, 854, 480, logFile); err != nil {
						log.Printf("Transcoding to 480p failed: %v", err)
						_ = msg.Ack(false)
						break
					}

					// Transcode to 1080p
					if err = scripts.TranscodeVideoHLS(input, "1080", outputDir, 1920, 1080, logFile); err != nil {
						log.Printf("Transcoding to 1080p failed: %v", err)
						_ = msg.Ack(false)
						break
					}

				}

			}

			if err != nil {
				err = scripts.ClearFiles(outputDir)
				if err != nil {
					log.Printf("Failed to remove files %v", err)
					fmt.Println("Transcode error")
				}
				_ = file.Close()
				_ = msg.Ack(false)
				continue
			}

			err = scripts.ConstructMasterFile(result.Streams, langMap, filmsRepo, outputDir, "5128000", "1920x1080", "2564000", "854x480")
			if err != nil {
				log.Printf("Constructing master file failed: %v", err)
				_ = file.Close()
				_ = msg.Ack(false)
				continue
			}

			err = os.Remove(input)
			if err != nil {
				log.Printf("Remove original file: %v", err)
				_ = file.Close()
				_ = msg.Ack(false)
				continue
			}

			fmt.Println(`Upload started!!!`)
			err = scripts.UploadFiles(outputDir, minio, *appConfig)
			if err != nil {
				log.Printf("Upload files failed: %v", err)
				_ = file.Close()
				_ = msg.Ack(false)
				continue
			}

			if err == nil {
				_ = msg.Ack(true)

				err = scripts.ClearFiles(outputDir)

				if err != nil {
					fmt.Printf(`Failed to clean files: %v`, err)
				}

				_ = file.Close()

				log.Println("Success")
				fmt.Println("Success")
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	fmt.Printf("Received signal: %v", sig)

}
