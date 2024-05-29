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
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var r6515fixtures = "../../fixtures/dell/r6515"

func newFakeDellDevice(logger *logrus.Logger) *dell {
	device := common.NewDevice()
	device.Oem = true

	// set device
	device.Model = "r6515"
	device.Vendor = "dell"

	hardware := model.NewHardware(&device)
	hardware.OEMComponents = []*model.Component{}

	return &dell{
		hw:     hardware,
		dnf:    utils.NewFakeDnf(),
		logger: logger,
	}
}

// Get inventory, not listing updates available
func TestGetInventory(t *testing.T) {
	expected := dellFixtures.R6515_inventory_lshw_smartctl
	// patch oemcomponent data for expected results
	expected.Oem = true
	expectedOemComponents := dellFixtures.R6515_oem_components

	logger, hook := test.NewNullLogger()
	defer hook.Reset()
	dell := newFakeDellDevice(logger)

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

	// setup fake collectors
	// lshw
	lshwb, err := os.ReadFile(r6515fixtures + "/lshw.json")
	if err != nil {
		t.Error(err)
	}

	lshw := utils.NewFakeLshw(bytes.NewReader(lshwb))

	// smartctl
	smartctl := utils.NewFakeSmartctl(r6515fixtures + "/smartctl")
	lsblk := utils.NewFakeLsblk()

	hdparm := utils.NewFakeHdparm()
	nvme := utils.NewFakeNvme()

	collectors := &actions.Collectors{
		InventoryCollector:          lshw,
		DriveCollectors:             []actions.DriveCollector{smartctl, lsblk},
		DriveCapabilitiesCollectors: []actions.DriveCapabilityCollector{hdparm, nvme},
	}

	device, err := dell.GetInventory(context.TODO(), actions.WithCollectors(collectors))
	if err != nil {
		t.Error(err)
	}

	err = dell.GetInventoryOEM(context.TODO(), device, &model.UpdateOptions{})
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, dellFixtures.R6515_inventory_lshw_smartctl, device)
	assert.Equal(t, expectedOemComponents, dell.hw.OEMComponents)
}

// Get inventory, not listing updates available
func TestListUpdates(t *testing.T) {
	logger, hook := test.NewNullLogger()
	defer hook.Reset()
	dell := newFakeDellDevice(logger)

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

	assert.Equal(t, dellFixtures.R6515_updatePreview, device)
}
