package utils

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/assert"
)

func Test_dmidecode_asrockrack_E3C246D4I_NL(t *testing.T) {
	dmi, err := InitFakeDmidecode("../fixtures/asrr/e3c246d4i-nl/dmidecode")
	if err != nil {
		t.Error(err)
	}

	v, err := dmi.Manufacturer()
	if err != nil {
		t.Error(err)
	}

	m, err := dmi.ProductName()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "Packet", v)
	assert.Equal(t, "c3.small.x86", m)

	bv, err := dmi.BaseBoardManufacturer()
	if err != nil {
		t.Error(err)
	}

	bm, err := dmi.BaseBoardProductName()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "ASRockRack", bv)
	assert.Equal(t, "E3C246D4I-NL", bm)
}

func Test_dmidecode_asrockrack_E3C246D4I_NL_TPM(t *testing.T) {
	expected := []*common.TPM{
		{
			Common: common.Common{
				Description: "INFINEON",
				Vendor:      "infineon",
				Firmware: &common.Firmware{
					Installed: "5.63",
				},
				Metadata: map[string]string{
					"Specification Version": "2.0",
				},
			},
		},
	}

	dmi, err := InitFakeDmidecode("../fixtures/asrr/e3c246d4i-nl/dmidecode")
	if err != nil {
		t.Error(err)
	}

	got, err := dmi.TPMs(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, got)
}
