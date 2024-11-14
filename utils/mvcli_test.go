package utils

import (
	"context"
	"os"
	"testing"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/stretchr/testify/assert"
)

func newFakeMvcli(t *testing.T, fixtureName string) *Mvcli {
	e := &Mvcli{
		Executor: NewFakeExecutor("mvcli"),
	}

	b, err := os.ReadFile("../fixtures/utils/mvcli/" + fixtureName)
	if err != nil {
		t.Error(err)
	}

	e.Executor.SetStdout(b)

	return e
}

func Test_MvcliStorageControllers(t *testing.T) {
	expected := []*common.StorageController{
		{
			Common: common.Common{
				Description: "1b4b-9230 (1028-1fe2)",
				Vendor:      "marvell",
				Model:       "1b4b-9230",
				Serial:      "1B4B:9230",
				Oem:         false,
				Metadata:    map[string]string{},
				Firmware: &common.Firmware{
					Installed: "2.5.13.1210",
					Metadata: map[string]string{
						"bios_version":        "1.0.13.1003",
						"rom_version":         "2.5.13.3016",
						"boot_loader_version": "2.1.0.1009",
					},
				},
			},

			SupportedRAIDTypes: "RAID1",
			SpeedGbps:          0,
			PhysicalID:         "",
		},
	}

	cli := newFakeMvcli(t, "info-hba")

	inventory, err := cli.StorageControllers(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, inventory)
}

func Test_MvcliDrives(t *testing.T) {
	expected := []*common.Drive{
		{
			Common: common.Common{
				Description: "MTFDDAV240TCB",
				Vendor:      "micron",
				Model:       "MTFDDAV240TCB",
				Serial:      "18341E6651A9",
				Oem:         false,
				Metadata:    map[string]string{},
				Firmware: &common.Firmware{
					Installed: "D0DE008",
					Metadata:  nil,
				},
			},
			CapacityBytes:            234365528 * 1000,
			BlockSizeBytes:           234431064 * 1000,
			Type:                     common.SlugDriveTypeSATASSD,
			NegotiatedSpeedGbps:      6,
			StorageControllerDriveID: 0,
		},
		{
			Common: common.Common{
				Description: "MTFDDAV240TCB",
				Vendor:      "micron",
				Model:       "MTFDDAV240TCB",
				Serial:      "18341E6651BA",
				Oem:         false,
				Metadata:    map[string]string{},
				Firmware: &common.Firmware{
					Installed: "D0DE008",
					Metadata:  nil,
				},
			},
			CapacityBytes:            234365528 * 1000,
			BlockSizeBytes:           234431064 * 1000,
			Type:                     common.SlugDriveTypeSATASSD,
			NegotiatedSpeedGbps:      6,
			StorageControllerDriveID: 1,
		},
	}

	cli := newFakeMvcli(t, "info-pd")

	inventory, err := cli.Drives(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, inventory)
}
