package utils

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

const ipmicfg = "/usr/sbin/smc-ipmicfg"

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
// note: the binary is expected to be available as smc-ipmicfg,
//       as setup in the fup firmware-update image
func NewIpmicfgCmd(trace bool) *Ipmicfg {
	e := NewExecutor(ipmicfg)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Ipmicfg{Executor: e}
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
func (i Ipmicfg) BMC(ctx context.Context) (*model.BMC, error) {
	summary, err := i.Summary()
	if err != nil {
		return nil, err
	}

	return &model.BMC{
		Vendor:      "Supermicro",
		Description: model.SlugBMC,
		Firmware: &model.Firmware{
			Installed: summary.FirmwareRevision,
			Metadata: map[string]string{
				"build_date": summary.FirmwareBuildDate,
			},
		},
	}, nil
}

// BIOS returns a SMC BIOS component
func (i *Ipmicfg) BIOS(ctx context.Context) (*model.BIOS, error) {
	summary, err := i.Summary()
	if err != nil {
		return nil, err
	}

	// add CPLD and BIOS firmware inventory
	return &model.BIOS{
		Vendor:      "Supermicro",
		Description: model.SlugBIOS,
		Firmware: &model.Firmware{
			Installed: summary.BIOSVersion,
			Metadata: map[string]string{
				"build_date": summary.BIOSBuildDate,
			},
		},
	}, nil
}

// CPLD returns a SMC CPLD component
func (i Ipmicfg) CPLD(ctx context.Context) (*model.CPLD, error) {
	summary, err := i.Summary()
	if err != nil {
		return nil, err
	}

	return &model.CPLD{
		Vendor:      "Supermicro",
		Description: model.SlugCPLD,
		Firmware: &model.Firmware{
			Installed: summary.CPLDVersion,
		},
	}, nil
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
