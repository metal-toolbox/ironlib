package generic

import (
	"context"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// A Generic device has methods to collect hardware inventory, regardless of the vendor
type Generic struct {
	ID               string
	Vendor           string
	Model            string
	Serial           string
	UpdatesAvailable int
	PendingReboot    bool // set when the device requires a reboot after update
	UpdatesInstalled bool // set when updates were installed on the device
	Collectors       map[string]utils.Collector
	Logger           *logrus.Logger
	Dmidecode        *utils.Dmidecode
}

func New(vendor, model string, l *logrus.Logger) (model.Manager, error) {
	return nil, nil
}

// Returns hardware inventory for the device
func (a *Generic) GetInventory(ctx context.Context, listUpdates bool) (*model.Device, error) {

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

func (a *Generic) GetModel() string {
	return a.Model
}

func (a *Generic) GetVendor() string {
	return a.Vendor
}

func (a *Generic) GetDeviceID() string {
	return a.ID
}

func (a *Generic) SetDeviceID(id string) {
	a.ID = id
}

func (a *Generic) RebootRequired() bool {
	return a.PendingReboot
}

func (a *Generic) SetFirmwareUpdateConfig(config *model.FirmwareUpdateConfig) {
}

func (a *Generic) SetOptions(options map[string]interface{}) error {
	return nil
}

func (a *Generic) UpdatesApplied() bool {
	return a.UpdatesInstalled
}

func (a *Generic) ApplyUpdatesAvailable(ctx context.Context, config *model.FirmwareUpdateConfig, dryRun bool) (err error) {
	return nil
}

func (a *Generic) GetDeviceFirmwareRevision(ctx context.Context) (string, error) {
	return "", nil
}

func (a *Generic) GetUpdatesAvailable(ctx context.Context) (*model.Device, error) {
	return &model.Device{}, nil
}
