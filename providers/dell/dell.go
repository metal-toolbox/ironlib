package dell

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
)

// For the future Joel, it may be worth parsing the inventory catalog
// /usr/libexec/dell_dup/inv.xml

type Dell struct {
	ID                      string
	PendingReboot           bool // set when the device requires a reboot after running an upgrade
	DsuPrequisitesInstalled bool
	Vendor                  string
	Model                   string
	Serial                  string
	Components              []*model.Component
	ComponentUpdates        []*model.Component
	Logger                  *logrus.Logger
	Dmidecode               *utils.Dmidecode
	Dnf                     *utils.Dnf
	Dsu                     *utils.Dsu
	FirmwareUpdateConfig    *model.FirmwareUpdateConfig
}

type Component struct {
	Serial   string
	Type     string
	Model    string
	Firmware string
}

func (d *Dell) GetModel() string {
	return d.Model
}

func (d *Dell) GetVendor() string {
	return d.Vendor
}

func (d *Dell) GetDeviceID() string {
	return d.ID
}

func (d *Dell) SetDeviceID(id string) {
	d.ID = id
}

func (d *Dell) SetFirmwareUpdateConfig(config *model.FirmwareUpdateConfig) {
	d.FirmwareUpdateConfig = config
}

func (d *Dell) RebootRequired() bool {
	return d.PendingReboot
}

// nolint: gocyclo
// Return device component inventory, including any update information
func (d *Dell) GetInventory(ctx context.Context, listUpdates bool) (*model.Device, error) {

	err := d.pre()
	if err != nil {
		return nil, err
	}

	// collect current component firmware versions
	d.Logger.Info("Collecting component firmware versions...")
	componentInventory, err := d.Dsu.ComponentInventory()
	if err != nil {
		return nil, err
	}

	device := &model.Device{
		ID:         d.ID,
		Serial:     d.Serial,
		Model:      d.Model,
		Vendor:     d.Vendor,
		Oem:        true,
		Components: componentInventory,
	}

	if !listUpdates {
		return device, nil
	}

	// collect firmware updates available for components
	d.Logger.Info("Identifying available component firmware updates...")
	// set d.ComponentUpdates with available updates
	err = d.listUpdatesAvailable()
	if err != nil {
		return nil, err
	}

	count := len(d.ComponentUpdates)
	if count > 0 {
		d.Logger.WithField("count", count).Info("device has updates available..")
	} else {
		d.Logger.Info("device has no available updates to install")
	}

	// converge component inventory data with firmware update data
	for _, component := range componentInventory {
		component.DeviceID = d.ID
		for _, update := range d.ComponentUpdates {
			if component.Slug == update.Slug {
				component.Metadata = update.Metadata
				if strings.TrimSpace(update.FirmwareAvailable) != "" {
					d.Logger.WithFields(logrus.Fields{"component slug": component.Slug, "update": update.FirmwareAvailable}).Trace("update available")
				}
				if component.Slug == "Unknown" {
					d.Logger.WithFields(logrus.Fields{"component name": component.Name}).Warn("component slug is 'Unknown', this needs to be fixed in componentNameSlug()")
				}
				component.FirmwareAvailable = update.FirmwareAvailable
			}
		}
	}

	device.Components = componentInventory
	return device, nil
}

// Return available firmware updates for device
func (d *Dell) GetUpdatesAvailable(ctx context.Context) (*model.Device, error) {

	err := d.pre()
	if err != nil {
		return nil, err
	}

	// collect firmware updates available for components
	d.Logger.Info("Identifying component firmware updates...")
	err = d.listUpdatesAvailable()
	if err != nil {
		return nil, err
	}

	count := len(d.ComponentUpdates)
	if count > 0 {
		d.Logger.WithField("count", count).Info("device has updates available..")
	} else {
		d.Logger.Info("device has no available updates to install")
	}

	device := &model.Device{
		ID:               d.ID,
		Serial:           d.Serial,
		Model:            d.Model,
		Vendor:           d.Vendor,
		Oem:              true,
		ComponentUpdates: d.ComponentUpdates,
	}

	return device, nil
}

// The installed DSU release is the firmware revision for dells
func (d *Dell) GetDeviceFirmwareRevision(ctx context.Context) (string, error) {
	return d.Dsu.Version()
}

// sets d.ComponentUpdates to the slice of updates identified by the dell-system-update utility
func (d *Dell) listUpdatesAvailable() error {

	err := d.pre()
	if err != nil {
		return err
	}

	// collect firmware updates available for components
	componentUpdates, exitCode, err := d.Dsu.ComponentFirmwareUpdatePreview()
	if err != nil && exitCode != utils.DSUExitCodeNoUpdatesAvailable {
		return err
	}

	d.ComponentUpdates = componentUpdates

	return nil
}

func (d *Dell) ApplyUpdatesAvailable(ctx context.Context, config *model.FirmwareUpdateConfig, dryRun bool) (err error) {

	if config == nil {
		return fmt.Errorf("ApplyUpdatesAvailable() requires a valid *model.FirmwareUpdateConfig")
	}

	d.FirmwareUpdateConfig = config

	// collect data before we proceed to apply updates
	steps := []func() error{d.pre}

	// lookup component updates if it wasn't done earlier
	if len(d.ComponentUpdates) == 0 {
		steps = append(steps, d.listUpdatesAvailable)
	}

	for _, step := range steps {
		err = step()
		if err != nil {
			return err
		}
	}

	if len(d.ComponentUpdates) == 0 {
		d.Logger.Info("No updates available to apply")
		return nil
	}

	for _, component := range d.ComponentUpdates {
		d.Logger.WithFields(logrus.Fields{"slug": component.Slug, "name": component.Name, "installed": component.FirmwareInstalled, "available": component.FirmwareAvailable}).Info("component update to be applied")
	}

	if dryRun {
		return nil
	}

	// log update process to stdout
	d.Dsu.Executor.SetVerbose()

	// apply updates
	exitCode, err := d.Dsu.ApplyUpdates()

	d.Dsu.Executor.SetQuiet()

	// check exit code - see dsu_return_codes.md
	switch exitCode {
	case utils.DSUExitCodeUpdatesApplied:
		d.Logger.Infoln("updates applied successfully")
		return nil
	case utils.DSUExitCodeNoUpdatesAvailable: // no applicable updates
		d.Logger.Infoln("no applicable updates")
		return nil
	case utils.DSUExitCodeRebootRequired: // updates applied, reboot required
		d.Logger.Infoln("updates applied, reboot required")
		d.PendingReboot = true
		return nil
	default:
		return fmt.Errorf("executing dsu returned error: %s, exit code: %d", err.Error(), exitCode)
	}

}

// pre sets up prequisites for dealing with updates
func (d *Dell) pre() (err error) {

	errPrefix := "dell dsu prereqs setup error: "

	if d.DsuPrequisitesInstalled {
		return nil
	}

	actions := []func() error{
		d.enableRepo, d.installPkgs, d.startSrvHelper,
	}

	for _, action := range actions {
		err := action()
		if err != nil {
			return fmt.Errorf(errPrefix + err.Error())
		}
	}

	d.Logger.Info("Dell DSU prerequisites setup complete")
	d.DsuPrequisitesInstalled = true

	return nil
}

func (d *Dell) installPkgs() error {

	// install dsu package
	var dsuPkgs []string

	if d.FirmwareUpdateConfig != nil && len(d.FirmwareUpdateConfig.Updates) > 0 {
		for _, version := range d.FirmwareUpdateConfig.Updates {
			dsuPkgs = append(dsuPkgs, "dell-system-update-"+version)
		}
	} else {
		dsuPkgs = append(dsuPkgs, "dell-system-update")
	}

	err := d.Dnf.InstallOneOf(dsuPkgs)
	if err != nil {
		return err
	}

	// install dsu related tools
	miscPkgs := []string{
		"srvadmin-idracadm7",
		"usbutils",
		"OpenIPMI",
		"net-tools",
	}

	err = d.Dnf.Install(miscPkgs)
	if err != nil {
		return fmt.Errorf("Attempts to install dsu related tools: " + err.Error())
	}

	return nil
}

func (d *Dell) enableRepo() error {

	// the update environment this dsu package is being installed
	// environment is one of production, vanguard, canary
	// the update environment is used by fup to segregate devices under upgrade for testing/production
	var updateEnv string

	// if a deploy manifest was defined
	if d.FirmwareUpdateConfig != nil && d.FirmwareUpdateConfig.UpdateEnv != "" {
		updateEnv = d.FirmwareUpdateConfig.UpdateEnv
	} else {
		updateEnv = "production"
	}

	repos := []string{updateEnv + "-dell-system-update_independent", updateEnv + "-dell-system-update_dependent"}

	return d.Dnf.EnableRepo(repos)

}

// startSrvHelper starts up the service that loads various ipmi modules,
// Since we're running dsu within a docker container on the target host,
// this was found to be required to ensure dsu was able to inventorize the host correctly.
// else it would not be able to retrieve data over IPMI
func (d *Dell) startSrvHelper() error {

	if os.Getenv("IRONLIB_TEST") != "" {
		return nil
	}

	e := utils.NewExecutor("/usr/libexec/instsvcdrv-helper")
	e.SetArgs([]string{"start"})

	result, err := e.ExecWithContext(context.Background())

	if err != nil || result.ExitCode != 0 {
		return err
	}

	return nil

}

//
