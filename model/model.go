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
	Oem               bool              `json:"oem"` // Component is an OEM component
	Serial            string            `json:"serial"`
	Vendor            string            `json:"vendor"`
	Type              string            `json:"type"`
	Model             string            `json:"model"`
	Name              string            `json:"name"`
	Slug              string            `json:"slug"`
	FirmwareInstalled string            `json:"firmware_installed"` // The firmware revision installed
	FirmwareAvailable string            `json:"firmware_available"` // The firmware revision available
	FirmwareManaged   bool              `json:"firmware_managed"`   // Firmware on the component is managed/unmanaged
	Metadata          map[string]string `json:"metadata"`           // Additional metadata if any
}

type FirmwareInventory struct {
	Component
}
