package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/model"
)

type Dsu struct {
	Executor Executor
}

const (
	// see test_data/dsu_return_codes.md
	DSUExitCodeUpdatesApplied     = 0
	DSUExitCodeRebootRequired     = 8
	DSUExitCodeNoUpdatesAvailable = 34
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

// Applies available updates
func (d *Dsu) ApplyUpdates() (int, error) {

	d.Executor.SetArgs([]string{"--non-interactive", "--log-level=4"})
	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		// our executor returns err if exitcode is not zero
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
