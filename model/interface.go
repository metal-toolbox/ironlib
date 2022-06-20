package model

import (
	"context"

	"github.com/bmc-toolbox/common"
)

type DeviceManager interface {
	Setter
	Getter
	Updater
}

type Setter interface {
	SetBIOSConfiguration(ctx context.Context, config map[string]string) error
}

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
	GetInventory(ctx context.Context) (*common.Device, error)
	// Retrieve inventory using the OEM tooling for the device,
	GetInventoryOEM(ctx context.Context, device *common.Device, options *UpdateOptions) error
	// List updates identifed by the vendor tooling (DSU for dells)
	ListAvailableUpdates(ctx context.Context, options *UpdateOptions) (*common.Device, error)
	// Retrieve BIOS configuration for device
	GetBIOSConfiguration(ctx context.Context) (map[string]string, error)
}

type Updater interface {
	// InstallUpdates installs updates based on the update options
	InstallUpdates(ctx context.Context, options *UpdateOptions) error
}
