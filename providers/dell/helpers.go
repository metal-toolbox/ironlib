package dell

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	BinUpdateExitCodeRebootRequired = 2
)

// nolint: gocyclo
// fetch component updates and install them
func (d *Dell) fetchAndApplyUpdates(componentUpdates []*model.Component, config *model.FirmwareUpdateConfig, fetchOnly bool) error {

	var err error
	var exitCode int

	// list of dell update bin files
	var updateFiles []string

	// if updates are pinned for this device
	var pinned bool

	if len(componentUpdates) == 0 {
		return fmt.Errorf("expected a list of component updates, got none")
	}

	// identify and fetch component updates
	if len(config.Components) > 0 {

		if config == nil {
			return fmt.Errorf("expected a valid firmware config to identify pinned updates, got nil")
		}

		updateFiles, err = d.fetchPinnedUpdates(d.DM.FirmwareUpdateConfig)
		if err != nil {
			return err
		}

		pinned = true

		d.Logger.WithFields(
			logrus.Fields{"count": len(updateFiles), "files": strings.Join(updateFiles, ",")},
		).Info("pinned updates fetched and ready for install")

	} else {

		// fetch update files for install
		exitCode, err = d.dsuFetchUpdates(utils.LocalUpdatesDirectory)
		if err != nil && exitCode != utils.DSUExitCodeNoUpdatesAvailable {
			return err
		}

		if exitCode == utils.DSUExitCodeNoUpdatesAvailable {
			d.Logger.Info("no DSU updates applicable")
			return nil
		}

		d.Logger.WithFields(
			logrus.Fields{"count": len(componentUpdates)},
		).Info("DSU updates fetched and ready for install")

	}

	if fetchOnly {
		return nil
	}

	d.Dsu.Executor.SetVerbose()

	// if updates were pinned we apply the given update files
	// else the cached updates fetched by DSU are applied
	if pinned {
		// install pinned updates - exec DUP files
		exitCode, err = d.installUpdateFiles(updateFiles, false)
	} else {
		// install updates with DSU
		exitCode, err = d.dsuApplyLocalUpdates(utils.LocalUpdatesDirectory)
	}

	if err != nil {
		d.Logger.WithFields(
			logrus.Fields{"updates": len(componentUpdates), "pinned": pinned, "exit code": exitCode, "err": err},
		).Error("error applying updates")
		return err
	}

	d.Logger.WithFields(
		logrus.Fields{"updates": len(componentUpdates), "pinned": pinned, "exit code": exitCode},
	).Trace("update apply complete")

	d.Dsu.Executor.SetQuiet()

	// check exit code - see dsu_return_codes.md
	switch exitCode {
	case utils.DSUExitCodeUpdatesApplied:
		// sometimes the installer does not indicate a reboot is required
		d.DM.UpdatesInstalled = true
		d.DM.PendingReboot = true
		d.Logger.Infoln("updates applied successfully")
		return nil
	case utils.DSUExitCodeRebootRequired, BinUpdateExitCodeRebootRequired: // updates applied, reboot required
		d.Logger.Infoln("updates applied, reboot required")
		d.DM.UpdatesInstalled = true
		d.DM.PendingReboot = true
		return nil
	case utils.DSUExitCodeNoUpdatesAvailable: // no applicable updates
		d.Logger.Infoln("no applicable updates")
		return nil
	default:
		return fmt.Errorf("unhandled dsu exit code: %d", exitCode)
	}

}

// Identify updates applicable based on the firmware configuration
// when the config lists component firmware pins - return the list of applicable updates
// else the list of DSU updates appliable are returned
func (d *Dell) identifyUpdatesApplicable(components []*model.Component, config *model.FirmwareUpdateConfig) ([]*model.Component, error) {

	var updates []*model.Component
	var err error

	// Components are pinned
	if config != nil && config.Components != nil {

		if len(components) == 0 {
			return nil, fmt.Errorf("expected a list of components to identify current firmware versions, got none")
		}

		if len(config.Components) == 0 {
			return nil, fmt.Errorf("expected a list of component firmware config, got none")
		}

		// list pinned updates
		d.Logger.Info("Identifying pinned component firmware updates...")
		updates, err = utils.ComponentsForUpdate(components, config)
		if err != nil {
			return nil, err
		}

		if len(updates) > 0 {
			d.Logger.WithField("count", len(updates)).Info("Identified pinned updates for install")
		} else {
			d.Logger.WithField("count", len(updates)).Info("Pinned updates listed in config not applicable")
		}

	} else {
		// list DSU updates
		d.Logger.Info("Identifying DSU component firmware updates...")
		updates, err = d.dsuListUpdates()
		if err != nil {
			return nil, err
		}
		if len(updates) > 0 {
			d.Logger.WithField("count", len(updates)).Info("applicable DSU updates to be applied")

		} else {
			d.Logger.WithField("count", len(updates)).Info("No DSU updates applicable for install")
		}
	}

	if len(updates) > 0 {
		d.DM.UpdatesAvailable = 1
	} else {
		d.DM.UpdatesAvailable = 0
	}

	return updates, nil

}

// Downloads update files and validates checksum - for update files listed in the configuration
func (d *Dell) fetchPinnedUpdates(config *model.FirmwareUpdateConfig) ([]string, error) {

	updateFiles := make([]string, 0)

	// retrieve update files
	for _, c := range config.Components {
		d.Logger.WithFields(logrus.Fields{"slug": c.Slug, "url": c.UpdateFileURL, "dst": "/tmp"}).Info("fetching pinned component update")
		// retrieve update file under target directory, validate checksum
		updateFile, err := utils.RetrieveUpdateFile(c.UpdateFileURL, "/tmp")
		if err != nil {
			return []string{}, err
		}
		updateFiles = append(updateFiles, updateFile)
	}

	return updateFiles, nil

}

// runs the dell-system-update utility to fetch update files identified by DSU
func (d *Dell) dsuFetchUpdates(dstDir string) (int, error) {

	err := d.pre()
	if err != nil {
		return 0, errors.Wrap(err, "error installing pre-requisites for DSU")
	}

	// Fetch DSU identified update files
	return d.Dsu.FetchUpdateFiles(dstDir)
}

// runs DSU to install update files available in the given directory
func (d *Dell) dsuApplyLocalUpdates(updatesDir string) (int, error) {

	err := d.pre()
	if err != nil {
		return 0, errors.Wrap(err, "error setting up DSU pre-requisites")
	}

	// install updates from the local directory
	return d.Dsu.ApplyLocalUpdates(updatesDir)
}

// install the given list of dell update files
func (d *Dell) installUpdateFiles(updateFiles []string, forceDowngrade bool) (int, error) {

	/*
		./BIOS_CR1K4_LN_2.9.4_01.BIN -h
		-c            : Determine if the update can be applied to the system (1)
		-f            : Force a downgrade to an older version. (1)(2)
		-q            : Execute the update package silently without user intervention
		-n            : Execute the update package without security verification
		-r            : Reboot if necessary after the update (2)
		-v,--version  : Display version information
		--list        : Display contents of package (3)
		--installpath=<path>    : Install the update package in the specified path only if options (2) and (8)  are supported.
	*/

	if os.Getenv("IRONLIB_TEST") != "" {
		return 0, nil
	}

	// install each update file
	for _, updateFile := range updateFiles {

		// set files executable
		err := os.Chmod(updateFile, 0744)
		if err != nil {
			return 0, err
		}

		// non-interactive
		args := []string{"-q"}

		if forceDowngrade {
			args = append(args, "-f")
		}

		e := utils.NewExecutor(updateFile)
		e.SetArgs(args)
		e.SetVerbose()

		d.Logger.WithFields(
			logrus.Fields{"file": updateFile},
		).Info("Installing Dell Update Bin file")

		result, err := e.ExecWithContext(context.Background())
		if err != nil {
			return result.ExitCode, err
		}

		d.Logger.WithFields(
			logrus.Fields{"file": updateFile},
		).Info("Installed")

	}

	d.DM.PendingReboot = true

	return 0, nil
}

// runs the dell-system-update utility to retrieve device inventory
func (d *Dell) dsuListUpdates() ([]*model.Component, error) {

	err := d.pre()
	if err != nil {
		return nil, errors.Wrap(err, "error ensuring prerequisites for dsu update list")
	}

	// collect firmware updates available for components
	updates, exitCode, err := d.Dsu.ComponentFirmwareUpdatePreview()
	if err != nil && exitCode != utils.DSUExitCodeNoUpdatesAvailable {
		return nil, errors.Wrap(err, "error running dsu update preview")
	}

	return updates, nil
}

// runs the dell-system-update utility to identify and list firmware updates available
func (d *Dell) dsuInventory() ([]*model.Component, error) {

	err := d.pre()
	if err != nil {
		return nil, err
	}

	inv, err := d.Dsu.ComponentInventory()
	if err != nil {
		return nil, err
	}

	return inv, nil
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

	if d.DM.FirmwareUpdateConfig != nil && len(d.DM.FirmwareUpdateConfig.Updates) > 0 {
		for _, version := range d.DM.FirmwareUpdateConfig.Updates {
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
	if d.DM.FirmwareUpdateConfig != nil && d.DM.FirmwareUpdateConfig.UpdateEnv != "" {
		updateEnv = d.DM.FirmwareUpdateConfig.UpdateEnv
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
