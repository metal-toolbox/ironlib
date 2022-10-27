package utils

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/assert"
)

func Test_lsblk_Drives(t *testing.T) {
	l := NewFakeLsblk()

	drives, err := l.Drives(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fixtureLsblkDrives, drives)
}

var (
	fixtureLsblkDrives = []*common.Drive{
		{
			Common: common.Common{
				Description: "Samsung SSD 850 EVO 2TB",
				Vendor:      "SSD",
				Model:       "Samsung SSD 850 EVO 2TB",
				Serial:      "S2RLNX0H700433B",
				Metadata:    map[string]string{},
				Firmware:    &common.Firmware{},
			},
			Protocol: "sata",
		},
		{
			Common: common.Common{
				Oem:         false,
				Description: "Samsung SSD 850 EVO 2TB",
				Vendor:      "SSD",
				Model:       "Samsung SSD 850 EVO 2TB",
				Serial:      "S2RLNX0H700428V",
				Metadata:    map[string]string{},
				Firmware:    &common.Firmware{},
			},
			Protocol: "sata",
		},
		{
			Common: common.Common{
				Oem:         false,
				Description: "Samsung SSD 970 PRO 512GB",
				Vendor:      "SSD",
				Model:       "Samsung SSD 970 PRO 512GB",
				Serial:      "S5JYNC0N102898N",
				Metadata:    map[string]string{},
				Firmware:    &common.Firmware{},
			},
			Protocol: "nvme",
		},
	}
)
