package utils

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/packethost/ironlib/config"
	"github.com/stretchr/testify/assert"
)

func Test_parseSMCBIOSConfig_X11SCHFF(t *testing.T) {
	expected := &config.SupermicroBIOS{
		BootMode:       "LEGACY",
		Hyperthreading: "Enabled",
		TPM:            "Enable",
		SecureBoot:     "Disabled",
		IntelSGX:       "Software Controlled",
	}

	b, err := ioutil.ReadFile("test_data/bios_configs/config_X11SCHF-F.xml")
	if err != nil {
		t.Error(err)
	}

	s := NewFakeSMCSum(bytes.NewReader(b))
	c, err := s.parseBIOSConfig(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, expected, c)
}

func Test_parseSMCBIOSConfig_X11DPHT(t *testing.T) {
	expected := &config.SupermicroBIOS{
		BootMode:       "DUAL",
		Hyperthreading: "Enable",
		TPM:            "Enable",
		SecureBoot:     "Disabled",
	}

	b, err := ioutil.ReadFile("test_data/bios_configs/config_X11DPH-T.xml")
	if err != nil {
		t.Error(err)
	}

	s := NewFakeSMCSum(bytes.NewReader(b))
	c, err := s.parseBIOSConfig(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, expected, c)
}
