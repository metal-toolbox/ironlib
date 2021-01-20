package dell

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newFakeDellDevice() *Dell {

	uid, _ := uuid.NewRandom()
	return &Dell{
		ID:        uid.String(),
		Vendor:    "dell",
		Model:     "R640",
		Dmidecode: utils.NewFakeDmidecode(),
		Dnf:       utils.NewFakeDnf(),
		Dsu:       utils.NewFakeDsu(),
		Logger:    logrus.New(),
	}
}

func TestGetInventory(t *testing.T) {

	dell := newFakeDellDevice()

	// skip "/usr/libexec/instsvcdrv-helper start" from being executed
	os.Setenv("IRONLIB_TEST", "1")

	device, err := dell.GetInventory(context.TODO(), true)
	if err != nil {
		t.Error(err)
	}

	expectedComponent0 := &model.Component{
		ID:                "b2b52416-3479-4a7a-8e73-25b43c69199b",
		DeviceID:          "66e73a31-6abd-4fb3-9f66-d98e83fb785f",
		Serial:            "",
		Vendor:            "",
		Type:              "",
		Model:             "",
		Name:              "BIOS",
		Slug:              "BIOS",
		FirmwareInstalled: "2.6.4",
		FirmwareAvailable: "2.8.1",
		Metadata:          map[string]string{"firmware_available_filename": "BIOS_RTWM9_LN_2.8.1"},
		Oem:               true,
		FirmwareManaged:   true,
	}

	expectedComponentUpdate0 := &model.Component{
		ID:                "d31b38cc-6640-4949-92ba-43367b5d93ef",
		DeviceID:          "",
		Serial:            "",
		Vendor:            "",
		Type:              "",
		Model:             "",
		Name:              "Dell HBA330 Mini Controller 0 Firmware",
		Slug:              "SAS HBA330 Controller",
		FirmwareInstalled: "",
		FirmwareAvailable: "16.17.01.00",
		Metadata:          map[string]string{"firmware_available_filename": "SAS-Non-RAID_Firmware_124X2_LN_16.17.01.00_A08"},
		Oem:               true,
		FirmwareManaged:   true,
	}

	assert.Equal(t, "dell", device.Vendor)
	assert.Equal(t, "R640", device.Model)

	// expect 18 components
	assert.Equal(t, len(device.Components), 18)
	// set uuids to match
	expectedComponent0.ID = device.Components[0].ID
	expectedComponent0.DeviceID = device.Components[0].DeviceID
	assert.Equal(t, expectedComponent0, device.Components[0])

	// expect 5 component updates
	assert.Equal(t, 5, len(device.ComponentUpdates))
	// set uuids to match
	expectedComponentUpdate0.ID = device.ComponentUpdates[0].ID
	assert.Equal(t, expectedComponentUpdate0, device.ComponentUpdates[0])

	//spewC := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true, DisableMethods: true}
	//spewC.Dump(device.ComponentUpdates)

}
