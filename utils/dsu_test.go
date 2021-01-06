package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_dsuParseInventoryBytes(t *testing.T) {

	d := NewFakeDsu()
	d.Executor.SetArgs([]string{"--import-public-key", "--inventory"})
	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Errorf(err.Error())
	}

	components := dsuParseInventoryBytes(result.Stdout)
	assert.Equal(t, 18, len(components))
	assert.Equal(t, "BOSS", components[1].Name)
	assert.Equal(t, "Boss Adapter", components[1].Slug)

}

func Test_dsuParsePreviewBytes(t *testing.T) {

	d := NewFakeDsu()
	d.Executor.SetArgs([]string{"--import-public-key", "--preview"})
	result, err := d.Executor.ExecWithContext(context.Background())
	if err != nil {
		t.Errorf(err.Error())
	}

	components := dsuParsePreviewBytes(result.Stdout)
	assert.Equal(t, 5, len(components))
	assert.Equal(t, "Dell HBA330 Mini Controller 0 Firmware", components[0].Name)
	assert.Equal(t, "SAS HBA330 Controller", components[0].Slug)
	assert.Equal(t, "16.17.01.00", components[0].FirmwareAvailable)
	assert.Equal(t, "SAS-Non-RAID_Firmware_124X2_LN_16.17.01.00_A08", components[0].Metadata["firmware_available_filename"])

}
