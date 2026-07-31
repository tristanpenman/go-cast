package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

const receiverConfigPath = "config.json"

type receiverConfig struct {
	DeviceID string `json:"deviceId"`
}

func resolveDeviceID(optionValue string, configPath string) (string, error) {
	if strings.TrimSpace(optionValue) != "" {
		return normalizeDeviceID(optionValue)
	}

	configBytes, err := os.ReadFile(configPath)
	if err == nil {
		var config receiverConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return "", fmt.Errorf("decode %s: %w", configPath, err)
		}
		if strings.TrimSpace(config.DeviceID) == "" {
			return "", fmt.Errorf("decode %s: deviceId is required", configPath)
		}
		deviceID, err := normalizeDeviceID(config.DeviceID)
		if err != nil {
			return "", fmt.Errorf("decode %s: %w", configPath, err)
		}
		return deviceID, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read %s: %w", configPath, err)
	}

	deviceID := uuid.NewString()
	configBytes, err = json.MarshalIndent(receiverConfig{DeviceID: deviceID}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode %s: %w", configPath, err)
	}
	configBytes = append(configBytes, '\n')

	// O_EXCL prevents two receivers starting concurrently from silently
	// replacing one another's identity.
	configFile, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return resolveDeviceID("", configPath)
		}
		return "", fmt.Errorf("create %s: %w", configPath, err)
	}
	if _, err := configFile.Write(configBytes); err != nil {
		_ = configFile.Close()
		_ = os.Remove(configPath)
		return "", fmt.Errorf("write %s: %w", configPath, err)
	}
	if err := configFile.Close(); err != nil {
		return "", fmt.Errorf("close %s: %w", configPath, err)
	}

	return deviceID, nil
}

func normalizeDeviceID(value string) (string, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("invalid device UUID %q: %w", value, err)
	}
	if parsed == uuid.Nil {
		return "", fmt.Errorf("invalid device UUID %q: UUID must not be nil", value)
	}
	return parsed.String(), nil
}
