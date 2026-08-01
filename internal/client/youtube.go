package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/tristanpenman/go-cast/internal/channel"
	"github.com/tristanpenman/go-cast/internal/common"
)

const youtubeNamespace = "urn:x-cast:com.google.youtube.mdx"

const youtubeBaseURL = "https://www.youtube.com/"

var youtubeLog = common.NewLogger("youtube")

var (
	youtubeSIDPattern      = regexp.MustCompile(`"c","([^"]+)",`)
	youtubeGSessionPattern = regexp.MustCompile(`"S","([^"]+)"\]`)
)

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

type youtubeMDXMessage struct {
	Type string `json:"type"`
	Data struct {
		ScreenID string `json:"screenId"`
	} `json:"data"`
}

// RequestYouTubeScreenID asks the running YouTube receiver for the screen ID
// required to establish a YouTube Lounge session.
func (s *Sender) RequestYouTubeScreenID(transportID string, timeout time.Duration) (string, error) {
	s.mu.Lock()
	s.youtubeScreenID = ""
	s.mu.Unlock()

	s.SendAppMessage(youtubeNamespace, transportID, `{"type":"getMdxSessionStatus"}`)
	deadline := time.Now().Add(timeout)
	timer := s.broadcastAtTimeout(timeout)
	defer timer.Stop()

	s.mu.Lock()
	defer s.mu.Unlock()
	for s.youtubeScreenID == "" {
		if err := s.waitErrorLocked(deadline, "YouTube screen ID"); err != nil {
			return "", err
		}
		s.cond.Wait()
	}
	return s.youtubeScreenID, nil
}

func (s *Sender) handleYouTubeMessage(castMessage *channel.CastMessage) {
	if castMessage.PayloadUtf8 == nil {
		return
	}
	var message youtubeMDXMessage
	if err := json.Unmarshal([]byte(*castMessage.PayloadUtf8), &message); err != nil {
		s.log.Warn("failed to parse YouTube MDX message", "err", err)
		return
	}
	if message.Type != "mdxSessionStatus" || message.Data.ScreenID == "" {
		return
	}
	s.mu.Lock()
	s.youtubeScreenID = message.Data.ScreenID
	s.cond.Broadcast()
	s.mu.Unlock()
}

// PlayYouTubeViaLounge binds to a YouTube receiver screen and starts a video.
func PlayYouTubeViaLounge(ctx context.Context, screenID, videoID string) error {
	youtubeLog.Info("starting YouTube Lounge playback", "screenId", screenID, "videoId", videoID)
	httpClient := &http.Client{Timeout: 15 * time.Second}
	if err := playYouTubeViaLounge(ctx, httpClient, youtubeBaseURL, screenID, videoID); err != nil {
		youtubeLog.Error("YouTube Lounge playback failed", "err", err)
		return err
	}
	youtubeLog.Info("YouTube Lounge playback command accepted", "videoId", videoID)
	return nil
}

func playYouTubeViaLounge(ctx context.Context, httpClient *http.Client, baseURL, screenID, videoID string) error {
	baseURL = strings.TrimRight(baseURL, "/") + "/"
	loungeToken, err := youtubeLoungeToken(ctx, httpClient, baseURL, screenID)
	if err != nil {
		return err
	}
	sid, gsessionID, err := youtubeLoungeBind(ctx, httpClient, baseURL, loungeToken)
	if err != nil {
		return err
	}
	return youtubeLoungeSetPlaylist(ctx, httpClient, baseURL, loungeToken, sid, gsessionID, videoID)
}

func youtubeLoungeToken(ctx context.Context, httpClient *http.Client, baseURL, screenID string) (string, error) {
	form := url.Values{"screen_ids": {screenID}}
	youtubeLog.Info("requesting YouTube lounge token", "screenId", screenID)
	response, err := youtubePost(ctx, httpClient, baseURL+"api/lounge/pairing/get_lounge_token_batch", nil, form, "")
	if err != nil {
		return "", fmt.Errorf("get YouTube lounge token: %w", err)
	}
	defer closeYouTubeResponseBody(response.Body)
	var payload struct {
		Screens []struct {
			LoungeToken string `json:"loungeToken"`
		} `json:"screens"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode YouTube lounge token: %w", err)
	}
	if len(payload.Screens) == 0 || payload.Screens[0].LoungeToken == "" {
		return "", errors.New("YouTube lounge token response is empty")
	}
	youtubeLog.Info("received YouTube lounge token", "screens", len(payload.Screens))
	return payload.Screens[0].LoungeToken, nil
}

func youtubeLoungeBind(ctx context.Context, httpClient *http.Client, baseURL, loungeToken string) (string, string, error) {
	params := url.Values{
		"RID": {"0"}, "VER": {"8"}, "CVER": {"1"},
		"TYPE": {"bind"}, "auth_failure_option": {"send_error"},
	}
	youtubeLog.Info("sending YouTube lounge bind request", "rid", "0", "mdxVersion", "3")
	response, err := youtubePost(ctx, httpClient, baseURL+"api/lounge/bc/bind", params, youtubeBindData(loungeToken), loungeToken)
	if err != nil {
		return "", "", fmt.Errorf("bind YouTube lounge session: %w", err)
	}
	defer closeYouTubeResponseBody(response.Body)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", fmt.Errorf("read YouTube lounge bind response: %w", err)
	}
	sidMatch := youtubeSIDPattern.FindSubmatch(body)
	gsessionMatch := youtubeGSessionPattern.FindSubmatch(body)
	if len(sidMatch) != 2 || len(gsessionMatch) != 2 {
		youtubeLog.Error("unexpected YouTube lounge bind response", "payload", truncateLogPayload(body))
		return "", "", errors.New("YouTube lounge bind response has no session identifiers")
	}
	youtubeLog.Info("bound YouTube lounge session", "payload", truncateLogPayload(body))
	return string(sidMatch[1]), string(gsessionMatch[1]), nil
}

func youtubeLoungeSetPlaylist(ctx context.Context, httpClient *http.Client, baseURL, loungeToken, sid, gsessionID, videoID string) error {
	params := url.Values{
		"SID": {sid}, "gsessionid": {gsessionID}, "RID": {"1"}, "VER": {"8"}, "CVER": {"1"},
		"v": {"2"}, "TYPE": {"bind"}, "t": {"1"}, "AID": {"0"}, "CI": {"0"},
		"name": {"GoCast Remote"}, "id": {"aaaaaaaaaaaaaaaaaaaaaaaaaa"}, "device": {"REMOTE_CONTROL"},
		"loungeIdToken": {loungeToken},
	}
	form := url.Values{
		"count":             {"1"},
		"ofs":               {"0"},
		"req0__sc":          {"setPlaylist"},
		"req0_listId":       {""},
		"req0_currentTime":  {"0"},
		"req0_currentIndex": {"-1"},
		"req0_audioOnly":    {"false"},
		"req0_videoId":      {videoID},
		"req0_params":       {""},
		"req0_playerParams": {""},
		"req0_prioritizeMobileSenderPlaybackStateOnConnection": {"true"},
	}
	youtubeLog.Info("sending YouTube setPlaylist request", "videoId", videoID, "rid", "1", "currentTime", "0")
	response, err := youtubePost(ctx, httpClient, baseURL+"api/lounge/bc/bind", params, form, loungeToken)
	if err != nil {
		return fmt.Errorf("set YouTube playlist: %w", err)
	}
	defer closeYouTubeResponseBody(response.Body)
	body, err := io.ReadAll(io.LimitReader(response.Body, 16*1024))
	if err != nil {
		return fmt.Errorf("read YouTube playlist response: %w", err)
	}
	youtubeLog.Info("YouTube setPlaylist response",
		"videoId", videoID,
		"status", response.Status,
		"payload", truncateLogPayload(body))
	return nil
}

func youtubeBindData(loungeToken string) url.Values {
	return url.Values{
		"app": {"web"}, "mdx-version": {"3"}, "name": {"GoCast Remote"},
		"id": {"aaaaaaaaaaaaaaaaaaaaaaaaaa"}, "device": {"REMOTE_CONTROL"},
		"capabilities": {"que,dsdtr,atp"}, "method": {"setPlaylist"},
		"magnaKey": {"cloudPairedDevice"}, "ui": {"false"}, "theme": {"cl"},
		"deviceContext": {"user_agent=dunno"}, "os_name": {"android"},
		"window_width_points": {""}, "window_height_points": {""}, "ms": {""},
		"loungeIdToken": {loungeToken},
	}
}

func closeYouTubeResponseBody(body io.Closer) {
	if err := body.Close(); err != nil {
		youtubeLog.Warn("failed to close YouTube response body", "err", err)
	}
}

func youtubePost(ctx context.Context, httpClient *http.Client, endpoint string, params, form url.Values, loungeToken string) (*http.Response, error) {
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Origin", youtubeBaseURL)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if loungeToken != "" {
		request.Header.Set("X-YouTube-LoungeId-Token", loungeToken)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		defer closeYouTubeResponseBody(response.Body)
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return nil, fmt.Errorf("YouTube returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return response, nil
}

func truncateLogPayload(payload []byte) string {
	const limit = 4096
	if len(payload) <= limit {
		return string(payload)
	}
	return string(payload[:limit]) + "... (truncated)"
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
