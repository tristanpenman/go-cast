package main

import (
	"flag"
	"os"
	"sync"
	"time"

	"github.com/tristanpenman/go-cast/internal"
	"github.com/tristanpenman/go-cast/internal/sender"
)

const (
	mirroringAppID = "0F5096E8"
	youtubeAppID   = "233637DE"
)

var log = internal.NewLogger("main")

func main() {
	var disableChallenge = flag.Bool("disable-challenge", false, "disable auth challenge")
	var hostname = flag.String("hostname", "", "receiver address or hostname")
	var port = flag.Uint("port", 8009, "receiver port")
	var appID = flag.String("app-id", "", "Chromecast app ID to launch")
	var videoPath = flag.String("video-path", "", "path to local video file; launches Chromecast mirroring app")
	var youtubeURL = flag.String("youtube-url", "", "YouTube video URL to play via the Chromecast YouTube app")

	flag.Parse()

	if *hostname == "" {
		flag.PrintDefaults()
		return
	}

	effectiveAppID := *appID
	youtubeVideoID := ""
	if *videoPath != "" {
		if _, err := os.Stat(*videoPath); err != nil {
			log.Error("video path is invalid", "path", *videoPath, "err", err)
			return
		}
		effectiveAppID = mirroringAppID
	}

	if *youtubeURL != "" {
		if *appID != "" || *videoPath != "" {
			log.Error("--youtube-url cannot be combined with --app-id or --video-path")
			return
		}

		videoID, err := sender.ParseYouTubeVideoID(*youtubeURL)
		if err != nil {
			log.Error("youtube URL is invalid", "youtube-url", *youtubeURL, "err", err)
			return
		}

		youtubeVideoID = videoID
		effectiveAppID = youtubeAppID
	}

	if effectiveAppID == "" {
		log.Error("one of --app-id, --video-path, or --youtube-url must be provided")
		return
	}

	log.Info("args",
		"disable-challenge", *disableChallenge,
		"hostname", *hostname,
		"port", *port,
		"app-id", effectiveAppID,
		"video-path", *videoPath,
		"youtube-url", *youtubeURL)

	var wg sync.WaitGroup
	wg.Add(1)

	client := internal.NewClient(*hostname, *port, !*disableChallenge, &wg)
	if client == nil {
		return
	}

	s := sender.New(client, log)
	s.Connect()
	s.RequestStatus()
	s.LaunchApp(effectiveAppID)

	transportID, err := s.WaitForApp(effectiveAppID, 10*time.Second)
	if err != nil {
		log.Error("failed to launch app", "err", err)
		_ = client.Close()
		wg.Wait()
		return
	}

	s.ConnectTransport(transportID)
	log.Info("app launched", "app-id", effectiveAppID, "transport-id", transportID)

	if *videoPath != "" {
		log.Warn("video-path support currently launches the Chromecast mirroring app only; stream upload is not implemented yet", "video-path", *videoPath)
	}

	if youtubeVideoID != "" {
		s.FlingYouTubeVideo(transportID, youtubeVideoID)
		log.Info("youtube video queued", "video-id", youtubeVideoID)
		time.Sleep(500 * time.Millisecond)
	}

	_ = client.Close()
	wg.Wait()
}
