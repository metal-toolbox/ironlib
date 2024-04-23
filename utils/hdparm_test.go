package utils

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/assert"
)

func Test_HdparmDriveCapabilities(t *testing.T) {
	h := NewFakeHdparm()

	device := "/dev/sda"

	capabilities, err := h.DriveCapabilities(context.TODO(), device)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fixtureHdparmDeviceCapabilities, capabilities)
}

var fixtureHdparmDeviceCapabilities = []*common.Capability{
	{
		Name:        "sf",
		Description: "SMART feature",
		Enabled:     true,
	},
	{
		Name:        "smf",
		Description: "Security Mode feature",
		Enabled:     false,
	},
	{
		Name:        "pmf",
		Description: "Power Management feature",
		Enabled:     true,
	},
	{
		Name:        "wc",
		Description: "Write cache",
		Enabled:     true,
	},
	{
		Name:        "la",
		Description: "Look ahead",
		Enabled:     true,
	},
	{
		Name:        "hpaf",
		Description: "Host Protected Area feature",
		Enabled:     true,
	},
	{
		Name:        "wb",
		Description: "WRITE BUFFER",
		Enabled:     true,
	},
	{
		Name:        "rb",
		Description: "READ BUFFER",
		Enabled:     true,
	},
	{
		Name:        "nc",
		Description: "NOP cmd",
		Enabled:     true,
	},
	{
		Name:        "dm",
		Description: "DOWNLOAD MICROCODE",
		Enabled:     true,
	},
	{
		Name:        "smse",
		Description: "SET MAX security extension",
		Enabled:     false,
	},
	{
		Name:        "4baf",
		Description: "48 bit Address feature",
		Enabled:     true,
	},
	{
		Name:        "dcof",
		Description: "Device Configuration Overlay feature",
		Enabled:     true,
	},
	{
		Name:        "mfc",
		Description: "Mandatory FLUSH CACHE",
		Enabled:     true,
	},
	{
		Name:        "fce",
		Description: "FLUSH CACHE EXT",
		Enabled:     true,
	},
	{
		Name:        "sel",
		Description: "SMART error logging",
		Enabled:     true,
	},
	{
		Name:        "sst",
		Description: "SMART self test",
		Enabled:     true,
	},
	{
		Name:        "gplf",
		Description: "General Purpose Logging feature",
		Enabled:     true,
	},
	{
		Name:        "wdfe",
		Description: "WRITE DMAMULTIPLE FUA EXT",
		Enabled:     true,
	},
	{
		Name:        "6bwwn",
		Description: "64 bit World wide name",
		Enabled:     true,
	},
	{
		Name:        "wrvf",
		Description: "Write Read Verify feature",
		Enabled:     false,
	},
	{
		Name:        "wue",
		Description: "WRITE UNCORRECTABLE EXT",
		Enabled:     true,
	},
	{
		Name:        "rdegs",
		Description: "READWRITE DMA EXT GPL s",
		Enabled:     true,
	},
	{
		Name:        "sdm",
		Description: "Segmented DOWNLOAD MICROCODE",
		Enabled:     true,
	},
	{
		Name:        "gss1",
		Description: "Gen1 signaling speed 1.5Gb/s",
		Enabled:     true,
	},
	{
		Name:        "gss3",
		Description: "Gen2 signaling speed 3.0Gb/s",
		Enabled:     true,
	},
	{
		Name:        "gss6",
		Description: "Gen3 signaling speed 6.0Gb/s",
		Enabled:     true,
	},
	{
		Name:        "ncqn",
		Description: "Native Command Queueing NCQ",
		Enabled:     true,
	},
	{
		Name:        "pec",
		Description: "Phy event counters",
		Enabled:     true,
	},
	{
		Name:        "rldeetrle",
		Description: "READ LOG DMA EXT equivalent to READ LOG EXT",
		Enabled:     true,
	},
	{
		Name:        "dsaao",
		Description: "DMA Setup Auto Activate optimization",
		Enabled:     true,
	},
	{
		Name:        "diipm",
		Description: "Device initiated interface power management",
		Enabled:     true,
	},
	{
		Name:        "anemc",
		Description: "Asynchronous notification eg. media change",
		Enabled:     true,
	},
	{
		Name:        "stp",
		Description: "Software tings preservation",
		Enabled:     true,
	},
	{
		Name:        "dsd",
		Description: "Device Sleep DEVSLP",
		Enabled:     false,
	},
	{
		Name:        "sctsf",
		Description: "SMART Command Transport SCT feature",
		Enabled:     true,
	},
	{
		Name:        "swsa",
		Description: "SCT Write Same AC2",
		Enabled:     true,
	},
	{
		Name:        "serca",
		Description: "SCT Error Recovery Control AC3",
		Enabled:     true,
	},
	{
		Name:        "sfca",
		Description: "SCT Features Control AC4",
		Enabled:     true,
	},
	{
		Name:        "sdta",
		Description: "SCT Data Tables AC5",
		Enabled:     true,
	},
	{
		Name:        "deaud",
		Description: "Device encrypts all user data",
		Enabled:     true,
	},
	{
		Name:        "dmd",
		Description: "DOWNLOAD MICROCODE DMA",
		Enabled:     true,
	},
	{
		Name:        "smsds",
		Description: "SET MAX SETPASSWORD/UNLOCK DMA s",
		Enabled:     true,
	},
	{
		Name:        "wbd",
		Description: "WRITE BUFFER DMA",
		Enabled:     true,
	},
	{
		Name:        "rbd",
		Description: "READ BUFFER DMA",
		Enabled:     true,
	},
	{
		Name:        "dsmtsl8b",
		Description: "Data Set Management TRIM supported limit 8 blocks",
		Enabled:     true,
	},
	{
		Name:        "pns",
		Description: "password not set",
		Enabled:     true,
	},
	{
		Name:        "es",
		Description: "encryption supported",
		Enabled:     true,
	},
	{
		Name:        "ena",
		Description: "encryption not active",
		Enabled:     true,
	},
	{
		Name:        "dnl",
		Description: "device is not locked",
		Enabled:     true,
	},
	{
		Name:        "dnf",
		Description: "device is not frozen",
		Enabled:     true,
	},
	{
		Name:        "ene",
		Description: "encryption not expired",
		Enabled:     true,
	},
	{
		Name:        "esee",
		Description: "encryption supports enhanced erase",
		Enabled:     true,
	},
	{
		Name:        "time2:8",
		Description: "erase time: 2m, 8m (enhanced)",
		Enabled:     false,
	},
}
