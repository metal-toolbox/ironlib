package utils

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

const EnvSmcIpmicfgUtility = "IRONLIB_UTIL_SMC_IPMICFG"

type Ipmicfg struct {
	Executor Executor
}

type IpmicfgSummary struct {
	FirmwareRevision  string // BMC
	FirmwareBuildDate string
	BIOSVersion       string
	BIOSBuildDate     string
	CPLDVersion       string
}

// Return a new Supermicro IPMICFG executor
func NewIpmicfgCmd(trace bool) *Ipmicfg {
	utility := "smc-ipmicfg"

	// lookup env var for util
	if eVar := os.Getenv(EnvSmcIpmicfgUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Ipmicfg{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (i *Ipmicfg) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := i.Executor.CheckExecutable()

	return "smc-ipmicfg", i.Executor.CmdPath(), er
}

// Fake IPMI executor for tests
func NewFakeIpmicfg(r io.Reader) *Ipmicfg {
	e := NewFakeExecutor("ipmicfg")
	e.SetStdin(r)

	return &Ipmicfg{
		Executor: e,
	}
}

// BMC returns a SMC BMC component
func (i Ipmicfg) BMC(ctx context.Context) (*common.BMC, error) {
	summary, err := i.Summary()
	if err != nil {
		return nil, err
	}

	return &common.BMC{
		Common: common.Common{
			Vendor:      "Supermicro",
			Description: common.SlugBMC,
			Firmware: &common.Firmware{
				Installed: summary.FirmwareRevision,
				Metadata: map[string]string{
					"build_date": summary.FirmwareBuildDate,
				},
			},
		},
	}, nil
}

// BIOS returns a SMC BIOS component
func (i *Ipmicfg) BIOS(ctx context.Context) (*common.BIOS, error) {
	summary, err := i.Summary()
	if err != nil {
		return nil, err
	}

	// add CPLD and BIOS firmware inventory
	return &common.BIOS{
		Common: common.Common{
			Vendor:      "Supermicro",
			Description: common.SlugBIOS,
			Firmware: &common.Firmware{
				Installed: summary.BIOSVersion,
				Metadata: map[string]string{
					"build_date": summary.BIOSBuildDate,
				},
			},
		},
	}, nil
}

// CPLDs returns a slice of SMC CPLD components
func (i Ipmicfg) CPLDs(ctx context.Context) ([]*common.CPLD, error) {
	summary, err := i.Summary()
	if err != nil {
		return nil, err
	}

	cplds := []*common.CPLD{}
	cplds = append(cplds, &common.CPLD{
		Common: common.Common{
			Vendor:      "Supermicro",
			Description: common.SlugCPLD,
			Firmware: &common.Firmware{
				Installed: summary.CPLDVersion,
			},
		},
	})

	return cplds, nil
}

func (i *Ipmicfg) Summary() (*IpmicfgSummary, error) {
	// smc-ipmicfg --summary
	i.Executor.SetArgs([]string{"-summary"})

	result, err := i.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrNoCommandOutput, i.Executor.GetCmd())
	}

	return i.parseQueryOutput(result.Stdout), nil
}

func (i *Ipmicfg) parseQueryOutput(bSlice []byte) *IpmicfgSummary {
	summary := &IpmicfgSummary{}

	lines := bytes.Split(bSlice, []byte("\n"))
	for _, line := range lines {
		s := string(line)

		cols := 2
		parts := strings.Split(s, ":")

		if len(parts) < cols {
			continue
		}

		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

		switch key {
		case "Firmware Revision":
			summary.FirmwareRevision = value
		case "Firmware Build Time":
			summary.FirmwareBuildDate = value
		case "BIOS Version":
			summary.BIOSVersion = value
		case "BIOS Build Time":
			summary.BIOSBuildDate = value
		case "CPLD Version":
			summary.CPLDVersion = value
		}
	}

	return summary
}

// NewFakeSMCIpmiCfg returns a fake lshw executor for testing
func NewFakeSMCIpmiCfg() *SupermicroSUM {
	executor := &FakeExecute{}
	return &SupermicroSUM{Executor: executor}
}
