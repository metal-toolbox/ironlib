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

// UpdateOptions sets firmware update options for a device component
type UpdateOptions struct {
	AllowDowngrade    bool // Allow firmware to be downgraded
	InstallAll        bool // Install all available updates (vendor tooling like DSU fetches the updates and installs them)
	DownloadOnly      bool // Only download updates, skip install - Works with InstallAll (where updates are fetched by the vendor tooling)
	Serial            string
	Vendor            string
	Model             string
	Name              string
	Slug              string
	UpdateFile        string // Location of the UpdateFile to be installed
	InstallerVersion  string // The all available updates installer version (specific to dell DSU)
	RepositoryVersion string // The update repository version to activate when defined
	BaseURL           string // The BaseURL for the updates
}

type CreateVirtualDiskOptions struct {
	RaidMode        string
	PhysicalDiskIDs []uint
	Name            string
	BlockSize       uint
}

type DestroyVirtualDiskOptions struct {
	VirtualDiskID int
}
