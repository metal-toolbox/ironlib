package utils

import (
	"context"

	"github.com/packethost/ironlib/model"
)

// InventoryCollector is to replace the Collector interface
type InventoryCollector interface {
	Inventory() (*model.Device, error)
}

type Collector interface {
	Components() ([]*model.Component, error)
}

type Updater interface {
	ApplyUpdate(ctx context.Context, updateFile, component string) error
}
