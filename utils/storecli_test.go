package utils

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func Test_StoreCLIDeviceAttributes(t *testing.T) {
	expected := []*model.StorageController{
		{Serial: "500304801c71e8d0", Vendor: "lsi", Model: "LSI3008-IT", Description: "LSI3008-IT", Metadata: map[string]string{"drives_attached": "12"}, Firmware: &model.Firmware{Installed: "16.00.01.00", Metadata: map[string]string{"bios_version": "08.37.00.00_18.00.00.00"}}},
	}

	b, err := os.ReadFile("../fixtures/utils/storecli/show.json")
	if err != nil {
		t.Error(err)
	}

	cli, err := NewFakeStoreCLI(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	inventory, err := cli.StorageControllers(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, inventory)
}

func Test_StoreCLIDeviceAttributesNoControllers(t *testing.T) {
	b, err := os.ReadFile("../fixtures/utils/storecli/show_nocontrollers.json")
	if err != nil {
		t.Error(err)
	}

	cli, err := NewFakeStoreCLI(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	inventory, err := cli.StorageControllers(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 0, len(inventory))
}
