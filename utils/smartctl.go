package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/packethost/ironlib/model"
	"github.com/pkg/errors"
)

const smartctlBin = "/usr/sbin/smartctl"

type Smartctl struct {
	Executor Executor
}

type SmartctlDriveAttributes struct {
	ModelName       string          `json:"model_name"`
	SerialNumber    string          `json:"serial_number"`
	FirmwareVersion string          `json:"firmware_version"`
	Status          *SmartctlStatus `json:"smart_status"`
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
	e := NewExecutor(smartctlBin)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Smartctl{Executor: e}
}

// Drives returns drives identified by smartctl
func (s *Smartctl) Drives(ctx context.Context) ([]*model.Drive, error) {
	drives := make([]*model.Drive, 0)

	DrivesList, err := s.Scan()
	if err != nil {
		return nil, err
	}

	for _, drive := range DrivesList.Drives {
		// collect drive information with smartctl -a <drive>
		smartctlAll, err := s.All(drive.Name)
		if err != nil {
			return nil, err
		}

		item := &model.Drive{
			Vendor:      model.VendorFromString(smartctlAll.ModelName),
			Model:       smartctlAll.ModelName,
			Serial:      smartctlAll.SerialNumber,
			ProductName: smartctlAll.ModelName,
			// TODO: use the smartctl form_factor/rotational attributes to determine the drive type
			Type: model.DriveTypeSlug(smartctlAll.ModelName),
			Firmware: &model.Firmware{
				Installed: smartctlAll.FirmwareVersion,
			},
			SmartStatus: "unknown",
		}

		if smartctlAll.Status != nil {
			if smartctlAll.Status.Passed {
				item.SmartStatus = "true"
			} else {
				item.SmartStatus = "false"
			}
		}

		drives = append(drives, item)
	}

	return drives, nil
}

// Scan runs smartctl scan -j and returns its value as an object
func (s *Smartctl) Scan() (*SmartctlScan, error) {
	s.Executor.SetArgs([]string{"--scan", "-j"})

	result, err := s.Executor.ExecWithContext(context.Background())
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
func (s *Smartctl) All(device string) (*SmartctlDriveAttributes, error) {
	// smartctl -a /dev/sda1 -j
	s.Executor.SetArgs([]string{"-a", device, "-j"})

	result, err := s.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrNoCommandOutput, s.Executor.GetCmd())
	}

	deviceAttributes := &SmartctlDriveAttributes{}

	err = json.Unmarshal(result.Stdout, deviceAttributes)
	if err != nil {
		return nil, err
	}

	return deviceAttributes, nil
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
	// Executor embedded in here to skip having to implement all the utils.Executor methods
	Executor
}

// NewFakeSmartctlExecutor returns a fake smartctl executor for tests
func NewFakeSmartctlExecutor(cmd, dir string) Executor {
	return &FakeSmartctlExecute{Cmd: cmd, JSONFilesDir: dir}
}

// NewFakeSmartctl returns a fake smartctl object for testing
func NewFakeSmartctl(dataDir string) *Smartctl {
	executor := NewFakeSmartctlExecutor("smartctl", dataDir)
	return &Smartctl{Executor: executor}
}

// nolint:gocyclo // test code
// ExecWithContext implements the utils.Executor interface
func (e *FakeSmartctlExecute) ExecWithContext(ctx context.Context) (*Result, error) {
	switch e.Args[0] {
	case "--scan":
		b, err := ioutil.ReadFile(e.JSONFilesDir + "/scan.json")
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

		b, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}

		e.Stdout = b
	}

	return &Result{Stdout: e.Stdout, Stderr: e.Stderr, ExitCode: 0}, nil
}

// SetStdin is to set input to the fake execute method
func (e *FakeSmartctlExecute) SetStdin(r io.Reader) {
	e.Stdin = r
}

// SetArgs is to set cmd args to the fake execute method
func (e *FakeSmartctlExecute) SetArgs(a []string) {
	e.Args = a
}
