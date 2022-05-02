package dell

import (
	"context"
	"os"
	"strconv"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	BinUpdateExitCodeRebootRequired = 2
)

var (
	ErrUnhandledDsuExitCode = errors.New("unhandled dell dsu exit code")
	ErrDellDnfRepoSetup     = errors.New("error setting up dell dnf repo")
)

// dsuInstallUpdates installs DSU identified updates
func (d *dell) dsuInstallUpdates(downloadOnly bool) (int, error) {
	// install pre-requisites
	err := d.pre()
	if err != nil {
		return 0, errors.Wrap(err, "error installing pre-requisites for DSU")
	}

	// Fetch DSU identified update files
	exitCode, err := d.dsu.FetchUpdateFiles(utils.LocalUpdatesDirectory)
	if err != nil {
		return exitCode, err
	}

	if downloadOnly {
		return exitCode, nil
	}

	// Install DSU fetched local update files
	exitCode, err = d.dsu.ApplyLocalUpdates(utils.LocalUpdatesDirectory)
	if err != nil {
		return exitCode, err
	}

	return exitCode, nil
}

// installUpdate installs a given dell update file (DUP)
func (d *dell) installUpdate(ctx context.Context, updateFile string, downgrade bool) (int, error) {
	//	./BIOS_CR1K4_LN_2.9.4_01.BIN -h
	//	-c            : Determine if the update can be applied to the system (1)
	//	-f            : Force a downgrade to an older version. (1)(2)
	//	-q            : Execute the update package silently without user intervention
	//	-n            : Execute the update package without security verification
	//	-r            : Reboot if necessary after the update (2)
	//	-v,--version  : Display version information
	//	--list        : Display contents of package (3)
	//	--installpath=<path>    : Install the update package in the specified path only if options (2) and (8)  are supported.
	if os.Getenv("IRONLIB_TEST") != "" {
		return 0, nil
	}

	// set files executable
	// nolint gocritic: this fs mode declaration is as clear as it gets
	err := os.Chmod(updateFile, 0744)
	if err != nil {
		return 0, err
	}

	// non-interactive
	args := []string{"-q"}

	if downgrade {
		args = append(args, "-f")
	}

	e := utils.NewExecutor(updateFile)
	e.SetArgs(args)

	if d.logger.Level == logrus.TraceLevel {
		e.SetVerbose()
	}

	d.logger.WithFields(
		logrus.Fields{"file": updateFile},
	).Info("Installing dell Update Bin file")

	result, err := e.ExecWithContext(ctx)
	if err != nil {
		return result.ExitCode, err
	}

	d.logger.WithFields(
		logrus.Fields{"file": updateFile},
	).Info("Installed")

	d.hw.PendingReboot = true

	return 0, nil
}

// dsuListUpdates runs the dell-system-update utility to retrieve device inventory
func (d *dell) dsuListUpdates() ([]*model.Component, error) {
	err := d.pre()
	if err != nil {
		return nil, errors.Wrap(err, "error ensuring prerequisites for dsu update list")
	}

	// collect firmware updates available for components
	updates, exitCode, err := d.dsu.ComponentFirmwareUpdatePreview()
	if err != nil && exitCode != utils.DSUExitCodeNoUpdatesAvailable {
		return nil, errors.Wrap(err, "error running dsu update preview")
	}

	return updates, nil
}

// runs the dell-system-update utility to identify and list firmware updates available
func (d *dell) dsuInventory() ([]*model.Component, error) {
	err := d.pre()
	if err != nil {
		return nil, err
	}

	return d.dsu.Inventory()
}

// pre sets up prequisites for dealing with updates
func (d *dell) pre() (err error) {
	errPrefix := "dell dsu prereqs setup error: "

	if d.DsuPrequisitesInstalled {
		return nil
	}

	actions := []func() error{
		d.addDsuRepo, d.installPkgs, d.startSrvHelper,
	}

	for _, action := range actions {
		err := action()
		if err != nil {
			return errors.Wrap(err, errPrefix)
		}
	}

	d.logger.Info("Dell DSU prerequisites setup complete")
	d.DsuPrequisitesInstalled = true

	return nil
}

func (d *dell) installPkgs() error {
	// install dsu package
	dsuPkg := "dell-system-update"

	if d.dsuPackageVersion != "" {
		dsuPkg += "-" + d.dsuPackageVersion
	}

	// install packages
	miscPkgs := []string{
		dsuPkg,
		"srvadmin-idracadm7",
		"usbutils",
		"OpenIPMI",
		"net-tools",
	}

	err := d.dnf.Install(miscPkgs)
	if err != nil {
		return errors.Wrap(err, "failed to install dsu and related tooling")
	}

	return nil
}

// addDsuRepo sets up the dell dnf repository
//
// expects d.updateBaseURL, d.RepoVersion is set before invocation
func (d *dell) addDsuRepo() error {
	params := &utils.DnfRepoParams{
		GPGCheck:    true,
		Name:        "dell",
		BaseURL:     d.updateBaseURL,
		RepoVersion: d.dsuReleaseVersion,
	}

	if os.Getenv("IRONLIB_TEST") != "" {
		return nil
	}

	err := d.dnf.AddRepo("/etc/yum.repos.d", params, []byte(utils.DellRepoTemplate))
	if err != nil {
		return errors.Wrap(err, ErrDellDnfRepoSetup.Error())
	}

	return nil
}

// startSrvHelper starts up the service that loads various ipmi modules,
// Since we're running dsu within a docker container on the target host,
// this was found to be required to ensure dsu was able to inventorize the host correctly.
// else it would not be able to retrieve data over IPMI
func (d *dell) startSrvHelper() error {
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

// checkExitCode looks up the various DSU/update bin exitcodes
// and returns an error if its an actual error
func (d *dell) checkExitCode(exitCode int) error {
	switch exitCode {
	// sometimes the installer does not indicate a reboot is required
	case utils.DSUExitCodeUpdatesApplied:
		d.hw.UpdatesInstalled = true
		d.hw.PendingReboot = true
		d.logger.Trace("update applied successfully")

		return nil
	case utils.DSUExitCodeRebootRequired, BinUpdateExitCodeRebootRequired: // updates applied, reboot required
		d.logger.Trace("update applied, reboot required")
		d.hw.UpdatesInstalled = true
		d.hw.PendingReboot = true

		return nil
	case utils.DSUExitCodeNoUpdatesAvailable: // no applicable updates
		d.logger.Trace("no pending/applicable update(s) for device")

		return nil
	default:
		return errors.Wrap(ErrUnhandledDsuExitCode, strconv.Itoa(exitCode))
	}
}
