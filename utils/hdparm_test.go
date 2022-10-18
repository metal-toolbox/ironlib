package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newFakeHdparm() *Hdparm {
	return &Hdparm{
		Executor: NewFakeExecutor("hdparm"),
	}
}

func Test_ParseHdparmFeatures(t *testing.T) {
	h := newFakeHdparm()

	device := "/dev/sda"

	features, err := h.ParseHdparmFeatures(device)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fixtureHdparmDevicefeatures, features)
}

var (
	fixtureHdparmDevicefeatures = []hdparmDeviceFeatures{
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
	}
)
