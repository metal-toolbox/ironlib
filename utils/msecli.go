package utils

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

var ErrMseCliDriveNotIdentified = errors.New("failed to identify drive for update")

const (
	EnvMsecliUtility = "IRONLIB_UTIL_MSECLI"
)

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

// NewMsecli returns a Msecli object to run msecli commands
func NewMsecli(trace bool) *Msecli {
	utility := "msecli"

	// lookup env var for util
	if eVar := os.Getenv(EnvMsecliUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Msecli{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (m *Msecli) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := m.Executor.CheckExecutable()

	return "msecli", m.Executor.CmdPath(), er
}

// Drives returns a slice of drive components identified
func (m *Msecli) Drives(ctx context.Context) ([]*common.Drive, error) {
	devices, err := m.Query(ctx)
	if err != nil {
		return nil, err
	}

	drives := []*common.Drive{}

	for _, d := range devices {
		item := &common.Drive{
			Common: common.Common{
				Model:       d.ModelNumber,
				Vendor:      common.VendorFromString(d.ModelNumber),
				Description: d.ModelNumber,
				Serial:      d.SerialNumber,
				Firmware:    &common.Firmware{Installed: d.FirmwareRevision},
				Metadata:    make(map[string]string),
			},

			Type: model.DriveTypeSlug(d.ModelNumber),
		}

		drives = append(drives, item)
	}

	return drives, nil
}

// UpdateDrive installs drive updates
func (m *Msecli) UpdateDrive(ctx context.Context, updateFile, modelNumber, serialNumber string) error {
	// query list of drives
	drives, err := m.Query(ctx)
	if err != nil {
		return err
	}

	// msecli expects the update file to be named 1.bin - don't ask
	expectedFileName := "1.bin"

	// rename update file
	if filepath.Base(updateFile) != expectedFileName {
		newName := filepath.Join(filepath.Dir(updateFile), expectedFileName)

		err := os.Rename(updateFile, newName)
		if err != nil {
			return err
		}

		updateFile = newName
	}

	for _, d := range drives {
		// filter by serial number
		if serialNumber != "" {
			if !strings.EqualFold(d.SerialNumber, serialNumber) {
				continue
			}
		}

		// filter by model number
		if modelNumber != "" {
			if !strings.EqualFold(d.ModelNumber, modelNumber) {
				continue
			}
		}

		return m.updateDrive(ctx, d.ModelNumber, updateFile)
	}

	return ErrMseCliDriveNotIdentified
}

// updateDrive installs the given updatefile
func (m *Msecli) updateDrive(ctx context.Context, modelNumber, updateFile string) error {
	// get the product name from the model number - msecli expects the product name
	modelNForMsecli, err := mseCLIModelType(modelNumber)
	if err != nil {
		return err
	}

	// echo 'y'
	m.Executor.SetStdin(bytes.NewReader([]byte("y\n")))
	m.Executor.SetArgs(
		"-U", // update
		"-m", // model
		modelNForMsecli,
		"-i", // directory containing the update file
		filepath.Dir(updateFile),
	)

	result, err := m.Executor.Exec(ctx)
	if err != nil {
		return newExecError(m.Executor.GetCmd(), result)
	}

	if result.ExitCode != 0 {
		return newExecError(m.Executor.GetCmd(), result)
	}

	return nil
}

// Query parses the output of mseli -L and returns a slice of *MsecliDevice's
func (m *Msecli) Query(ctx context.Context) ([]*MsecliDevice, error) {
	m.Executor.SetArgs("-L")

	result, err := m.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrNoCommandOutput, m.Executor.GetCmd())
	}

	return m.parseMsecliQueryOutput(result.Stdout), nil
}

// Parse msecli -L output into []*MsecliDevice
func (m *Msecli) parseMsecliQueryOutput(b []byte) []*MsecliDevice {
	devices := []*MsecliDevice{}

	// split
	byteSlice := bytes.Split(b, []byte("\n\n"))
	for _, sl := range byteSlice {
		if !bytes.Contains(sl, []byte("Device Name")) {
			continue
		}

		devices = append(devices, parseMsecliDeviceAttributes(sl))
	}

	return devices
}

// parse a Device information section into *MsecliDevice
func parseMsecliDeviceAttributes(bSlice []byte) *MsecliDevice {
	device := &MsecliDevice{}

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
		case "Model No":
			device.ModelNumber = value
		case "FW-Rev":
			device.FirmwareRevision = value
		case "Serial No":
			device.SerialNumber = value
		}
	}

	return device
}

// msecli expects a model type which we derive from the model number
//
//	Invalid model type MICRON_5200_MTFDDAK480TDN! Valid options are:
//	M500, M510, M550, MX100, M600, M500DC, MX200, BX100, P400M, P400E, M510DC,
//	M500IT, BX200, BX300, S610DC, S630DC, S650DC, S655DC, 9100PRO, 9100MAX, 2100,
//	TX3, MX300, 1100, M500ITL, 7100ECO, 7100MAX, 5100ECO, 5100PRO, 5100MAX, 5300PRO,
//	5300MAX, 5300BOOT, 5200ECO, 5200PRO, 5200MAX, 9200ECO, 9200MAX, 9200PRO, 9300ECO, 9300MAX,
//	9300PRO, 7300ECO, 7300MAX, 7300PRO, MX500, 5210, 1300, MX600, BX500, P1,
//	2200, 2200S, P4, 2300, 2300V, 2210, P1W2, 2100IT, 3400, 2450,
//	BX503, BX504, P7
func mseCLIModelType(mseModel string) (string, error) {
	errUnsupported := errors.New("unsupported model number: " + mseModel)
	switch mseModel {
	case "Micron_5300_MTFDDAK480TDT":
		return "5300MAX", nil
	case "Micron_5200_MTFDDAK480TDN":
		return "5200MAX", nil
	default:
		return "", errors.Wrap(errUnsupported, mseModel)
	}
}
