package internal

import "errors"

const androidMirroringAppId = "674A0243"
const backdropAppId = "E8C28D3C"
const chromeMirroringAppId = "0F5096E8"

type Application struct {
	AppId       string
	DisplayName string
	Namespaces  []string
	SessionId   string
	StatusText  string
	TransportId string
}

type Device struct {
	Applications  map[string]*Application
	AvailableApps []string
	DeviceModel   string
	FriendlyName  string
	Id            string
}

func (device *Device) startApplication(appId string) error {
	for _, application := range device.Applications {
		if application.AppId == appId {
			return errors.New("application already started")
		}
	}

	namespaces := make([]string, 2)
	namespaces[0] = receiverNamespace
	namespaces[1] = debugNamespace

	var application Application
	switch appId {
	case androidMirroringAppId:
		application = Application{
			AppId:       androidMirroringAppId,
			DisplayName: "Android Mirroring",
			Namespaces:  namespaces,
			SessionId:   "835ff891-f76f-4a04-8618-a5dc95477075",
			StatusText:  "",
			TransportId: "web-5",
		}
		break
	case chromeMirroringAppId:
		application = Application{
			AppId:       chromeMirroringAppId,
			DisplayName: "Chrome Mirroring",
			Namespaces:  namespaces,
			SessionId:   "7E2FF513-CDF6-9A91-2B28-3E3DE7BAC174",
			StatusText:  "",
			TransportId: "web-5",
		}
		break
	default:
		return errors.New("unsupported app")
	}

	device.Applications[application.SessionId] = &application

	return nil
}

func NewDevice(deviceModel string, friendlyName string, id string) *Device {
	// Namespaces covered by the Backdrop appllication
	namespaces := make([]string, 3)
	namespaces[0] = receiverNamespace
	namespaces[1] = debugNamespace
	namespaces[2] = remotingNamespace

	// Create session for Backdrop application
	sessionId := "AD3DFC60-A6AE-4532-87AF-18504DA22607"
	applications := make(map[string]*Application, 1)
	applications[sessionId] = &Application{
		AppId:       backdropAppId,
		DisplayName: "Backdrop",
		Namespaces:  namespaces,
		SessionId:   sessionId,
		StatusText:  "",
		TransportId: "pid-22607",
	}

	// Allow clients to start Android or Chrome mirroring apps
	availableApps := make([]string, 2)
	availableApps[0] = androidMirroringAppId
	availableApps[1] = chromeMirroringAppId

	return &Device{
		Applications:  applications,
		AvailableApps: availableApps,
		DeviceModel:   deviceModel,
		FriendlyName:  friendlyName,
		Id:            id,
	}
}
