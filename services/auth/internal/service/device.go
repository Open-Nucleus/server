package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
)

// DeviceRecord is the JSON document stored in Git for each device.
type DeviceRecord struct {
	DeviceID          string `json:"device_id"`
	PublicKey         string `json:"public_key"`
	PractitionerID   string `json:"practitioner_id"`
	SiteID           string `json:"site_id"`
	DeviceName       string `json:"device_name"`
	Role             string `json:"role"`
	Status           string `json:"status"`
	RegisteredAt     string `json:"registered_at"`
	RevokedAt        string `json:"revoked_at,omitempty"`
	RevokedBy        string `json:"revoked_by,omitempty"`
	RevocationReason string `json:"revocation_reason,omitempty"`
}

func deviceGitPath(devicesDir, deviceID string) string {
	return fmt.Sprintf("%s/%s.json", devicesDir, deviceID)
}

func loadDevice(git gitstore.Store, devicesDir, deviceID string) (*DeviceRecord, error) {
	data, err := git.Read(deviceGitPath(devicesDir, deviceID))
	if err != nil {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	var device DeviceRecord
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, fmt.Errorf("unmarshal device %s: %w", deviceID, err)
	}
	return &device, nil
}

func saveDevice(git gitstore.Store, devicesDir string, device *DeviceRecord, operation, author string) (string, error) {
	data, err := json.MarshalIndent(device, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal device: %w", err)
	}

	commit, err := git.WriteAndCommit(
		deviceGitPath(devicesDir, device.DeviceID),
		data,
		gitstore.CommitMessage{
			ResourceType: "Device",
			Operation:    operation,
			ResourceID:   device.DeviceID,
			Author:       author,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return "", fmt.Errorf("save device to git: %w", err)
	}
	return commit, nil
}

func listAllDevices(git gitstore.Store, devicesDir string) ([]*DeviceRecord, error) {
	var devices []*DeviceRecord

	err := git.TreeWalk(func(path string, data []byte) error {
		if len(path) <= len(devicesDir) {
			return nil
		}
		if path[:len(devicesDir)] != devicesDir {
			return nil
		}

		var device DeviceRecord
		if err := json.Unmarshal(data, &device); err != nil {
			return nil // skip malformed entries
		}
		devices = append(devices, &device)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return devices, nil
}
