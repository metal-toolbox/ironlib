package utils

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/packethost/ironlib/model"

	asrrFixtures "github.com/packethost/ironlib/fixtures/asrr"
	dellFixtures "github.com/packethost/ironlib/fixtures/dell"

	"github.com/stretchr/testify/assert"
)

func Test_lshw_asrr(t *testing.T) {
	b, err := os.ReadFile("../fixtures/asrr/e3c246d4i-nl/lshw.json")
	if err != nil {
		t.Error(err)
	}

	l := NewFakeLshw(bytes.NewReader(b))
	device := model.NewDevice()

	err = l.Collect(context.TODO(), device)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, asrrFixtures.E3C246D4INL, device)
}

func Test_lshw_dell(t *testing.T) {
	b, err := os.ReadFile("../fixtures/dell/r6515/lshw.json")
	if err != nil {
		t.Error(err)
	}

	l := NewFakeLshw(bytes.NewReader(b))
	device := model.NewDevice()

	err = l.Collect(context.TODO(), device)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, dellFixtures.R6515_inventory_lshw, device)
}

func Test_lshwNicFwStringParse(t *testing.T) {
	type tester struct {
		vendor   string
		fw       string
		expected string
		testName string
	}

	tests := []*tester{
		{
			"",
			"",
			"",
			"empty fw string returns empty version",
		},
		{
			"foo",
			"1.0.1",
			"1.0.1",
			"fw string from unknown vendor is returned as is",
		},
		{
			"Intel Corporation",
			"7.10 0x800075df 19.5.12",
			"19.5.12",
			"intel fw string returns version",
		},
		{
			"Mellanox Technologies",
			"14.27.1016 (MT_2420110034)",
			"14.27.1016",
			"mlx fw string returns version",
		},
		{
			"Broadcom Inc. and subsidiaries",
			"5719-v1.46 NCSI v1.5.1.0",
			"v1.5.1.0",
			"broadcom fw string returns version",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, lshwNicFwStringParse(tt.fw, tt.vendor), tt.testName)
	}
}
