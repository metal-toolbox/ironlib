package model

// Component is a low level device component - before its classified into a device type (BMC{}, BIOS{}, NIC{})
type Component struct {
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
