package utils

import (
	"bytes"
	"context"
	"os"
	"testing"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/stretchr/testify/assert"
)

func initFakeIPMICfg() (*Ipmicfg, error) {
	b, err := os.ReadFile("../fixtures/utils/ipmicfg/summary")
	if err != nil {
		return nil, err
	}

	i := NewFakeIpmicfg(bytes.NewReader(b))
	i.Executor.SetArgs("-summary")

	return i, nil
}

func Test_Ipmicfg_BIOS(t *testing.T) {
	expected := &common.BIOS{Common: common.Common{Description: "BIOS", Vendor: "Supermicro", Firmware: &common.Firmware{Installed: "3.3", Metadata: map[string]string{"build_date": "02/24/2020"}}}}

	i, err := initFakeIPMICfg()
	if err != nil {
		t.Error(err)
	}

	bios, err := i.BIOS(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, bios)
}

func Test_Ipmicfg_BMC(t *testing.T) {
	expected := &common.BMC{Common: common.Common{Description: "BMC", Vendor: "Supermicro", Firmware: &common.Firmware{Installed: "1.71.11", Metadata: map[string]string{"build_date": "10/25/2019"}}}}

	i, err := initFakeIPMICfg()
	if err != nil {
		t.Error(err)
	}

	bmc, err := i.BMC(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, bmc)
}

func Test_Ipmicfg_CPLD(t *testing.T) {
	expected := []*common.CPLD{{Common: common.Common{Description: "CPLD", Vendor: "Supermicro", Firmware: &common.Firmware{Installed: "02.b6.04"}}}}

	i, err := initFakeIPMICfg()
	if err != nil {
		t.Error(err)
	}

	cpld, err := i.CPLDs(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, cpld)
}

func Test_IpmicfgParseSummaryOutput(t *testing.T) {
	expected := &IpmicfgSummary{FirmwareRevision: "1.71.11", FirmwareBuildDate: "10/25/2019", BIOSVersion: "3.3", BIOSBuildDate: "02/24/2020", CPLDVersion: "02.b6.04"}

	i, err := initFakeIPMICfg()
	if err != nil {
		t.Error(err)
	}

	result, err := i.Executor.Exec(context.Background())
	if err != nil {
		t.Error(err)
	}

	summary := i.parseQueryOutput(result.Stdout)

	assert.Equal(t, expected, summary)
}
