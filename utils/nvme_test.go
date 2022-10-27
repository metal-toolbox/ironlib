package utils

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/assert"
)

func Test_NvmeComponents(t *testing.T) {
	expected := []*common.Drive{
		{Common: common.Common{Serial: "Z9DF70I8FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}, ProductName: "NULL",
			Metadata: map[string]string{
				"Additional media modification after sanitize operation completes successfully is not defined": "false",
				"Block Erase Sanitize Operation Not Supported":                                                 "false",
				"Crypto Erase Applies to Single Namespace(s)":                                                  "false",
				"Crypto Erase Sanitize Operation Not Supported":                                                "false",
				"Crypto Erase Support":                                                                         "true",
				"Crypto Erase Supported as part of Secure Erase":                                               "true",
				"Format Applies to Single Namespace(s)":                                                        "false",
				"No-Deallocate After Sanitize bit in Sanitize command Supported":                               "false",
				"Overwrite Sanitize Operation Not Supported":                                                   "false",
				"Sanitize Support": "false",
			},
		}},
		{Common: common.Common{Serial: "Z9DF70I9FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}, ProductName: "NULL",
			Metadata: map[string]string{
				"Additional media modification after sanitize operation completes successfully is not defined": "false",
				"Block Erase Sanitize Operation Not Supported":                                                 "false",
				"Crypto Erase Applies to Single Namespace(s)":                                                  "false",
				"Crypto Erase Sanitize Operation Not Supported":                                                "false",
				"Crypto Erase Support":                                                                         "true",
				"Crypto Erase Supported as part of Secure Erase":                                               "true",
				"Format Applies to Single Namespace(s)":                                                        "false",
				"No-Deallocate After Sanitize bit in Sanitize command Supported":                               "false",
				"Overwrite Sanitize Operation Not Supported":                                                   "false",
				"Sanitize Support": "false",
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

var (
	fixtureNvmeDeviceCapabilities = []*common.Capability{
		{
			Name:        "sanicap",
			Description: "Sanitize Support",
			Enabled:     false,
		},
		{
			Name:        "ammasocsind",
			Description: "Additional media modification after sanitize operation completes successfully is not defined",
			Enabled:     false,
		},
		{
			Name:        "nasbiscs",
			Description: "No-Deallocate After Sanitize bit in Sanitize command Supported",
			Enabled:     false,
		},
		{
			Name:        "osons",
			Description: "Overwrite Sanitize Operation Not Supported",
			Enabled:     false,
		},
		{
			Name:        "besons",
			Description: "Block Erase Sanitize Operation Not Supported",
			Enabled:     false,
		},
		{
			Name:        "cesons",
			Description: "Crypto Erase Sanitize Operation Not Supported",
			Enabled:     false,
		},
		{
			Name:        "fna",
			Description: "Crypto Erase Support",
			Enabled:     true,
		},
		{
			Name:        "cesapose",
			Description: "Crypto Erase Supported as part of Secure Erase",
			Enabled:     true,
		},
		{
			Name:        "ceatsn",
			Description: "Crypto Erase Applies to Single Namespace(s)",
			Enabled:     false,
		},
		{
			Name:        "fatsn",
			Description: "Format Applies to Single Namespace(s)",
			Enabled:     false,
		},
	}
)
