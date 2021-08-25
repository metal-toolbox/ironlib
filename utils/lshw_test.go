package utils

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/packethost/ironlib/model"

	asrrFixtures "github.com/packethost/ironlib/fixtures/asrr"
	dellFixtures "github.com/packethost/ironlib/fixtures/dell"

	"github.com/stretchr/testify/assert"
)

func Test_lshw_asrr(t *testing.T) {
	b, err := ioutil.ReadFile("../fixtures/asrr/e3c246d4i-nl/lshw.json")
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
	b, err := ioutil.ReadFile("../fixtures/dell/r6515/lshw.json")
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
		input    string
		output   string
		testName string
	}

	tests := []*tester{
		{
			"",
			"",
			"empty fw string returns empty version",
		},
		{
			"1.0.1",
			"1.0.1",
			"fw string with no parts is returns as is",
		},
		{
			"7.10 0x800075df 19.5.12",
			"19.5.12",
			"intel fw string returns version",
		},
		{
			"14.27.1016 (MT_2420110034)",
			"14.27.1016",
			"mlx fw string returns version",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.output, lshwNicFwStringParse(tt.input), tt.testName)
	}
}
