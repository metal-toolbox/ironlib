package utils

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/assert"
)

func newFakeNvme() *Nvme {
	return &Nvme{
		Executor: NewFakeExecutor("nvme"),
	}
}

func Test_NvmeComponents(t *testing.T) {
	expected := []*common.Drive{
		{Common: common.Common{Serial: "Z9DF70I8FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}, ProductName: "NULL"}},
		{Common: common.Common{Serial: "Z9DF70I9FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}, ProductName: "NULL"}},
	}

	n := newFakeNvme()

	drives, err := n.Drives(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, drives)
}

func Test_ParseNvmeFeatures(t *testing.T) {
	n := newFakeNvme()

	d := &nvmeDeviceAttributes{DevicePath: "/dev/nvme0"}

	features, err := n.ParseNvmeFeatures(d.DevicePath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fixtureNvmeDevicefeatures, features)
}

var (
	fixtureNvmeDevicefeatures = []nvmeDeviceFeatures{
		{
			Name:        "sanicap",
			Description: "Sanitize Support",
			Enabled:     false,
		},
		{
			Name:        "",
			Description: "Additional media modification after sanitize operation completes successfully is not defined",
			Enabled:     false,
		},
		{
			Name:        "",
			Description: "No-Deallocate After Sanitize bit in Sanitize command Supported",
			Enabled:     false,
		},
		{
			Name:        "",
			Description: "Overwrite Sanitize Operation Not Supported",
			Enabled:     false,
		},
		{
			Name:        "",
			Description: "Block Erase Sanitize Operation Not Supported",
			Enabled:     false,
		},
		{
			Name:        "",
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
