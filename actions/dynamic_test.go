package actions

import (
	"testing"

	"github.com/metal-toolbox/ironlib/utils"
	"github.com/stretchr/testify/assert"
)

func Test_driveCapabilityCollectorByLogicalName(t *testing.T) {
	testcases := []struct {
		testname           string
		logicalName        string
		collectors         []DriveCapabilityCollector
		expectedCollectors DriveCapabilityCollector
	}{
		{
			"given a scsi drive and two collectors a hdparm collector is returned",
			"/dev/sda1",
			[]DriveCapabilityCollector{utils.NewNvmeCmd(false), utils.NewHdparmCmd(false)},
			utils.NewHdparmCmd(false),
		},
		{
			"given a nvme drive and two collectors a nvme collector is returned",
			"/dev/nvme0",
			[]DriveCapabilityCollector{utils.NewHdparmCmd(false), utils.NewNvmeCmd(false)},
			utils.NewNvmeCmd(false),
		},
		{
			"given a scsi drive, an hdparm collector is returned",
			"/dev/sda1",
			[]DriveCapabilityCollector{},
			utils.NewHdparmCmd(false),
		},
		{
			"given a nvme drive, an nvme collector is returned",
			"/dev/nvme0",
			[]DriveCapabilityCollector{},
			utils.NewNvmeCmd(false),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.testname, func(t *testing.T) {
			got := driveCapabilityCollectorByLogicalName(tc.logicalName, false, tc.collectors)
			assert.EqualValues(t, tc.expectedCollectors, got)
		})
	}
}
