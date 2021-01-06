package model

type Device struct {
	ID               string       `json:"id"`
	HWType           string       `json:"hwtype"`
	Vendor           string       `json:"vendor"`
	Model            string       `json:"model"`
	Serial           string       `json:"serial"`
	FirmwareVersion  string       `json:"firmware_revision"` // The device firmware revision
	Components       []*Component `json:"components"`
	ComponentUpdates []*Component `json:"component_updates"`
	Oem              bool         `json:"oem"` // Device is an OEM device
}

type Component struct {
	ID                string            `json:"id"`
	DeviceID          string            `json:"device_id"`
	Serial            string            `json:"serial"`
	Vendor            string            `json:"vendor"`
	Type              string            `json:"type"`
	Model             string            `json:"model"`
	Name              string            `json:"name"`
	Slug              string            `json:"slug"`
	FirmwareInstalled string            `json:"firmware_installed"` // The firmware revision installed
	FirmwareAvailable string            `json:"firmware_available"` // The firmware revision available
	Metadata          map[string]string `json:"metadata"`           // Additional metadata if any
	Oem               bool              `json:"oem"`                // Component is an OEM component
	FirmwareManaged   bool              `json:"firmware_managed"`   // Firmware on the component is managed/unmanaged
}

// The firmware update configuration applicable for the device
type FirmwareUpdateConfig struct {
	UpdateEnv      string   `yaml:"update_env" json:"update_env"` // fup specific update environment - production/canary/vanguard
	Method         string   `yaml:"method"      json:"method"`
	DeviceType     string   `yaml:"deviceType"  json:"deviceType"`
	Updates        []string `yaml:"updates"     json:"updates"`
	Vendor         string   `yaml:"vendor"      json:"vendor"`
	VendorURI      string   `yaml:"vendorURI"   json:"vendorURI"`
	UpdateFileURL  string   `yaml:"updateFileURL"  json:"updateFileURL"`
	UpdateFileSHA1 string   `yaml:"updateFileSHA1" json:"updateFileSHA1"`
}
