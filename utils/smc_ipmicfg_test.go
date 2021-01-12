package utils

import (
	"context"
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func Test_IpmicfgDeviceAttributes(t *testing.T) {

	expected := []*model.Component{
		{Vendor: "Supermicro", Model: "Supermicro", Name: "CPLD", Slug: "CPLD", FirmwareInstalled: "02.b6.04"},
		{Vendor: "Supermicro", Model: "Supermicro", Name: "BIOS", Slug: "BIOS", FirmwareInstalled: "3.3"},
		{Vendor: "Supermicro", Model: "Supermicro", Name: "BMC", Slug: "BMC", FirmwareInstalled: "1.71.11"},
	}
	i := NewFakeIpmicfg()
	inventory, err := i.Components()
	if err != nil {
		t.Error(err)
	}
	// since the component IDs are unique
	inventory = purgeTestComponentID(inventory)
	assert.Equal(t, expected, inventory)
}

func Test_IpmicfgParseSummaryOutput(t *testing.T) {
	expected := &IpmicfgSummary{FirmwareRevision: "1.71.11", BIOSVersion: "3.3", CPLDVersion: "02.b6.04"}

	i := NewFakeIpmicfg()
	i.Executor.SetArgs([]string{"-summary"})
	result, err := i.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Error(err)
	}

	summary := i.parseQueryOutput(result.Stdout)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, summary)
}
