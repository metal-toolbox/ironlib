package utils

import (
	"context"
	"os"
	"testing"

	"github.com/packethost/ironlib/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func newFakeMsecli() *Msecli {
	return &Msecli{
		Executor: NewFakeExecutor("msecli"),
	}
}

func Test_MsecliComponents(t *testing.T) {
	expected := []*model.Component{
		{
			Serial:            "193423710BDA",
			Vendor:            "Micron",
			Type:              model.SlugDriveTypeSATASSD,
			Model:             "Micron_5200_MTFDDAK480TDN",
			Name:              "Micron_5200_MTFDDAK480TDN",
			Slug:              model.SlugDrive,
			FirmwareInstalled: "D1MU020",
			FirmwareManaged:   true,
			Metadata:          map[string]string{},
		},
		{
			Serial:            "193423711167",
			Vendor:            "Micron",
			Type:              model.SlugDriveTypeSATASSD,
			Model:             "Micron_5200_MTFDDAK480TDN",
			Name:              "Micron_5200_MTFDDAK480TDN",
			Slug:              model.SlugDrive,
			FirmwareInstalled: "D1MU020",
			FirmwareManaged:   true,
			Metadata:          map[string]string{},
		},
	}

	m := newFakeMsecli()

	inventory, err := m.Components()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, inventory)
}

func Test_parseMsecliQueryOutput(t *testing.T) {
	expected := []*MsecliDevice{
		{
			ModelNumber:      "Micron_5200_MTFDDAK480TDN",
			SerialNumber:     "193423710BDA",
			FirmwareRevision: "D1MU020",
		},
		{
			ModelNumber:      "Micron_5200_MTFDDAK480TDN",
			SerialNumber:     "193423711167",
			FirmwareRevision: "D1MU020",
		},
	}

	m := newFakeMsecli()
	m.Executor.SetArgs([]string{"-L"})

	result, err := m.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Error(err)
	}

	devices := m.parseMsecliQueryOutput(result.Stdout)

	assert.Equal(t, expected, devices)
}

func Test_parseMsecliQueryOutputCmdFailure(t *testing.T) {
	os.Setenv("FAIL_MICRON_UPDATE", "1")

	m := newFakeMsecli()
	m.Executor.SetArgs([]string{"-L"})

	result, err := m.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Error(err)
	}

	devices := m.parseMsecliQueryOutput(result.Stdout)
	assert.Equal(t, 0, len(devices))
}

func Test_QueryOutputEmpty(t *testing.T) {
	os.Setenv("FAIL_MICRON_QUERY", "1")

	m := newFakeMsecli()

	_, err := m.Query()
	assert.Equal(t, ErrNoCommandOutput, errors.Cause(err))
}
