package utils

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/assert"
)

func Test_NvmeComponents(t *testing.T) {
	expected := []*common.Drive{
		{Common: common.Common{
			Serial: "Z9DF70I8FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}, ProductName: "NULL",
			Metadata: map[string]string{
				"Block Erase Sanitize Operation Supported":                          "false",
				"Crypto Erase Applies to All/Single Namespace(s) (t:All, f:Single)": "false",
				"Crypto Erase Sanitize Operation Supported":                         "false",
				"Crypto Erase Supported as part of Secure Erase":                    "true",
				"Format Applies to All/Single Namespace(s) (t:All, f:Single)":       "false",
				"No-Deallocate After Sanitize bit in Sanitize command Supported":    "false",
				"Overwrite Sanitize Operation Supported":                            "false",
			},
		}},
		{Common: common.Common{
			Serial: "Z9DF70I9FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}, ProductName: "NULL",
			Metadata: map[string]string{
				"Block Erase Sanitize Operation Supported":                          "false",
				"Crypto Erase Applies to All/Single Namespace(s) (t:All, f:Single)": "false",
				"Crypto Erase Sanitize Operation Supported":                         "false",
				"Crypto Erase Supported as part of Secure Erase":                    "true",
				"Format Applies to All/Single Namespace(s) (t:All, f:Single)":       "false",
				"No-Deallocate After Sanitize bit in Sanitize command Supported":    "false",
				"Overwrite Sanitize Operation Supported":                            "false",
			},
		}},
	}

	n := NewFakeNvme()

	drives, err := n.Drives(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, drives)
}

func Test_NvmeDriveCapabilities(t *testing.T) {
	n := NewFakeNvme()

	d := &nvmeDeviceAttributes{DevicePath: "/dev/nvme0"}

	capabilities, err := n.DriveCapabilities(context.TODO(), d.DevicePath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fixtureNvmeDeviceCapabilities, capabilities)
}

var fixtureNvmeDeviceCapabilities = []*common.Capability{
	{Name: "fmns", Description: "Format Applies to All/Single Namespace(s) (t:All, f:Single)", Enabled: false},
	{Name: "cens", Description: "Crypto Erase Applies to All/Single Namespace(s) (t:All, f:Single)", Enabled: false},
	{Name: "cese", Description: "Crypto Erase Supported as part of Secure Erase", Enabled: true},
	{Name: "cer", Description: "Crypto Erase Sanitize Operation Supported", Enabled: false},
	{Name: "ber", Description: "Block Erase Sanitize Operation Supported", Enabled: false},
	{Name: "owr", Description: "Overwrite Sanitize Operation Supported", Enabled: false},
	{Name: "ndi", Description: "No-Deallocate After Sanitize bit in Sanitize command Supported", Enabled: false},
}
