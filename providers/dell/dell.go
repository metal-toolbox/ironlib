package dell

import (
	"context"
	"fmt"
	"strings"

	"github.com/packethost/ironlib/errs"
	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// The dell device provider struct
type Dell struct {
	DM                      *model.DeviceManager // the device manager holds common fields that is shared among providers
	DsuPrequisitesInstalled bool
	Dnf                     *utils.Dnf
	Dsu                     *utils.Dsu
	Dmidecode               *utils.Dmidecode
	Logger                  *logrus.Logger
}

func New(deviceVendor, deviceModel string, l *logrus.Logger) (model.Manager, error) {

	var trace bool

	if l.GetLevel().String() == "trace" {
		trace = true
	}

	dmidecode, err := utils.NewDmidecode()
	if err != nil {
		errors.Wrap(err, "erorr in dmidecode init")
	}

	// set device
	device := &model.Device{
		Model:            deviceModel,
		Vendor:           deviceVendor,
		Oem:              true,
		Components:       []*model.Component{},
		ComponentUpdates: []*model.Component{},
	}

	// set device manager
	dm := &Dell{
		DM:        &model.DeviceManager{Device: device},
		Dmidecode: dmidecode,
		Dnf:       utils.NewDnf(trace),
		Dsu:       utils.NewDsu(trace),
		Logger:    l,
	}

	return dm, nil
}

type Component struct {
	Serial   string
	Type     string
	Model    string
	Firmware string
}

func (d *Dell) GetModel() string {
	return d.DM.Device.Model
}

func (d *Dell) GetVendor() string {
	return d.DM.Device.Vendor
}

func (d *Dell) GetDeviceID() string {
	return d.DM.Device.ID
}

func (d *Dell) SetDeviceID(id string) {
	d.DM.Device.ID = id
}

func (d *Dell) SetFirmwareUpdateConfig(config *model.FirmwareUpdateConfig) {
	d.FirmwareUpdateConfig = config
}

func (d *Dell) RebootRequired() bool {
	return d.DM.PendingReboot
}

func (d *Dell) UpdatesApplied() bool {
	return d.DM.UpdatesInstalled
}

// nolint: gocyclo
// Return device component inventory, including any update information
func (d *Dell) GetInventory(ctx context.Context, listUpdates bool) (*model.Device, error) {

	var err error

	// dsu list inventory
	d.Logger.Info("Collecting DSU component inventory...")
	d.DM.Device.Components, err = d.dsuInventory()
	if err != nil {
		return nil, err
	}

	if len(d.DM.Device.Components) == 0 {
		d.Logger.Warn("No device components returned by dsu inventory")
	}

	if !listUpdates {
		return d.DM.Device, nil
	}

	// Identify updates to install
	updates, err := d.identifyUpdatesApplicable(d.DM.Device.Components, d.DM.FirmwareUpdateConfig)
	if err != nil {
		return nil, err
	}

	count := len(updates)
	if count == 0 {
		return d.DM.Device, nil
	}

	// converge component inventory data with firmware update information
	d.DM.Device.ComponentUpdates = updates
	for _, component := range d.DM.Device.Components {
		component.DeviceID = d.DM.Device.ID
		for _, update := range d.DM.Device.ComponentUpdates {
			if strings.EqualFold(component.Slug, update.Slug) {
				component.Metadata = update.Metadata
				if strings.TrimSpace(update.FirmwareAvailable) != "" {
					d.Logger.WithFields(logrus.Fields{"component slug": component.Slug, "installed": component.FirmwareInstalled, "update": update.FirmwareAvailable}).Trace("update available")
				}
				if component.Slug == "Unknown" {
					d.Logger.WithFields(logrus.Fields{"component name": component.Name}).Warn("component slug is 'Unknown', this needs to be fixed in componentNameSlug()")
				}
				component.FirmwareAvailable = update.FirmwareAvailable
			}
		}
	}

	return d.DM.Device, nil
}

// Return available firmware updates for device
func (d *Dell) GetUpdatesAvailable(ctx context.Context) (*model.Device, error) {

	// collect firmware updates available for components
	d.Logger.Info("Identifying component firmware updates...")
	updates, err := d.dsuListUpdates()
	if err != nil {
		return nil, err
	}

	count := len(updates)
	if count > 0 {
		d.DM.Device.ComponentUpdates = append(d.DM.Device.ComponentUpdates, updates...)
		d.Logger.WithField("count", count).Info("updates available..")
	} else {
		d.Logger.Info("no available updates")
	}

	return d.DM.Device, nil
}

// The installed DSU release is the firmware revision for dells
func (d *Dell) GetDeviceFirmwareRevision(ctx context.Context) (string, error) {
	return d.Dsu.Version()
}

// nolint: gocyclo
// Installs either all updates identified by DSU, OR component level updates if any listed
func (d *Dell) ApplyUpdatesAvailable(ctx context.Context, config *model.FirmwareUpdateConfig, dryRun bool) (err error) {

	if config == nil {
		return fmt.Errorf("ApplyUpdatesAvailable() requires a valid *model.FirmwareUpdateConfig")
	}

	// atleast one of these is required
	if len(config.Updates) == 0 && len(config.Components) == 0 {
		return fmt.Errorf("expected a valid 'updates:' and/or 'components:' list in the firmware config")
	}

	d.DM.FirmwareUpdateConfig = config

	// fetch firmware inventory if updates were not yet identified
	if d.DM.UpdatesAvailable == -1 {
		_, err = d.GetInventory(ctx, true)
		if err != nil {
			return err
		}
	}

	if len(d.DM.Device.ComponentUpdates) == 0 {
		d.DM.PendingReboot = false
		d.DM.UpdatesInstalled = false
		d.DM.UpdatesAvailable = 0
		return errs.ErrNoUpdatesApplicable
	}

	// fetch applicable updates and install
	err = d.fetchAndApplyUpdates(d.DM.Device.ComponentUpdates, config, dryRun)
	if err != nil {
		return err
	}

	return nil
}
