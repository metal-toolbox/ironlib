package supermicro

import (
	"context"
	"strings"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type supermicro struct {
	hw         *model.Hardware
	logger     *logrus.Logger
	lshw       *utils.Lshw
	collectors map[string]utils.Collector
}

func New(deviceVendor, deviceModel string, l *logrus.Logger) (model.DeviceManager, error) {
	var trace bool

	if l.GetLevel().String() == "trace" {
		trace = true
	}

	// register inventory collectors
	collectors := map[string]utils.Collector{
		"ipmi":     utils.NewIpmicfgCmd(trace),
		"smartctl": utils.NewSmartctlCmd(trace),
		"storecli": utils.NewStoreCLICmd(trace),
		"mlxup":    utils.NewMlxupCmd(trace),
	}

	device := &model.Device{
		Model:  deviceModel,
		Vendor: deviceVendor,
	}

	return &supermicro{
		hw:         model.NewHardware(device),
		lshw:       utils.NewLshwCmd(trace),
		collectors: collectors,
		logger:     l,
	}, nil
}

func (s *supermicro) GetModel() string {
	return s.hw.Device.Model
}

func (s *supermicro) GetVendor() string {
	return s.hw.Device.Vendor
}

func (s *supermicro) RebootRequired() bool {
	return s.hw.PendingReboot
}

func (s *supermicro) UpdatesApplied() bool {
	return s.hw.UpdatesInstalled
}

// GetInventory collects hardware inventory along with the firmware installed and returns a Device object
func (s *supermicro) GetInventory(ctx context.Context) (*model.Device, error) {
	inventory := make([]*model.Component, 0)

	// Collect device inventory from lshw
	s.logger.Info("Collecting inventory with lshw")

	s.hw.Device = model.NewDevice()

	err := s.lshw.Inventory(s.hw.Device)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving device inventory")
	}

	// collect current component firmware versions
	s.logger.Info("Identifying component firmware versions...")

	for cmd, collector := range s.collectors {
		components, err := collector.Components()
		if err != nil {
			s.logger.WithFields(logrus.Fields{"cmd": cmd, "err": err}).Error("inventory collector error")
		}

		if len(components) == 0 {
			s.logger.WithFields(logrus.Fields{"cmd": cmd}).Trace("inventory collector returned no items")
			continue
		}

		inventory = append(inventory, components...)
	}

	// update device with the components retrieved from inventory
	model.SetDeviceComponents(s.hw.Device, inventory)

	return s.hw.Device, nil
}

// ListUpdatesAvailable does nothing on a SMC device
func (s *supermicro) ListUpdatesAvailable(ctx context.Context) (*model.Device, error) {
	return nil, nil
}

// InstallUpdates for Supermicros based on the given options
func (s *supermicro) InstallUpdates(ctx context.Context, options *model.UpdateOptions) (err error) {
	var updater utils.Updater

	var trace bool
	if s.logger.Level == logrus.TraceLevel {
		trace = true
	}

	// set updater based on the component slug, vendor
	switch options.Slug {
	case model.SlugBIOS, model.SlugBMC:
		// setup SMC sum for executing
		updater = utils.NewSupermicroSUM(trace)

	case model.SlugNIC:
		// mellanox NIC update
		updater = utils.NewMlxupUpdater(trace)

	case model.SlugDrive:
		// micron disk
		if strings.EqualFold(options.Vendor, model.VendorMicron) {
			updater = utils.NewMsecliUpdater(trace)
		} else {
			s.logger.WithFields(
				logrus.Fields{"slug": options.Slug, "name": options.Name, "vendor": options.Vendor},
			).Warn("unsupported disk vendor")
		}
	}

	err = updater.ApplyUpdate(ctx, options.UpdateFile, options.Slug)
	if err != nil {
		s.logger.WithFields(
			logrus.Fields{"component": options.Slug, "err": err},
		).Warn("component update error")

		return err
	}

	// this flag can be optimized further
	// BMC updates don't require a reboot
	s.hw.PendingReboot = true
	s.hw.UpdatesInstalled = true

	return nil
}
