package dell

import (
	"context"

	"github.com/packethost/ironlib/actions"
	"github.com/packethost/ironlib/errs"
	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// The dell device provider struct
type dell struct {
	trace                   bool
	hw                      *model.Hardware
	dnf                     *utils.Dnf
	dsu                     *utils.Dsu
	logger                  *logrus.Logger
	dsuVersion              string
	collectors              *actions.Collectors
	DsuPrequisitesInstalled bool
}

// New returns a new Dell device manager
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
	device.Oem = true
	device.OemComponents = &model.OemComponents{Dell: []*model.Component{}}

	// set device manager
	dm := &dell{
		hw:     model.NewHardware(device),
		dnf:    utils.NewDnf(trace),
		dsu:    utils.NewDsu(trace),
		logger: l,
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
	// Collect device inventory from lshw
	d.logger.Info("Collecting hardware inventory")

	err := actions.Collect(ctx, d.hw.Device, d.collectors, d.trace)
	if err != nil {
		return nil, err
	}

	// collect dell component info
	d.logger.Info("Collecting dell OEM component inventory with DSU")

	return d.hw.Device, nil
}

// GetInventoryOEM collects device inventory using vendor specific tooling
// and updates the given device.OemComponents object with the OEM inventory
func (d *dell) GetInventoryOEM(ctx context.Context, device *model.Device, options *model.UpdateOptions) error {
	d.setUpdateOptions(options)

	oemComponents, err := d.dsuInventory()
	if err != nil {
		return err
	}

	d.hw.Device.OemComponents.Dell = append(d.hw.Device.OemComponents.Dell, oemComponents...)

	return nil
}

// ListAvailableUpdates runs the vendor tooling (dsu) to identify updates available
func (d *dell) ListAvailableUpdates(ctx context.Context, options *model.UpdateOptions) (*model.Device, error) {
	// collect firmware updates available for components
	d.logger.Info("Identifying component firmware updates...")

	d.setUpdateOptions(options)

	oemUpdates, err := d.dsuListUpdates()
	if err != nil {
		return nil, err
	}

	count := len(oemUpdates)
	if count == 0 {
		d.logger.Info("no available dell Oem updates")
		return nil, nil
	}

	d.logger.WithField("count", count).Info("component updates identified..")

	d.hw.Device.OemComponents.Dell = append(d.hw.Device.OemComponents.Dell, oemUpdates...)

	return d.hw.Device, nil
}

// InstallUpdates for Dells based on updateOptions
func (d *dell) InstallUpdates(ctx context.Context, options *model.UpdateOptions) error {
	if options.InstallAll {
		return d.installAvailableUpdates(ctx, options.InstallerVersion, options.DownloadOnly)
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
	// the installer returns non-zero return code on failure,
	// when no updates are available
	// or when the device requires a reboot
	exitCode, err := d.dsuInstallUpdates(revision, downloadOnly)
	if err != nil {
		switch exitCode {
		case utils.DSUExitCodeNoUpdatesAvailable:
			d.logger.Debug("update(s) not applicable for this device")
			return errs.ErrNoUpdatesApplicable
		case utils.DSUExitCodeRebootRequired:
			d.logger.Debug("update(s) applied, device requires a reboot")
			d.hw.PendingReboot = true
		default:
			return err
		}
	}

	d.hw.UpdatesInstalled = true

	return nil
}
