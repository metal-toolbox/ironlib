package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/model"
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

func (s *Smartctl) Components() ([]*model.Component, error) {

	components := make([]*model.Component, 0)
	DrivesList, err := s.Scan()
	if err != nil {
		return nil, err
	}

	for idx, drive := range DrivesList.Drives {
		// collect drive information with smartctl -a <drive>
		smartctlAll, err := s.All(drive.Name)
		if err != nil {
			return nil, err
		}

		uid, _ := uuid.NewRandom()
		item := &model.Component{
			ID:                uid.String(),
			Vendor:            vendorFromString(smartctlAll.ModelName),
			Model:             smartctlAll.ModelName,
			Serial:            smartctlAll.SerialNumber,
			Slug:              prefixIndex(idx, drive.Type),
			Name:              drive.Type,
			FirmwareInstalled: smartctlAll.FirmwareVersion,
		}

		components = append(components, item)
	}

	return components, nil
}

func (s *Smartctl) Scan() (*SmartctlScan, error) {

	// smartctl scan -j
	s.Executor.SetArgs([]string{"--scan", "-j"})

	result, err := s.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, fmt.Errorf("no output from command: %s", s.Executor.GetCmd())
	}

	list := &SmartctlScan{Drives: []*SmartctlDrive{}}
	err = json.Unmarshal(result.Stdout, list)
	if err != nil {
		return nil, err
	}

	return list, nil

}

func (s *Smartctl) All(device string) (*SmartctlDriveAttributes, error) {

	// smartctl -a /dev/sda1 -j
	s.Executor.SetArgs([]string{"-a", device, "-j"})

	result, err := s.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, fmt.Errorf("no output from command: %s", s.Executor.GetCmd())
	}

	deviceAttributes := &SmartctlDriveAttributes{}
	err = json.Unmarshal(result.Stdout, deviceAttributes)
	if err != nil {
		return nil, err
	}

	return deviceAttributes, nil

}
