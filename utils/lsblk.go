package utils

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

const (
	EnvLsblkUtility                        = "IRONLIB_UTIL_LSBLK"
	LsblkUtility    model.CollectorUtility = "lsblk"
)

var ErrLsblkTransportUnsupported = errors.New("Unsupported transport type")

type Lsblk struct {
	Executor Executor
}

type lsblkDeviceAttributes struct {
	Name      string `json:"name"`
	Device    string `json:"kname"`
	Model     string `json:"model"`
	Serial    string `json:"serial"`
	Firmware  string `json:"rev:"`
	Transport string `json:"tran"`
}

// Return a new lsblk executor
func NewLsblkCmd(trace bool) *Lsblk {
	utility := string(LsblkUtility)

	// lookup env var for util
	if eVar := os.Getenv(EnvLsblkUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Lsblk{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (l *Lsblk) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := l.Executor.CheckExecutable()

	return LsblkUtility, l.Executor.CmdPath(), er
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
			Protocol: strings.ToLower(d.Transport),
			Common: common.Common{
				LogicalName: strings.TrimSpace(d.Device),
				Serial:      strings.TrimSpace(d.Serial),
				Vendor:      strings.TrimSpace(vendor),
				Model:       strings.TrimSpace(dModel),
			},
			StorageControllerDriveID: -1,
		}

		drives = append(drives, drive)
	}

	return drives, nil
}

func (l *Lsblk) list(ctx context.Context) ([]byte, error) {
	// lsblk --json --nodeps --output name,path,model,serial,tran -e1,7,11
	l.Executor.SetArgs("--json", "--nodeps", "-p", "--output", "kname,name,model,serial,rev,tran", "-e1,7,11")

	result, err := l.Executor.Exec(ctx)
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
