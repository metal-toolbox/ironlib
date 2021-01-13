package utils

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/model"
)

const storecli = "/opt/MegaRAID/storcli/storcli64"

type StoreCLI struct {
	Executor Executor
}

type ShowController struct {
	Controllers []*Controller `json:"Controllers"`
}

type Controller struct {
	CommandStatus *CommandStatus `json:"Command Status"`
	ResponseData  *ResponseData  `json:"Response Data"`
}

type CommandStatus struct {
	Status      string `json:"Status"`
	Description string `json:"Description"`
}

type ResponseData struct {
	ProductName     string `json:"Product Name"`
	SerialNumber    string `json:"Serial Number"`
	FirmwareVersion string `json:"FW Version"`
	BIOSVersion     string `json:"BIOS Version"`
	PhysicalDrives  int    `json:"Physical Drives"`
}

// Return a new nvme executor
func NewStoreCLICmd(trace bool) Collector {

	e := NewExecutor(storecli)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	if !trace {
		e.SetQuiet()
	}

	return &StoreCLI{Executor: e}
}

// Executes nvme list, parses the output and returns a slice of model.Component
func (s *StoreCLI) Components() ([]*model.Component, error) {

	inv := make([]*model.Component, 0)

	out, err := s.ShowController0()
	if err != nil {
		return nil, err
	}

	list := &ShowController{Controllers: []*Controller{}}
	err = json.Unmarshal(out, list)
	if err != nil {
		return nil, err
	}

	uid, _ := uuid.NewRandom()
	for idx, c := range list.Controllers {

		if strings.Contains(c.CommandStatus.Description, "not found") || c.CommandStatus.Status == "Failure" {
			continue
		}

		item := &model.Component{
			ID:                uid.String(),
			Serial:            c.ResponseData.SerialNumber,
			Vendor:            vendorFromString(c.ResponseData.ProductName),
			Model:             c.ResponseData.ProductName,
			FirmwareInstalled: c.ResponseData.FirmwareVersion,
			Slug:              prefixIndex(idx, "Serial Attached SCSI controller"),
			Name:              "Serial Attached SCSI controller", // based on lspci
		}
		inv = append(inv, item)
	}

	return inv, nil
}

func (s *StoreCLI) ShowController0() ([]byte, error) {
	// /opt/MegaRAID/storcli/storcli64 /c0 show J
	s.Executor.SetArgs([]string{"/c0", "show", "J"})

	result, err := s.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}
