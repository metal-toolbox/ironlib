package dell

import (
	"context"
	"fmt"
	"os"

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
	DsuVersions             []string
	UpdateEnvironment       string
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

func (d *Dell) RebootRequired() bool {
	return d.PendingReboot
}

// Sets various options for the dell system update utility and fup
func (d *Dell) SetOptions(options map[string]interface{}) error {

	// the dell-system-update versions that should be installed
	dsuVersions, exists := options["dsu_versions"]
	if !exists {
		return fmt.Errorf("dell provider expects 'dsu_versions' option to be set with SetOptions()")
	}

	d.DsuVersions = dsuVersions.([]string)

	// the update environment this device is part of
	// update_environment is one of production, vanguard, canary
	// this is a fup thing and could be moved out from this lib at some point
	updateEnvironment, exists := options["update_environment"]
	if !exists {
		return fmt.Errorf("dell provider expects 'update_environment' options to be set with SetOptions()")
	}

	switch updateEnvironment {
	case "production", "vanguard", "canary":
		d.UpdateEnvironment = updateEnvironment.(string)
	default:
		return fmt.Errorf("dell provider expects a valid 'update_environment' - one of production, vanguard, canary")
	}

	return nil
}

// Return device component inventory
func (d *Dell) GetInventory(ctx context.Context) (*model.Device, error) {

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

func (d *Dell) ApplyUpdatesAvailable() (err error) {

	// collect data before we proceed to apply updates
	steps := []func() error{d.pre}

	// Skip component inventory collection if we have an update count
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

	// log update process to stdout
	d.Dsu.Executor.SetVerbose()
	d.Logger.Info("applying available updates")

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

//
//// pre sets up prequisites for dealing with updates
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

	if len(d.DsuVersions) > 0 {
		for _, version := range d.DsuVersions {
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
	var e string

	// if a deploy manifest was defined
	if d.UpdateEnvironment != "" {
		e = d.UpdateEnvironment
	} else {
		e = "production"
	}

	repos := []string{e + "-dell-system-update_independent", e + "-dell-system-update_dependent"}

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
