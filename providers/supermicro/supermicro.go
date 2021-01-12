package supermicro

import (
	"context"
	"fmt"
	"strings"

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
	Updater              utils.Updater
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
		components, err := collector.Components()
		if err != nil {
			s.Logger.WithFields(logrus.Fields{"cmd": cmd, "err": err}).Error("inventory collector error")
		}

		if len(components) == 0 {
			s.Logger.WithFields(logrus.Fields{"cmd": cmd}).Trace("inventory collector returned no items")
			continue
		}
		inventory = append(inventory, components...)
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

	// identify components eligible for updates
	if listUpdates {
		components, err := s.componentsForUpdate(s.Components, s.FirmwareUpdateConfig)
		if err != nil {
			s.Logger.WithFields(logrus.Fields{"err": err}).Warn("failed to identify components for update")
		} else {
			device.ComponentUpdates = components
		}
	}

	return device, nil
}

func (s *Supermicro) GetUpdatesAvailable(ctx context.Context) (*model.Device, error) {
	return &model.Device{}, nil
}

// TODO: decide how we calculate the device firmware revision
// as of now the firmware revision is set to the first part of the firmware config uuID
func (s *Supermicro) GetDeviceFirmwareRevision(ctx context.Context) (string, error) {
	if s.FirmwareUpdateConfig == nil {
		return "", fmt.Errorf("GetDeviceFirmwareRevision requires a valid *model.FirmwareUpdateConfig")
	}

	tokens := strings.Split(s.FirmwareUpdateConfig.ID, "-")
	if len(tokens) == 0 {
		return "", fmt.Errorf("GetDeviceFirmwareRevision expects *model.FirmwareUpdateConfig.ID to be a valid UUID")
	}

	return tokens[0], nil
}

// Identify components firmware revisions and apply updates
func (s *Supermicro) ApplyUpdatesAvailable(ctx context.Context, config *model.FirmwareUpdateConfig, dryRun bool) (err error) {

	if config == nil {
		return fmt.Errorf("ApplyUpdatesAvailable() requires a valid *model.FirmwareUpdateConfig")
	}

	s.FirmwareUpdateConfig = config

	// get component firmware inventory
	device, err := s.GetInventory(ctx, false)
	if err != nil {
		return fmt.Errorf("Failed to get inventory for device before upgrade: " + err.Error())
	}

	components, err := s.componentsForUpdate(device.Components, config)
	if err != nil {
		return err
	}

	if len(components) == 0 {
		s.Logger.Info("No components identified for updates")
	}

	// fetch and apply component updates
	for _, component := range components {

		s.Logger.WithFields(logrus.Fields{"slug": component.Slug, "name": component.Name, "url": component.Config.UpdateFileURL, "dst": "/tmp"}).Info("fetching component update")
		// retrieve update file under target directory, validate checksum
		updateFile, err := utils.RetrieveUpdateFile(component.Config.UpdateFileURL, "/tmp")
		if err != nil {
			return err
		}

		s.Logger.WithFields(logrus.Fields{"slug": component.Slug, "name": component.Name, "installed": component.FirmwareInstalled, "update": component.Config.Updates[0]}).Info("component update to be applied")
		if dryRun {
			continue
		}

		// setup SMC sum for executing
		sum := utils.NewSupermicroSUM(true)

		// execute sum and apply update for component
		err = sum.ApplyUpdate(ctx, updateFile, strings.ToLower(component.Slug))
		if err != nil {
			s.Logger.WithFields(logrus.Fields{"component": component.Slug, "err": err}).Warn("component update error")
			return err
		}
	}

	return nil
}

// Given a slice of components and the firmware config,
// compares current installed firmware with the version listed in the config and
// returns a slice of *model.Component's which are eligible for updates
// sets Component.Config to the config identified for the component
// the component config is matched by the Slug attribute
func (s *Supermicro) componentsForUpdate(components []*model.Component, config *model.FirmwareUpdateConfig) ([]*model.Component, error) {

	forUpdate := make([]*model.Component, 0)

	// identify and apply update
	for _, component := range components {

		// identify component firmware config
		componentConfig := s.componentConfig(component.Slug)
		if componentConfig == nil {
			continue
		}

		// version compare current firmware version with the configuration
		hasUpdate, err := utils.VersionIsNewer(componentConfig.Updates[0], component.FirmwareInstalled)
		if err != nil {
			return nil, fmt.Errorf("version compare error: component '%s' installed '%s', update '%s': error %s",
				component.Slug, component.FirmwareInstalled, componentConfig.Updates[0], err.Error())
		}

		if !hasUpdate {
			continue
		}

		component.Config = componentConfig
		forUpdate = append(forUpdate, component)

	}

	return forUpdate, nil

}

// Returns the configuration that is valid for the component
// compares the given slug to the component slug in the component firmware configuration
func (s *Supermicro) componentConfig(slug string) *model.ComponentFirmwareConfig {

	for _, config := range s.FirmwareUpdateConfig.Components {
		if strings.ToLower(slug) == strings.ToLower(config.Slug) {
			return config
		}
	}

	return nil
}
