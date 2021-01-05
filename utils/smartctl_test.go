package utils

import (
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeSmartctl() *Smartctl {
	return &Smartctl{
		Executor: NewFakeExecutor("smartctl"),
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

	expected := &SmartctlDriveAttributes{ModelName: "Micron_5200_MTFDDAK960TDN", SerialNumber: "2013273A99BD", FirmwareVersion: "D1MU020"}
	s := newFakeSmartctl()
	results, err := s.All("/dev/sda")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, results)

}

func Test_SmartctlAllNVME(t *testing.T) {

	expected := &SmartctlDriveAttributes{ModelName: "KXG60ZNV256G TOSHIBA", SerialNumber: "Z9DF70I8FY3L", FirmwareVersion: "AGGA4104"}
	s := newFakeSmartctl()
	results, err := s.All("/dev/nvme0")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, results)

}

func Test_SmartctlDeviceAttributes(t *testing.T) {
	expected := []*model.Component{
		{Serial: "2013273A99BD", Vendor: "Micron", Model: "Micron_5200_MTFDDAK960TDN", Name: "scsi", Slug: "scsi", FirmwareInstalled: "D1MU020"},
		{Serial: "2013273A99BD", Vendor: "Micron", Model: "Micron_5200_MTFDDAK960TDN", Name: "scsi", Slug: "scsi", FirmwareInstalled: "D1MU020"},
		{Serial: "Z9DF70I8FY3L", Vendor: "Toshiba", Model: "KXG60ZNV256G TOSHIBA", Name: "nvme", Slug: "nvme", FirmwareInstalled: "AGGA4104"},
		{Serial: "Z9DF70I8FY3L", Vendor: "Toshiba", Model: "KXG60ZNV256G TOSHIBA", Name: "nvme", Slug: "nvme", FirmwareInstalled: "AGGA4104"},
	}
	s := newFakeSmartctl()
	inv, err := s.Components()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, inv)
}
