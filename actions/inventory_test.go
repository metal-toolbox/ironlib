package actions

import (
	"bytes"
	"context"
	"os"
	"testing"

	"gotest.tools/assert"

	"github.com/bmc-toolbox/common"
	dellFixtures "github.com/metal-toolbox/ironlib/fixtures/dell"
	smcFixtures "github.com/metal-toolbox/ironlib/fixtures/supermicro"
	"github.com/metal-toolbox/ironlib/utils"
)

func Test_Inventory_dell(t *testing.T) {
	device := common.NewDevice()

	// set device
	device.Model = "r6515"
	device.Vendor = "dell"

	lshwb, err := os.ReadFile("../fixtures/dell/r6515/lshw.json")
	if err != nil {
		t.Error(err)
	}

	lshw := utils.NewFakeLshw(bytes.NewReader(lshwb))
	smartctl := utils.NewFakeSmartctl("../fixtures/dell/r6515/smartctl")
	collectors := &Collectors{
		Inventory: lshw,
		Drives:    []DriveCollector{smartctl},
	}

	err = Collect(context.TODO(), &device, collectors, true, false)
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, dellFixtures.R6515_inventory_lshw_smartctl, &device)
}

func Test_Inventory_smc(t *testing.T) {
	device := common.NewDevice()
	// set device
	device.Model = "x11dph-t"
	device.Vendor = "supermicro"

	// setup fake collectors with fixture data
	fixturesDir := "../fixtures/supermicro/x11dph-t"
	// lshw
	lshwb, err := os.ReadFile(fixturesDir + "/lshw.json")
	if err != nil {
		t.Error(err)
	}

	lshw := utils.NewFakeLshw(bytes.NewReader(lshwb))

	// smartctl
	smartctl := utils.NewFakeSmartctl(fixturesDir + "/smartctl")

	// mlxup
	mlxupb, err := os.ReadFile(fixturesDir + "/mlxup")
	if err != nil {
		t.Error(err)
	}

	// register mlx NIC collector
	mlxup, err := utils.NewFakeMlxup(bytes.NewReader(mlxupb))
	if err != nil {
		t.Error(err)
	}

	// storecli
	storeclib, err := os.ReadFile(fixturesDir + "/storecli.json")
	if err != nil {
		t.Error(err)
	}

	storecli, err := utils.NewFakeStoreCLI(bytes.NewReader(storeclib))
	if err != nil {
		t.Error(err)
	}

	// ipmicfg
	ipmicfgb, err := os.ReadFile(fixturesDir + "/ipmicfg-summary")
	if err != nil {
		t.Error(err)
	}

	ipmicfg0 := utils.NewFakeIpmicfg(bytes.NewReader(ipmicfgb))
	ipmicfg1 := utils.NewFakeIpmicfg(bytes.NewReader(ipmicfgb))
	ipmicfg2 := utils.NewFakeIpmicfg(bytes.NewReader(ipmicfgb))

	// tpms
	dmi, err := utils.InitFakeDmidecode("../fixtures/supermicro/x11dph-t/dmidecode/tpm")
	if err != nil {
		t.Error(err)
	}

	collectors := &Collectors{
		Inventory:          lshw,
		Drives:             []DriveCollector{smartctl},
		NICs:               mlxup,
		CPLDs:              ipmicfg0,
		BIOS:               ipmicfg1,
		BMC:                ipmicfg2,
		TPMs:               dmi,
		StorageControllers: storecli,
	}

	err = Collect(context.TODO(), &device, collectors, true, false)
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, smcFixtures.Testdata_X11DPH_T_Inventory, &device)
}
