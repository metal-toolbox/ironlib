package utils

import (
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeStoreCLI(cmd string) *StoreCLI {
	return &StoreCLI{
		Executor: NewFakeExecutor(cmd),
	}
}

func Test_StoreCLIDeviceAttributes(t *testing.T) {

	expected := []*model.Component{
		{Serial: "500304801c71e8d0", Vendor: "LSI", Model: "LSI3008-IT", Name: "Serial Attached SCSI controller", Slug: "[0] Serial Attached SCSI controller", FirmwareInstalled: "16.00.01.00"},
	}

	n := newFakeStoreCLI("storecli")
	inventory, err := n.Components()
	if err != nil {
		t.Error(err)
	}

	// since the component IDs are unique
	inventory = purgeTestComponentID(inventory)
	assert.Equal(t, expected, inventory)
}

func Test_StoreCLIDeviceAttributesNoControllers(t *testing.T) {

	// see fake_executor.go for how this works
	n := newFakeStoreCLI("storecli-nocontrollers")
	inventory, err := n.Components()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 0, len(inventory))
}
