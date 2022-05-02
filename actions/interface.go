package actions

import (
	"context"

	"github.com/packethost/ironlib/model"
)

// Utility interface provides methods to configure, collect and update
type Utility interface {
	BIOSConfiguror
	InventoryCollector
	Updater
}

type BIOSConfiguror interface {
	// GetBIOSConfiguration returns the BIOS configuration for the device
	// deviceModel is an optional parameter depending on the hardware variants
	GetBIOSConfiguration(ctx context.Context, deviceModel string) (map[string]string, error)
}

// Updater runs updates
type Updater interface {
	ApplyUpdate(ctx context.Context, updateFile, component string) error
}

// Collectors

// InventoryCollector populates device inventory
type InventoryCollector interface {
	Collect(ctx context.Context, device *model.Device) error
}

// DriveCollector returns drive inventory
type DriveCollector interface {
	Drives(ctx context.Context) ([]*model.Drive, error)
}

// NICCollector returns NIC inventory
type NICCollector interface {
	NICs(ctx context.Context) ([]*model.NIC, error)
}

// BMCCollector returns BMC inventory
type BMCCollector interface {
	BMC(ctx context.Context) (*model.BMC, error)
}

// CPLDCollector returns CPLD inventory
type CPLDCollector interface {
	CPLD(ctx context.Context) (*model.CPLD, error)
}

// BIOSCollector returns BIOS inventory
type BIOSCollector interface {
	BIOS(ctx context.Context) (*model.BIOS, error)
}

// StorageControllerCollect returns storage controllers inventory
type StorageControllerCollector interface {
	StorageControllers(ctx context.Context) ([]*model.StorageController, error)
}

// Updaters

// DriveUpdater updates drive firmware
type DriveUpdater interface {
	UpdateDrive(ctx context.Context, updateFile, modelNumber, serialNumber string) error
}

// NICUpdater updates NIC firmware
type NICUpdater interface {
	UpdateNIC(ctx context.Context, updateFile, modelNumber string) error
}

// BMCUpdater updates BMC firmware
type BMCUpdater interface {
	UpdateBMC(ctx context.Context, updateFile, modelNumber string) error
}

// CPLDUpdater updates CPLD firmware
type CPLDUpdater interface {
	UpdateCPLD() error
}

// BIOSUpdater updates BIOS firmware
type BIOSUpdater interface {
	UpdateBIOS(ctx context.Context, updateFile, modelNumber string) error
}

// StorageControllerUpdater updates storage controller firmware
type StorageControllerUpdater interface {
	UpdateStorageController() error
}
