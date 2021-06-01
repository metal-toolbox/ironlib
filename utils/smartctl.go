package utils

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/packethost/ironlib/model"
	"github.com/pkg/errors"
)

const smartctl = "/usr/sbin/smartctl"

type Smartctl struct {
	Executor Executor
}

type SmartctlDriveAttributes struct {
	ModelName       string `json:"model_name"`
	SerialNumber    string `json:"serial_number"`
	FirmwareVersion string `json:"firmware_version"`
}

type SmartctlScan struct {
	Drives []*SmartctlDrive `json:"Devices"`
}

type SmartctlDrive struct {
	Name     string `json:"name"`     // /dev/sdX
	Type     string `json:"type"`     // scsi / nvme
	Protocol string `json:"protocol"` // SCSI / NVMe
}

// Return a new smartctl executor
func NewSmartctlCmd(trace bool) Collector {
	e := NewExecutor(smartctl)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Smartctl{Executor: e}
}

// Components returns drives identified by smartctl
func (s *Smartctl) Components() ([]*model.Component, error) {
	components := make([]*model.Component, 0)

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

		item := &model.Component{
			Vendor:            model.VendorFromString(smartctlAll.ModelName),
			Model:             smartctlAll.ModelName,
			Serial:            smartctlAll.SerialNumber,
			Name:              smartctlAll.ModelName,
			Type:              model.DriveTypeSlug(smartctlAll.ModelName),
			Slug:              model.SlugDrive,
			FirmwareInstalled: smartctlAll.FirmwareVersion,
		}

		components = append(components, item)
	}

	return components, nil
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

// ExecWithContext implements the utils.Executor interface
func (e *FakeSmartctlExecute) ExecWithContext(ctx context.Context) (*Result, error) {
	switch e.Args[0] {
	case "--scan":
		b, err := ioutil.ReadFile(e.JSONFilesDir + "/smartctl_scan.json")
		if err != nil {
			return nil, err
		}

		e.Stdout = b
	case "-a":
		if strings.Join(e.Args, " ") == "-a /dev/sda -j" {
			b, err := ioutil.ReadFile(e.JSONFilesDir + "/smartctl_sda.json")
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		}

		if strings.Join(e.Args, " ") == "-a /dev/nvme0 -j" {
			b, err := ioutil.ReadFile(e.JSONFilesDir + "/smartctl_nvme0.json")
			if err != nil {
				return nil, err
			}

			e.Stdout = b
		}
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
