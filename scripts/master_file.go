package scripts

import (
	"fmt"
	"os"
	"strings"
	"transcode/models"
	"transcode/repository"
)

const audioTemplate = `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",NAME="%s",LANGUAGE="%s",AUTOSELECT=%s,DEFAULT=%s,URI="%s"`

const subtitlesTemplate = `#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subtitles",NAME="%s",LANGUAGE="%s",AUTOSELECT=%s,DEFAULT=%s,FORCED=%s,URI="%s"`

const videoTemplate = `
#EXT-X-STREAM-INF:BANDWIDTH=%s,RESOLUTION=%s,AUDIO="audio",SUBTITLES="subtitles"
%s`

// #EXT-X-STREAM-INF:BANDWIDTH=%s,RESOLUTION=%s,CODECS="avc1.42e01e,mp4a.40.2",AUDIO="audio",SUBTITLES="subtitles"

func ConstructMasterFile(streams []models.Stream, records map[string]models.Language, filmsRepo *repository.FilmsRepository, outputDir string, bandwidth1080 string, resolution1080 string, bandwidth480 string, resolution480 string) error {
	var audioLines, subtitleLines []string

	for _, stream := range streams {
		if stream.CodecType == "audio" {

			arr := strings.Split(stream.Tags.HandlerName, "-")
			autoSelect := "NO"

			defaultTrack := "NO"
			uri := "audio/"
			switch arr[0] {
			case "1":
				uri += "original/"
				autoSelect = "YES"
			case "2":
				uri += "dubbings/"
			case "3":
				uri += "studios/"
				studio, err := filmsRepo.GetStudioByID(arr[1])
				if err != nil {
					return fmt.Errorf("error getting studio by ID %s: %v", arr[1], err)
				}
				uri += studio.Abbreviated + "/"

			}
			uri += arr[2] + "/audio.m3u8"

			audioLine := fmt.Sprintf(audioTemplate, records[arr[2]].Name, records[arr[2]].Alpha2, autoSelect, defaultTrack, uri)
			audioLines = append(audioLines, audioLine)

		} else if stream.CodecType == "subtitle" {
			autoSelect := "NO"
			if stream.Tags.HandlerName == "eng" {
				autoSelect = "YES"
			}
			defaultTrack := "NO"
			forced := "NO"
			uri := fmt.Sprintf("subtitle/%s/subtitle.m3u8", stream.Tags.HandlerName)
			subtitleLine := fmt.Sprintf(subtitlesTemplate, records[stream.Tags.HandlerName].Name, records[stream.Tags.HandlerName].Alpha2, autoSelect, defaultTrack, forced, uri)

			subtitleLines = append(subtitleLines, subtitleLine)
		}

	}

	file, err := os.Create(outputDir + "/master.m3u8")
	if err != nil {
		return fmt.Errorf("failed to create master.m3u8 file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString("#EXTM3U\n")
	if err != nil {
		return fmt.Errorf("failed to write header to master.m3u8: %w", err)
	}

	_, err = file.WriteString("# AUDIO\n")
	if err != nil {
		return fmt.Errorf("failed to write audio section header: %w", err)
	}
	for _, audioLine := range audioLines {
		if _, err := file.WriteString(audioLine + "\n"); err != nil {
			return fmt.Errorf("failed to write audio line: %w", err)
		}
	}

	_, err = file.WriteString("\n# SUBTITLES\n")
	if err != nil {
		return fmt.Errorf("failed to write subtitles section header: %w", err)
	}
	for _, subtitleLine := range subtitleLines {
		if _, err := file.WriteString(subtitleLine + "\n"); err != nil {
			return fmt.Errorf("failed to write subtitle line: %w", err)
		}
	}

	_, err = file.WriteString("# VIDEO")
	if err != nil {
		return fmt.Errorf("failed to write video section header: %w", err)
	}

	_, err = file.WriteString(fmt.Sprintf(videoTemplate, bandwidth480, resolution480, "480p/playlist.m3u8"))
	if err != nil {
		return fmt.Errorf("failed to write 480p video line: %w", err)
	}

	// _, err = file.WriteString(fmt.Sprintf(videoTemplate, bandwidth1080, resolution1080, "1080p/playlist.m3u8"))
	// if err != nil {
	// 	return fmt.Errorf("failed to write 1080p video line: %w", err)
	// }

	return nil
}
