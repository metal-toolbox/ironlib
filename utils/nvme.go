package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/sirupsen/logrus"
)

const EnvNvmeUtility = "IRONLIB_UTIL_NVME"

var (
	errSanicapNODMMASReserved = errors.New("sanicap nodmmas reserved bits set, not sure what to do with them")
	errSanitizeInvalidAction  = errors.New("invalid sanitize action")
)

type Nvme struct {
	Executor Executor
}

type nvmeDeviceAttributes struct {
	Namespace    int    `json:"Namespace"`
	DevicePath   string `json:"DevicePath"`
	Firmware     string `json:"Firmware"`
	Index        int    `json:"Index"`
	ModelNumber  string `json:"ModelNumber"`
	ProductName  string `json:"ProductName"`
	SerialNumber string `json:"SerialNumber"`
}

type nvmeList struct {
	Devices []*nvmeDeviceAttributes `json:"Devices"`
}

// Return a new nvme executor
func NewNvmeCmd(trace bool) *Nvme {
	utility := "nvme"

	// lookup env var for util
	if eVar := os.Getenv(EnvNvmeUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Nvme{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (n *Nvme) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := n.Executor.CheckExecutable()

	return "nvme", n.Executor.CmdPath(), er
}

// Executes nvme list, parses the output and returns a slice of *common.Drive
func (n *Nvme) Drives(ctx context.Context) ([]*common.Drive, error) {
	drives := make([]*common.Drive, 0)

	out, err := n.list(ctx)
	if err != nil {
		return nil, err
	}

	list := &nvmeList{Devices: []*nvmeDeviceAttributes{}}

	err = json.Unmarshal(out, list)
	if err != nil {
		return nil, err
	}

	for _, d := range list.Devices {
		dModel := d.ModelNumber

		var vendor string

		modelTokens := strings.Split(d.ModelNumber, " ")

		if len(modelTokens) > 1 {
			vendor = modelTokens[1]
		}
		drive := &common.Drive{
			Common: common.Common{
				Serial:      d.SerialNumber,
				Vendor:      vendor,
				Model:       dModel,
				ProductName: d.ProductName,
				Description: d.ModelNumber,
				Firmware: &common.Firmware{
					Installed: d.Firmware,
				},
				Metadata: map[string]string{},
			},
		}

		// Collect drive capabilitiesFound
		capabilitiesFound, err := n.DriveCapabilities(ctx, d.DevicePath)
		if err != nil {
			return nil, err
		}

		for _, f := range capabilitiesFound {
			drive.Common.Metadata[f.Description] = strconv.FormatBool(f.Enabled)
		}

		drives = append(drives, drive)
	}

	return drives, nil
}

func (n *Nvme) list(ctx context.Context) ([]byte, error) {
	// nvme list --output-format=json
	n.Executor.SetArgs("list", "--output-format=json")

	result, err := n.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

func (n *Nvme) cmdListCapabilities(ctx context.Context, device string) ([]byte, error) {
	// nvme id-ctrl --output-format=json devicepath
	n.Executor.SetArgs("id-ctrl", "--output-format=json", device)
	result, err := n.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

// DriveCapabilities returns the drive capability attributes obtained through nvme
//
// The device is the kernel/OS assigned drive name - /dev/nvmeX
//
// This method implements the actions.DriveCapabilityCollector interface.
func (n *Nvme) DriveCapabilities(ctx context.Context, device string) ([]*common.Capability, error) {
	out, err := n.cmdListCapabilities(ctx, device)
	if err != nil {
		return nil, err
	}

	var caps struct {
		FNA     uint `json:"fna"`
		SANICAP uint `json:"sanicap"`
	}

	err = json.Unmarshal(out, &caps)
	if err != nil {
		return nil, err
	}

	var capabilitiesFound []*common.Capability
	capabilitiesFound = append(capabilitiesFound, parseFna(caps.FNA)...)

	var parsedCaps []*common.Capability
	parsedCaps, err = parseSanicap(caps.SANICAP)
	if err != nil {
		return nil, err
	}
	capabilitiesFound = append(capabilitiesFound, parsedCaps...)

	return capabilitiesFound, nil
}

func parseFna(fna uint) []*common.Capability {
	// Bit masks values came from nvme-cli repo
	// All names come from internal nvme-cli names
	// We will *not* keep in sync as these names form our API
	// https: // github.com/linux-nvme/nvme-cli/blob/v2.8/nvme-print-stdout.c#L2199-L2217
	//
	// This function uses go's binary notation + bitshifts instead of masking of hex numbers
	// I think its easier to see whats going on this way vs hex masks
	// Refresher:
	//   1<<N makes gives us all zeros except for the Nth bit which is a one
	//   fna & 1<<N is bitwise and, the result will be 1 if fna had a 1 in Nth bi

	return []*common.Capability{
		{
			Name:        "fmns",
			Description: "Format Applies to All/Single Namespace(s) (t:All, f:Single)",
			Enabled:     fna&(0b1<<0) != 0,
		},
		{
			Name:        "cens",
			Description: "Crypto Erase Applies to All/Single Namespace(s) (t:All, f:Single)",
			Enabled:     fna&(0b1<<1) != 0,
		},
		{
			Name:        "cese",
			Description: "Crypto Erase Supported as part of Secure Erase",
			Enabled:     fna&(0b1<<2) != 0,
		},
	}
}

func parseSanicap(sanicap uint) ([]*common.Capability, error) {
	// Bit masks values came from nvme-cli repo
	// All names come from internal nvme-cli names
	// We will *not* keep in sync as these names form our API
	// https://github.com/linux-nvme/nvme-cli/blob/v2.8/nvme-print-stdout.c#L2064-L2093
	//
	// This function uses go's binary notation + bitshifts instead of masking of hex numbers
	// I think its easier to see whats going on this way vs hex masks
	// Refresher:
	//   1<<N makes gives us all zeros except for the Nth bit which is a one
	//   sanicap & 1<<N is bitwise and, the result will be 1 if sanicap had a 1 in Nth bit

	caps := []*common.Capability{
		{
			Name:        "cer",
			Description: "Crypto Erase Sanitize Operation Supported",
			Enabled:     sanicap&(0b1<<0) != 0,
		},
		{
			Name:        "ber",
			Description: "Block Erase Sanitize Operation Supported",
			Enabled:     sanicap&(0b1<<1) != 0,
		},
		{
			Name:        "owr",
			Description: "Overwrite Sanitize Operation Supported",
			Enabled:     sanicap&(0b1<<2) != 0,
		},
		{
			Name:        "ndi",
			Description: "No-Deallocate After Sanitize bit in Sanitize command Supported",
			Enabled:     sanicap&(0b1<<29) != 0,
		},
	}

	switch sanicap & (0b11 << 30) >> 30 {
	case 0b00:
		// nvme prints this out for 0b00:
		//   "Additional media modification after sanitize operation completes successfully is not defined"
		// So I'm taking "not defined" literally since we can't really represent 2 bits in a bool
		// If we ever want this as a bool we could maybe call it "dmmas" maybe?
	case 0b01, 0b10:
		caps = append(caps, &common.Capability{
			Name:        "nodmmas",
			Description: "Media is additionally modified after sanitize operation completes successfully",
			Enabled:     sanicap&(0b11<<30)>>30 == 0b10,
		})
	case 0b11:
		return nil, errSanicapNODMMASReserved
	}

	return caps, nil
}

//go:generate stringer -type SanitizeAction
type SanitizeAction uint8

const (
	Invalid SanitizeAction = iota
	ExitFailureMode
	BlockErase
	Overwrite
	CryptoErase
)

// WipeDisk implements DiskWiper by running nvme sanitize
func (n *Nvme) WipeDisk(ctx context.Context, logger *logrus.Logger, device string) error {
	caps, err := n.DriveCapabilities(ctx, device)
	if err != nil {
		return fmt.Errorf("WipeDisk: %w", err)
	}
	return n.wipe(ctx, logger, device, caps)
}

func (n *Nvme) wipe(ctx context.Context, logger *logrus.Logger, device string, caps []*common.Capability) error {
	var ber bool
	var cer bool
	for _, cap := range caps {
		switch cap.Name {
		case "ber":
			ber = cap.Enabled
		case "cer":
			cer = cap.Enabled
		}
	}

	if cer {
		l := logger.WithField("method", "sanitize").WithField("action", CryptoErase)
		l.Info("trying wipe")
		err := n.Sanitize(ctx, device, CryptoErase)
		if err == nil {
			return nil
		}
		l.WithError(err).Info("failed")
	}
	if ber {
		l := logger.WithField("method", "sanitize").WithField("action", BlockErase)
		l.Info("trying wipe")
		err := n.Sanitize(ctx, device, BlockErase)
		if err == nil {
			return nil
		}
		l.WithError(err).Info("failed")
	}
	return ErrIneffectiveWipe
}

func (n *Nvme) Sanitize(ctx context.Context, device string, sanact SanitizeAction) error {
	switch sanact { // nolint:exhaustive
	case BlockErase, CryptoErase:
	default:
		return fmt.Errorf("%w: %v", errSanitizeInvalidAction, sanact)
	}

	verify, err := ApplyWatermarks(device)
	if err != nil {
		return err
	}

	n.Executor.SetArgs("sanitize", "--sanact="+strconv.Itoa(int(sanact)), device)
	_, err = n.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	// now we loop until sanitize-log reports that sanitization is complete
	dev := path.Base(device)
	var log map[string]struct {
		Progress uint16 `json:"sprog"`
	}
	for {
		n.Executor.SetArgs("sanitize-log", "--output-format=json", device)
		result, err := n.Executor.Exec(ctx)
		if err != nil {
			return err
		}
		err = json.Unmarshal(result.Stdout, &log)
		if err != nil {
			return err
		}

		l, ok := log[dev]
		if !ok {
			return fmt.Errorf("%s: device not present in sanitize-log: %w: %s", dev, io.ErrUnexpectedEOF, result.Stdout)
		}

		if l.Progress == 65535 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return verify()
}

// NewFakeNvme returns a mock nvme collector that returns mock data for use in tests.
func NewFakeNvme() *Nvme {
	return &Nvme{
		Executor: NewFakeExecutor("nvme"),
	}
}
