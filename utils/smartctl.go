package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

const EnvSmartctlUtility = "IRONLIB_UTIL_SMARTCTL"

// map of smartctl error bits to explanation - man 8 smartctl
var smartclErrors = map[int]string{
	0: "Command line did not parse",
	1: "Device open failed, device did not return an IDENTIFY DEVICE structure, or device is in a low-power mode",
	2: "Some SMART or other ATA command to the disk failed, or there was a checksum error in a SMART data structure",
	3: "SMART status check returned 'DISK FAILING'",
	4: "We found prefail Attributes <= threshold",
	5: "SMART status check returned 'DISK OK' but we found that some (usage or prefail) Attributes have been <= threshold at some time in the past",
	6: "The device error log contains records of errors",
	7: "The device self-test log contains records of errors. [ATA only] Failed self-tests outdated by a newer successful extended self-test are ignored",
}

type Smartctl struct {
	Executor Executor
}

type SmartctlDriveAttributes struct {
	ModelName       string          `json:"model_name"`
	OemProductID    string          `json:"ata_additional_product_id"`
	ModelFamily     string          `json:"model_family"`
	SerialNumber    string          `json:"serial_number"`
	FirmwareVersion string          `json:"firmware_version"`
	Status          *SmartctlStatus `json:"smart_status"`
	Errors          []string        `json:"-"`
}

type SmartctlScan struct {
	Drives []*SmartctlDrive `json:"Devices"`
}

type SmartctlDrive struct {
	Name     string `json:"name"`     // /dev/sdX
	Type     string `json:"type"`     // scsi / nvme
	Protocol string `json:"protocol"` // SCSI / NVMe
}

type SmartctlStatus struct {
	Passed bool `json:"passed"`
}

// Return a new smartctl executor
func NewSmartctlCmd(trace bool) *Smartctl {
	utility := "smartctl"

	// lookup env var for util
	if eVar := os.Getenv(EnvSmartctlUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Smartctl{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (s *Smartctl) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := s.Executor.CheckExecutable()

	return "smartctl", s.Executor.CmdPath(), er
}

// Drives returns drives identified by smartctl
func (s *Smartctl) Drives(ctx context.Context) ([]*common.Drive, error) {
	drives := make([]*common.Drive, 0)

	DrivesList, err := s.Scan(ctx)
	if err != nil {
		return nil, err
	}

	for _, drive := range DrivesList.Drives {
		// collect drive information with smartctl -a <drive>
		smartctlAll, err := s.All(ctx, drive.Name)
		if err != nil {
			return nil, err
		}

		item := &common.Drive{
			Common: common.Common{
				Vendor:      common.VendorFromString(smartctlAll.ModelName),
				Model:       smartctlAll.ModelName,
				Serial:      smartctlAll.SerialNumber,
				ProductName: smartctlAll.ModelName,
				// TODO: use the smartctl form_factor/rotational attributes to determine the drive type

				Firmware: &common.Firmware{
					Installed: smartctlAll.FirmwareVersion,
				},
			},

			Type:                     model.DriveTypeSlug(smartctlAll.ModelName),
			SmartStatus:              common.SmartStatusUnknown,
			StorageControllerDriveID: -1,
		}

		if item.Vendor == "" {
			item.Vendor = common.VendorFromString(smartctlAll.ModelFamily)
		}

		item.OemID = strings.TrimSpace(smartctlAll.OemProductID)

		if smartctlAll.Status != nil {
			if smartctlAll.Status.Passed {
				item.SmartStatus = common.SmartStatusOK
			} else {
				item.SmartStatus = common.SmartStatusFailed
			}
		}

		if len(item.SmartErrors) > 0 {
			item.SmartErrors = smartctlAll.Errors
		}

		drives = append(drives, item)
	}

	return drives, nil
}

// Scan runs smartctl scan -j and returns its value as an object
func (s *Smartctl) Scan(ctx context.Context) (*SmartctlScan, error) {
	s.Executor.SetArgs([]string{"--scan", "-j"})

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrNoCommandOutput, s.Executor.GetCmd())
	}

	list := &SmartctlScan{Drives: []*SmartctlDrive{}}

	err = json.Unmarshal(result.Stdout, list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// All runs smartctl -a /dev/<device> and returns its value as an object
func (s *Smartctl) All(ctx context.Context, device string) (*SmartctlDriveAttributes, error) {
	// smartctl -a /dev/sda1 -j
	s.Executor.SetArgs([]string{"-a", device, "-j"})

	// smartctl can exit with a non-zero status based on drive smart data
	result, _ := s.Executor.ExecWithContext(ctx)
	// determine the errors if any based on the exit code
	smartCtlErrs := smartCtlExitStatus(result.ExitCode)

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrNoCommandOutput, s.Executor.GetCmd())
	}

	deviceAttributes := &SmartctlDriveAttributes{}

	err := json.Unmarshal(result.Stdout, deviceAttributes)
	if err != nil {
		return nil, err
	}

	// smartctl identified errors are included in returned attributes
	if len(smartCtlErrs) > 0 {
		deviceAttributes.Errors = smartCtlErrs
	}

	return deviceAttributes, nil
}

// smartCtlExitStatus identifies the error bits in the smartctl exitcode
// and returns a slice of strings containing one or more errors identified - if any.
// an empty slice is returned if there are no errors
func smartCtlExitStatus(exitCode int) []string {
	// man 8 smartctl
	// Return Values
	// The return values of smartctl are defined by a bitmask.
	// If all is well with the disk, the return value (exit status) of smartctl is 0 (all bits turned off).
	// If a problem occurs, or an error, potential error, or fault is detected, then a non-zero status is returned.
	// In this case, the eight different bits in the return value have the following meanings for ATA disks,
	// some of these values may also be returned for SCSI disks.
	e := []string{}

	// identify smartctl error bits set
	bits := maskExitCode(exitCode)
	if len(bits) == 0 {
		return e
	}

	for _, i := range bits {
		e = append(e, smartclErrors[i])
	}

	return e
}

// maskExitCode identifies error bits set based on the smartctl exit code
// it returns a slice of bits that are set to a value > 0
// see man 8 smartctl for details
// nolint:gomnd // comments clarify magic numbers
func maskExitCode(e int) []int {
	set := []int{}

	if e == 0 {
		return set
	}

	for i := 0; i < 8; i++ {
		// check exit code has bit set
		x := (e & exponentInt(2, i))
		if x != 0 {
			set = append(set, i)
		}
	}

	return set
}

// exponentInt returns the exponent of n to the power of a
func exponentInt(a, n int) int {
	var i, result int
	result = 1

	for i = 0; i < n; i++ {
		result *= a
	}

	return result
}

// FakeSmartctlExecute implements the utils.Executor interface for testing
type FakeSmartctlExecute struct {
	Cmd          string
	Args         []string
	Env          []string
	Stdin        io.Reader
	Stdout       []byte // Set this for the dummy data to be returned
	Stderr       []byte // Set this for the dummy data to be returned
	Quiet        bool
	JSONFilesDir string
	ExitCode     int
	CheckBin     bool
	// Executor embedded in here to skip having to implement all the utils.Executor methods
	Executor
}

// NewFakeSmartctlExecutor returns a fake smartctl executor for tests
func NewFakeSmartctlExecutor(cmd, dir string) Executor {
	return &FakeSmartctlExecute{Cmd: cmd, JSONFilesDir: dir, CheckBin: false}
}

// NewFakeSmartctl returns a fake smartctl object for testing
func NewFakeSmartctl(dataDir string) *Smartctl {
	executor := NewFakeSmartctlExecutor("smartctl", dataDir)
	return &Smartctl{Executor: executor}
}

// nolint:gocyclo // test code
// ExecWithContext implements the utils.Executor interface
func (e *FakeSmartctlExecute) ExecWithContext(context.Context) (*Result, error) {
	switch e.Args[0] {
	case "--scan":
		b, err := os.ReadFile(e.JSONFilesDir + "/scan.json")
		if err != nil {
			return nil, err
		}

		e.Stdout = b
	case "-a":
		// -a /dev/sdg -j
		argLength := 3
		if len(e.Args) < argLength {
			return nil, ErrFakeExecutorInvalidArgs
		}

		driveName := path.Base(e.Args[1])
		f := fmt.Sprintf("%s/%s.json", e.JSONFilesDir, driveName)

		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		e.Stdout = b
	}

	return &Result{Stdout: e.Stdout, Stderr: e.Stderr, ExitCode: e.ExitCode}, nil
}

// CheckExecutable implements the Executor interface
func (e *FakeSmartctlExecute) CheckExecutable() error {
	return nil
}

// SetStdin is to set input to the fake execute method
func (e *FakeSmartctlExecute) SetStdin(r io.Reader) {
	e.Stdin = r
}

// SetArgs is to set cmd args to the fake execute method
func (e *FakeSmartctlExecute) SetArgs(a []string) {
	e.Args = a
}

func (e *FakeSmartctlExecute) SetExitCode(i int) {
	e.ExitCode = i
}

func (e *FakeSmartctlExecute) CmdPath() string {
	return e.Cmd
}
