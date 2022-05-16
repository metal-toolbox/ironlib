package utils

import (
	"context"
	"testing"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeNvme() *Nvme {
	return &Nvme{
		Executor: NewFakeExecutor("nvme"),
	}
}

func Test_NvmeComponents(t *testing.T) {
	expected := []*model.Drive{
		{Serial: "Z9DF70I8FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &model.Firmware{Installed: "AGGA4104"}, ProductName: "NULL"},
		{Serial: "Z9DF70I9FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &model.Firmware{Installed: "AGGA4104"}, ProductName: "NULL"},
	}

	n := newFakeNvme()

	drives, err := n.Drives(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, drives)
}
