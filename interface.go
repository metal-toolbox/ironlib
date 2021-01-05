package ironlib

import (
	"github.com/packethost/ironlib/model"
	"golang.org/x/net/context"
)

type Setter interface {
	SetOptions(map[string]interface{}) error
}

type Getter interface {
	GetModel() string
	GetVendor() string
	RebootRequired() bool
	GetInventory(ctx context.Context) (*model.Device, error)
	GetUpdatesAvailable(ctx context.Context) (*model.Device, error)
}

//type Configurer interface {
//	ConfigureBMC(ctx context.Context) error
//	ConfigureBIOS(ctx context.Context) error
//}
//

type Updater interface {
	ApplyUpdatesAvailable(ctx context.Context) error
	//UpdateBMC(ctx context.Context) error
	//UpdateBIOS(ctx context.Context) error
}

type Manager interface {
	Getter
	Setter
	//	Configurer
	Updater
}
