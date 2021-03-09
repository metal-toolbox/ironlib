package model

func NewDevice() *Device {
	return &Device{Components: []*Component{}, ComponentUpdates: []*Component{}}
}

type Device struct {
	ID                string               `json:"id"`
	HardwareType      string               `json:"hardware_type,omitempty"`
	Vendor            string               `json:"vendor,omitempty"`
	Model             string               `json:"model,omitempty"`
	Serial            string               `json:"serial,omitempty"`
	Chassis           string               `json:"chassis,omitempty"`
	FirmwareVersion   string               `json:"firmware_revision,omitempty"` // The device firmware revision
	BIOS              *BIOS                `json:"bios,omitempty"`
	BMC               *BMC                 `json:"bmc,omitempty"`
	TPM               *TPM                 `json:"tpm,omitempty"`
	Mainboard         *Mainboard           `json:"mainboard,omitempty"`
	GPUs              []*GPU               `json:"gpu,omitempty"`
	CPUs              []*CPU               `json:"cpu,omitempty"`
	Memory            []*Memory            `json:"memory,omitempty"`
	NICs              []*NIC               `json:"nics,omitempty"`
	Drives            []*Drive             `json:"drives,omitempty"`
	StorageController []*StorageController `json:"storage_controller,omitempty"`
	// These fields are to be deprecated once the above fields are populated with firmware data
	Components       []*Component `json:"components"`
	ComponentUpdates []*Component `json:"component_updates"`
	Oem              bool         `json:"oem"`
}

type GPU struct {
}

type TPM struct {
}

type CPLD struct {
}

type BIOS struct {
	Description       string `json:"description,omitempty"`
	Vendor            string `json:"vendor,omitempty"`
	SizeBytes         int64  `json:"size_bytes,omitempty"`
	CapacityBytes     int64  `json:"capacity_bytes,omitempty"`
	FirmwareDate      string `json:"firmware_date,omitempty"`
	FirmwareInstalled string `json:"firmware_installed,omitempty"` // The firmware revision installed
	FirmwareAvailable string `json:"firmware_available,omitempty"` // The firmware revision available
	FirmwareManaged   bool   `json:"firmware_managed,omitempty"`   // Firmware on the component is managed/unmanaged
}

type BMC struct {
	Vendor            string `json:"vendor,omitempty"`
	FirmwareInstalled string `json:"firmware_installed,omitempty"` // The firmware revision installed
	FirmwareAvailable string `json:"firmware_available,omitempty"` // The firmware revision available
	FirmwareManaged   bool   `json:"firmware_managed,omitempty"`   // Firmware on the component is managed/unmanaged
}

type CPU struct {
	Description     string `json:"description,omitempty"`
	Vendor          string `json:"vendor,omitempty"`
	Model           string `json:"model,omitempty"`
	Serial          string `json:"serial,omitempty"`
	Slot            string `json:"slot,omitempty"`
	ClockSpeedHz    int64  `json:"clock_speeed_hz,omitempty"`
	Cores           int    `json:"cores,omitempty"`
	Threads         int    `json:"threads,omitempty"`
	FirmwareManaged bool   `json:"firmware_managed,omitempty"` // Firmware on the component is managed/unmanaged
}

type Memory struct {
	Description     string `json:"description,omitempty"`
	Slot            string `json:"slot,omitempty"`
	Type            string `json:"type,omitempty"`
	Vendor          string `json:"vendor,omitempty"`
	Model           string `json:"model,omitempty"`
	Serial          string `json:"serial,omitempty"`
	SizeBytes       int64  `json:"size_bytes,omitempty"`
	FormFactor      string `json:"form_factor,omitempty"`
	PartNumber      string `json:"part_number,omitempty"`
	ClockSpeedHz    int64  `json:"clock_speed_hz,omitempty"`
	FirmwareManaged bool   `json:"firmware_managed,omitempty"` // Firmware on the component is managed/unmanaged
}

type NIC struct {
	Description     string `json:"description,omitempty"`
	Vendor          string `json:"vendor,omitempty"`
	Model           string `json:"model,omitempty"`
	Serial          string `json:"serial,omitempty"`
	SpeedBits       int64  `json:"speed_bits,omitempty"`
	PhysicalID      string `json:"physid,omitempty"`
	FirmwareManaged bool   `json:"firmware_managed,omitempty"` // Firmware on the component is managed/unmanaged
}

type StorageController struct {
	Description     string `json:"description,omitempty"`
	Vendor          string `json:"vendor,omitempty"`
	Model           string `json:"model,omitempty"`
	Serial          string `json:"serial,omitempty"`
	PhysicalID      string `json:"physid,omitempty"`
	FirmwareManaged bool   `json:"firmware_managed,omitempty"` // Firmware on the component is managed/unmanaged
}

type Mainboard struct {
	Description     string `json:"description,omitempty"`
	Vendor          string `json:"vendor,omitempty"`
	Model           string `json:"model,omitempty"`
	Serial          string `json:"serial,omitempty"`
	PhysicalID      string `json:"physid,omitempty"`
	FirmwareManaged bool   `json:"firmware_managed,omitempty"` // Firmware on the component is managed/unmanaged
}

type Drive struct {
	Description       string            `json:"description,omitempty"`
	Serial            string            `json:"serial,omitempty"`
	StorageController string            `json:"storage_controller,omitempty"`
	Vendor            string            `json:"vendor,omitempty"`
	Model             string            `json:"model,omitempty"`
	Name              string            `json:"name,omitempty"`
	FirmwareInstalled string            `json:"firmware_installed,omitempty"` // The firmware revision installed
	FirmwareAvailable string            `json:"firmware_available,omitempty"` // The firmware revision available
	WWN               string            `json:"wwn,omitempty"`
	SizeBytes         int64             `json:"size_bytes,omitempty"`
	CapacityBytes     int64             `json:"capacity_bytes,omitempty"`
	BlockSize         int64             `json:"block_size,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`         // Additional metadata if any
	FirmwareManaged   bool              `json:"firmware_managed,omitempty"` // Firmware on the component is managed/unmanaged
	Oem               bool              `json:"oem,omitempty"`              // Component is an OEM component
}
