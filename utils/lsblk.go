package utils

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
)

const lsblk = "/usr/bin/lsblk"

var ErrLsblkTransportUnsupported = errors.New("Unsupported transport type")

type Lsblk struct {
	Executor Executor
}

type lsblkDeviceAttributes struct {
	Name      string `json:"name"`
	Device    string `json:"path"`
	Model     string `json:"model"`
	Serial    string `json:"serial"`
	Firmware  string `json:"rev:"`
	Transport string `json:"tran"`
}

type lsblkList struct {
	Devices []*lsblkDeviceAttributes `json:"Devices"`
}

// Return a new lsblk executor
func NewLsblkCmd(trace bool) *Lsblk {
	e := NewExecutor(lsblk)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Lsblk{Executor: e}
}

// Executes lsblk list, parses the output and returns a slice of *common.Drive
func (l *Lsblk) Drives(ctx context.Context) ([]*common.Drive, error) {
	drives := make([]*common.Drive, 0)

	out, err := l.List()
	if err != nil {
		return nil, err
	}

	list := &lsblkList{Devices: []*lsblkDeviceAttributes{}}

	err = json.Unmarshal(out, list)
	if err != nil {
		return nil, err
	}

	for _, d := range list.Devices {
		dModel := d.Model

		var vendor string
		var product []string

		modelTokens := strings.Split(d.Model, " ")

		if len(modelTokens) > 1 {
			vendor = modelTokens[1]
			product = append(modelTokens[:1], modelTokens[2:]...)
		}

		productName := strings.Join(product, " ")

		drive := &common.Drive{
			Protocol: d.Transport,
			Common: common.Common{
				Serial:      d.Serial,
				Vendor:      vendor,
				Model:       dModel,
				ProductName: productName,
				Description: d.Model,
				Firmware: &common.Firmware{
					Installed: d.Firmware,
				},
				Metadata: map[string]string{},
			},
		}

		drives = append(drives, drive)
	}

	return drives, nil
}

func (l *Lsblk) List() ([]byte, error) {
	// lsblk -Jdo name,model,serial,tran -e1,7,11
	l.Executor.SetArgs([]string{"-Jdo", "name,path,model,serial,rev,tran", "-e1,7,11"})

	result, err := l.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}
