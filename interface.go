package ironlib

import (
	"github.com/packethost/ironlib/model"
	"golang.org/x/net/context"
)

type Setter interface {
	SetDeviceID(string)
	SetFirmwareUpdateConfig(*model.FirmwareUpdateConfig)
}

type Getter interface {
	GetDeviceID() string
	GetModel() string
	GetVendor() string
	RebootRequired() bool
	UpdatesApplied() bool
	GetInventory(ctx context.Context, listUpdates bool) (*model.Device, error)
	GetUpdatesAvailable(ctx context.Context) (*model.Device, error)
	GetDeviceFirmwareRevision(ctx context.Context) (string, error)
}

//type Configurer interface {
//	ConfigureBMC(ctx context.Context) error
//	ConfigureBIOS(ctx context.Context) error
//}
//

type Updater interface {
	ApplyUpdatesAvailable(ctx context.Context, config *model.FirmwareUpdateConfig, dryRun bool) error
	//UpdateBMC(ctx context.Context) error
	//UpdateBIOS(ctx context.Context) error
}

type Manager interface {
	Getter
	Setter
	//	Configurer
	Updater
}
