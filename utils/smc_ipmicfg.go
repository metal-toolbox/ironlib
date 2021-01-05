package utils

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/packethost/ironlib/model"
)

const ipmicfg = "/usr/sbin/smc-ipmicfg"

type Ipmicfg struct {
	Executor Executor
}

type IpmicfgSummary struct {
	FirmwareRevision string
	BIOSVersion      string
	CPLDVersion      string
}

// Return a new Supermicro IPMICFG executor
// note: the binary is expected to be available as smc-ipmicfg,
//       as setup in the fup firmware-update image
func NewIpmicfgCmd(trace bool) Collector {

	e := NewExecutor(ipmicfg)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	if !trace {
		e.SetQuiet()
	}

	return &Ipmicfg{Executor: e}
}

func (i *Ipmicfg) Components() ([]*model.Component, error) {

	summary, err := i.Summary()
	if err != nil {
		return nil, err
	}

	// add CPLD and BIOS firmware inventory
	inv := []*model.Component{
		{
			Model:             "Supermicro",
			Vendor:            "Supermicro",
			Name:              "CPLD",
			Slug:              "CPLD",
			FirmwareInstalled: summary.CPLDVersion,
		},
		{
			Model:             "Supermicro",
			Vendor:            "Supermicro",
			Name:              "BIOS",
			Slug:              "BIOS",
			FirmwareInstalled: summary.BIOSVersion,
		},
	}

	return inv, nil
}

func (i *Ipmicfg) Summary() (*IpmicfgSummary, error) {

	// smc-ipmicfg --summary
	i.Executor.SetArgs([]string{"-summary"})
	result, err := i.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, fmt.Errorf("no output from command: %s", i.Executor.GetCmd())
	}

	return i.parseQueryOutput(result.Stdout), nil
}

func (i *Ipmicfg) parseQueryOutput(b []byte) *IpmicfgSummary {

	summary := &IpmicfgSummary{}

	byteSlice := bytes.Split(b, []byte("\n"))
	for _, line := range byteSlice {

		s := string(line)

		if strings.Contains(s, "Firmware Revision") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				summary.FirmwareRevision = strings.TrimSpace(t[1])
			}
			continue
		}

		if strings.Contains(s, "BIOS Version") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				summary.BIOSVersion = strings.TrimSpace(t[1])
			}
			continue
		}

		if strings.Contains(s, "CPLD Version") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				summary.CPLDVersion = strings.TrimSpace(t[1])
			}
			continue
		}
	}

	return summary
}
