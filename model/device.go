package model

// Hardware is a base struct that various providers inherit
type Hardware struct {
	PendingReboot    bool // set when the device requires a reboot after running an upgrade
	UpdatesInstalled bool // set when updates were installed on the device
	UpdatesAvailable int  // -1 == no update lookup as yet,  0 == no updates available, 1 == updates available
	Device           *Device
}

// NewHardware returns the base Hardware struct that various providers inherit
func NewHardware(d *Device) *Hardware {
	return &Hardware{Device: d, UpdatesAvailable: -1}
}

func NewDevice() *Device {
	return &Device{
		BMC:                &BMC{},
		BIOS:               &BIOS{},
		Mainboard:          &Mainboard{},
		TPM:                &TPM{},
		CPLD:               &CPLD{},
		PSUs:               []*PSU{},
		NICs:               []*NIC{},
		GPUs:               []*GPU{},
		CPUs:               []*CPU{},
		Memory:             []*Memory{},
		Drives:             []*Drive{},
		StorageControllers: []*StorageController{},
	}
}

type Device struct {
	HardwareType       string               `json:"hardware_type,omitempty"`
	Vendor             string               `json:"vendor,omitempty"`
	Model              string               `json:"model,omitempty"`
	Serial             string               `json:"serial,omitempty"`
	Chassis            string               `json:"chassis,omitempty"`
	BIOS               *BIOS                `json:"bios,omitempty"`
	BMC                *BMC                 `json:"bmc,omitempty"`
	TPM                *TPM                 `json:"tpm,omitempty"`
	Mainboard          *Mainboard           `json:"mainboard,omitempty"`
	CPLD               *CPLD                `json:"cpld"`
	GPUs               []*GPU               `json:"gpu,omitempty"`
	CPUs               []*CPU               `json:"cpu,omitempty"`
	Memory             []*Memory            `json:"memory,omitempty"`
	NICs               []*NIC               `json:"nics,omitempty"`
	Drives             []*Drive             `json:"drives,omitempty"`
	StorageControllers []*StorageController `json:"storage_controller,omitempty"`
	Oem                bool                 `json:"oem"`
	OemComponents      *OemComponents       `json:"oem_components,omitempty"`
	PSUs               []*PSU               `json:"power_supply,omitempty"`
}

// Firmware struct holds firmware attributes of a device component
type Firmware struct {
	Available string            `json:"available,omitempty"`
	Installed string            `json:"installed,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewFirmwareObj returns a *Firmware object
func NewFirmwareObj() *Firmware {
	return &Firmware{Metadata: make(map[string]string)}
}

type GPU struct {
}

type TPM struct {
}

type CPLD struct {
	Description string    `json:"description,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	Model       string    `json:"model,omitempty"`
	Serial      string    `json:"serial,omitempty"`
	Firmware    *Firmware `json:"firmware,omitempty"`
}

type PSU struct {
	Description string    `json:"description,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	Model       string    `json:"model,omitempty"`
	Serial      string    `json:"serial,omitempty"`
	Oem         bool      `json:"oem"`
	Firmware    *Firmware `json:"firmware,omitempty"`
}

// BIOS component
type BIOS struct {
	Description   string    `json:"description,omitempty"`
	Vendor        string    `json:"vendor,omitempty"`
	SizeBytes     int64     `json:"size_bytes,omitempty"`
	CapacityBytes int64     `json:"capacity_bytes,omitempty" diff:"immutable"`
	Firmware      *Firmware `json:"firmware,omitempty"`
}

// BMC component
type BMC struct {
	Description string    `json:"description,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	Firmware    *Firmware `json:"firmware,omitempty"`
}

// CPU component
type CPU struct {
	Description  string    `json:"description,omitempty"`
	Vendor       string    `json:"vendor,omitempty"`
	Model        string    `json:"model,omitempty"`
	Serial       string    `json:"serial,omitempty"`
	Slot         string    `json:"slot,omitempty"`
	ClockSpeedHz int64     `json:"clock_speeed_hz,omitempty"`
	Cores        int       `json:"cores,omitempty"`
	Threads      int       `json:"threads,omitempty"`
	Firmware     *Firmware `json:"firmware,omitempty"`
}

type Memory struct {
	Description  string    `json:"description,omitempty"`
	Slot         string    `json:"slot,omitempty"`
	Type         string    `json:"type,omitempty"`
	Vendor       string    `json:"vendor,omitempty"`
	Model        string    `json:"model,omitempty"`
	Serial       string    `json:"serial,omitempty"`
	SizeBytes    int64     `json:"size_bytes,omitempty"`
	FormFactor   string    `json:"form_factor,omitempty"`
	PartNumber   string    `json:"part_number,omitempty"`
	ClockSpeedHz int64     `json:"clock_speed_hz,omitempty"`
	Firmware     *Firmware `json:"firmware,omitempty"`
}

type NIC struct {
	Description string            `json:"description,omitempty"`
	Vendor      string            `json:"vendor,omitempty"`
	Model       string            `json:"model,omitempty"`
	Serial      string            `json:"serial,omitempty" diff:"identifier"`
	SpeedBits   int64             `json:"speed_bits,omitempty"`
	PhysicalID  string            `json:"physid,omitempty"`
	Oem         bool              `json:"oem"`
	Metadata    map[string]string `json:"metadata"`
	Firmware    *Firmware         `json:"firmware,omitempty"`
}

type StorageController struct {
	Description string            `json:"description,omitempty"`
	Vendor      string            `json:"vendor,omitempty"`
	Model       string            `json:"model,omitempty"`
	Serial      string            `json:"serial,omitempty"`
	Interface   string            `json:"interface,omitempty"` // SATA | SAS
	PhysicalID  string            `json:"physid,omitempty"`
	Oem         bool              `json:"oem"`
	Metadata    map[string]string `json:"metadata"`
	Firmware    *Firmware         `json:"firmware,omitempty"`
}

type Mainboard struct {
	ProductName string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	Model       string    `json:"model,omitempty"`
	Serial      string    `json:"serial,omitempty"`
	PhysicalID  string    `json:"physid,omitempty"`
	Firmware    *Firmware `json:"firmware,omitempty"`
}

type Drive struct {
	ProductName       string            `json:"name,omitempty"`
	Type              string            `json:"drive_type,omitempty"`
	Description       string            `json:"description,omitempty"`
	Serial            string            `json:"serial,omitempty" diff:"identifier"`
	StorageController string            `json:"storage_controller,omitempty"`
	Vendor            string            `json:"vendor,omitempty"`
	Model             string            `json:"model,omitempty"`
	WWN               string            `json:"wwn,omitempty"`
	SizeBytes         int64             `json:"size_bytes,omitempty"`
	CapacityBytes     int64             `json:"capacity_bytes,omitempty"`
	BlockSize         int64             `json:"block_size,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"` // Additional metadata if any
	Oem               bool              `json:"oem,omitempty"`      // Component is an OEM component
	Firmware          *Firmware         `json:"firmware,omitempty"`
	SmartStatus       string            `json:"smart_status,omitempty"`
}
