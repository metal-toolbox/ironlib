package dell

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

func newFakeDellDevice() *Dell {
	device := model.NewDevice()
	device.Oem = true
	device.OemComponents = &model.OemComponents{Dell: []*model.Component{}}

	b, err := ioutil.ReadFile("../../utils/test_data/r6515/lshw.json")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fakeLshw := utils.NewFakeLshw(bytes.NewReader(b))
	fakeSmartctl := utils.NewFakeSmartctl("test_data/r6515")

	return &Dell{
		hw:       model.NewHardware(device),
		dnf:      utils.NewFakeDnf(),
		dsu:      utils.NewFakeDsu(),
		lshw:     fakeLshw,
		smartctl: fakeSmartctl,
		logger:   logrus.New(),
	}
}

// Get inventory, not listing updates available
func TestGetInventory(t *testing.T) {
	dell := newFakeDellDevice()

	// skip "/usr/libexec/instsvcdrv-helper start" from being executed
	os.Setenv("IRONLIB_TEST", "1")

	device, err := dell.GetInventory(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, testdata_R6515Inventory, device)
}

// Get inventory, not listing updates available
func TestListUpdates(t *testing.T) {
	dell := newFakeDellDevice()

	// skip "/usr/libexec/instsvcdrv-helper start" from being executed
	os.Setenv("IRONLIB_TEST", "1")

	device, err := dell.ListUpdatesAvailable(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, testdata_R6515UpdatePreview, device)
}
