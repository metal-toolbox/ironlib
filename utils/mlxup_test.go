package utils

import (
	"context"
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeMlxup() *Mlxup {
	return &Mlxup{
		Executor: NewFakeExecutor("mlxup"),
	}
}

func Test_MlxupComponents(t *testing.T) {

	expected := []*model.Component{
		{Serial: "MT_2420110034", Vendor: "Mellanox", Model: "ConnectX4LX", Name: "ConnectX-4 Lx EN network interface card; 25GbE dual-port SFP28; PCIe3.0 x8; ROHS R6", Slug: "[0] NIC", FirmwareManaged: true, FirmwareInstalled: "14.27.1016", FirmwareAvailable: "14.28.2006", Metadata: map[string]string{"firmware_pxe_installed": "3.5.0901", "firmware_pxe_available": "3.6.0102", "firmware_uefi_installed": "14.20.0019", "firmware_uefi_available": "14.21.0017"}, Oem: false},
	}

	m := newFakeMlxup()
	inventory, err := m.Components()
	if err != nil {
		t.Error(err)
	}

	// since the component IDs are unique
	inventory = purgeTestComponentID(inventory)
	assert.Equal(t, expected, inventory)
}

func Test_MlxupParseQueryOutput(t *testing.T) {

	expected := []*MlxupDevices{
		{DeviceType: "ConnectX4LX", PartNumber: "MCX4121A-ACA_Ax", Description: "ConnectX-4 Lx EN network interface card; 25GbE dual-port SFP28; PCIe3.0 x8; ROHS R6", PSID: "MT_2420110034", BaseMAC: "b8599fde86f8", Firmware: []string{"14.27.1016", "14.28.2006"}, FirmwarePXE: []string{"3.5.0901", "3.6.0102"}, FirmwareUEFI: []string{"14.20.0019", "14.21.0017"}, Status: "Update required"},
	}

	m := newFakeMlxup()
	m.Executor.SetArgs([]string{"--query"})
	result, err := m.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Error(err)
	}

	mlxDevices := m.parseQueryOutput(result.Stdout)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, mlxDevices)
}
