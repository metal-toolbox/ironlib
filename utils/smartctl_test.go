package utils

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/common"
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
	expected := []*common.Drive{
		{Common: common.Common{Serial: "2013273A99BD", Vendor: common.VendorMicron, Model: "Micron_5200_MTFDDAK960TDN", ProductName: "Micron_5200_MTFDDAK960TDN", Firmware: &common.Firmware{Installed: "D1MU020"}}, Type: common.SlugDriveTypeSATASSD, SmartStatus: "ok", StorageControllerDriveID: -1},
		{Common: common.Common{Serial: "VDJ6SU9K", Vendor: common.VendorHGST, Model: "HGST HUS728T8TALE6L4", ProductName: "HGST HUS728T8TALE6L4", Firmware: &common.Firmware{Installed: "V8GNW460"}}, Type: common.SlugDriveTypeSATAHDD, SmartStatus: "ok", StorageControllerDriveID: -1},
		{Common: common.Common{Serial: "PHYH1016001D240J", Vendor: common.VendorDell, Model: "SSDSCKKB240G8R", ProductName: "SSDSCKKB240G8R", Firmware: &common.Firmware{Installed: "XC31DL6R"}}, Type: "Unknown", SmartStatus: "ok", OemID: "DELL(tm)", StorageControllerDriveID: -1},
		{Common: common.Common{Serial: "Z9DF70I8FY3L", Vendor: common.VendorToshiba, Model: "KXG60ZNV256G TOSHIBA", ProductName: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}}, Type: common.SlugDriveTypePCIeNVMEeSSD, SmartStatus: "ok", StorageControllerDriveID: -1},
		{Common: common.Common{Serial: "Z9DF70I9FY3L", Vendor: common.VendorToshiba, Model: "KXG60ZNV256G TOSHIBA", ProductName: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}}, Type: common.SlugDriveTypePCIeNVMEeSSD, SmartStatus: "ok", StorageControllerDriveID: -1},
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
