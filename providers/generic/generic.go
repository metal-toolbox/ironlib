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
	trace      bool
	hw         *model.Hardware
	logger     *logrus.Logger
	collectors *actions.Collectors
}

// New returns a generic device manager
func New(dmidecode *utils.Dmidecode, l *logrus.Logger) (model.DeviceManager, error) {
	var trace bool

	if l.GetLevel().String() == "trace" {
		trace = true
	}

	deviceVendor, err := dmidecode.Manufacturer()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("manufacturer", ""), err.Error())
	}

	deviceModel, err := dmidecode.ProductName()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("Product name", ""), err.Error())
	}

	serial, err := dmidecode.SerialNumber()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("Serial", ""), err.Error())
	}

	// set device
	device := model.NewDevice()
	device.Model = deviceModel
	device.Vendor = deviceVendor
	device.Serial = serial

	// set device manager
	dm := &Generic{
		hw:     model.NewHardware(device),
		logger: l,
		trace:  trace,
	}

	return dm, nil
}

// Returns hardware inventory for the device
func (a *Generic) GetInventory(ctx context.Context) (*model.Device, error) {
	// Collect device inventory from lshw
	a.logger.Info("Collecting inventory")

	err := actions.Collect(ctx, a.hw.Device, a.collectors, a.trace)
	if err != nil {
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
func (a *Generic) ListAvailableUpdates(ctx context.Context, options *model.UpdateOptions) (*model.Device, error) {
	return nil, nil
}

// InstallUpdates installs updates based on updateOptions
func (a *Generic) InstallUpdates(ctx context.Context, options *model.UpdateOptions) error {
	return nil
}

// GetInventoryOEM collects device inventory using vendor specific tooling
// and updates the given device.OemComponents object with the OEM inventory
func (a *Generic) GetInventoryOEM(ctx context.Context, device *model.Device, options *model.UpdateOptions) error {
	return nil
}
