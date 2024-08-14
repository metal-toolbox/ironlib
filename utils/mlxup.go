package utils

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

const (
	EnvMlxupUtility = "IRONLIB_UTIL_MLXUP"
)

// Mlxup is a mlxup command executor object
type Mlxup struct {
	Executor Executor
}

// MlxupDevice is a mellanox device object
type MlxupDevice struct {
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
func NewMlxupCmd(trace bool) *Mlxup {
	utility := "mlxup"

	// lookup env var for util
	if eVar := os.Getenv(EnvMlxupUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Mlxup{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (m *Mlxup) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := m.Executor.CheckExecutable()

	return "mlxup", m.Executor.CmdPath(), er
}

// NICs returns a slice of mellanox components as *common.NIC's
func (m *Mlxup) NICs(ctx context.Context) ([]*common.NIC, error) {
	devices, err := m.Query(ctx)
	if err != nil {
		return nil, err
	}

	nics := []*common.NIC{}

	// serials is a map of serials added to the nics slice
	serials := map[string]bool{}

	for _, d := range devices {
		// skip NICs without serials
		serial := strings.ToLower(d.BaseMAC)
		if serial == "" {
			log.Printf("Warn: NIC component without serial, ignored: %+v\n", d)
			continue
		}

		// skip NICs with duplicate serials
		if serials[serial] {
			continue
		}

		serials[serial] = true

		nic := &common.NIC{
			Common: common.Common{
				Model:       d.PartNumber,
				Vendor:      common.VendorFromString(d.DeviceType),
				Description: d.Description,
				Serial:      d.BaseMAC,
				Metadata:    make(map[string]string),
				Firmware:    common.NewFirmwareObj(),
			},
			NICPorts: []*common.NICPort{
				{
					BusInfo:    d.PCIDeviceName,
					MacAddress: d.BaseMAC,
				},
			},
		}

		// populate NIC firmware attributes
		setNICFirmware(d, nic.Firmware)

		nics = append(nics, nic)
	}

	return nics, nil
}

// setNICFirmware populates the NIC firmware object
func setNICFirmware(d *MlxupDevice, firmware *common.Firmware) {
	// [vInstalled, vAvailable]
	if len(d.Firmware) > 0 {
		firmware.Installed = d.Firmware[0]
		fwAvailableElem := 2

		if len(d.Firmware) == fwAvailableElem {
			firmware.Available = d.Firmware[1]
		}
	}

	// [vInstalled, vAvailable]
	if len(d.FirmwarePXE) > 0 {
		firmware.Metadata["firmware_pxe_installed"] = d.FirmwarePXE[0]
		fwAvailableElem := 2

		if len(d.FirmwarePXE) == fwAvailableElem {
			firmware.Metadata["firmware_pxe_available"] = d.FirmwarePXE[1]
		}
	}

	// [vInstalled, vAvailable]
	if len(d.FirmwareUEFI) > 0 {
		firmware.Metadata["firmware_uefi_installed"] = d.FirmwareUEFI[0]
		fwAvailableElem := 2

		if len(d.FirmwareUEFI) == fwAvailableElem {
			firmware.Metadata["firmware_uefi_available"] = d.FirmwareUEFI[1]
		}
	}
}

// UpdateRequirements implements the actions/NICUpdater interface to return any pre/post firmware install requirements.
func (m *Mlxup) UpdateRequirements(_ string) model.UpdateRequirements {
	return model.UpdateRequirements{PostInstallHostPowercycle: true}
}

// UpdateNIC updates mellanox NIC with the given update file
func (m *Mlxup) UpdateNIC(ctx context.Context, updateFile, modelNumber string, force bool) error {
	// query list of nics
	nics, err := m.Query(ctx)
	if err != nil {
		return err
	}

	// apply update
	for _, nic := range nics {
		if modelNumber != "" {
			if !strings.EqualFold(nic.PartNumber, modelNumber) {
				continue
			}
		}

		args := []string{"--yes", "--dev", nic.PCIDeviceName, "--image-file", updateFile}

		if force {
			args = append(args, "--force")
		}

		m.Executor.SetArgs(args...)
		result, err := m.Executor.Exec(ctx)
		if err != nil {
			if result != nil && result.ExitCode != 0 {
				resetRequiredStr := "The firmware image was already updated on flash, pending reset"
				if result.Stdout != nil && strings.Contains(string(result.Stdout), resetRequiredStr) {
					return errors.Wrap(ErrRebootRequired, resetRequiredStr)
				}
				return newExecError(m.Executor.GetCmd(), result)
			}
			return err
		}
	}

	return nil
}

// Query returns a slice of mellanox devices
func (m *Mlxup) Query(ctx context.Context) ([]*MlxupDevice, error) {
	// mlxup --query
	m.Executor.SetArgs("--query")

	result, err := m.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrNoCommandOutput, m.Executor.GetCmd())
	}

	return m.parseMlxQueryOutput(result.Stdout), nil
}

// parseMlxQueryOutput parses the mlxup --query output into a slice of []*MlxupDevice
func (m *Mlxup) parseMlxQueryOutput(b []byte) []*MlxupDevice {
	devices := []*MlxupDevice{}

	byteSlice := bytes.Split(b, []byte("\n"))
	for idx, sl := range byteSlice {
		s := string(sl)
		if strings.Contains(s, "Device #") {
			device := parseMlxDeviceAttributes(byteSlice[idx+1:])
			if device != nil && len(device.Firmware) > 0 {
				devices = append(devices, device)
			}
		}
	}

	return devices
}

// parseMlxDeviceAttributes a mlx device information section into *MlxupDevice
func parseMlxDeviceAttributes(bSlice [][]byte) *MlxupDevice {
	device := &MlxupDevice{}

	for sidx, line := range bSlice {
		// return if we're on a line that indicates a new device
		if bytes.Contains(line, []byte(`Device #`)) {
			return device
		}

		s := string(line)

		cols := 2
		parts := strings.Split(s, ": ")

		if len(parts) < cols {
			continue
		}

		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

		switch key {
		case "Device Type":
			device.DeviceType = value
		case "Part Number":
			device.PartNumber = value
		case "PSID":
			device.PSID = value
		case "PCI Device Name":
			device.PCIDeviceName = value
		case "Description":
			device.Description = value
		case "Base MAC":
			device.BaseMAC = formatBaseMacAddress(value)
		case "Status":
			device.Status = value
		case "Versions":
			versions := parseMlxVersions(bSlice[sidx:])
			device.Firmware = versions["FW"]
			device.FirmwarePXE = versions["PXE"]
			device.FirmwareUEFI = versions["UEFI"]

			continue
		}
	}

	return device
}

// parseMlxVersions parses the firmware version info
// and returns a map of versions
//
// version["FW"][0] indicates fw version installed
// version["FW"][1] indicates fw version available
func parseMlxVersions(bSlice [][]byte) map[string][]string {
	versions := map[string][]string{
		"FW":   make([]string, 0),
		"PXE":  make([]string, 0),
		"UEFI": make([]string, 0),
	}

	for _, s := range bSlice {
		// skip line with header
		// Versions:         Current        Available
		if bytes.Contains(s, []byte("Versions:")) {
			continue
		}

		switch {
		// line indicating the firmware installed,
		// ignore line in the form, "FW (Running)   14.27.1016     N/A)"
		// which indicates NIC has been updated but is running the older firmware.
		case bytes.Contains(s, []byte(" FW ")) && !bytes.Contains(s, []byte("Running")):
			fields := strings.Fields(string(s))
			if len(fields) == 0 {
				continue
			}

			// version installed, available
			versions["FW"] = append(versions["FW"], fields[1:]...)
		case bytes.Contains(s, []byte(" PXE ")):
			fields := strings.Fields(string(s))
			if len(fields) == 0 {
				continue
			}

			// version installed, available
			versions["PXE"] = append(versions["PXE"], fields[1:]...)
		case bytes.Contains(s, []byte(" UEFI ")):
			fields := strings.Fields(string(s))
			if len(fields) == 0 {
				continue
			}

			// version installed, available
			versions["UEFI"] = append(versions["UEFI"], fields[1:]...)
		}

		// break on line with status
		// Status:           Update required
		if bytes.Contains(s, []byte(`Status`)) {
			return versions
		}
	}

	return versions
}

// formatBaseMacAddress accepts a mac address string in the form "ac1f6bdc19c2"
// returns it in the delimited format "ac:1f:6b:dc:19:c2"
//
// returns an empty string if the hw address isn't 12 chars in length
func formatBaseMacAddress(str string) string {
	// length of a hwaddress without the separators
	minLen := 12
	if len(str) < minLen {
		return ""
	}

	if strings.ContainsAny(str, ":") {
		return str
	}

	n := 2

	n1 := n - 1
	l1 := len(str) - 1

	var buffer bytes.Buffer
	for i, r := range str {
		buffer.WriteRune(r)

		if i%n == n1 && i != l1 {
			buffer.WriteRune(':')
		}
	}

	return buffer.String()
}

func NewFakeMlxup(r io.Reader) (*Mlxup, error) {
	e := NewFakeExecutor("mlxup")
	b := bytes.Buffer{}

	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	e.SetStdout(b.Bytes())

	return &Mlxup{Executor: e}, nil
}
