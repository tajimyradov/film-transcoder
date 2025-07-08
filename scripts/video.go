package scripts

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/tajimyradov/transcoder/utils"
)

func getBitrate(resolution string) string {
	switch resolution {
	case "480":
		return "800k"
	case "1080":
		return "3000k"
	default:
		return "1000k"
	}
}

func TranscodeVideoHLS(input, resolution, outputDir string, width, height int, logFile *os.File) error {

	fullDir := fmt.Sprintf("%s/%sp", outputDir, resolution)
	if err := os.MkdirAll(fullDir, 0777); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	segmentPattern := fmt.Sprintf("%s/%s_%%03d.ts", fullDir, utils.GenerateRandomCode(4))
	outputM3U8 := fmt.Sprintf("%s/playlist.m3u8", fullDir)

	// args := []string{
	// 	"-i", input,
	// 	"-map", "0:v:0",
	// 	"-c:v", "libx264",
	// 	"-b:v", getBitrate(resolution),
	// 	"-vf", fmt.Sprintf("scale=w=%d:h=%d:force_original_aspect_ratio=decrease:force_divisible_by=2", width, height),
	// 	"-preset", "fast",
	// 	"-profile:v", "main",
	// 	"-g", "48", // GOP size = 2s (for 24fps)
	// 	"-keyint_min", "48",
	// 	"-sc_threshold", "0",
	// 	"-force_key_frames", "expr:gte(t,n_forced*6)", // Keyframes every 6 seconds
	// 	"-start_at_zero",
	// 	"-copyts",
	// 	"-hls_time", "10",
	// 	"-hls_playlist_type", "vod",
	// 	"-hls_segment_filename", segmentPattern, // e.g., "hls/480p/segment_%03d.ts"
	// 	outputM3U8,
	// }

	args := []string{
		"-i", input,
		"-map", "0:v:0",
		"-c:v", "libx264",
		"-b:v", getBitrate(resolution),
		"-vf", fmt.Sprintf("scale=w=%d:h=%d:force_original_aspect_ratio=decrease:force_divisible_by=2", width, height),
		"-preset", "fast",
		"-profile:v", "main",
		"-g", "48",
		"-keyint_min", "48",
		"-sc_threshold", "0",
		"-start_at_zero", // ensure starts from 0
		"-copyts",        // preserve timestamps
		"-vsync", "1",    // sync output frames
		"-hls_time", "10",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segmentPattern,
		outputM3U8,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	return cmd.Run()
}
