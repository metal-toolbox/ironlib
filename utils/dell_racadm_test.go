package utils

import (
	"context"
	"testing"

	"github.com/packethost/ironlib/config"
	"github.com/stretchr/testify/assert"
)

// Fake Dell Racadm executor for tests
func NewFakeDellRacadm() *DellRacadm {
	return &DellRacadm{
		Executor: NewFakeExecutor("racadm"),
	}
}

func Test_GetBIOSConfiguration(t *testing.T) {
	expected := &config.BIOSConfiguration{
		Dell: &config.DellBIOS{
			AMDSev:         1,
			BootMode:       "Bios",
			Hyperthreading: "Enabled",
			SRIOV:          "Enabled",
			TPM:            "On",
		},
	}
	d := NewFakeDellRacadm()
	cfg, err := d.GetBIOSConfiguration(context.TODO())
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, cfg)
}

func Test_parseRacadmBIOSConfig(t *testing.T) {
	expected := &config.DellBIOS{
		AMDSev:         1,
		BootMode:       "Bios",
		Hyperthreading: "Enabled",
		SRIOV:          "Enabled",
		TPM:            "On",
	}
	d := NewFakeDellRacadm()
	_, err := d.Executor.ExecWithContext(context.TODO())
	if err != nil {
		return
	}

	config, err := d.parseRacadmBIOSConfig(context.TODO())
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, expected, config)
}
