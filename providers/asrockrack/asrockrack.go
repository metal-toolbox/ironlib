package asrockrack

import (
	"context"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

// A asrockrack device has methods to collect hardware inventory, regardless of the vendor
type asrockrack struct {
	trace      bool
	hw         *model.Hardware
	logger     *logrus.Logger
	collectors *actions.Collectors
}

// New returns a ASRockRack device manager
func New(dmidecode *utils.Dmidecode, l *logrus.Logger) (model.DeviceManager, error) {
	var trace bool

	if l.Level == logrus.TraceLevel {
		trace = true
	}

	var err error

	// set device
	device := common.NewDevice()

	identifiers, err := utils.IdentifyVendorModel(dmidecode)
	if err != nil {
		return nil, err
	}

	device.Vendor = identifiers.Vendor
	device.Model = identifiers.Model
	device.Serial = identifiers.Serial

	// set device manager
	dm := &asrockrack{
		hw:     model.NewHardware(&device),
		logger: l,
		trace:  trace,
	}

	return dm, nil
}

// Returns hardware inventory for the device
func (a *asrockrack) GetInventory(ctx context.Context, dynamic bool) (*common.Device, error) {
	// Collect device inventory from lshw
	a.logger.Info("Collecting inventory")

	deviceObj := common.NewDevice()
	a.hw.Device = &deviceObj

	err := actions.Collect(ctx, a.hw.Device, a.collectors, a.trace, false, dynamic)
	if err != nil {
		return nil, err
	}

	return a.hw.Device, nil
}

func (a *asrockrack) GetModel() string {
	return a.hw.Device.Model
}

func (a *asrockrack) GetVendor() string {
	return a.hw.Device.Vendor
}

func (a *asrockrack) RebootRequired() bool {
	return a.hw.PendingReboot
}

func (a *asrockrack) UpdatesApplied() bool {
	return a.hw.UpdatesInstalled
}

// ListAvailableUpdates runs the vendor tooling (dsu) to identify updates available
func (a *asrockrack) ListAvailableUpdates(ctx context.Context, options *model.UpdateOptions) (*common.Device, error) {
	return nil, nil
}

// InstallUpdates for asrockrack based on updateOptions
func (a *asrockrack) InstallUpdates(ctx context.Context, options *model.UpdateOptions) error {
	return nil
}

// GetInventoryOEM collects device inventory using vendor specific tooling
// and updates the given device.OemComponents object with the OEM inventory
func (a *asrockrack) GetInventoryOEM(ctx context.Context, device *common.Device, options *model.UpdateOptions) error {
	return nil
}
