package dell

import (
	"context"
	"os"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// EnvUpdateStoreURL defines up the update store base URL prefix
const EnvUpdateBaseURL = "UPDATE_BASE_URL"

// The dell device provider struct
type dell struct {
	DsuPrequisitesInstalled bool
	hw                      *model.Hardware
	dnf                     *utils.Dnf
	dsu                     *utils.Dsu
	logger                  *logrus.Logger
	trace                   bool
	updateBaseURL           string

	// The DSU package version
	// for example 1.9.1.0-21.03.00 from https://linux.dell.com/repo/hardware/DSU_21.05.01/os_independent/x86_64/dell-system-update-1.9.1.0-21.03.00.x86_64.rpm
	dsuPackageVersion string

	// The DSU release version
	// for example: 21.05.01, from https://linux.dell.com/repo/hardware/DSU_21.05.01
	dsuReleaseVersion string
}

// New returns a new Dell device manager
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
	device := common.NewDevice()
	device.Model = deviceModel
	device.Vendor = deviceVendor
	device.Serial = serial
	device.Oem = true

	// when default, the repo URL will point to the default repository
	// this expects a EnvUpdateStoreURL/dell/default/ is made available
	dsuReleaseVersion := os.Getenv(utils.EnvDellDSURelease)
	if dsuReleaseVersion == "" {
		dsuReleaseVersion = "default"
	}

	// when default, whichever version of DSU is available will be installed
	dsuPackageVersion := os.Getenv(utils.EnvDellDSUVersion)
	if dsuPackageVersion == "" {
		dsuPackageVersion = "default"
	}

	// the base url for updates
	updateBaseURL := os.Getenv(EnvUpdateBaseURL)

	// set device manager
	trace := l.Level >= logrus.TraceLevel
	dm := &dell{
		hw:                model.NewHardware(&device),
		dnf:               utils.NewDnf(trace),
		dsu:               utils.NewDsu(trace),
		dsuReleaseVersion: dsuReleaseVersion,
		dsuPackageVersion: dsuPackageVersion,
		updateBaseURL:     updateBaseURL,
		logger:            l,
		trace:             trace,
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
func (d *dell) GetInventory(ctx context.Context, options ...actions.Option) (*common.Device, error) {
	// Collect device inventory
	d.logger.Info("Collecting hardware inventory")

	collector := actions.NewInventoryCollectorAction(d.logger, options...)
	if err := collector.Collect(ctx, d.hw.Device); err != nil {
		return nil, err
	}

	return d.hw.Device, nil
}

// GetInventoryOEM collects device inventory using vendor specific tooling
// and updates the given device.OemComponents object with the OEM inventory
func (d *dell) GetInventoryOEM(ctx context.Context, _ *common.Device, options *model.UpdateOptions) error {
	d.setUpdateOptions(options)

	oemComponents, err := d.dsuInventory(ctx)
	if err != nil {
		return err
	}

	d.hw.OEMComponents = append(d.hw.OEMComponents, oemComponents...)

	return nil
}

// ListAvailableUpdates runs the vendor tooling (dsu) to identify updates available
func (d *dell) ListAvailableUpdates(ctx context.Context, options *model.UpdateOptions) (*common.Device, error) {
	// collect firmware updates available for components
	d.logger.Info("Identifying component firmware updates...")

	d.setUpdateOptions(options)

	oemUpdates, err := d.dsuListUpdates(ctx)
	if err != nil {
		return nil, err
	}

	count := len(oemUpdates)
	if count == 0 {
		d.logger.Info("no available dell Oem updates")
		return nil, nil
	}

	d.logger.WithField("count", count).Info("component updates identified..")

	d.hw.OEMComponents = append(d.hw.OEMComponents, oemUpdates...)

	return d.hw.Device, nil
}

// UpdateRequirements returns requirements to be met before and after a firmware install,
// the caller may use the information to determine if a powercycle, reconfiguration or other actions are required on the component.
func (a *dell) UpdateRequirements(ctx context.Context, option *model.UpdateOptions) (*model.UpdateRequirements, error) {
	return nil, errors.Wrap(errs.ErrUpdateReqNotImplemented, "provider: dell")
}

// InstallUpdates for Dells based on updateOptions
func (d *dell) InstallUpdates(ctx context.Context, options *model.UpdateOptions) error {
	d.setUpdateOptions(options)

	if options.InstallAll {
		return d.installAvailableUpdates(ctx, options.DownloadOnly)
	}

	exitCode, err := d.installUpdate(ctx, options.Slug, options.ForceInstall)
	if err != nil {
		return err
	}

	return d.checkExitCode(exitCode)
}

// installAvailableUpdates runs DSU to install all available updates
// revision is the Dell DSU version to ensure installed
func (d *dell) installAvailableUpdates(ctx context.Context, downloadOnly bool) error {
	// the installer returns non-zero return code on failure,
	// when no updates are available
	// or when the device requires a reboot
	exitCode, err := d.dsuInstallUpdates(ctx, downloadOnly)
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

// setUpdateOptions overrides set the DSU version and repository parameters
func (d *dell) setUpdateOptions(options *model.UpdateOptions) {
	if options.InstallerVersion != "" {
		d.dsuPackageVersion = options.InstallerVersion
	}

	if options.RepositoryVersion != "" {
		d.dsuReleaseVersion = options.RepositoryVersion
	}

	if options.BaseURL != "" {
		d.updateBaseURL = options.BaseURL
	}

	d.logger.WithFields(
		logrus.Fields{
			"dsu version": d.dsuPackageVersion,
			"dsu repo":    d.dsuReleaseVersion,
			"base url":    d.updateBaseURL,
		},
	).Info("update parameters")
}

// ApplyUpdate is here to satisfy the actions.Updater interface
// it is to be deprecated in favor of InstallUpdates.
func (d *dell) ApplyUpdate(context.Context, string, string) error {
	return nil
}
