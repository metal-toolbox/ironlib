package generic

import (
	"context"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// A Generic device has methods to collect hardware inventory, regardless of the vendor
type Generic struct {
	hw       *model.Hardware
	lshw     *utils.Lshw
	logger   *logrus.Logger
	smartctl utils.Collector
}

// New returns a generic device manager
func New(deviceVendor, deviceModel string, l *logrus.Logger) (model.DeviceManager, error) {
	var trace bool

	if l.GetLevel().String() == "trace" {
		trace = true
	}

	// set device
	device := &model.Device{
		Model:  deviceModel,
		Vendor: deviceVendor,
	}

	// set device manager
	dm := &Generic{
		hw:       model.NewHardware(device),
		lshw:     utils.NewLshwCmd(trace),
		smartctl: utils.NewSmartctlCmd(trace),
		logger:   l,
	}

	return dm, nil
}

// Returns hardware inventory for the device
func (a *Generic) GetInventory(ctx context.Context) (*model.Device, error) {
	// Collect device inventory from lshw
	a.logger.Info("Collecting inventory with lshw")

	a.hw.Device = model.NewDevice()

	err := a.lshw.Inventory(a.hw.Device)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving device inventory")
	}

	// collect drive information
	drives, err := a.smartctl.Components()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving drive information")
	}

	// update drive information
	model.ComponentFirmwareDrives(a.hw.Device.Drives, drives, true)

	// update device with the components retrieved from inventory
	model.SetDeviceComponents(a.hw.Device, drives)

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

// ListUpdatesAvailable runs the vendor tooling (dsu) to identify updates available
func (a *Generic) ListUpdatesAvailable(ctx context.Context) (*model.Device, error) {
	return nil, nil
}

// InstallUpdates installs updates based on updateOptions
func (a *Generic) InstallUpdates(ctx context.Context, options *model.UpdateOptions) error {
	return nil
}
