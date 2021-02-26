package model

// Device Manager is a base struct that various providers inherit
type DeviceManager struct {
	PendingReboot        bool // set when the device requires a reboot after running an upgrade
	UpdatesInstalled     bool // set when updates were installed on the device
	UpdatesAvailable     int  // -1 == no update lookup as yet,  0 == no updates available, 1 == updates available
	Device               *Device
	FirmwareUpdateConfig *FirmwareUpdateConfig
}

// New Device manager constructor
func NewDeviceManager(d *Device) *DeviceManager {
	return &DeviceManager{Device: d, UpdatesAvailable: -1}
}

type Component struct {
	ID                string                   `json:"id"`
	DeviceID          string                   `json:"device_id"`
	Serial            string                   `json:"serial"`
	Vendor            string                   `json:"vendor"`
	Type              string                   `json:"type"`
	Model             string                   `json:"model"`
	Name              string                   `json:"name"`
	Slug              string                   `json:"slug"`
	FirmwareInstalled string                   `json:"firmware_installed"` // The firmware revision installed
	FirmwareAvailable string                   `json:"firmware_available"` // The firmware revision available
	Metadata          map[string]string        `json:"metadata"`           // Additional metadata if any
	Oem               bool                     `json:"oem"`                // Component is an OEM component
	FirmwareManaged   bool                     `json:"firmware_managed"`   // Firmware on the component is managed/unmanaged
	Config            *ComponentFirmwareConfig `json:"config"`             // The component firmware config
}

// Component specific firmware config
// each of the fields enable targeting the configuration to specific components
type ComponentFirmwareConfig struct {
	Slug           string   `yaml:"slug"        json:"slug"` // component name
	Vendor         string   `yaml:"vendor"      json:"vendor"`
	Model          string   `yaml:"model"       json:"model"`
	Serial         string   `yaml:"serial"      json:"serial"`
	Updates        []string `yaml:"updates"     json:"updates"`
	Method         string   `yaml:"method"      json:"method"`
	VendorURI      string   `yaml:"vendorURI"   json:"vendorURI"`
	UpdateFileURL  string   `yaml:"updateFileURL"  json:"updateFileURL"`
	UpdateFileSHA1 string   `yaml:"updateFileSHA1" json:"updateFileSHA1"`
}

// The firmware update configuration applicable for the device
type FirmwareUpdateConfig struct {
	ID         string                     `yaml:"id"          json:"id"`         // fup specific firmware config ID
	UpdateEnv  string                     `yaml:"update_env"  json:"update_env"` // fup specific update environment - production/canary/vanguard
	Method     string                     `yaml:"method"      json:"method"`
	Updates    []string                   `yaml:"updates"     json:"updates"`
	Vendor     string                     `yaml:"vendor"      json:"vendor"`
	Components []*ComponentFirmwareConfig `yaml:"components" json:"components"`
}
