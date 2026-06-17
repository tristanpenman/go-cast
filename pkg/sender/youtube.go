package sender

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

const youtubeNamespace = "urn:x-cast:com.google.youtube.mdx"

type youtubePayload struct {
	Type string      `json:"type"`
	Data youtubeData `json:"data"`
}

type youtubeData struct {
	VideoID     string  `json:"videoId"`
	CurrentTime float64 `json:"currentTime"`
	DoSeek      bool    `json:"doSeek"`
}

// FlingYouTubeVideo sends a flingVideo message on the YouTube namespace to the
// given transport, asking the YouTube app to play a video.
func (s *Sender) FlingYouTubeVideo(transportID, videoID string) {
	payload := youtubePayload{
		Type: "flingVideo",
		Data: youtubeData{
			VideoID:     videoID,
			CurrentTime: 0,
			DoSeek:      true,
		},
	}
	payloadBytes, _ := json.Marshal(payload)
	s.SendAppMessage(youtubeNamespace, transportID, string(payloadBytes))
}

// ParseYouTubeVideoID extracts a video ID from a common YouTube URL form.
func ParseYouTubeVideoID(rawURL string) (string, error) {
	if rawURL == "" {
		return "", errors.New("url is empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}

	host := strings.ToLower(parsedURL.Host)
	host = strings.TrimPrefix(host, "www.")

	switch host {
	case "youtu.be":
		videoID := strings.Trim(parsedURL.Path, "/")
		if videoID != "" {
			return videoID, nil
		}
	case "youtube.com", "m.youtube.com", "music.youtube.com":
		if id := parsedURL.Query().Get("v"); id != "" {
			return id, nil
		}

		cleanPath := path.Clean(parsedURL.Path)
		parts := strings.Split(strings.Trim(cleanPath, "/"), "/")
		if len(parts) == 2 {
			switch parts[0] {
			case "embed", "shorts", "live":
				if parts[1] != "" {
					return parts[1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("unsupported YouTube URL: %s", rawURL)
}
