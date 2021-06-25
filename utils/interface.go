package utils

import (
	"context"

	"github.com/packethost/ironlib/config"
	"github.com/packethost/ironlib/model"
)

// Utility interface provides methods to configure, collect and update
type Utility interface {
	BIOSConfiguror
	InventoryCollector
	Collector
	Updater
}

type BIOSConfiguror interface {
	// GetBIOSConfiguration returns the BIOS configuration for the device
	// deviceModel is an optional parameter depending on the hardware variants
	GetBIOSConfiguration(ctx context.Context, deviceModel string) (*config.BIOSConfiguration, error)
}

// InventoryCollector returns device inventory
type InventoryCollector interface {
	Inventory() (*model.Device, error)
}

// Collector returns device components
type Collector interface {
	Components() ([]*model.Component, error)
}

// Updater runs updates
type Updater interface {
	ApplyUpdate(ctx context.Context, updateFile, component string) error
}
