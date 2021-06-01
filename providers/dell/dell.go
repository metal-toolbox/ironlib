package dell

import (
	"context"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// The dell device provider struct
type dell struct {
	hw                      *model.Hardware
	dnf                     *utils.Dnf
	dsu                     *utils.Dsu
	lshw                    *utils.Lshw
	smartctl                utils.Collector
	logger                  *logrus.Logger
	dsuVersion              string
	DsuPrequisitesInstalled bool
}

// New returns a new Dell device manager
func New(deviceVendor, deviceModel string, l *logrus.Logger) (model.DeviceManager, error) {
	var trace bool

	if l.GetLevel().String() == "trace" {
		trace = true
	}

	// set device
	device := &model.Device{
		Model:         deviceModel,
		Vendor:        deviceVendor,
		Oem:           true,
		OemComponents: &model.OemComponents{Dell: []*model.Component{}},
	}

	// set device manager
	dm := &dell{
		hw:       model.NewHardware(device),
		dnf:      utils.NewDnf(trace),
		dsu:      utils.NewDsu(trace),
		lshw:     utils.NewLshwCmd(trace),
		smartctl: utils.NewSmartctlCmd(trace),
		logger:   l,
	}

	return dm, nil
}

// GetModel returns the device model
func (d *dell) GetModel() string {
	return d.hw.Device.Model
}

// GetVendor returns the device model
func (d *dell) GetVendor() string {
	return d.hw.Device.Vendor
}

// RebootRequired returns a bool value for when a device may be pending a reboot
func (d *dell) RebootRequired() bool {
	return d.hw.PendingReboot
}

// UpdatesApplied returns a bool value when updates were applied on a device
func (d *dell) UpdatesApplied() bool {
	return d.hw.UpdatesInstalled
}

// GetInventory collects hardware inventory along with the firmware installed and returns a Device object
func (d *dell) GetInventory(ctx context.Context) (*model.Device, error) {
	var err error

	// Collect device inventory from lshw
	d.logger.Info("Collecting inventory with lshw")

	d.hw.Device = model.NewDevice()

	err = d.lshw.Inventory(d.hw.Device)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving device inventory")
	}

	// collect drive information
	drives, err := d.smartctl.Components()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving drive information")
	}

	// update drive information
	model.ComponentFirmwareDrives(d.hw.Device.Drives, drives, true)

	// setup slice for oem components
	d.hw.Device.OemComponents = &model.OemComponents{Dell: []*model.Component{}}

	// collect dell component info
	d.logger.Info("Collecting dell specific component inventory with DSU")

	components, err := d.dsuInventory()
	if err != nil {
		return nil, err
	}

	// update device with the components retrieved from inventory
	model.SetDeviceComponents(d.hw.Device, components)

	return d.hw.Device, nil
}

// ListUpdatesAvailable runs the vendor tooling (dsu) to identify updates available
func (d *dell) ListUpdatesAvailable(ctx context.Context) (*model.Device, error) {
	// collect firmware updates available for components
	d.logger.Info("Identifying component firmware updates...")

	updates, err := d.dsuListUpdates()
	if err != nil {
		return nil, err
	}

	count := len(updates)
	if count == 0 {
		d.logger.Info("no available updates")
		return nil, nil
	}

	d.logger.WithField("count", count).Info("component updates identified..")

	model.SetDeviceComponents(d.hw.Device, updates)

	return d.hw.Device, nil
}

// InstallUpdates for Dells based on updateOptions
func (d *dell) InstallUpdates(ctx context.Context, options *model.UpdateOptions) error {
	if options.InstallAll {
		return d.installAvailableUpdates(ctx, options.InstallerVersion, options.AllowDowngrade)
	}

	exitCode, err := d.installUpdate(ctx, options.Slug, options.AllowDowngrade)
	if err != nil {
		return err
	}

	return d.checkExitCode(exitCode)
}

// installAvailableUpdates runs DSU to install all available updates
// revision is the Dell DSU version to ensure installed
func (d *dell) installAvailableUpdates(ctx context.Context, revision string, downloadOnly bool) error {
	exitCode, err := d.dsuInstallUpdates(revision, downloadOnly)
	if err != nil {
		if exitCode == utils.DSUExitCodeNoUpdatesAvailable {
			d.logger.Trace("update(s) not applicable for this device")
			return nil
		}

		return err
	}

	d.hw.UpdatesInstalled = true

	return nil
}
