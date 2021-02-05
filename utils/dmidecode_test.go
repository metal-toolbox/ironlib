package utils

import (
	"io/ioutil"
	"testing"

	"github.com/dselans/dmidecode"
	"github.com/stretchr/testify/assert"
)

// Given the test data file returns a Dmidecode with the test dmidecode output loaded
func InitTestDmidecode(testFile string) (*Dmidecode, error) {

	b, err := ioutil.ReadFile(testFile)
	if err != nil {
		return nil, err
	}

	// setup a dmidecode instance
	d := dmidecode.New()
	err = d.ParseDmidecode(string(b))
	if err != nil {
		return nil, err
	}

	// wrap the dmidecode instance in our Dmidecode wrapper
	return &Dmidecode{dmi: d}, nil
}

func Test_dmidecode_asrockrack_E3C246D4I_NL(t *testing.T) {

	dmi, err := InitTestDmidecode("test_data/dmidecode/asrockrack-E3C246D4I-NL")
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
