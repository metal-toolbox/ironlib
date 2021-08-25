package utils

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeMlxup() *Mlxup {
	return &Mlxup{
		Executor: NewFakeExecutor("mlxup"),
	}
}

func Test_MlxupNICs(t *testing.T) {
	expected := []*model.NIC{
		{
			Description: "ConnectX-4 Lx EN network interface card; 25GbE dual-port SFP28; PCIe3.0 x8; ROHS R6",
			Vendor:      "Mellanox",
			Model:       "MCX4121A-ACA_Ax",
			Serial:      "b8:59:9f:de:86:fd",
			SpeedBits:   0,
			PhysicalID:  "",
			Oem:         false,
			Metadata:    map[string]string{},
			Firmware: &model.Firmware{
				Available: "14.28.2007",
				Installed: "14.27.1016",
				Metadata: map[string]string{
					"firmware_pxe_available":  "3.6.0102",
					"firmware_pxe_installed":  "3.5.0901",
					"firmware_uefi_available": "14.21.0017",
					"firmware_uefi_installed": "14.20.0019",
				},
			},
		},
		{
			Description: "ConnectX-4 Lx EN network interface card; 25GbE dual-port SFP28; PCIe3.0 x8; ROHS R6",
			Vendor:      "Mellanox",
			Model:       "MCX4121A-ACA_Ax",
			Serial:      "b8:59:9f:de:86:f8",
			SpeedBits:   0,
			PhysicalID:  "",
			Oem:         false,
			Metadata:    map[string]string{},
			Firmware: &model.Firmware{
				Available: "14.28.2006",
				Installed: "14.27.1016",
				Metadata: map[string]string{
					"firmware_pxe_available":  "3.6.0102",
					"firmware_pxe_installed":  "3.5.0901",
					"firmware_uefi_available": "14.21.0017",
					"firmware_uefi_installed": "14.20.0019",
				},
			},
		},
	}

	cli := newFakeMlxup()

	b, err := ioutil.ReadFile("../fixtures/utils/mlxup/query")
	if err != nil {
		t.Error(err)
	}

	cli.Executor.SetStdout(b)

	inventory, err := cli.NICs(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, inventory)
}

func Test_MlxupParseQueryOutput(t *testing.T) {
	expected := []*MlxupDevice{
		{DeviceType: "ConnectX4LX", PartNumber: "MCX4121A-ACA_Ax", Description: "ConnectX-4 Lx EN network interface card; 25GbE dual-port SFP28; PCIe3.0 x8; ROHS R6", PSID: "MT_2420110034", BaseMAC: "b8:59:9f:de:86:fd", PCIDeviceName: "0000:d8:00.0", Firmware: []string{"14.27.1016", "14.28.2007"}, FirmwarePXE: []string{"3.5.0901", "3.6.0102"}, FirmwareUEFI: []string{"14.20.0019", "14.21.0017"}, Status: "Update required"},
		{DeviceType: "ConnectX4LX", PartNumber: "MCX4121A-ACA_Ax", Description: "ConnectX-4 Lx EN network interface card; 25GbE dual-port SFP28; PCIe3.0 x8; ROHS R6", PSID: "MT_2420110034", BaseMAC: "b8:59:9f:de:86:f8", PCIDeviceName: "0000:d8:00.1", Firmware: []string{"14.27.1016", "14.28.2006"}, FirmwarePXE: []string{"3.5.0901", "3.6.0102"}, FirmwareUEFI: []string{"14.20.0019", "14.21.0017"}, Status: "Update required"},
	}

	cli := newFakeMlxup()

	b, err := ioutil.ReadFile("../fixtures/utils/mlxup/query")
	if err != nil {
		t.Error(err)
	}

	cli.Executor.SetStdout(b)

	result, err := cli.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Error(err)
	}

	mlxDevices := cli.parseMlxQueryOutput(result.Stdout)

	assert.Equal(t, expected, mlxDevices)
}

func Test_FormatHWAddress(t *testing.T) {
	assert.Equal(t, "b8:59:9f:de:86:fd", formatBaseMacAddress("b8599fde86fd"))
	assert.Equal(t, "b8:59:9f:de:86:fd", formatBaseMacAddress("b8:59:9f:de:86:fd"))
	assert.Equal(t, "", formatBaseMacAddress("foo"))
}
