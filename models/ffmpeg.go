package models

type Stream struct {
	Index     int    `json:"index"`
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Tags      struct {
		Language    string `json:"language"`
		HandlerName string `json:"handler_name"`
		Title       string `json:"title"`
	} `json:"tags"`
}

type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
}
