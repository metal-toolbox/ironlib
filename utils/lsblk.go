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

	out, err := l.list(ctx)
	if err != nil {
		return nil, err
	}

	items := map[string][]*lsblkDeviceAttributes{}

	err = json.Unmarshal(out, &items)
	if err != nil {
		return nil, err
	}

	for _, d := range items["blockdevices"] {
		dModel := d.Model

		var vendor string

		modelTokens := strings.Split(d.Model, " ")

		if len(modelTokens) > 1 {
			vendor = modelTokens[1]
		}

		drive := &common.Drive{
			Protocol: d.Transport,
			Common: common.Common{
				Serial:      d.Serial,
				Vendor:      vendor,
				Model:       dModel,
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

func (l *Lsblk) list(ctx context.Context) ([]byte, error) {
	// lsblk --json --nodeps --output name,model,serial,tran -e1,7,11
	l.Executor.SetArgs([]string{"--json", "--nodeps", "--output", "name,path,model,serial,rev,tran", "-e1,7,11"})

	result, err := l.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

// NewFakeLsblk returns a mock lsblk collector that returns mock data for use in tests.
func NewFakeLsblk() *Lsblk {
	return &Lsblk{
		Executor: NewFakeExecutor("lsblk"),
	}
}
