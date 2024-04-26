package actions

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/bmc-toolbox/common"
	dellFixtures "github.com/metal-toolbox/ironlib/fixtures/dell"
	smcFixtures "github.com/metal-toolbox/ironlib/fixtures/supermicro"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/assert"
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
	lsblk := utils.NewFakeLsblk()

	hdparm := utils.NewFakeHdparm()
	nvme := utils.NewFakeNvme()

	collectors := &Collectors{
		InventoryCollector:          lshw,
		DriveCollectors:             []DriveCollector{smartctl, lsblk},
		DriveCapabilitiesCollectors: []DriveCapabilityCollector{hdparm, nvme},
	}

	options := []Option{
		WithCollectors(collectors),
		WithTraceLevel(),
		WithFailOnError(),
		WithDisabledCollectorUtilities([]model.CollectorUtility{"dmidecode"}),
	}

	collector := NewInventoryCollectorAction(slogt.New(t), options...)
	if err := collector.Collect(context.TODO(), &device); err != nil {
		t.Error(err)
	}

	assert.Equal(t, dellFixtures.R6515_inventory_lshw_smartctl, &device)
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
		InventoryCollector:          lshw,
		DriveCollectors:             []DriveCollector{smartctl},
		NICCollector:                mlxup,
		CPLDCollector:               ipmicfg0,
		BIOSCollector:               ipmicfg1,
		BMCCollector:                ipmicfg2,
		TPMCollector:                dmi,
		StorageControllerCollectors: []StorageControllerCollector{storecli},
	}

	collector := NewInventoryCollectorAction(slogt.New(t), WithCollectors(collectors), WithTraceLevel())
	if err := collector.Collect(context.TODO(), &device); err != nil {
		t.Error(err)
	}

	assert.Equal(t, smcFixtures.Testdata_X11DPH_T_Inventory, &device)
}

// nolint:gocyclo // Test code isn't pretty
func TestNewInventoryCollectorAction(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		want    *InventoryCollectorAction
	}{
		{
			"trace-enabled",
			[]Option{WithTraceLevel()},
			&InventoryCollectorAction{trace: true},
		},
		{
			"trace-disabled",
			[]Option{},
			&InventoryCollectorAction{},
		},
		{
			"dynamic-collectors",
			[]Option{WithDynamicCollection()},
			&InventoryCollectorAction{dynamicCollection: true},
		},
		{
			"fail-on-error",
			[]Option{WithFailOnError()},
			&InventoryCollectorAction{failOnError: true},
		},
		{
			"collectors-empty",
			[]Option{},
			&InventoryCollectorAction{},
		},
		{
			"default-collectors-set",
			[]Option{},
			&InventoryCollectorAction{},
		},
		{
			"default-lshw-collector",
			[]Option{},
			&InventoryCollectorAction{},
		},
		{
			"default-drive-collectors",
			[]Option{},
			&InventoryCollectorAction{},
		},
		{
			"default-drive-capabilities-collector",
			[]Option{},
			&InventoryCollectorAction{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInventoryCollectorAction(slogt.New(t), tt.options...)

			switch tt.name {
			case "trace-enabled":
				assert.Equal(t, true, got.trace)
			case "trace-disabled":
				assert.Equal(t, false, got.trace)
			case "dynamic-collectors":
				assert.Equal(t, true, got.dynamicCollection)
			case "fail-on-error":
				assert.Equal(t, true, got.failOnError)
			case "collectors-empty":
				assert.Equal(t, true, tt.want.collectors.Empty())
			case "default-collectors-set":
				assert.Equal(t, false, got.collectors.Empty())
			case "default-lshw-collector":
				assert.Equal(t, false, got.collectors.InventoryCollector == nil)
			case "default-drive-collector":
				assert.Equal(t, false, len(got.collectors.DriveCollectors) == 0)
			case "default-drive-capabilities-collector":
				assert.Equal(t, false, len(got.collectors.DriveCapabilitiesCollectors) == 0)
			}
		})
	}
}
