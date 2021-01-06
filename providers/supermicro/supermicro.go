package supermicro

import (
	"context"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
)

type Supermicro struct {
	ID                   string
	PendingReboot        bool // set when the device requires a reboot after update
	UpdatesAvailable     int
	Vendor               string
	Model                string
	Serial               string
	Components           []*model.Component
	Collectors           map[string]utils.Collector
	Logger               *logrus.Logger
	Dmidecode            *utils.Dmidecode
	FirmwareUpdateConfig *model.FirmwareUpdateConfig
}

func (s *Supermicro) GetModel() string {
	return s.Model
}

func (s *Supermicro) GetVendor() string {
	return s.Vendor
}

func (s *Supermicro) GetDeviceID() string {
	return s.ID
}

func (s *Supermicro) SetDeviceID(id string) {
	s.ID = id
}

func (s *Supermicro) RebootRequired() bool {
	return s.PendingReboot
}

func (s *Supermicro) SetFirmwareUpdateConfig(config *model.FirmwareUpdateConfig) {
	s.FirmwareUpdateConfig = config
}

func (s *Supermicro) SetOptions(options map[string]interface{}) error {
	return nil
}

// Returns hardware inventory for the device
func (s *Supermicro) GetInventory(ctx context.Context, listUpdates bool) (*model.Device, error) {

	inventory := make([]*model.Component, 0)
	// collect current component firmware versions
	s.Logger.Info("Identifying component firmware versions...")
	for cmd, collector := range s.Collectors {
		inv, err := collector.Components()
		if err != nil {
			s.Logger.WithFields(logrus.Fields{"cmd": cmd, "err": err}).Error("inventory collector error")
		}

		if len(inv) == 0 {
			s.Logger.WithFields(logrus.Fields{"cmd": cmd}).Trace("inventory collector returned no items")
			continue
		}

		inventory = append(inventory, inv...)
	}

	s.Components = inventory
	device := &model.Device{
		ID:         s.ID,
		Serial:     s.Serial,
		Model:      s.Model,
		Vendor:     s.Vendor,
		Oem:        true,
		Components: s.Components,
	}

	return device, nil
}

func (s *Supermicro) GetUpdatesAvailable(ctx context.Context) (*model.Device, error) {
	return &model.Device{}, nil
}

// TODO: decide how we calculate the device firmware revision
//  one possibility is to SHA the firmware versions of all components
func (s *Supermicro) GetDeviceFirmwareRevision(ctx context.Context) (string, error) {
	return "", nil
}

func (s *Supermicro) ApplyUpdatesAvailable(ctx context.Context) (err error) {

	return nil
}
