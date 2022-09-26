package dell

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/actions"
	dellFixtures "github.com/metal-toolbox/ironlib/fixtures/dell"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
)

var (
	r6515fixtures = "../../fixtures/dell/r6515"
)

func newFakeDellDevice() (*dell, error) {
	device := common.NewDevice()
	device.Oem = true

	// set device
	device.Model = "r6515"
	device.Vendor = "dell"

	// lshw
	lshwb, err := os.ReadFile(r6515fixtures + "/lshw.json")
	if err != nil {
		return nil, err
	}

	lshw := utils.NewFakeLshw(bytes.NewReader(lshwb))

	// smartctl
	smartctl := utils.NewFakeSmartctl(r6515fixtures + "/smartctl")

	collectors := &actions.Collectors{
		Inventory: lshw,
		Drives:    []actions.DriveCollector{smartctl},
	}

	hardware := model.NewHardware(&device)
	hardware.OemComponents = &model.OemComponents{Dell: []*model.Component{}}

	return &dell{
		hw:         hardware,
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
	expectedOemComponents := dellFixtures.R6515_oem_components

	dell, err := newFakeDellDevice()
	if err != nil {
		t.Error(err)
	}

	// dsu
	b, err := os.ReadFile(r6515fixtures + "/dsu_inventory")
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

	device, err := dell.GetInventory(context.TODO(), false)
	if err != nil {
		t.Error(err)
	}

	err = dell.GetInventoryOEM(context.TODO(), device, &model.UpdateOptions{})
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, dellFixtures.R6515_inventory_lshw_smartctl, device)
	assert.DeepEqual(t, expectedOemComponents, dell.hw.OemComponents)
}

// Get inventory, not listing updates available
func TestListUpdates(t *testing.T) {
	dell, err := newFakeDellDevice()
	if err != nil {
		t.Error(err)
	}

	// dsu
	b, err := os.ReadFile(r6515fixtures + "/dsu_preview")
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

	device, err := dell.ListAvailableUpdates(context.TODO(), &model.UpdateOptions{})
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, dellFixtures.R6515_updatePreview, device)
}
