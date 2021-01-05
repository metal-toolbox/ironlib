package utils

import (
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/stretchr/testify/assert"
)

func newFakeNvme() *Nvme {
	return &Nvme{
		Executor: NewFakeExecutor("nvme"),
	}
}

func Test_NvmeComponents(t *testing.T) {

	expected := []*model.Component{
		{Serial: "Z9DF70I8FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", FirmwareInstalled: "AGGA4104", Slug: "NVME drive", Name: "NVME drive"},
		{Serial: "Z9DF70I9FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", FirmwareInstalled: "AGGA4104", Slug: "NVME drive", Name: "NVME drive"},
	}

	n := newFakeNvme()
	components, err := n.Components()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, components)
}
