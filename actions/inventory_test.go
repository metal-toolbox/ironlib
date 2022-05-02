package actions

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	dellFixtures "github.com/packethost/ironlib/fixtures/dell"
	smcFixtures "github.com/packethost/ironlib/fixtures/supermicro"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"gotest.tools/assert"
)

func Test_Inventory_dell(t *testing.T) {
	device := model.NewDevice()

	// set device
	device.Model = "r6515"
	device.Vendor = "dell"

	lshwb, err := ioutil.ReadFile("../fixtures/dell/r6515/lshw.json")
	if err != nil {
		t.Error(err)
	}

	lshw := utils.NewFakeLshw(bytes.NewReader(lshwb))
	smartctl := utils.NewFakeSmartctl("../fixtures/dell/r6515/smartctl")

	collectors := &Collectors{
		Inventory: lshw,
		Drives:    smartctl,
	}

	err = Collect(context.TODO(), device, collectors, true)
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, dellFixtures.R6515_inventory_lshw_smartctl, device)
}

func Test_Inventory_smc(t *testing.T) {
	device := model.NewDevice()
	// set device
	device.Model = "x11dph-t"
	device.Vendor = "supermicro"

	// setup fake collectors with fixture data
	fixturesDir := "../fixtures/supermicro/x11dph-t"
	// lshw
	lshwb, err := ioutil.ReadFile(fixturesDir + "/lshw.json")
	if err != nil {
		t.Error(err)
	}

	lshw := utils.NewFakeLshw(bytes.NewReader(lshwb))

	// smartctl
	smartctl := utils.NewFakeSmartctl(fixturesDir + "/smartctl")

	// mlxup
	mlxupb, err := ioutil.ReadFile(fixturesDir + "/mlxup")
	if err != nil {
		t.Error(err)
	}

	// register mlx NIC collector
	mlxup, err := utils.NewFakeMlxup(bytes.NewReader(mlxupb))
	if err != nil {
		t.Error(err)
	}

	// storecli
	storeclib, err := ioutil.ReadFile(fixturesDir + "/storecli.json")
	if err != nil {
		t.Error(err)
	}

	storecli, err := utils.NewFakeStoreCLI(bytes.NewReader(storeclib))
	if err != nil {
		t.Error(err)
	}

	// ipmicfg
	ipmicfgb, err := ioutil.ReadFile(fixturesDir + "/ipmicfg-summary")
	if err != nil {
		t.Error(err)
	}

	ipmicfg0 := utils.NewFakeIpmicfg(bytes.NewReader(ipmicfgb))
	ipmicfg1 := utils.NewFakeIpmicfg(bytes.NewReader(ipmicfgb))
	ipmicfg2 := utils.NewFakeIpmicfg(bytes.NewReader(ipmicfgb))

	collectors := &Collectors{
		Inventory:          lshw,
		Drives:             smartctl,
		NICs:               mlxup,
		CPLD:               ipmicfg0,
		BIOS:               ipmicfg1,
		BMC:                ipmicfg2,
		StorageControllers: storecli,
	}

	err = Collect(context.TODO(), device, collectors, true)
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, smcFixtures.Testdata_X11DPH_T_Inventory, device)
}
