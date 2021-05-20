package utils

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func Test_lshw_e3c246d4Inl(t *testing.T) {
	b, err := ioutil.ReadFile("test_data/lshw_e3c246d4I-nl.json")
	if err != nil {
		t.Error(err)
	}

	l := NewFakeLshw(bytes.NewReader(b))
	device := model.NewDevice()

	err = l.Inventory(device)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, e3c246d4INL, device)
}

func Test_lshw(t *testing.T) {
	b, err := ioutil.ReadFile("test_data/r6515/lshw.json")
	if err != nil {
		t.Error(err)
	}

	l := NewFakeLshw(bytes.NewReader(b))
	device := model.NewDevice()

	err = l.Inventory(device)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, r6515, device)
}

// from test_data/lshw_e3c246d4l-nl.json
var e3c246d4INL = &model.Device{
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
			Managed:   true,
		},
	},
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
		},
		{
			Description: "Ethernet interface",
			Vendor:      "Intel Corporation",
			Model:       "Ethernet Controller X710 for 10GbE SFP+",
			Serial:      "b4:96:91:70:26:c8",
			SpeedBits:   10000000000,
			PhysicalID:  "0.1",
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

var r6515 = &model.Device{
	Vendor:       "Dell Inc.",
	Model:        "PowerEdge R6515 (SKU=NotProvided;ModelName=PowerEdge R6515)",
	Serial:       "11WLK93",
	GPUs:         []*model.GPU{},
	BMC:          &model.BMC{},
	TPM:          &model.TPM{},
	PSUs:         []*model.PSU{},
	HardwareType: "",
	Chassis:      "",
	BIOS: &model.BIOS{
		Description:   "BIOS",
		Vendor:        "Dell Inc.",
		SizeBytes:     65536,
		CapacityBytes: 33554432,
		Firmware: &model.Firmware{
			Available: "",
			Installed: "1.7.4",
			Managed:   true,
		},
	},
	Mainboard: &model.Mainboard{
		Description: "Motherboard",
		Vendor:      "Dell Inc.",
		Model:       "0R4CNN",
		Serial:      ".11WLK93.CNCMS0009G0078.",
		PhysicalID:  "0",
	},
	CPUs: []*model.CPU{{
		Description:  "AMD EPYC 7502P 32-Core Processor",
		Vendor:       "Advanced Micro Devices [AMD]",
		Model:        "AMD EPYC 7502P 32-Core Processor",
		Serial:       "",
		Slot:         "CPU1",
		ClockSpeedHz: 2000000000,
		Cores:        32,
		Threads:      64,
	},
	},
	Memory: []*model.Memory{
		{
			Description:  "DIMM DDR4 Synchronous Registered (Buffered) 3200 MHz (0.3 ns)",
			Slot:         "A1",
			Type:         "",
			Vendor:       "Hynix Semiconductor (Hyundai Electronics)",
			Model:        "HMA84GR7DJR4N-XN",
			Serial:       "3510CB7F",
			SizeBytes:    34359738368,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 3200000000,
		},
		{
			Description:  "DIMM DDR4 Synchronous Registered (Buffered) 3200 MHz (0.3 ns)",
			Slot:         "A2",
			Type:         "",
			Vendor:       "Hynix Semiconductor (Hyundai Electronics)",
			Model:        "HMA84GR7DJR4N-XN",
			Serial:       "3510CBDD",
			SizeBytes:    34359738368,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 3200000000,
		},
		{
			Description:  "DIMM DDR4 Synchronous Registered (Buffered) 3200 MHz (0.3 ns)",
			Slot:         "A3",
			Type:         "",
			Vendor:       "Hynix Semiconductor (Hyundai Electronics)",
			Model:        "HMA84GR7DJR4N-XN",
			Serial:       "3510CC97",
			SizeBytes:    34359738368,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 3200000000,
		},
		{
			Description:  "DIMM DDR4 Synchronous Registered (Buffered) 3200 MHz (0.3 ns)",
			Slot:         "A4",
			Type:         "",
			Vendor:       "Hynix Semiconductor (Hyundai Electronics)",
			Model:        "HMA84GR7DJR4N-XN",
			Serial:       "3510CC8C",
			SizeBytes:    34359738368,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 3200000000,
		},
		{
			Description:  "DIMM DDR4 Synchronous Registered (Buffered) 3200 MHz (0.3 ns)",
			Slot:         "A5",
			Type:         "",
			Vendor:       "Hynix Semiconductor (Hyundai Electronics)",
			Model:        "HMA84GR7DJR4N-XN",
			Serial:       "3510CBDE",
			SizeBytes:    34359738368,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 3200000000,
		},
		{
			Description:  "DIMM DDR4 Synchronous Registered (Buffered) 3200 MHz (0.3 ns)",
			Slot:         "A6",
			Type:         "",
			Vendor:       "Hynix Semiconductor (Hyundai Electronics)",
			Model:        "HMA84GR7DJR4N-XN",
			Serial:       "3510CB77",
			SizeBytes:    34359738368,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 3200000000,
		},
		{
			Description:  "DIMM DDR4 Synchronous Registered (Buffered) 3200 MHz (0.3 ns)",
			Slot:         "A7",
			Type:         "",
			Vendor:       "Hynix Semiconductor (Hyundai Electronics)",
			Model:        "HMA84GR7DJR4N-XN",
			Serial:       "3510CC78",
			SizeBytes:    34359738368,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 3200000000,
		},
		{
			Description:  "DIMM DDR4 Synchronous Registered (Buffered) 3200 MHz (0.3 ns)",
			Slot:         "A8",
			Type:         "",
			Vendor:       "Hynix Semiconductor (Hyundai Electronics)",
			Model:        "HMA84GR7DJR4N-XN",
			Serial:       "3510CC93",
			SizeBytes:    34359738368,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 3200000000,
		},
		{
			Description:  "[empty]",
			Slot:         "A9",
			Type:         "",
			Vendor:       "",
			Model:        "",
			Serial:       "",
			SizeBytes:    0,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 0,
		},
		{
			Description:  "[empty]",
			Slot:         "A10",
			Type:         "",
			Vendor:       "",
			Model:        "",
			Serial:       "",
			SizeBytes:    0,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 0,
		},
		{
			Description:  "[empty]",
			Slot:         "A11",
			Type:         "",
			Vendor:       "",
			Model:        "",
			Serial:       "",
			SizeBytes:    0,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 0,
		},
		{
			Description:  "[empty]",
			Slot:         "A12",
			Type:         "",
			Vendor:       "",
			Model:        "",
			Serial:       "",
			SizeBytes:    0,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 0,
		},
		{
			Description:  "[empty]",
			Slot:         "A13",
			Type:         "",
			Vendor:       "",
			Model:        "",
			Serial:       "",
			SizeBytes:    0,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 0,
		},
		{
			Description:  "[empty]",
			Slot:         "A14",
			Type:         "",
			Vendor:       "",
			Model:        "",
			Serial:       "",
			SizeBytes:    0,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 0,
		},
		{
			Description:  "[empty]",
			Slot:         "A15",
			Type:         "",
			Vendor:       "",
			Model:        "",
			Serial:       "",
			SizeBytes:    0,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 0,
		},
		{
			Description:  "[empty]",
			Slot:         "A16",
			Type:         "",
			Vendor:       "",
			Model:        "",
			Serial:       "",
			SizeBytes:    0,
			FormFactor:   "",
			PartNumber:   "",
			ClockSpeedHz: 0,
		},
	},
	NICs: []*model.NIC{
		{
			Description: "Ethernet interface",
			Vendor:      "Intel Corporation",
			Model:       "Ethernet Controller XXV710 for 25GbE SFP28",
			Serial:      "40:a6:b7:4e:8a:a0",
			SpeedBits:   0,
			PhysicalID:  "0",
		},
		{
			Description: "Ethernet interface",
			Vendor:      "Intel Corporation",
			Model:       "Ethernet Controller XXV710 for 25GbE SFP28",
			Serial:      "40:a6:b7:4e:8a:a0",
			SpeedBits:   0,
			PhysicalID:  "0.1",
		},
	},
	Drives: []*model.Drive{
		{
			Description:       "NVMe device",
			Serial:            "202728F691F5",
			StorageController: "",
			Vendor:            "Micron Technology Inc",
			Model:             "Micron_9300_MTFDHAL3T8TDP",
			WWN:               "",
			SizeBytes:         0,
			CapacityBytes:     0,
			BlockSize:         0,
		},
		{
			Description:       "NVMe device",
			Serial:            "202728F691C6",
			StorageController: "",
			Vendor:            "Micron Technology Inc",
			Model:             "Micron_9300_MTFDHAL3T8TDP",
			WWN:               "",
			SizeBytes:         0,
			CapacityBytes:     0,
			BlockSize:         0,
		},
		{
			Description:       "ATA Disk",
			Serial:            "203329F89392",
			StorageController: "",
			Vendor:            "",
			Model:             "MTFDDAV240TDU",
			WWN:               "",
			SizeBytes:         240057409536,
			CapacityBytes:     0,
			BlockSize:         0,
		},
		{
			Description:       "ATA Disk",
			Serial:            "203329F89796",
			StorageController: "",
			Vendor:            "",
			Model:             "MTFDDAV240TDU",
			WWN:               "",
			SizeBytes:         240057409536,
			CapacityBytes:     0,
			BlockSize:         0,
		},
	},
	StorageControllers: []*model.StorageController{
		{
			Description: "Serial Attached SCSI controller",
			Vendor:      "Broadcom / LSI",
			Model:       "SAS3008 PCI-Express Fusion-MPT SAS-3",
			Serial:      "",
			Interface:   "SAS",
			PhysicalID:  "0",
		},
		{
			Description: "SATA controller",
			Vendor:      "Advanced Micro Devices, Inc. [AMD]",
			Model:       "FCH SATA Controller [AHCI mode]",
			Serial:      "",
			Interface:   "SATA",
			PhysicalID:  "0",
		},
		{
			Description: "SATA controller",
			Vendor:      "Advanced Micro Devices, Inc. [AMD]",
			Model:       "FCH SATA Controller [AHCI mode]",
			Serial:      "",
			Interface:   "SATA",
			PhysicalID:  "0",
		},
		{
			Description: "SATA controller",
			Vendor:      "Marvell Technology Group Ltd.",
			Model:       "88SE9230 PCIe SATA 6Gb/s Controller",
			Serial:      "",
			Interface:   "SATA",
			PhysicalID:  "0",
		},
	},
}
