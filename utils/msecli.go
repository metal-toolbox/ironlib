package utils

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/model"
	"github.com/pkg/errors"
)

//
const msecli = "/usr/bin/msecli"

// Msecli is an msecli executor
type Msecli struct {
	Executor Executor
}

// MseclieDevice is a Micron disk device object
type MsecliDevice struct {
	ModelNumber      string // Micron_5200_MTFDDAK480TDN
	SerialNumber     string
	FirmwareRevision string
}

// NewMseclicmd returns a new cli command runner
func NewMsecliCmd(trace bool) Collector {

	e := NewExecutor(msecli)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	if !trace {
		e.SetQuiet()
	}

	return &Mlxup{Executor: e}
}

// Components returns a slice of drive components identified
func (m *Msecli) Components() ([]*model.Component, error) {

	devices, err := m.Query()
	if err != nil {
		return nil, err
	}

	inv := []*model.Component{}

	for _, d := range devices {
		uid, _ := uuid.NewRandom()
		item := &model.Component{
			ID:                uid.String(),
			Model:             d.ModelNumber,
			Vendor:            vendorFromString(d.ModelNumber),
			Slug:              model.SlugDiskSataSsd,
			Name:              model.SlugDiskSataSsd,
			Serial:            d.SerialNumber,
			FirmwareInstalled: d.FirmwareRevision,
			FirmwareManaged:   true,
			Metadata:          make(map[string]string),
		}
		inv = append(inv, item)
	}

	return inv, nil
}

// ApplyUpdate installs the updateFile
func (m *Msecli) ApplyUpdate(ctx context.Context, updateFile, componetSlug string) error {

	// query list of drives
	drives, err := m.Query()
	if err != nil {
		return err
	}

	for _, d := range drives {
		// echo 'y'
		m.Executor.SetStdin(bytes.NewReader([]byte("y\n")))
		m.Executor.SetArgs([]string{
			"-U", // update
			"-m", // model
			// get the product name from the model number
			FormatProductName(d.ModelNumber),
			"-i", // directory containing the update file
			filepath.Dir(updateFile),
		},
		)

		result, err := m.Executor.ExecWithContext(ctx)
		if err != nil {
			return newUtilsExecError(m.Executor.GetCmd(), result)
		}

		if result.ExitCode != 0 {
			return newUtilsExecError(m.Executor.GetCmd(), result)
		}
	}

	return nil
}

// Query parses the output of mseli -L and returns a slice of *MsecliDevice's
func (m *Msecli) Query() ([]*MsecliDevice, error) {
	m.Executor.SetArgs([]string{"-L"})

	result, err := m.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrUtilsNoCommandOutput, m.Executor.GetCmd())
	}

	return m.parseMsecliQueryOutput(result.Stdout), nil
}

// Parse msecli -L output into []*MsecliDevice
// see tests for details
func (m *Msecli) parseMsecliQueryOutput(b []byte) []*MsecliDevice {

	devices := []*MsecliDevice{}

	byteSlice := bytes.Split(b, []byte("\n"))
	for idx, sl := range byteSlice {
		s := string(sl)
		if strings.Contains(s, "Device Name") {
			device := parseMsecliDeviceAttributes(byteSlice[idx:])
			if device != nil && len(device.FirmwareRevision) > 0 {
				devices = append(devices, device)
			}
		}
	}

	return devices
}

// nolint: gocyclo
func parseMsecliDeviceAttributes(byteSlice [][]byte) *MsecliDevice {

	device := &MsecliDevice{}

	for _, line := range byteSlice {

		s := string(line)

		// Parse Model number
		if strings.Contains(s, "Model No") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				device.ModelNumber = strings.TrimSpace(t[1])
			}

			continue
		}

		if strings.Contains(s, "Serial No") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				device.SerialNumber = strings.TrimSpace(t[1])
			}

			continue
		}

		if strings.Contains(s, "FW-Rev") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				device.FirmwareRevision = strings.TrimSpace(t[1])
			}

			break
		}

	}
	return device
}
