package utils

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/model"
)

const mlxup = "/usr/sbin/mlxup"

type Mlxup struct {
	Executor Executor
}

type MlxupDevices struct {
	PartNumber    string
	DeviceType    string
	Description   string
	PCIDeviceName string
	PSID          string
	BaseMAC       string
	Firmware      []string // [version_current, version_available]
	FirmwarePXE   []string
	FirmwareUEFI  []string
	Status        string
}

// Return a new mellanox mlxup command executor
func NewMlxupCmd(trace bool) Collector {

	e := NewExecutor(mlxup)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	if !trace {
		e.SetQuiet()
	}

	return &Mlxup{Executor: e}
}

func NewMlxupUpdater(trace bool) Updater {
	e := NewExecutor(mlxup)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	if !trace {
		e.SetQuiet()
	}

	return &Mlxup{Executor: e}
}

// Get a list of components
func (m *Mlxup) Components() ([]*model.Component, error) {

	devices, err := m.Query()
	if err != nil {
		return nil, err
	}

	inv := []*model.Component{}
	for _, d := range devices {

		uid, _ := uuid.NewRandom()
		item := &model.Component{
			ID:              uid.String(),
			Model:           d.DeviceType,
			Vendor:          vendorFromString(d.DeviceType),
			Slug:            model.SlugNIC,
			Name:            d.Description,
			Serial:          d.BaseMAC,
			FirmwareManaged: true,
			Metadata:        make(map[string]string),
		}

		// [vInstalled, vAvailable]
		if len(d.Firmware) > 0 {
			item.FirmwareInstalled = d.Firmware[0]
			if len(d.Firmware) == 2 {
				item.FirmwareAvailable = d.Firmware[1]
			}
		} else {
			item.FirmwareInstalled = "unknown"
		}

		// [vInstalled, vAvailable]
		if len(d.FirmwarePXE) > 0 {
			item.Metadata["firmware_pxe_installed"] = d.FirmwarePXE[0]
			if len(d.FirmwarePXE) == 2 {
				item.Metadata["firmware_pxe_available"] = d.FirmwarePXE[1]
			}
		}

		// [vInstalled, vAvailable]
		if len(d.FirmwareUEFI) > 0 {
			item.Metadata["firmware_uefi_installed"] = d.FirmwareUEFI[0]
			if len(d.FirmwareUEFI) == 2 {
				item.Metadata["firmware_uefi_available"] = d.FirmwareUEFI[1]
			}
		}

		inv = append(inv, item)
	}
	return inv, nil
}

// Update mellanox component
func (m *Mlxup) ApplyUpdate(ctx context.Context, updateFile, componentSlug string) error {
	// query list of nics
	nics, err := m.Query()
	if err != nil {
		return err
	}

	// apply update
	for _, nic := range nics {
		m.Executor.SetArgs([]string{
			"--yes",
			"--dev",
			nic.PCIDeviceName,
			"--image-file",
			updateFile,
		})

		result, err := m.Executor.ExecWithContext(ctx)
		if err != nil {
			return err
		}

		if result.ExitCode != 0 {
			return fmt.Errorf("%s returned non-zero exit code: %d, stderr: %s", m.Executor.GetCmd(), result.ExitCode, result.Stderr)
		}
	}

	return nil
}

// Query returns a slice of mellanox devices
func (m *Mlxup) Query() ([]*MlxupDevices, error) {

	// mlxup --query
	m.Executor.SetArgs([]string{"--query"})

	result, err := m.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, fmt.Errorf("no output from command: %s", m.Executor.GetCmd())
	}

	return m.parseQueryOutput(result.Stdout), nil

}

// Parse mlxup --query output into MlxupDevices
// see tests for details
func (m *Mlxup) parseQueryOutput(b []byte) []*MlxupDevices {

	devices := []*MlxupDevices{}
	byteSlice := bytes.Split(b, []byte("\n"))
	for idx, sl := range byteSlice {
		s := string(sl)
		if strings.Contains(s, "Device #") {
			device := parseDeviceAttributes(byteSlice[idx:])
			if device != nil && len(device.Firmware) > 0 {
				devices = append(devices, device)
			}
		}
	}

	return devices
}

// nolint: gocyclo
func parseDeviceAttributes(byteSlice [][]byte) *MlxupDevices {

	device := &MlxupDevices{}

	for _, line := range byteSlice {

		s := string(line)

		// Parse device type
		if strings.Contains(s, "Device Type:") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				device.DeviceType = strings.TrimSpace(t[1])
			}
			continue
		}

		// Parse part number
		if strings.Contains(s, "Part Number:") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				device.PartNumber = strings.TrimSpace(t[1])
			}
			continue
		}

		// Parse serial
		if strings.Contains(s, "PSID:") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				device.PSID = strings.TrimSpace(t[1])
			}
			continue
		}

		// Parse PCI device name
		if strings.Contains(s, "PCI Device Name:") {
			t := strings.Split(s, "PCI Device Name:")
			if len(t) > 0 {
				device.PCIDeviceName = strings.TrimSpace(t[1])
			}
			continue
		}

		// Parse description
		if strings.Contains(s, "Description:") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				device.Description = strings.TrimSpace(t[1])
			}
			continue
		}

		// Parse MAC
		if strings.Contains(s, "Base MAC:") {
			t := strings.Split(s, ":")
			if len(t) > 0 {
				device.BaseMAC = strings.TrimSpace(t[1])
			}
			continue
		}

		// Parse current, available firmware versions
		if strings.Contains(s, " FW ") {
			fields := strings.Fields(s)
			if len(fields) == 0 {
				continue
			}

			device.Firmware = fields[1:]
			continue
		}

		// Parse current, available PXE versions
		if strings.Contains(s, " PXE ") {
			fields := strings.Fields(s)
			if len(fields) == 0 {
				continue
			}

			device.FirmwarePXE = fields[1:]
			continue
		}

		// Parse current, available UEFI versions
		if strings.Contains(s, " UEFI ") {
			fields := strings.Fields(s)
			if len(fields) == 0 {
				continue
			}

			device.FirmwareUEFI = fields[1:]
			continue
		}

		// Parse status line
		if strings.Contains(s, " Status:") {
			status := strings.Split(s, ":")
			if len(status) > 0 {
				device.Status = strings.TrimSpace(status[1])
			}
			break
		}
	}

	return device
}
