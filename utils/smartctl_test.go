package utils

import (
	"context"
	"testing"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeSmartctl() *Smartctl {
	return &Smartctl{
		Executor: NewFakeSmartctlExecutor("smartctl", "../fixtures/utils/smartctl"),
	}
}

func Test_SmartctlScan(t *testing.T) {
	expected := &SmartctlScan{
		Drives: []*SmartctlDrive{
			{Name: "/dev/sda", Type: "scsi", Protocol: "SCSI"},
			{Name: "/dev/sdb", Type: "scsi", Protocol: "SCSI"},
			{Name: "/dev/sdc", Type: "scsi", Protocol: "SCSI"},
			{Name: "/dev/nvme0", Type: "nvme", Protocol: "NVMe"},
			{Name: "/dev/nvme1", Type: "nvme", Protocol: "NVMe"},
		},
	}

	s := newFakeSmartctl()

	scan, err := s.Scan()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, scan)
}

func Test_SmartctlAllSCSI(t *testing.T) {
	expected := &SmartctlDriveAttributes{ModelName: "Micron_5200_MTFDDAK960TDN", ModelFamily: "Micron 5100 Pro / 5200 SSDs", SerialNumber: "2013273A99BD", FirmwareVersion: "D1MU020", Status: &SmartctlStatus{Passed: true}}
	s := newFakeSmartctl()

	results, err := s.All("/dev/sda")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, results)
}

func Test_SmartctlAllNVME(t *testing.T) {
	expected := &SmartctlDriveAttributes{ModelName: "KXG60ZNV256G TOSHIBA", SerialNumber: "Z9DF70I8FY3L", FirmwareVersion: "AGGA4104", Status: &SmartctlStatus{Passed: true}}
	s := newFakeSmartctl()

	results, err := s.All("/dev/nvme0")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, results)
}

func Test_SmartctlDeviceAttributes(t *testing.T) {
	expected := []*model.Drive{
		{Serial: "2013273A99BD", Vendor: model.VendorMicron, Model: "Micron_5200_MTFDDAK960TDN", ProductName: "Micron_5200_MTFDDAK960TDN", Type: model.SlugDriveTypeSATASSD, Firmware: &model.Firmware{Installed: "D1MU020"}, SmartStatus: "ok"},
		{Serial: "VDJ6SU9K", Vendor: model.VendorHGST, Model: "HGST HUS728T8TALE6L4", ProductName: "HGST HUS728T8TALE6L4", Type: model.SlugDriveTypeSATAHDD, Firmware: &model.Firmware{Installed: "V8GNW460"}, SmartStatus: "ok"},
		{Serial: "PHYH1016001D240J", Vendor: model.VendorDell, Model: "SSDSCKKB240G8R", ProductName: "SSDSCKKB240G8R", Type: "Unknown", Firmware: &model.Firmware{Installed: "XC31DL6R"}, SmartStatus: "ok", OemID: "DELL(tm)"},
		{Serial: "Z9DF70I8FY3L", Vendor: model.VendorToshiba, Model: "KXG60ZNV256G TOSHIBA", ProductName: "KXG60ZNV256G TOSHIBA", Type: model.SlugDriveTypePCIeNVMEeSSD, Firmware: &model.Firmware{Installed: "AGGA4104"}, SmartStatus: "ok"},
		{Serial: "Z9DF70I9FY3L", Vendor: model.VendorToshiba, Model: "KXG60ZNV256G TOSHIBA", ProductName: "KXG60ZNV256G TOSHIBA", Type: model.SlugDriveTypePCIeNVMEeSSD, Firmware: &model.Firmware{Installed: "AGGA4104"}, SmartStatus: "ok"},
	}
	s := newFakeSmartctl()

	drives, err := s.Drives(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, drives)
}

func Test_SmartctlNonZeroExit(t *testing.T) {
	expected := &SmartctlDriveAttributes{
		ModelName:       "SSDSCKKB240G8R",
		OemProductID:    "DELL(tm)",
		ModelFamily:     "Dell Certified Intel S4x00/D3-S4x10 Series SSDs",
		SerialNumber:    "PHYH1016001D240J",
		FirmwareVersion: "XC31DL6R",
		Status: &SmartctlStatus{
			Passed: true,
		},
		Errors: []string{
			"Some SMART or other ATA command to the disk failed, or there was a checksum error in a SMART data structure",
		},
	}

	s := newFakeSmartctl()
	s.Executor.SetExitCode(4)

	results, err := s.All("/dev/sdc")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, results)
}

func Test_exponentInt(t *testing.T) {
	m := map[int]int{
		0: 1,
		1: 2,
		2: 4,
		3: 8,
		4: 16,
		5: 32,
	}
	for k, v := range m {
		i := exponentInt(2, k)
		assert.Equal(t, v, i)
	}
}

func Test_maskExitCode(t *testing.T) {
	m := map[int][]int{
		0:  {},
		1:  {0},
		2:  {1},
		3:  {0, 1},
		4:  {2},
		10: {1, 3},
	}
	for k, v := range m {
		i := maskExitCode(k)
		assert.Equal(t, v, i)
	}
}
