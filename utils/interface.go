package utils

import (
	"context"

	"github.com/packethost/ironlib/config"
	"github.com/packethost/ironlib/model"
)

type BIOSConfiguror interface {
	GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error)
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
