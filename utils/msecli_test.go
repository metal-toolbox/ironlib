package utils

import (
	"context"
	"os"
	"testing"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func newFakeMsecli() (*Msecli, error) {
	m := &Msecli{Executor: NewFakeExecutor("msecli")}

	b, err := os.ReadFile("../fixtures/utils/msecli/list")
	if err != nil {
		return nil, err
	}

	m.Executor.SetStdout(b)

	return m, nil
}

func Test_MsecliDrives(t *testing.T) {
	expected := []*model.Drive{
		{
			Serial:      "193423710BDA",
			Vendor:      "Micron",
			Type:        model.SlugDriveTypeSATASSD,
			Model:       "Micron_5200_MTFDDAK480TDN",
			Description: "Micron_5200_MTFDDAK480TDN",
			Firmware: &model.Firmware{
				Installed: "D1MU020",
			},
			Metadata: map[string]string{},
		},
		{
			Serial:      "193423711167",
			Vendor:      "Micron",
			Type:        model.SlugDriveTypeSATASSD,
			Model:       "Micron_5200_MTFDDAK480TDN",
			Description: "Micron_5200_MTFDDAK480TDN",
			Firmware: &model.Firmware{
				Installed: "D1MU020",
			},
			Metadata: map[string]string{},
		},
	}

	m, err := newFakeMsecli()
	if err != nil {
		t.Error(err)
	}

	drives, err := m.Drives(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, drives)
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

	m, err := newFakeMsecli()
	if err != nil {
		t.Error(err)
	}

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

	m, err := newFakeMsecli()
	if err != nil {
		t.Error(err)
	}

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

	m, err := newFakeMsecli()
	if err != nil {
		t.Error(err)
	}

	m.Executor.SetArgs([]string{"-L"})

	_, err = m.Query()
	assert.Equal(t, ErrNoCommandOutput, errors.Cause(err))
}
