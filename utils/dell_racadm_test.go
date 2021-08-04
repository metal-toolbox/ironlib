package utils

import (
	"context"
	"testing"

	"github.com/packethost/ironlib/config"
	"github.com/stretchr/testify/assert"
)

const (
	biosConfigR6515 = "../fixtures/dell/r6515/bios.json"
	biosConfigC6320 = "../fixtures/dell/c6320/bios.xml"
)

// Fake Dell Racadm executor for tests
func NewFakeDellRacadm() *DellRacadm {
	return &DellRacadm{
		Executor:       NewFakeExecutor("racadm"),
		KeepConfigFile: true,
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
	d.BIOSCfgTmpFile = biosConfigR6515

	cfg, err := d.GetBIOSConfiguration(context.TODO(), "")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, cfg)
}

func Test_RacadmBIOSConfigJSON(t *testing.T) {
	expected := &config.DellBIOS{
		AMDSev:         1,
		BootMode:       "Bios",
		Hyperthreading: "Enabled",
		SRIOV:          "Enabled",
		TPM:            "On",
	}

	// setup fake racadm, pass the bios config file
	r := NewFakeRacadm(biosConfigR6515)

	c, err := r.racadmBIOSConfigJSON(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, c)
}

func Test_RacadmBIOSConfigXML(t *testing.T) {
	expected := &config.DellBIOS{
		BootMode:       "Bios",
		Hyperthreading: "Enabled",
		SRIOV:          "Disabled",
		TPM:            "On",
	}

	// setup fake racadm, pass in the read bios config
	r := NewFakeRacadm(biosConfigC6320)

	c, err := r.racadmBIOSConfigXML(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, c)
}
