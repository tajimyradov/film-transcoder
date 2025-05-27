package scripts

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"transcode/models"
	"transcode/utils"
)

func TranscodeSubtitle(input string, stream models.Stream, outputDir string) error {
	handlerName := stream.Tags.HandlerName

	folder := "subtitle/" + handlerName

	hlsFolder := fmt.Sprintf("%s/%s", outputDir, folder)
	if err := os.MkdirAll(hlsFolder, 0777); err != nil {
		return errors.New("Failed to create HLS folder: " + err.Error())
	}

	playlistPath := fmt.Sprintf("%s/%s/subtitle.m3u8", outputDir, folder)
	segmentPattern := fmt.Sprintf("%s/%s/%s_%%03d.vtt", outputDir, folder, utils.GenerateRandomCode(4))

	// args := []string{
	// 	"-i", input,
	// 	"-map", fmt.Sprintf("0:%d", stream.Index),
	// 	"-c:s", "webvtt",
	// 	"-copyts",
	// 	"-start_at_zero",
	// 	"-f", "segment",
	// 	"-segment_time", "10",
	// 	"-segment_list", playlistPath,
	// 	"-segment_list_type", "m3u8",

	// 	segmentPattern,
	// }

	args := []string{
		"-i", input,
		"-map", fmt.Sprintf("0:%d", stream.Index),
		"-c:s", "webvtt",
		"-start_at_zero",
		"-copyts", // critical
		"-f", "segment",
		"-segment_time", "10", // match audio/video
		"-segment_list", playlistPath,
		"-segment_list_type", "m3u8",
		segmentPattern,
	}

	ffmpeg := exec.Command("ffmpeg", args...)
	ffmpeg.Stdout = os.Stdout
	ffmpeg.Stderr = os.Stderr

	return ffmpeg.Run()
}
