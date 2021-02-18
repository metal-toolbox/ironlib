package asrockrack

import (
	"context"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ASRockRack struct {
	ID                   string
	Vendor               string
	Model                string
	Serial               string
	Updater              utils.Updater
	UpdatesAvailable     int
	PendingReboot        bool // set when the device requires a reboot after update
	UpdatesInstalled     bool // set when updates were installed on the device
	Components           []*model.Component
	Logger               *logrus.Logger
	Dmidecode            *utils.Dmidecode
	FirmwareUpdateConfig *model.FirmwareUpdateConfig
}

func New(vendor, model string, l *logrus.Logger) (model.Manager, error) {

	dmidecode, err := utils.NewDmidecode()
	if err != nil {
		errors.Wrap(err, "erorr in dmidecode init")
	}

	uid, _ := uuid.NewRandom()
	return &ASRockRack{
		ID:        uid.String(),
		Vendor:    vendor,
		Model:     model,
		Dmidecode: dmidecode,
		Logger:    l,
	}, nil
}

// Returns hardware inventory for the device
func (a *ASRockRack) GetInventory(ctx context.Context, listUpdates bool) (*model.Device, error) {

	var trace bool
	if a.Logger.GetLevel().String() == "trace" {
		trace = true
	}

	// Collect device inventory from lshw
	lshw := utils.NewLshwCmd(trace)
	device, err := lshw.Inventory()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving device inventory")
	}

	// Collect additional drives data from smartctl
	smartctl := utils.NewSmartctlCmd(trace)
	drives, err := smartctl.Components()
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving device inventory")
	}

	device = utils.UpdateComponentData(device, drives)

	return device, nil
}

func (a *ASRockRack) GetModel() string {
	return a.Model
}

func (a *ASRockRack) GetVendor() string {
	return a.Vendor
}

func (a *ASRockRack) GetDeviceID() string {
	return a.ID
}

func (a *ASRockRack) SetDeviceID(id string) {
	a.ID = id
}

func (a *ASRockRack) RebootRequired() bool {
	return a.PendingReboot
}

func (a *ASRockRack) SetFirmwareUpdateConfig(config *model.FirmwareUpdateConfig) {
	a.FirmwareUpdateConfig = config
}

func (a *ASRockRack) SetOptions(options map[string]interface{}) error {
	return nil
}

func (a *ASRockRack) UpdatesApplied() bool {
	return a.UpdatesInstalled
}

func (a *ASRockRack) ApplyUpdatesAvailable(ctx context.Context, config *model.FirmwareUpdateConfig, dryRun bool) (err error) {
	return nil
}

func (a *ASRockRack) GetDeviceFirmwareRevision(ctx context.Context) (string, error) {
	return "", nil
}

func (a *ASRockRack) GetUpdatesAvailable(ctx context.Context) (*model.Device, error) {
	return &model.Device{}, nil
}
