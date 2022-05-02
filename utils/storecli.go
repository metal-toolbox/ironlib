package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"

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

// Return a new storecli executor
func NewStoreCLICmd(trace bool) *StoreCLI {
	e := NewExecutor(storecli)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &StoreCLI{Executor: e}
}

// Return a Fake storecli executor for tests
func NewFakeStoreCLI(r io.Reader) (*StoreCLI, error) {
	e := NewFakeExecutor("storecli")
	b := bytes.Buffer{}

	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	e.SetStdout(b.Bytes())

	return &StoreCLI{
		Executor: e,
	}, nil
}

// Executes nvme list, parses the output and returns a slice of model.StorageControllers
func (s *StoreCLI) StorageControllers(ctx context.Context) ([]*model.StorageController, error) {
	controllers := make([]*model.StorageController, 0)

	out, err := s.ShowController0()
	if err != nil {
		return nil, err
	}

	list := &ShowController{Controllers: []*Controller{}}

	err = json.Unmarshal(out, list)
	if err != nil {
		return nil, err
	}

	for _, c := range list.Controllers {
		if strings.Contains(c.CommandStatus.Description, "not found") || c.CommandStatus.Status == "Failure" {
			continue
		}

		item := &model.StorageController{
			Serial:      c.ResponseData.SerialNumber,
			Vendor:      model.VendorFromString(c.ResponseData.ProductName),
			Model:       c.ResponseData.ProductName,
			Description: c.ResponseData.ProductName,
			Metadata:    map[string]string{"drives_attached": strconv.Itoa(c.ResponseData.PhysicalDrives)},
			Firmware: &model.Firmware{
				Installed: c.ResponseData.FirmwareVersion,
				Metadata:  map[string]string{"bios_version": c.ResponseData.BIOSVersion},
			},
		}
		controllers = append(controllers, item)
	}

	return controllers, nil
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
