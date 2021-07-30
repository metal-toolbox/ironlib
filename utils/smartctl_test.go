package utils

import (
	"context"
	"testing"

	"github.com/packethost/ironlib/model"
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
	expected := &SmartctlDriveAttributes{ModelName: "Micron_5200_MTFDDAK960TDN", SerialNumber: "2013273A99BD", FirmwareVersion: "D1MU020", Status: &SmartctlStatus{Passed: true}}
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
		{Serial: "2013273A99BD", Vendor: "Micron", Model: "Micron_5200_MTFDDAK960TDN", ProductName: "Micron_5200_MTFDDAK960TDN", Type: model.SlugDriveTypeSATASSD, Firmware: &model.Firmware{Installed: "D1MU020"}, SmartStatus: "true"},
		{Serial: "VDJ6SU9K", Vendor: "HGST", Model: "HGST HUS728T8TALE6L4", ProductName: "HGST HUS728T8TALE6L4", Type: model.SlugDriveTypeSATAHDD, Firmware: &model.Firmware{Installed: "V8GNW460"}, SmartStatus: "true"},
		{Serial: "Z9DF70I8FY3L", Vendor: "Toshiba", Model: "KXG60ZNV256G TOSHIBA", ProductName: "KXG60ZNV256G TOSHIBA", Type: model.SlugDriveTypePCIeNVMEeSSD, Firmware: &model.Firmware{Installed: "AGGA4104"}, SmartStatus: "true"},
		{Serial: "Z9DF70I9FY3L", Vendor: "Toshiba", Model: "KXG60ZNV256G TOSHIBA", ProductName: "KXG60ZNV256G TOSHIBA", Type: model.SlugDriveTypePCIeNVMEeSSD, Firmware: &model.Firmware{Installed: "AGGA4104"}, SmartStatus: "true"},
	}
	s := newFakeSmartctl()

	drives, err := s.Drives(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, drives)
}
