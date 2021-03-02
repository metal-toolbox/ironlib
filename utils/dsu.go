package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/model"
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

	LocalUpdatesDirectory = "/root/dsu"
)

// Returns a executor to run dsu commands
// if trace is enabled, stdout is printed to the terminal
func NewDsu(trace bool) *Dsu {

	e := NewExecutor("dsu")
	if !trace {
		e.SetQuiet()
	}

	return &Dsu{
		Executor: e,
	}
}

// Returns a dsu instance with a fake executor for tests
func NewFakeDsu() *Dsu {
	return &Dsu{
		Executor: NewFakeExecutor("dsu"),
	}
}

// Fetch updates to local directory - these updates are DSU specific
// returns the exitcode and error if any
func (d *Dsu) FetchUpdateFiles() (int, error) {

	// purge any existing update file/directory with the same name
	_ = os.Remove(LocalUpdatesDirectory)

	d.Executor.SetArgs([]string{"--destination-type=CBD", "--destination-location=" + LocalUpdatesDirectory})

	// because... yeah dsu wants to fetch updates interactively
	d.Executor.SetStdin(bytes.NewReader([]byte(`a\nc\n`)))

	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		return result.ExitCode, err
	}

	return result.ExitCode, nil
}

// Apply local update files - works for files fetched by FetchUpdateFiles()
// DSU needs to be pointed to the right inventory bin or it barfs
// returns the resulting exitcode and error if any
func (d *Dsu) ApplyLocalUpdates(updateDir string) (int, error) {

	// ensure the updates directory exists
	_, err := os.Stat(updateDir)
	if err != nil {
		return 0, errors.Wrap(err, "expected updates directory not present")
	}

	// identify the inventory collector bin
	matches, err := filepath.Glob(fmt.Sprintf("%s/invcol_*.BIN", updateDir))
	if err != nil {
		return 0, err
	}

	if matches == nil || len(matches) == 0 {
		return 0, fmt.Errorf("inventory collector bin missing from: %s", updateDir)
	}

	if len(matches) > 1 {
		return 0, fmt.Errorf("expected a single inventory collector bin, found multiple: %s", strings.Join(matches, ","))
	}

	invcol := matches[0]

	//dsu --log-level=4 --non-interactive --source-type=REPOSITORY --source-location=/root/dsu/dellupdates --ic-location=/root/dsu/invcol_5N2WM_LN64_20_09_200_921_A00.BIN
	d.Executor.SetArgs([]string{"--non-interactive", "--log-level=4", "--source-type=REPOSITORY", "--source-location=" + updateDir, "--ic-location=" + invcol})
	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		return result.ExitCode, err
	}

	return result.ExitCode, nil
}

// Returns components identified by dell system update and their current firmware revisions
func (d *Dsu) ComponentInventory() ([]*model.Component, error) {

	d.Executor.SetArgs([]string{"--import-public-key", "--inventory"})

	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return dsuParseInventoryBytes(result.Stdout), nil
}

// Returns component firmware updates available based on the dell system update
func (d *Dsu) ComponentFirmwareUpdatePreview() ([]*model.Component, int, error) {

	d.Executor.SetArgs([]string{"--import-public-key", "--preview"})

	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, result.ExitCode, err
	}

	return dsuParsePreviewBytes(result.Stdout), result.ExitCode, nil
}

// Run DSU to apply all available updates
func (d *Dsu) ApplyUpdates() (int, error) {

	args := []string{"--non-interactive", "--log-level=4"}
	d.Executor.SetArgs(args)
	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		// our executor returns err if exitcode is not zero
		// 34 - no updates applicable
		return result.ExitCode, err
	}

	return result.ExitCode, nil
}

// Returns the version of dsu currently installed
func (d *Dsu) Version() (string, error) {

	e := NewExecutor("rpm")
	e.SetArgs([]string{"-q", "dell-system-update", "--queryformat=%{VERSION}-%{RELEASE}"})
	result, err := e.ExecWithContext(context.Background())
	e.SetVerbose()
	if err != nil {
		// our executor returns err if exitcode is not zero
		return "", fmt.Errorf("error querying dsu version: " + err.Error())
	}

	return string(result.Stdout), nil
}

// *** dsu output parser helpers **

// Parse dsu -i output and return a slice of Component
func dsuParseInventoryBytes(in []byte) []*model.Component {

	components := make([]*model.Component, 0)

	// see test test file for sample data
	r := regexp.MustCompile(`(?m)^\d+\. \w+(:?|, (.*) \( Version : (.*) \))$`)
	matches := r.FindAllSubmatch(in, -1)
	for _, m := range matches {
		if len(m) == 4 {
			uid, _ := uuid.NewRandom()
			component := &model.Component{
				ID:                uid.String(),
				Slug:              componentNameToSlug(trimBytes(m[2])),
				Name:              trimBytes(m[2]),
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
	for _, m := range matches {
		s := strings.Split(string(m[0]), ":")
		if len(s) == 5 {
			uid, _ := uuid.NewRandom()
			component := &model.Component{
				ID:                uid.String(),
				Slug:              componentNameToSlug(strings.TrimSpace(s[2])),
				Name:              strings.TrimSpace(s[2]),
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

// nolint: gocyclo
// returns the component name normalized
// The component name slug is used in an index along with the device ID
// to uniquely identify the component and update its records, each time fup sends across inventory data
// it is required that components marked as Unknown are identified and given a unique identifier
func componentNameToSlug(n string) string {
	switch name := strings.ToLower(n); {
	case strings.Contains(name, "ethernet"):
		return "NIC"
	case strings.Contains(name, "idrac service module"):
		return "iDrac Service Module"
	case strings.Contains(name, "idrac"), strings.Contains(name, "integrated dell remote access controller"):
		return "BMC"
	case strings.Contains(name, "backplane"):
		return "Backplane Expander"
	case strings.Contains(name, "power supply"):
		return "Power Supply"
	case strings.Contains(name, "disk 0 of boss adapter"):
		return "Boss Adapter - Disk 0"
	case strings.Contains(name, "boss"):
		return "Boss Adapter"
	case strings.Contains(name, "bios"):
		return "BIOS"
	case strings.Contains(name, "hba330"):
		return "SAS HBA330 Controller"
	case strings.Contains(name, "nvmepcissd"):
		return "Disk - NVME PCI SSD"
	case strings.Contains(name, "system cpld"):
		return "System CPLD"
	case strings.Contains(name, "sep firmware"):
		return "Non-expander Storage Backplane (SEP)"
	case strings.Contains(name, "lifecycle controller"):
		return "Lifecycle Controller"
	case strings.Contains(name, "os collector"):
		return "OS collector"
	case strings.Contains(name, "dell 64 bit uefi diagnostics"):
		return "Dell 64 bit uEFI diagnostics"
	default:
		return "Unknown"
	}
}
