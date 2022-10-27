package actions

import (
	"context"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/utils"
)

// Utility interface couples the configuration, collection and update interfaces
type Utility interface {
	BIOSConfiguror
	InventoryCollector
	Updater
}

// BIOSConfiguror defines an interface to collect BIOS configuration
type BIOSConfiguror interface {
	// GetBIOSConfiguration returns the BIOS configuration for the device
	// deviceModel is an optional parameter depending on the hardware variants
	GetBIOSConfiguration(ctx context.Context, deviceModel string) (map[string]string, error)
}

// Updater defines an interface to install an update file
type Updater interface {
	ApplyUpdate(ctx context.Context, updateFile, component string) error
}

// InventoryCollector defines an interface to collect all device inventory
type InventoryCollector interface {
	Collect(ctx context.Context, device *common.Device) error
}

// DriveCollector defines an interface to return disk drive inventory
type DriveCollector interface {
	Drives(ctx context.Context) ([]*common.Drive, error)
}

// DriveCapabilityCollector defines an interface to collect disk drive capability attributes
//
// The logicalName is the kernel/OS assigned drive name - /dev/sdX or /dev/nvmeX
type DriveCapabilityCollector interface {
	DriveCapabilities(ctx context.Context, logicalName string) ([]*common.Capability, error)
}

// NICCollector defines an interface to returns NIC inventory
type NICCollector interface {
	NICs(ctx context.Context) ([]*common.NIC, error)
}

// BMCCollector defines an interface to return BMC inventory
type BMCCollector interface {
	BMC(ctx context.Context) (*common.BMC, error)
}

// CPLDCollector defines an interface to return CPLD inventory
type CPLDCollector interface {
	CPLDs(ctx context.Context) ([]*common.CPLD, error)
}

// BIOSCollector defines an interface to return BIOS inventory
type BIOSCollector interface {
	BIOS(ctx context.Context) (*common.BIOS, error)
}

// StorageControllerCollector defines an interface to returns storage controllers inventory
type StorageControllerCollector interface {
	StorageControllers(ctx context.Context) ([]*common.StorageController, error)
}

// TPMCollector defines an interface to collect TPM device inventory
type TPMCollector interface {
	TPMs(ctx context.Context) ([]*common.TPM, error)
}

// Updaters

// DriveUpdater defines an interface to update drive firmware
type DriveUpdater interface {
	UpdateDrive(ctx context.Context, updateFile, modelNumber, serialNumber string) error
}

// NICUpdater defines an interface to update NIC firmware
type NICUpdater interface {
	UpdateNIC(ctx context.Context, updateFile, modelNumber string) error
}

// BMCUpdater defines an interface to update BMC firmware
type BMCUpdater interface {
	UpdateBMC(ctx context.Context, updateFile, modelNumber string) error
}

// CPLDUpdater defines an interface to update CPLD firmware
type CPLDUpdater interface {
	UpdateCPLD() error
}

// BIOSUpdater defines an interface to update BIOS firmware
type BIOSUpdater interface {
	UpdateBIOS(ctx context.Context, updateFile, modelNumber string) error
}

// StorageControllerUpdater defines an interface to update storage controller firmware
type StorageControllerUpdater interface {
	UpdateStorageController() error
}

// VirtualDiskCreator defines an interface to create virtual disks, generally via a StorageController
type VirtualDiskCreator interface {
	CreateVirtualDisk(ctx context.Context, raidMode string, physicalDisks []uint, name string, blockSize uint) error
}

// VirtualDiskCreator defines an interface to destroy virtual disks, generally via a StorageController
type VirtualDiskDestroyer interface {
	DestroyVirtualDisk(ctx context.Context, virtualDiskID int) error
}

type VirtualDiskManager interface {
	VirtualDiskCreator
	VirtualDiskDestroyer
	VirtualDisks(ctx context.Context) ([]*utils.MvcliDevice, error)
}
