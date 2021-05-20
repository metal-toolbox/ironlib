package model

import (
	"context"

	"github.com/packethost/ironlib/config"
)

type DeviceManager interface {
	Setter
	Getter
	Updater
}

type Setter interface {
	SetBIOSConfiguration(ctx context.Context, config *config.BIOSConfiguration) error
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
	GetInventory(ctx context.Context) (*Device, error)
	// List updates identifed by the vendor tooling (DSU for dells)
	ListUpdatesAvailable(ctx context.Context) (*Device, error)
	GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error)
}

type Updater interface {
	// ApplyAvailableUpdates invokes the vendor update tooling (DSU) for the given revision
	// and applies any and all available updates
	InstallAvailableUpdates(ctx context.Context, revision string, downloadOnly bool) error
	// InstallUpdate applies the given update
	InstallUpdate(ctx context.Context, slug string, updateFile string, downgrade bool) error
}
