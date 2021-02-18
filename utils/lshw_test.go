package utils

import (
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeLshw() *Lshw {
	return &Lshw{
		Executor: NewFakeExecutor("lshw"),
	}
}

// from test_data/lshw_e3c246d4l-nl.json
var e3c246d4INL = &model.Device{
	BIOS: &model.BIOS{
		Description:       "BIOS",
		Vendor:            "American Megatrends Inc.",
		SizeBytes:         65536,
		CapacityBytes:     33554432,
		FirmwareDate:      "",
		FirmwareInstalled: "L2.07B",
		FirmwareAvailable: "",
		FirmwareManaged:   true,
	},
	Mainboard: &model.Mainboard{
		Description:     "Motherboard",
		Vendor:          "ASRockRack",
		Model:           "E3C246D4I-NL",
		Serial:          "196231220000153",
		PhysicalID:      "0",
		FirmwareManaged: true,
	},
	CPUs: []*model.CPU{
		{
			Description:     "Intel(R) Xeon(R) E-2278G CPU @ 3.40GHz",
			Vendor:          "Intel Corp.",
			Model:           "Intel(R) Xeon(R) E-2278G CPU @ 3.40GHz",
			Serial:          "",
			Slot:            "CPU1",
			ClockSpeedHz:    100000000,
			Cores:           8,
			Threads:         16,
			FirmwareManaged: false,
		},
	},
	Memory: []*model.Memory{
		{
			Description:     "SODIMM DDR4 Synchronous 2666 MHz (0.4 ns)",
			Slot:            "ChannelA-DIMM0",
			Type:            "",
			Vendor:          "Micron",
			Model:           "18ASF2G72HZ-2G6E1",
			Serial:          "F0F9053F",
			SizeBytes:       17179869184,
			FormFactor:      "",
			PartNumber:      "",
			ClockSpeedHz:    2666000000,
			FirmwareManaged: false,
		},
		{
			Description:     "SODIMM DDR4 Synchronous 2666 MHz (0.4 ns)",
			Slot:            "ChannelB-DIMM0",
			Type:            "",
			Vendor:          "Micron",
			Model:           "18ASF2G72HZ-2G6E1",
			Serial:          "F0F90894",
			SizeBytes:       17179869184,
			FormFactor:      "",
			PartNumber:      "",
			ClockSpeedHz:    2666000000,
			FirmwareManaged: false,
		},
	},
	NICs: []*model.NIC{
		{
			Description:     "Ethernet interface",
			Vendor:          "Intel Corporation",
			Model:           "Ethernet Controller X710 for 10GbE SFP+",
			Serial:          "b4:96:91:70:26:c8",
			SpeedBits:       10000000000,
			PhysicalID:      "0",
			FirmwareManaged: false,
		},
		{
			Description:     "Ethernet interface",
			Vendor:          "Intel Corporation",
			Model:           "Ethernet Controller X710 for 10GbE SFP+",
			Serial:          "b4:96:91:70:26:c8",
			SpeedBits:       10000000000,
			PhysicalID:      "0.1",
			FirmwareManaged: false,
		},
	},
	Drives: []*model.Drive{
		{
			Description:       "ATA Disk",
			Serial:            "PHYF001300HB480BGN",
			StorageController: "",
			Vendor:            "",
			Model:             "INTEL SSDSC2KB48",
			Name:              "",
			FirmwareInstalled: "",
			FirmwareAvailable: "",
			WWN:               "",
			SizeBytes:         480103981056,
			CapacityBytes:     0,
			BlockSize:         0,
			Metadata:          nil,
			FirmwareManaged:   false,
			Oem:               false,
		},
		{
			Description:       "ATA Disk",
			Serial:            "PHYF001209KL480BGN",
			StorageController: "",
			Vendor:            "",
			Model:             "INTEL SSDSC2KB48",
			Name:              "",
			FirmwareInstalled: "",
			FirmwareAvailable: "",
			WWN:               "",
			SizeBytes:         480103981056,
			CapacityBytes:     0,
			BlockSize:         0,
			Metadata:          nil,
			FirmwareManaged:   false,
			Oem:               false,
		},
	},
	StorageController: []*model.StorageController{
		{
			Description:     "SATA controller",
			Vendor:          "Intel Corporation",
			Model:           "Cannon Lake PCH SATA AHCI Controller",
			Serial:          "",
			PhysicalID:      "17",
			FirmwareManaged: false,
		},
	},
	GPUs: nil,
	BMC:  nil,
	TPM:  nil,
}

func Test_lshw(t *testing.T) {
	l := newFakeLshw()
	device, err := l.Inventory()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, e3c246d4INL, device)
}
