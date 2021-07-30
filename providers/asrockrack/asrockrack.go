package asrockrack

import (
	"context"

	"github.com/packethost/ironlib/actions"
	"github.com/packethost/ironlib/model"
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
func New(deviceVendor, deviceModel string, l *logrus.Logger) (model.DeviceManager, error) {
	var trace bool

	if l.GetLevel().String() == "trace" {
		trace = true
	}

	// set device
	device := model.NewDevice()
	device.Model = deviceModel
	device.Vendor = deviceVendor

	// set device manager
	dm := &asrockrack{
		hw:     model.NewHardware(device),
		logger: l,
		trace:  trace,
	}

	return dm, nil
}

// Returns hardware inventory for the device
func (a *asrockrack) GetInventory(ctx context.Context) (*model.Device, error) {
	// Collect device inventory from lshw
	a.logger.Info("Collecting inventory")

	a.hw.Device = model.NewDevice()

	err := actions.Collect(ctx, a.hw.Device, a.collectors, a.trace)
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

// ListUpdatesAvailable runs the vendor tooling (dsu) to identify updates available
func (a *asrockrack) ListUpdatesAvailable(ctx context.Context) (*model.Device, error) {
	return nil, nil
}

// InstallUpdates for asrockrack based on updateOptions
func (a *asrockrack) InstallUpdates(ctx context.Context, options *model.UpdateOptions) error {
	return nil
}
