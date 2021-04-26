package model

import (
	"context"

	"github.com/packethost/ironlib/config"
)

type Setter interface {
	SetDeviceID(string)
	SetFirmwareUpdateConfig(*FirmwareUpdateConfig)
	SetBIOSConfiguration(ctx context.Context, config *config.BIOSConfiguration) error
}

type Getter interface {
	GetDeviceID() string
	GetModel() string
	GetVendor() string
	RebootRequired() bool
	UpdatesApplied() bool
	GetInventory(ctx context.Context, listUpdates bool) (*Device, error)
	GetUpdatesAvailable(ctx context.Context) (*Device, error)
	GetDeviceFirmwareRevision(ctx context.Context) (string, error)
	GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error)
}

type Updater interface {
	ApplyUpdatesAvailable(ctx context.Context, config *FirmwareUpdateConfig, dryRun bool) error
	//UpdateBMC(ctx context.Context) error
	//UpdateBIOS(ctx context.Context) error
}

type Manager interface {
	Getter
	Setter
	//	Configurer
	Updater
}
