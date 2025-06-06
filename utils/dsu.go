package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

type Dsu struct {
	Executor Executor
}

const (
	// see test_data/dsu_return_codes.md
	DSUExitCodeUpdatesApplied     = 0
	DSUExitCodeRebootRequired     = 8
	DSUExitCodeNoUpdatesAvailable = 34

	// Dell specific component slugs
	slugDellSystemCPLD                  = "Dell System CPLD"
	slugDellBossAdapter                 = "Boss Adapter"
	slugDellBossAdapterDisk0            = "Boss Adapter - Disk 0"
	slugDellBossAdapterDisk1            = "Boss Adapter - Disk 1"
	slugDellIdracServiceModule          = "IDrac Service Module"
	slugDellLifeCycleController         = "Lifecycle Controller"
	slugDellOSCollector                 = "OS Collector"
	slugDell64bitUefiDiagnostics        = "Dell 64 bit uEFI diagnostics"
	slugDellBackplaneExpander           = "Backplane-Expander"
	slugDellNonExpanderStorageBackplane = "Non-Expander Storage Backplane (SEP)"

	// EnvDellDSURelease is the Dell DSU release version
	//
	// e.g: 21.11.12 from https://linux.dell.com/repo/hardware/DSU_21.11.12/
	EnvDellDSURelease = "DELL_DSU_RELEASE"

	// EnvDellDSUVersion is the Dell DSU utility package version
	//
	// e.g: 1.9.2.0-21.07.00 from https://linux.dell.com/repo/hardware/DSU_21.11.12/os_independent/x86_64/dell-system-update-1.9.2.0-21.07.00.x86_64.rpm
	EnvDellDSUVersion = "DELL_DSU_VERSION"

	LocalUpdatesDirectory = "/root/dsu"

	EnvDsuUtility = "IRONLIB_UTIL_DSU"
)

// DellComponentSlug is an ordered list of dell component identifiers to component slug
// To identify components correctly, if two components contain a similar string
// e.g: "idrac", "dell emc idrac service module" the latter should be positioned before the former in the list.
var dellComponentSlug = [][]string{
	{"bios", common.SlugBIOS},
	{"ethernet", common.SlugNIC},
	{"dell emc idrac service module", slugDellIdracServiceModule},
	{"idrac", common.SlugBMC},
	{"backplane", common.SlugBackplaneExpander},
	{"power supply", common.SlugPSU},
	{"hba330", common.SlugStorageController},
	{"nvmepcissd", common.SlugDrive},
	{"system cpld", slugDellSystemCPLD},
	{"sep firmware", slugDellNonExpanderStorageBackplane},
	{"lifecycle controller", slugDellLifeCycleController},
	{"os collector", slugDellOSCollector},
	{"disk 0 of boss adapter", slugDellBossAdapterDisk0},
	{"disk 1 of boss adapter", slugDellBossAdapterDisk1},
	{"boss", slugDellBossAdapter},
	{"dell 64 bit uefi diagnostics", slugDell64bitUefiDiagnostics},
	{"integrated dell remote access controller", common.SlugBMC},
}

var (
	ErrDsuInventoryCollectorBinMissing   = errors.New("dsu inventory collector executable missing 'invcol_*_*.BIN'")
	ErrMultipleDsuInventoryCollectorBins = errors.New("multiple inventory collector bins found")
	ErrDsuUpdatesDirectoryMissing        = errors.New("dsu updates directory missing")
	ErrDsuVersionQuery                   = errors.New("dsu version query error")
)

// NewDsu returns a executor to run dsu commands
// if trace is enabled, stdout is printed to the terminal
func NewDsu(trace bool) *Dsu {
	utility := "dsu"

	// lookup env var for util
	if eVar := os.Getenv(EnvDsuUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	if !trace {
		e.SetQuiet()
	}

	return &Dsu{
		Executor: e,
	}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (d *Dsu) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := d.Executor.CheckExecutable()

	return "dsu", d.Executor.CmdPath(), er
}

// Returns a dsu instance with a fake executor for tests
func NewFakeDsu(r io.Reader) (*Dsu, error) {
	dsu := &Dsu{
		Executor: NewFakeExecutor("dsu"),
	}

	b := bytes.Buffer{}

	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	dsu.Executor.SetStdout(b.Bytes())

	return dsu, nil
}

// FetchUpdateFiles executes dsu to fetch applicable updates into to local directory
// returns the exitcode and error if any
// NOTE:
// dsu 1.8 drops update files under the given $updateDir
// dsu 1.9 creates a directory '$updateDir/dellupdates' and drops the updates in there
func (d *Dsu) FetchUpdateFiles(ctx context.Context, dstDir string) (int, error) {
	// purge any existing update file/directory with the same name
	_ = os.Remove(dstDir)

	d.Executor.SetArgs("--destination-type=CBD", "--destination-location="+dstDir)

	// because... yeah dsu wants to fetch updates interactively
	d.Executor.SetStdin(bytes.NewReader([]byte("a\nc\n")))

	result, err := d.Executor.Exec(ctx)

	return result.ExitCode, err
}

// ApplyLocalUpdates installs update files fetched by FetchUpdateFiles()
// DSU needs to be pointed to the right inventory bin or it barfs
// returns the resulting exitcode and error if any
func (d *Dsu) ApplyLocalUpdates(ctx context.Context, updateDir string) (int, error) {
	// ensure the updates directory exists
	_, err := os.Stat(updateDir)
	if err != nil {
		return 0, errors.Wrap(err, ErrDsuUpdatesDirectoryMissing.Error())
	}

	// identify the inventory collector bin
	// dsu 1.8 drops update files under the given $updateDir
	// dsu 1.9 creates a directory '$updateDir/dellupdates' and drops the updates in there
	matches := findDSUInventoryCollector(updateDir)

	if len(matches) == 0 {
		return 0, errors.Wrap(ErrDsuInventoryCollectorBinMissing, updateDir)
	}

	if len(matches) > 1 {
		return 0, errors.Wrap(ErrMultipleDsuInventoryCollectorBins, strings.Join(matches, ","))
	}

	invcol := matches[0]
	// the updates directory is where the inventory collector bin is located
	updateDir = filepath.Dir(invcol)

	// dsu --log-level=4 --non-interactive --source-type=REPOSITORY --source-location=/root/dsu/dellupdates --ic-location=/root/dsu/dellupdates/invcol_5N2WM_LN64_20_09_200_921_A00.BIN
	d.Executor.SetArgs("--non-interactive", "--log-level=4", "--source-type=REPOSITORY", "--source-location="+updateDir, "--ic-location="+invcol)
	result, err := d.Executor.Exec(ctx)

	return result.ExitCode, err
}

// Inventory collects inventory with the dell-system-update utility and
// updates device component firmware based on data listed by the dell system update tool
func (d *Dsu) Inventory(ctx context.Context) ([]*model.Component, error) {
	d.Executor.SetArgs("--import-public-key", "--inventory")
	result, err := d.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	components := dsuParseInventoryBytes(result.Stdout)
	if len(components) == 0 {
		return nil, errors.Wrap(errs.ErrDeviceInventory, "no components returned by dsuParseInventoryBytes()")
	}

	return components, nil
}

// Returns component firmware updates available based on the dell system update
func (d *Dsu) ComponentFirmwareUpdatePreview(ctx context.Context) ([]*model.Component, int, error) {
	d.Executor.SetArgs("--import-public-key", "--preview")
	result, err := d.Executor.Exec(ctx)
	if err != nil {
		return nil, result.ExitCode, err
	}

	return dsuParsePreviewBytes(result.Stdout), result.ExitCode, nil
}

// ApplyUpdates installs all available updates
func (d *Dsu) ApplyUpdates() (int, error) {
	d.Executor.SetArgs("--non-interactive", "--log-level=4")
	result, err := d.Executor.Exec(context.Background())
	// our executor returns err if exitcode is not zero
	// 34 - no updates applicable
	return result.ExitCode, err
}

// Version returns the dsu currently installed
func (d *Dsu) Version() (string, error) {
	e := NewExecutor("rpm")
	e.SetArgs("-q", "dell-system-update", "--queryformat=%{VERSION}-%{RELEASE}")
	e.SetVerbose()

	result, err := e.Exec(context.Background())
	if err != nil {
		// our executor returns err if exitcode is not zero
		return "", errors.Wrap(ErrDsuVersionQuery, err.Error())
	}

	return string(result.Stdout), nil
}

// *** dsu output parser helpers **

// Parse dsu -i output and return a slice of Component
func dsuParseInventoryBytes(in []byte) []*model.Component {
	components := make([]*model.Component, 0)

	// see test file for sample data
	r := regexp.MustCompile(`(?m)^\d+\. \w+(:?|, (.*) \( Version : (.*) \))$`)
	matches := r.FindAllSubmatch(in, -1)

	// each matched line is expected to have 4 parts
	// 1. BIOS, BIOS  ( Version : 2.6.4 )
	cols := 4

	for _, m := range matches {
		if len(m) == cols {
			component := &model.Component{
				Slug:              dsuComponentNameToSlug(trimBytes(m[2])),
				Name:              trimBytes(m[2]),
				Vendor:            "dell",
				FirmwareInstalled: trimBytes(m[3]),
				Oem:               true,
				FirmwareManaged:   true,
			}

			components = append(components, component)
		}
	}

	return components
}

func dsuParsePreviewBytes(in []byte) []*model.Component {
	components := make([]*model.Component, 0)

	// see test file for sample data
	r := regexp.MustCompile(`(?m)^\d : \w+.*`)
	matches := r.FindAllSubmatch(in, -1)
	// each matched line is expected to have 5 parts
	// 3 : BIOS : BIOS : 2.8.1 : BIOS_RTWM9_LN_2.8.1
	cols := 5

	for _, m := range matches {
		s := strings.Split(string(m[0]), ":")
		if len(s) == cols {
			component := &model.Component{
				Slug:              dsuComponentNameToSlug(strings.TrimSpace(s[2])),
				Name:              strings.TrimSpace(s[2]),
				Vendor:            "dell",
				FirmwareAvailable: strings.TrimSpace(s[3]),
				Metadata:          make(map[string]string),
				Oem:               true,
				FirmwareManaged:   true,
			}
			component.Metadata["firmware_available_filename"] = strings.TrimSpace(s[4])
			components = append(components, component)
		}
	}

	return components
}

func trimBytes(b []byte) string {
	return strings.TrimSpace(string(b))
}

// Find the DSU inventory collector bin
func findDSUInventoryCollector(path string) []string {
	var found []string

	globs := []string{
		fmt.Sprintf("%s/invcol_*.BIN", path),
		fmt.Sprintf("%s/dellupdates/invcol_*.BIN", path),
	}

	for _, g := range globs {
		matches, err := filepath.Glob(g)
		if err == nil {
			found = append(found, matches...)
		}
	}

	return found
}

// returns the component slug for the given dell component name
//
// since the component name exposed by the dsu command doesn't tell the component name in a unique manner,
// the model.DellComponentSlug list has be ordered to ensure we don't have incorrect identification.
// Attempts were made to use fuzzy matching and levenstiens distance, to identify the components correctly,
// although none seemed to work as well as an ordered list.
func dsuComponentNameToSlug(n string) string {
	componentName := strings.ToLower(n)

	for _, componentSlug := range dellComponentSlug {
		identifier, slug := componentSlug[0], componentSlug[1]
		if strings.EqualFold(componentName, identifier) {
			return slug
		}

		if strings.Contains(componentName, identifier) {
			return slug
		}
	}

	return "unknown"
}
