package utils

import (
	"context"

	"github.com/packethost/ironlib/model"
)

type Collector interface {
	Components() ([]*model.Component, error)
}

type Updater interface {
	ApplyUpdate(ctx context.Context, updateFile, component string) error
}
