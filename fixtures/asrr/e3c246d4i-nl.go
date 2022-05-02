package asrr

import "github.com/packethost/ironlib/model"

// nolint //testcode
// inventory taken with lshw
var E3C246D4INL = &model.Device{
	Vendor: "Packet",
	Model:  "c3.small.x86 (To Be Filled By O.E.M.)",
	Serial: "D5S0R8000105",
	BIOS: &model.BIOS{
		Description:   "BIOS",
		Vendor:        "American Megatrends Inc.",
		SizeBytes:     65536,
		CapacityBytes: 33554432,
		Firmware: &model.Firmware{
			Installed: "L2.07B",
			Available: "",
		},
	},
	CPLD: &model.CPLD{},
	Mainboard: &model.Mainboard{
		Description: "Motherboard",
		Vendor:      "ASRockRack",
		Model:       "E3C246D4I-NL",
		Serial:      "196231220000153",
		PhysicalID:  "0",
	},
	CPUs: []*model.CPU{
		{
			Description:  "Intel(R) Xeon(R) E-2278G CPU @ 3.40GHz",
			Vendor:       "Intel Corp.",
			Model:        "Intel(R) Xeon(R) E-2278G CPU @ 3.40GHz",
			Serial:       "",
			Slot:         "CPU1",
			ClockSpeedHz: 100000000,
			Cores:        8,
			Threads:      16,
		},
	},
	Memory: []*model.Memory{
		{
			Description:  "SODIMM DDR4 Synchronous 2666 MHz (0.4 ns)",
			Slot:         "ChannelA-DIMM0",
			Type:         "",
			Vendor:       "Micron",
			Model:        "18ASF2G72HZ-2G6E1",
			Serial:       "F0F9053F",
			SizeBytes:    17179869184,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 2666000000,
		},
		{
			Description:  "SODIMM DDR4 Synchronous 2666 MHz (0.4 ns)",
			Slot:         "ChannelB-DIMM0",
			Type:         "",
			Vendor:       "Micron",
			Model:        "18ASF2G72HZ-2G6E1",
			Serial:       "F0F90894",
			SizeBytes:    17179869184,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 2666000000,
		},
	},
	NICs: []*model.NIC{
		{
			Description: "Ethernet interface",
			Vendor:      "Intel Corporation",
			Model:       "Ethernet Controller X710 for 10GbE SFP+",
			Serial:      "b4:96:91:70:26:c8",
			SpeedBits:   10000000000,
			PhysicalID:  "0",
			Firmware: &model.Firmware{
				Available: "",
				Installed: "1.1853.0",
			},
		},
	},
	Drives: []*model.Drive{
		{
			Description:       "ATA Disk",
			Serial:            "PHYF001300HB480BGN",
			StorageController: "",
			Vendor:            "",
			Model:             "INTEL SSDSC2KB48",
			WWN:               "",
			SizeBytes:         480103981056,
			CapacityBytes:     0,
			BlockSize:         0,
			Metadata:          nil,
			Oem:               false,
		},
		{
			Description:       "ATA Disk",
			Serial:            "PHYF001209KL480BGN",
			StorageController: "",
			Vendor:            "",
			Model:             "INTEL SSDSC2KB48",
			WWN:               "",
			SizeBytes:         480103981056,
			CapacityBytes:     0,
			BlockSize:         0,
			Metadata:          nil,
			Oem:               false,
		},
	},
	StorageControllers: []*model.StorageController{
		{
			Description: "SATA controller",
			Vendor:      "Intel Corporation",
			Model:       "Cannon Lake PCH SATA AHCI Controller",
			Serial:      "",
			Interface:   "SATA",
			PhysicalID:  "17",
		},
	},
	GPUs: []*model.GPU{},
	BMC:  &model.BMC{},
	TPM:  &model.TPM{},
	PSUs: []*model.PSU{},
}
