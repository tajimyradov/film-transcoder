package scripts

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/tajimyradov/transcoder/models"
	"github.com/tajimyradov/transcoder/repository"
	"github.com/tajimyradov/transcoder/utils"
)

func TranscodeAudio(input string, stream models.Stream, outputDir string, filmsRepo *repository.VideosRepository, logFile *os.File) error {

	handlerName := stream.Tags.HandlerName
	arr := strings.Split(handlerName, "-")

	if len(arr[2]) != 3 {
		log.Printf("Invalid handler: %s", handlerName)
		return errors.New("invalid handler")
	}

	folder := ""

	if arr[0] == "1" {
		folder = "audio/original/" + arr[2]

		hlsFolder := fmt.Sprintf("%s/%s", outputDir, folder)
		if err := os.MkdirAll(hlsFolder, 0777); err != nil {
			log.Printf("Failed to create HLS folder: %v", err)
			return err
		}

	}

	if arr[0] == "2" {
		folder = "audio/dubbings/" + arr[2]

		hlsFolder := fmt.Sprintf("%s/%s", outputDir, folder)
		if err := os.MkdirAll(hlsFolder, 0777); err != nil {
			log.Printf("Failed to create HLS folder: %v", err)
			return err
		}

	}

	if arr[0] == "3" {
		studio, err := filmsRepo.GetStudioByID(arr[1])

		if err != nil {
			log.Printf(`Failed to get studio by id "%s"`, arr[1])
			return err
		}

		folder = "audio/studios/" + studio.Abbreviated + "/" + arr[2]

		hlsFolder := fmt.Sprintf("%s/%s", outputDir, folder)
		if err = os.MkdirAll(hlsFolder, 0777); err != nil {
			log.Printf("Failed to create HLS folder: %v", err)
			return err
		}

	}

	outputM3U8 := fmt.Sprintf("%s/%s/audio.m3u8", outputDir, folder)
	segmentPattern := fmt.Sprintf("%s/%s/%s_%%03d.aac", outputDir, folder, utils.GenerateRandomCode(4))

	// args := []string{
	// 	"-i", input,
	// 	"-map", fmt.Sprintf("0:%d", stream.Index),
	// 	"-c:a", "aac",
	// 	"-b:a", "192k",
	// 	"-vn",
	// 	"-start_at_zero",
	// 	"-copyts",
	// 	"-hls_time", "10",
	// 	"-hls_playlist_type", "vod",
	// 	"-hls_segment_filename", segmentPattern, // e.g., "hls_audio/eng/segment_%03d.aac"
	// 	outputM3U8,
	// }

	args := []string{
		"-i", input,
		"-map", fmt.Sprintf("0:%d", stream.Index),
		"-c:a", "aac",
		"-b:a", "192k",
		"-vn",
		"-start_at_zero",
		"-copyts",
		"-vsync", "1", // ensures alignment
		"-hls_time", "10",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segmentPattern,
		outputM3U8,
	}

	ffmpeg := exec.Command("ffmpeg", args...)
	ffmpeg.Stdout = logFile
	ffmpeg.Stderr = logFile

	return ffmpeg.Run()
}
