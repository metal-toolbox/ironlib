package utils

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
)

const EnvNvmeUtility = "IRONLIB_UTIL_NVME"

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
	n.Executor.SetArgs([]string{"list", "--output-format=json"})

	result, err := n.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

func (n *Nvme) cmdListCapabilities(ctx context.Context, logicalPath string) ([]byte, error) {
	// nvme id-ctrl --output-format=json devicepath
	n.Executor.SetArgs([]string{"id-ctrl", "--output-format=json", logicalPath})
	result, err := n.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

// DriveCapabilities returns the drive capability attributes obtained through nvme
//
// The logicalName is the kernel/OS assigned drive name - /dev/nvmeX
//
// This method implements the actions.DriveCapabilityCollector interface.
func (n *Nvme) DriveCapabilities(ctx context.Context, logicalName string) ([]*common.Capability, error) {
	out, err := n.cmdListCapabilities(ctx, logicalName)
	if err != nil {
		return nil, err
	}

	var caps struct {
		FNA     uint `json:"fna"`
		SANICAP uint `json:"sanicap"`
	}

	if err := json.Unmarshal(out, &caps); err != nil {
		return nil, err
	}

	var capabilitiesFound []*common.Capability
	capabilitiesFound = append(capabilitiesFound, parseSanicap(caps.SANICAP)...)
	capabilitiesFound = append(capabilitiesFound, parseFna(caps.FNA)...)

	return capabilitiesFound, nil
}

func parseFna(fna uint) []*common.Capability {
	// Bit masks values came from nvme-cli repo:
	// https://github.com/linux-nvme/nvme-cli/blob/v2.8/nvme-print-stdout.c#L2199-L2217

	caps := make([]*common.Capability, 0, 4)

	caps = append(caps, &common.Capability{
		Name:        "fna",
		Description: "Crypto Erase Support",
		Enabled:     fna != 0,
	})

	if (fna&(0b1<<2))>>2 == 0 {
		caps = append(caps, &common.Capability{
			Name:        "censapose",
			Description: "Crypto Erase Not Supported as part of Secure Erase",
			Enabled:     false,
		})
	} else {
		caps = append(caps, &common.Capability{
			Name:        "cesapose",
			Description: "Crypto Erase Supported as part of Secure Erase",
			Enabled:     true,
		})
	}

	// nvme: cens
	if (fna&(0b1<<1))>>1 == 0 {
		caps = append(caps, &common.Capability{
			Name:        "ceatsn",
			Description: "Crypto Erase Applies to Single Namespace(s)",
			Enabled:     false,
		})
	} else {
		caps = append(caps, &common.Capability{
			Name:        "ceatan",
			Description: "Crypto Erase Applies to All Namespace(s)",
			Enabled:     true,
		})
	}

	// nvme: fmns
	if (fna&(0b1<<0))>>0 == 0 {
		caps = append(caps, &common.Capability{
			Name:        "fatsn",
			Description: "Format Applies to Single Namespace(s)",
			Enabled:     false,
		})
	} else {
		caps = append(caps, &common.Capability{
			Name:        "fatan",
			Description: "Format Applies to All Namespace(s)",
			Enabled:     true,
		})
	}

	return caps
}

func parseSanicap(sanicap uint) []*common.Capability {
	// Bit masks values came from nvme-cli repo:
	// https://github.com/linux-nvme/nvme-cli/blob/v2.8/nvme-print-stdout.c#L2064-L2093

	caps := make([]*common.Capability, 0, 6)

	caps = append(caps, &common.Capability{
		Name:        "sanicap",
		Description: "Sanitize Support",
		Enabled:     sanicap != 0,
	})

	// nvme: nodmmas
	switch (sanicap & (0b11 << 30)) >> 30 {
	case 0b00:
		caps = append(caps, &common.Capability{
			Name:        "ammasocsind",
			Description: "Additional media modification after sanitize operation completes successfully is not defined",
			Enabled:     false,
		})
	case 0b01:
		caps = append(caps, &common.Capability{
			Name:        "minamasocs",
			Description: "Media is not additionally modified after sanitize operation completes successfully",
			Enabled:     true,
		})
	case 0b10:
		caps = append(caps, &common.Capability{
			Name:        "miamasocs",
			Description: "Media is additionally modified after sanitize operation completes successfully",
			Enabled:     true,
		})
	case 0b11:
		caps = append(caps, &common.Capability{
			Name:        "",
			Description: "Reserved",
			Enabled:     true,
		})
	default:
		panic("unreachable")
	}

	// nvme: ndi
	if (sanicap&(0b1<<29))>>29 == 0 {
		caps = append(caps, &common.Capability{
			Name:        "nasbiscs",
			Description: "No-Deallocate After Sanitize bit in Sanitize command Supported",
			Enabled:     false,
		})
	} else {
		caps = append(caps, &common.Capability{
			Name:        "nasbiscns",
			Description: "No-Deallocate After Sanitize bit in Sanitize command Not Supported",
			Enabled:     true,
		})
	}

	// nvme: owr
	if (sanicap&(0b1<<2))>>2 == 0 {
		caps = append(caps, &common.Capability{
			Name:        "osons",
			Description: "Overwrite Sanitize Operation Not Supported",
			Enabled:     false,
		})
	} else {
		caps = append(caps, &common.Capability{
			Name:        "osos",
			Description: "Overwrite Sanitize Operation Supported",
			Enabled:     true,
		})
	}

	// nvme: ber
	if (sanicap&(0b1<<1))>>1 == 0 {
		caps = append(caps, &common.Capability{
			Name:        "besons",
			Description: "Block Erase Sanitize Operation Not Supported",
			Enabled:     false,
		})
	} else {
		caps = append(caps, &common.Capability{
			Name:        "besos",
			Description: "Block Erase Sanitize Operation Supported",
			Enabled:     true,
		})
	}

	// nvme: cer
	if (sanicap&(0b1<<0))>>0 == 0 {
		caps = append(caps, &common.Capability{
			Name:        "cesons",
			Description: "Crypto Erase Sanitize Operation Not Supported",
			Enabled:     false,
		})
	} else {
		caps = append(caps, &common.Capability{
			Name:        "cesos",
			Description: "Crypto Erase Sanitize Operation Supported",
			Enabled:     true,
		})
	}

	return caps
}

// NewFakeNvme returns a mock nvme collector that returns mock data for use in tests.
func NewFakeNvme() *Nvme {
	return &Nvme{
		Executor: NewFakeExecutor("nvme"),
	}
}
