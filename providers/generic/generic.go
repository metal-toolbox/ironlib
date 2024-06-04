package generic

import (
	"context"

	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// A Generic device has methods to collect hardware inventory, regardless of the vendor
type Generic struct {
	trace  bool
	hw     *model.Hardware
	logger *logrus.Logger
}

// New returns a generic device manager
func New(dmidecode *utils.Dmidecode, l *logrus.Logger) (actions.DeviceManager, error) {
	deviceVendor, err := dmidecode.Manufacturer()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("manufacturer", "", 0), err.Error())
	}

	deviceModel, err := dmidecode.ProductName()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("Product name", "", 0), err.Error())
	}

	serial, err := dmidecode.SerialNumber()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("Serial", "", 0), err.Error())
	}

	// set device
	device := &model.Device{}
	device.Model = deviceModel
	device.Vendor = deviceVendor
	device.Serial = serial

	// set device manager
	return &Generic{
		hw:     model.NewHardware(device),
		logger: l,
		trace:  l.Level >= logrus.TraceLevel,
	}, nil
}

// Returns hardware inventory for the device
func (a *Generic) GetInventory(ctx context.Context, options ...actions.Option) (*model.Device, error) {
	// Collect device inventory
	a.logger.Info("Collecting inventory")

	collector := actions.NewInventoryCollectorAction(a.logger, options...)
	if err := collector.Collect(ctx, a.hw.Device); err != nil {
		return nil, err
	}

	return a.hw.Device, nil
}

func (a *Generic) GetModel() string {
	return a.hw.Device.Model
}

func (a *Generic) GetVendor() string {
	return a.hw.Device.Vendor
}

func (a *Generic) RebootRequired() bool {
	return a.hw.PendingReboot
}

func (a *Generic) UpdatesApplied() bool {
	return a.hw.UpdatesInstalled
}

// ListAvailableUpdates runs the vendor tooling (dsu) to identify updates available
func (a *Generic) ListAvailableUpdates(_ context.Context, _ *model.UpdateOptions) (*model.Device, error) {
	return nil, nil
}

// InstallUpdates installs updates based on updateOptions
func (a *Generic) InstallUpdates(_ context.Context, _ *model.UpdateOptions) error {
	return nil
}

// ApplyUpdate is here to satisfy the actions.Updater interface
// it is to be deprecated in favor of InstallUpdates.
func (a *Generic) ApplyUpdate(_ context.Context, _, _ string) error {
	return nil
}

// GetInventoryOEM collects device inventory using vendor specific tooling
// and updates the given device.OemComponents object with the OEM inventory
func (a *Generic) GetInventoryOEM(context.Context, *model.Device, *model.UpdateOptions) error {
	return nil
}
