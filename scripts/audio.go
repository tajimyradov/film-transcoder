package scripts

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"transcode/models"
	"transcode/repository"
	"transcode/utils"
)

func TranscodeAudio(input string, stream models.Stream, outputDir string, filmsRepo *repository.FilmsRepository) error {

	handlerName := stream.Tags.HandlerName
	arr := strings.Split(handlerName, "-")

	if len(arr[2]) != 3 {
		log.Fatalf("Invalid handler: %s", handlerName)
	}

	folder := ""

	if arr[0] == "1" {
		folder = "audio/original/" + arr[2]

		hlsFolder := fmt.Sprintf("%s/%s", outputDir, folder)
		if err := os.MkdirAll(hlsFolder, 0777); err != nil {
			log.Fatalf("Failed to create HLS folder: %v", err)
		}

	}

	if arr[0] == "2" {
		folder = "audio/dubbings/" + arr[2]

		hlsFolder := fmt.Sprintf("%s/%s", outputDir, folder)
		if err := os.MkdirAll(hlsFolder, 0777); err != nil {
			log.Fatalf("Failed to create HLS folder: %v", err)
		}

	}

	if arr[0] == "3" {
		studio, err := filmsRepo.GetStudioByID(arr[1])

		if err != nil {
			log.Fatalf(`Failed to get studio by id "%s"`, arr[2])
		}

		folder = "audio/studios/" + studio.Abbreviated + "/" + arr[2]

		hlsFolder := fmt.Sprintf("%s/%s", outputDir, folder)
		if err = os.MkdirAll(hlsFolder, 0777); err != nil {
			log.Fatalf("Failed to create HLS folder: %v", err)
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
	ffmpeg.Stdout = os.Stdout
	ffmpeg.Stderr = os.Stderr

	return ffmpeg.Run()
}
