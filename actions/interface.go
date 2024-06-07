package actions

import (
	"context"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

// DeviceManager interface is returned to the caller when calling ironlib.New()
type DeviceManager interface {
	Setter
	Getter
	Updater
}

// RaidController interface declares methods to manage virtual disks.
type RaidController interface {
	VirtualDiskCreator
	VirtualDiskDestroyer
}

// Setter interface declares methods to set attributes on a system.
type Setter interface {
	SetBIOSConfiguration(ctx context.Context, config map[string]string) error
}

// Getter interface declares methods implemented by providers to return various attributes.
type Getter interface {
	// Get device model
	GetModel() string
	// Get device vendor
	GetVendor() string
	// Check the device reboot required flag
	RebootRequired() bool
	// Check if any updates were applied
	UpdatesApplied() bool
	// Retrieve inventory for the device
	GetInventory(ctx context.Context, options ...Option) (*common.Device, error)
	// Retrieve inventory using the OEM tooling for the device,
	GetInventoryOEM(ctx context.Context, device *common.Device, options *model.UpdateOptions) error
	// List updates identifed by the vendor tooling (DSU for dells)
	ListAvailableUpdates(ctx context.Context, options *model.UpdateOptions) (*common.Device, error)
	// Retrieve BIOS configuration for device
	GetBIOSConfiguration(ctx context.Context) (map[string]string, error)
}

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

// UtilAttributeGetter defines methods to retrieve utility attributes.
type UtilAttributeGetter interface {
	Attributes() (utilName model.CollectorUtility, absolutePath string, err error)
}

// Updater defines an interface to install an update file
type Updater interface {
	// ApplyUpdate is to be deprecated in favor of InstallUpdates, its implementations are to be moved
	// to use InstallUpdates.
	ApplyUpdate(ctx context.Context, updateFile, component string) error

	// InstallUpdates installs updates based on the update options
	InstallUpdates(ctx context.Context, options *model.UpdateOptions) error
}

// InventoryCollector defines an interface to collect all device inventory
type InventoryCollector interface {
	UtilAttributeGetter
	Collect(ctx context.Context, device *common.Device) error
}

// DriveCollector defines an interface to return disk drive inventory
type DriveCollector interface {
	UtilAttributeGetter
	Drives(ctx context.Context) ([]*common.Drive, error)
}

// DriveCapabilityCollector defines an interface to collect disk drive capability attributes
//
// The logicalName is the kernel/OS assigned drive name - /dev/sdX or /dev/nvmeX
type DriveCapabilityCollector interface {
	UtilAttributeGetter
	DriveCapabilities(ctx context.Context, logicalName string) ([]*common.Capability, error)
}

// NICCollector defines an interface to returns NIC inventory
type NICCollector interface {
	UtilAttributeGetter
	NICs(ctx context.Context) ([]*common.NIC, error)
}

// BMCCollector defines an interface to return BMC inventory
type BMCCollector interface {
	UtilAttributeGetter
	BMC(ctx context.Context) (*common.BMC, error)
}

// CPLDCollector defines an interface to return CPLD inventory
type CPLDCollector interface {
	UtilAttributeGetter
	CPLDs(ctx context.Context) ([]*common.CPLD, error)
}

// BIOSCollector defines an interface to return BIOS inventory
type BIOSCollector interface {
	UtilAttributeGetter
	BIOS(ctx context.Context) (*common.BIOS, error)
}

// StorageControllerCollector defines an interface to returns storage controllers inventory
type StorageControllerCollector interface {
	UtilAttributeGetter
	StorageControllers(ctx context.Context) ([]*common.StorageController, error)
}

// TPMCollector defines an interface to collect TPM device inventory
type TPMCollector interface {
	UtilAttributeGetter
	TPMs(ctx context.Context) ([]*common.TPM, error)
}

// Checksum collectors

// FirmwareChecksumCollector defines an interface to collect firmware checksums
type FirmwareChecksumCollector interface {
	UtilAttributeGetter
	// return the sha-256 of the BIOS logo as a string, or the associated error
	BIOSLogoChecksum(ctx context.Context) (string, error)
}

// UEFIVarsCollector defines an interface to collect EFI variables
type UEFIVarsCollector interface {
	UtilAttributeGetter
	GetUEFIVars(ctx context.Context) (utils.UEFIVars, error)
}

// Updaters

// DriveUpdater defines an interface to update drive firmware
type DriveUpdater interface {
	UtilAttributeGetter
	UpdateDrive(ctx context.Context, updateFile, modelNumber, serialNumber string) error
}

// NICUpdater defines an interface to update NIC firmware
type NICUpdater interface {
	UtilAttributeGetter
	UpdateNIC(ctx context.Context, updateFile, modelNumber string) error
}

// BMCUpdater defines an interface to update BMC firmware
type BMCUpdater interface {
	UtilAttributeGetter
	UpdateBMC(ctx context.Context, updateFile, modelNumber string) error
}

// CPLDUpdater defines an interface to update CPLD firmware
type CPLDUpdater interface {
	UtilAttributeGetter
	UpdateCPLD() error
}

// BIOSUpdater defines an interface to update BIOS firmware
type BIOSUpdater interface {
	UtilAttributeGetter
	UpdateBIOS(ctx context.Context, updateFile, modelNumber string) error
}

// StorageControllerUpdater defines an interface to update storage controller firmware
type StorageControllerUpdater interface {
	UtilAttributeGetter
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

// DriveWiper defines an interface to override disk data
type DriveWiper interface {
	WipeDrive(ctx context.Context, log *logrus.Logger, logicalName string) error
}
