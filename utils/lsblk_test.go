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
				Model:  "MTFDDAV240TDU",
				Serial: "203329F89392",
			},
			Protocol: "sata",
		},
		{
			Common: common.Common{
				Model:  "MTFDDAV240TDU",
				Serial: "203329F89796",
			},
			Protocol: "sata",
		},
		{
			Common: common.Common{
				Model:  "Micron_9300_MTFDHAL3T8TDP",
				Serial: "202728F691F5",
			},
			Protocol: "nvme",
		},
		{
			Common: common.Common{
				Model:  "Micron_9300_MTFDHAL3T8TDP",
				Serial: "202728F691C6",
			},
			Protocol: "nvme",
		},
	}
)
