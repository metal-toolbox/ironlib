package dell

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/packethost/ironlib/actions"
	dellFixtures "github.com/packethost/ironlib/fixtures/dell"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

var (
	r6515fixtures = "../../fixtures/dell/r6515"
)

func newFakeDellDevice() (*dell, error) {
	device := model.NewDevice()
	device.Oem = true
	device.OemComponents = &model.OemComponents{Dell: []*model.Component{}}

	// set device
	device.Model = "r6515"
	device.Vendor = "dell"

	// lshw
	lshwb, err := ioutil.ReadFile(r6515fixtures + "/lshw.json")
	if err != nil {
		return nil, err
	}

	lshw := utils.NewFakeLshw(bytes.NewReader(lshwb))

	// smartctl
	smartctl := utils.NewFakeSmartctl(r6515fixtures + "/smartctl")

	collectors := &actions.Collectors{
		Inventory: lshw,
		Drives:    smartctl,
	}

	return &dell{
		hw:         model.NewHardware(device),
		dnf:        utils.NewFakeDnf(),
		logger:     logrus.New(),
		collectors: collectors,
	}, nil
}

// Get inventory, not listing updates available
func TestGetInventory(t *testing.T) {
	expected := dellFixtures.R6515_inventory_lshw_smartctl
	// patch oemcomponent data for expected results
	expected.Oem = true
	expected.OemComponents = dellFixtures.R6515_oem_components

	dell, err := newFakeDellDevice()
	if err != nil {
		t.Error(err)
	}

	// dsu
	b, err := ioutil.ReadFile(r6515fixtures + "/dsu_inventory")
	if err != nil {
		t.Error(err)
	}

	dsu, err := utils.NewFakeDsu(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	dell.dsu = dsu

	// skip "/usr/libexec/instsvcdrv-helper start" from being executed
	os.Setenv("IRONLIB_TEST", "1")

	device, err := dell.GetInventory(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, dellFixtures.R6515_inventory_lshw_smartctl, device)
}

// Get inventory, not listing updates available
func TestListUpdates(t *testing.T) {
	dell, err := newFakeDellDevice()
	if err != nil {
		t.Error(err)
	}

	// dsu
	b, err := ioutil.ReadFile(r6515fixtures + "/dsu_preview")
	if err != nil {
		t.Error(err)
	}

	dsu, err := utils.NewFakeDsu(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	dell.dsu = dsu

	// skip "/usr/libexec/instsvcdrv-helper start" from being executed
	os.Setenv("IRONLIB_TEST", "1")

	device, err := dell.ListUpdatesAvailable(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, dellFixtures.R6515_updatePreview, device)
}
