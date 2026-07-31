package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDeviceIDCreatesAndReusesConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")

	first, err := resolveDeviceID("", configPath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := resolveDeviceID("", configPath)
	if err != nil {
		t.Fatal(err)
	}
	if second != first {
		t.Fatalf("device ID changed: first=%q second=%q", first, second)
	}

	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var config receiverConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		t.Fatal(err)
	}
	if config.DeviceID != first {
		t.Fatalf("persisted device ID %q, want %q", config.DeviceID, first)
	}
}

func TestResolveDeviceIDOptionOverridesConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{"deviceId":"not-a-uuid"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	const optionID = "550e8400-e29b-41d4-a716-446655440000"
	got, err := resolveDeviceID(optionID, configPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != optionID {
		t.Fatalf("device ID %q, want %q", got, optionID)
	}
}

func TestResolveDeviceIDRejectsInvalidOption(t *testing.T) {
	if _, err := resolveDeviceID("not-a-uuid", filepath.Join(t.TempDir(), "config.json")); err == nil {
		t.Fatal("expected invalid --device-id to fail")
	}
}

func TestResolveDeviceIDRejectsNilUUID(t *testing.T) {
	if _, err := resolveDeviceID("00000000-0000-0000-0000-000000000000", filepath.Join(t.TempDir(), "config.json")); err == nil {
		t.Fatal("expected nil --device-id to fail")
	}
}

func TestResolveDeviceIDRejectsInvalidConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(`{"deviceId":"not-a-uuid"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveDeviceID("", configPath); err == nil {
		t.Fatal("expected invalid config deviceId to fail")
	}
}
