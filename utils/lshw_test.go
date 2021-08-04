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
