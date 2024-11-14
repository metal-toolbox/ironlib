package utils

import (
	"os"
	"strings"
	"testing"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/stretchr/testify/assert"
)

func Test_CopyFile(t *testing.T) {
	srcFile := "/tmp/foobar"
	dstFile := "/tmp/barfoo"
	content := []byte(`meh`)

	f, err := os.Create(srcFile)
	if err != nil {
		t.Error(err)
	}

	_, err = f.Write(content)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(srcFile)

	err = copyFile(srcFile, dstFile)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(dstFile)

	stat, err := os.Stat(dstFile)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "-rw-------", stat.Mode().String())

	b, err := os.ReadFile(dstFile)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, content, b)
}

func Test_IdentifyVendorModel(t *testing.T) {
	dmi, err := InitFakeDmidecode("../fixtures/asrr/e3c246d4i-nl/dmidecode-non-packet")
	if err != nil {
		t.Error(err)
	}

	device, err := IdentifyVendorModel(dmi)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, common.VendorAsrockrack, strings.ToLower(device.Vendor))
	assert.Equal(t, "E3C246D4I-NL", device.Model)
}
