package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
)

const EnvStorecliUtility = "IRONLIB_UTIL_STORECLI"

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
	utility := "/opt/MegaRAID/storcli/storcli64"

	// lookup env var for util
	if eVar := os.Getenv(EnvStorecliUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &StoreCLI{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (s *StoreCLI) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := s.Executor.CheckExecutable()

	return "storecli", s.Executor.CmdPath(), er
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

// StorageControllers returns a slice of model.StorageControllers from the output of nvme list
func (s *StoreCLI) StorageControllers(ctx context.Context) ([]*common.StorageController, error) {
	controllers := make([]*common.StorageController, 0)

	out, err := s.ShowController0(ctx)
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

		item := &common.StorageController{
			Common: common.Common{
				Serial:      c.ResponseData.SerialNumber,
				Vendor:      common.VendorFromString(c.ResponseData.ProductName),
				Model:       c.ResponseData.ProductName,
				Description: c.ResponseData.ProductName,
				Metadata:    map[string]string{"drives_attached": strconv.Itoa(c.ResponseData.PhysicalDrives)},
				Firmware: &common.Firmware{
					Installed: c.ResponseData.FirmwareVersion,
					Metadata:  map[string]string{"bios_version": c.ResponseData.BIOSVersion},
				},
			},
		}
		controllers = append(controllers, item)
	}

	return controllers, nil
}

// ShowController0 runs storecli to list controller 0
func (s *StoreCLI) ShowController0(ctx context.Context) ([]byte, error) {
	// /opt/MegaRAID/storcli/storcli64 /c0 show J
	s.Executor.SetArgs([]string{"/c0", "show", "J"})

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}
