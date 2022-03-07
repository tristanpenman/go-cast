package internal

type Device struct {
	DeviceModel  string
	FriendlyName string
	Id           string
}

func NewDevice(deviceModel string, friendlyName string) *Device {
	return &Device{
		DeviceModel:  deviceModel,
		FriendlyName: friendlyName,
		Id:           "a98a257b4a3dd84392a34bd0",
	}
}
