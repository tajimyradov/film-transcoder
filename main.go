package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
	"transcode/database"
	"transcode/models"
	"transcode/repository"
	"transcode/scripts"
)

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
		log.Fatal(err)
	}

	db, err := database.NewPostgresDB(appConfig.FilmsDB)
	if err != nil {
		log.Fatal(err)
	}

	filmsRepo := repository.NewFilmsRepository(db)

	input := "input.mp4"
	outputDir := "hls"

	if _, err := exec.LookPath("ffprobe"); err != nil {
		log.Fatal("ffprobe not found in PATH")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("ffmpeg not found in PATH")
	}

	if err = os.MkdirAll(outputDir, 0777); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("ffprobe failed: %v\n%s", err, string(out))
	}

	var result models.FFProbeOutput
	if err = json.Unmarshal(out, &result); err != nil {
		log.Fatalf("JSON parse error: %v", err)
	}

	wg := sync.WaitGroup{}

	for _, stream := range result.Streams {
		if stream.CodecType == "audio" {
			wg.Add(1)
			go func(s models.Stream) {
				defer wg.Done()
				if err = scripts.TranscodeAudio(input, s, outputDir, filmsRepo); err != nil {
					log.Fatalf("Transcoding audio failed: %v", err)
				}
			}(stream)

		}

		if stream.CodecType == "subtitle" {

			wg.Add(1)
			go func(s models.Stream) {
				defer wg.Done()
				if err = scripts.TranscodeSubtitle(input, s, outputDir); err != nil {
					log.Fatalf("Transcoding subtitle failed: %v", err)
				}
			}(stream)
		}

		if stream.CodecType == "video" {
			// Transcode to 480p
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err = scripts.TranscodeVideoHLS(input, "480", outputDir, 854, 480); err != nil {
					log.Fatalf("Transcoding to 480p failed: %v", err)
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				// Transcode to 1080p
				if err = scripts.TranscodeVideoHLS(input, "1080", outputDir, 1920, 1080); err != nil {
					log.Fatalf("Transcoding to 1080p failed: %v", err)
				}
			}()

		}

	}

	wg.Wait()

	err = scripts.ConstructMasterFile(result.Streams, langMap, filmsRepo, outputDir, "5128000", "1920x1080", "2564000", "854x480")
	if err != nil {
		log.Fatalf("Constructing master file failed: %v", err)
	}

}
