package utils

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func Test_dsuComponentNameToSlug(t *testing.T) {
	kv := map[string]string{
		"BIOS":                    common.SlugBIOS,
		"Power Supply":            common.SlugPSU,
		"Disk 0 of BOSS Adapter ": model.SlugDellBossAdapterDisk0,
		"BOSS":                    model.SlugDellBossAdapter,
		"Dell HBA330 Mini Controller 0 Firmware ":                    common.SlugStorageController,
		"Backplane Expander FW ":                                     model.SlugDellBackplaneExpander,
		"Intel(R) Ethernet 10G 4P X710 SFP+ rNDC":                    common.SlugNIC,
		"Intel(R) Ethernet 10G X710 rNDC ":                           common.SlugNIC,
		"Intel(R) Ethernet 10G X710 rNDC":                            common.SlugNIC,
		"iDRAC":                                                      common.SlugBMC,
		"NVMePCISSD Model Number: Micron_9200_MTFDHAL3T8TCT":         common.SlugDrive,
		"Lifecycle Controller":                                       model.SlugDellLifeCycleController,
		"System CPLD":                                                model.SlugDellSystemCPLD,
		"Dell EMC iDRAC Service Module Embedded Package v3.5.0, A00": model.SlugDellIdracServiceModule,
	}

	for componentName, expectedSlug := range kv {
		slug := dsuComponentNameToSlug(componentName)
		assert.Equal(t, expectedSlug, slug, "component Name: "+componentName)
	}
}

func Test_dsuParseInventoryBytes(t *testing.T) {
	b, err := os.ReadFile("../fixtures/utils/dsu/inventory")
	if err != nil {
		t.Error(err)
	}

	d, err := NewFakeDsu(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	d.Executor.SetArgs([]string{"--import-public-key", "--inventory"})

	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Errorf(err.Error())
	}

	components := dsuParseInventoryBytes(result.Stdout)
	assert.Equal(t, 18, len(components))
	assert.Equal(t, "BOSS", components[1].Name)
	assert.Equal(t, "Boss Adapter", components[1].Slug)
}

func Test_dsuParsePreviewBytes(t *testing.T) {
	b, err := os.ReadFile("../fixtures/utils/dsu/preview")
	if err != nil {
		t.Error(err)
	}

	d, err := NewFakeDsu(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Errorf(err.Error())
	}

	components := dsuParsePreviewBytes(result.Stdout)
	assert.Equal(t, 5, len(components))
	assert.Equal(t, "Dell HBA330 Mini Controller 0 Firmware", components[0].Name)
	assert.Equal(t, common.SlugStorageController, components[0].Slug)
	assert.Equal(t, "16.17.01.00", components[0].FirmwareAvailable)
	assert.Equal(t, "SAS-Non-RAID_Firmware_124X2_LN_16.17.01.00_A08", components[0].Metadata["firmware_available_filename"])
}

func Test_findDSUInventoryCollector(t *testing.T) {
	invb := "invcol_5N2WM_LN64_20_09_200_921_A00.BIN"

	tmpDir := t.TempDir()

	dirs := []string{
		tmpDir + "/foo/dsu",
		tmpDir + "/foo/dsu" + "/dellupdates",
	}

	expected := []string{}

	for _, d := range dirs {
		err := os.MkdirAll(d, 0o744)
		if err != nil {
			t.Error(err)
		}

		f := d + "/" + invb
		expected = append(expected, f)

		_, err = os.Create(f)
		if err != nil {
			t.Error(err)
		}
	}

	assert.ElementsMatch(t, expected, findDSUInventoryCollector(dirs[0]))
}
