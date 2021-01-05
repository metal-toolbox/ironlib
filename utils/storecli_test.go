package utils

import (
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeStoreCLI() *StoreCLI {
	return &StoreCLI{
		Executor: NewFakeExecutor("storecli"),
	}
}

func Test_StoreCLIDeviceAttributes(t *testing.T) {

	expected := []*model.Component{
		{Serial: "500304801c71e8d0", Vendor: "LSI", Model: "LSI3008-IT", Name: "Serial Attached SCSI controller", Slug: "Serial Attached SCSI controller", FirmwareInstalled: "16.00.01.00"},
	}

	n := newFakeStoreCLI()
	inventory, err := n.Components()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, inventory)
}
