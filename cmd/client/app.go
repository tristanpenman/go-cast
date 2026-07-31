package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	castclient "github.com/tristanpenman/go-cast/internal/client"
	"github.com/tristanpenman/go-cast/internal/discovery"
)

const receiverTimeout = 10 * time.Second
const youtubeLaunchTimeout = 30 * time.Second
const youtubeAppID = "233637DE"

type knownApplication struct {
	ID   string
	Name string
}

var knownApplications = []knownApplication{
	{ID: youtubeAppID, Name: "YouTube"},
	{ID: "0F5096E8", Name: "Chrome mirroring"},
	{ID: "674A0243", Name: "Android mirroring"},
}

// DeviceApp is the frontend view of an application supported by a receiver.
type DeviceApp struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	StatusText string `json:"statusText"`
	Running    bool   `json:"running"`
}

// App contains the backend methods exposed to the Wails frontend.
type App struct {
	discover func(time.Duration) ([]discovery.Device, error)

	mu         sync.Mutex
	castClient *castclient.Client
	sender     *castclient.Sender
}

func NewApp() *App {
	return &App{discover: discovery.Discover}
}

// DiscoverDevices performs a bounded mDNS scan for Google Cast receivers.
func (a *App) DiscoverDevices() ([]discovery.Device, error) {
	return a.discover(5 * time.Second)
}

// SelectDevice connects to a receiver and returns its available applications.
func (a *App) SelectDevice(device discovery.Device) ([]DeviceApp, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if device.Host == "" || device.Port < 1 || device.Port > 65535 {
		return nil, fmt.Errorf("device has an invalid address")
	}

	a.closeLocked()
	castClient, err := castclient.NewClient(device.Host, uint(device.Port), true, nil)
	if err != nil {
		return nil, err
	}

	sender := castclient.NewSender(castClient, nil)
	a.castClient = castClient
	a.sender = sender

	sender.Connect()
	appIDs := make([]string, len(knownApplications))
	for index, app := range knownApplications {
		appIDs[index] = app.ID
	}
	sender.RequestAppAvailability(appIDs)
	sender.RequestStatus()

	if _, err := sender.WaitForAvailability(receiverTimeout); err != nil {
		a.closeLocked()
		return nil, fmt.Errorf("query available apps: %w", err)
	}
	if _, err := sender.WaitForStatus(receiverTimeout); err != nil {
		a.closeLocked()
		return nil, fmt.Errorf("query receiver status: %w", err)
	}

	return deviceApps(sender.Availability(), sender.Status()), nil
}

// LaunchApp launches an available application on the selected receiver.
func (a *App) LaunchApp(appID string) ([]DeviceApp, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.sender == nil {
		return nil, fmt.Errorf("no device selected")
	}
	if a.sender.Availability()[appID] != "APP_AVAILABLE" {
		return nil, fmt.Errorf("app %s is not available on this device", appID)
	}

	a.sender.LaunchApp(appID)
	if _, err := a.sender.WaitForApp(appID, receiverTimeout); err != nil {
		return nil, fmt.Errorf("launch app: %w", err)
	}
	return deviceApps(a.sender.Availability(), a.sender.Status()), nil
}

// PlayYouTube launches YouTube if needed, connects to its application
// transport, and asks it to play the supplied video URL.
func (a *App) PlayYouTube(rawURL string) ([]DeviceApp, error) {
	videoID, err := castclient.ParseYouTubeVideoID(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid YouTube URL: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.sender == nil {
		return nil, fmt.Errorf("no device selected")
	}

	// A ready-to-cast YouTube session may be marked as idle while retaining a
	// usable app transport. Reuse that transport instead of launching again.
	transportID := a.sender.SessionTransportID(youtubeAppID)
	if transportID == "" {
		if a.sender.Availability()[youtubeAppID] != "APP_AVAILABLE" {
			return nil, fmt.Errorf("YouTube is not available on this device")
		}
		a.sender.LaunchApp(youtubeAppID)
		// Some receivers take longer than the general control timeout to cold
		// start YouTube. Requesting status also covers devices that don't send an
		// unsolicited status update immediately after LAUNCH.
		a.sender.RequestStatus()
		transportID, err = a.sender.WaitForAppTransport(youtubeAppID, youtubeLaunchTimeout)
		if err != nil {
			return nil, fmt.Errorf("launch YouTube: %w", err)
		}
	}

	a.sender.ConnectTransport(transportID)
	screenID, err := a.sender.RequestYouTubeScreenID(transportID, receiverTimeout)
	if err != nil {
		return nil, fmt.Errorf("query YouTube screen: %w", err)
	}
	if err := castclient.PlayYouTubeViaLounge(context.Background(), screenID, videoID); err != nil {
		return nil, fmt.Errorf("start YouTube playback: %w", err)
	}
	return deviceApps(a.sender.Availability(), a.sender.Status()), nil
}

// TerminateApp stops a running application on the selected receiver.
func (a *App) TerminateApp(appID string) ([]DeviceApp, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.sender == nil {
		return nil, fmt.Errorf("no device selected")
	}

	status := a.sender.Status()
	if status == nil {
		return nil, fmt.Errorf("receiver status is unavailable")
	}
	for _, runningApp := range status.Applications {
		if runningApp.AppID != appID || runningApp.IsIdleScreen {
			continue
		}
		if runningApp.SessionID == "" {
			return nil, fmt.Errorf("running app %s has no session ID", appID)
		}
		a.sender.StopApp(runningApp.SessionID)
		if err := a.sender.WaitForAppStopped(appID, receiverTimeout); err != nil {
			return nil, fmt.Errorf("terminate app: %w", err)
		}
		return deviceApps(a.sender.Availability(), a.sender.Status()), nil
	}

	return nil, fmt.Errorf("app %s is not running", appID)
}

// shutdown closes any active receiver connection.
func (a *App) shutdown(context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closeLocked()
}

func (a *App) closeLocked() {
	if a.castClient != nil {
		_ = a.castClient.Close()
	}
	a.castClient = nil
	a.sender = nil
}

func deviceApps(availability map[string]string, status *castclient.ReceiverStatus) []DeviceApp {
	running := make(map[string]castclient.Application)
	if status != nil {
		for _, app := range status.Applications {
			if app.IsIdleScreen {
				continue
			}
			running[app.AppID] = app
		}
	}

	apps := make([]DeviceApp, 0, len(knownApplications)+len(running))
	for _, known := range knownApplications {
		runningApp, isRunning := running[known.ID]
		if availability[known.ID] != "APP_AVAILABLE" && !isRunning {
			continue
		}
		apps = append(apps, DeviceApp{
			ID:         known.ID,
			Name:       known.Name,
			StatusText: runningApp.StatusText,
			Running:    isRunning,
		})
		delete(running, known.ID)
	}

	unknownIDs := make([]string, 0, len(running))
	for appID := range running {
		unknownIDs = append(unknownIDs, appID)
	}
	sort.Strings(unknownIDs)
	for _, appID := range unknownIDs {
		runningApp := running[appID]
		name := runningApp.DisplayName
		if name == "" {
			name = runningApp.AppID
		}
		apps = append(apps, DeviceApp{
			ID:         runningApp.AppID,
			Name:       name,
			StatusText: runningApp.StatusText,
			Running:    true,
		})
	}
	return apps
}
