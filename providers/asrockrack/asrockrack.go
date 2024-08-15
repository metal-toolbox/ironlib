package asrockrack

import (
	"context"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// A asrockrack device has methods to collect hardware inventory, regardless of the vendor
type asrockrack struct {
	trace  bool
	hw     *model.Hardware
	logger *logrus.Logger
}

// New returns a ASRockRack device manager
func New(dmidecode *utils.Dmidecode, l *logrus.Logger) (actions.DeviceManager, error) {
	var trace bool

	if l.Level == logrus.TraceLevel {
		trace = true
	}

	var err error

	// set device
	device := common.NewDevice()

	identifiers, err := utils.IdentifyVendorModel(dmidecode)
	if err != nil {
		return nil, err
	}

	device.Vendor = identifiers.Vendor
	device.Model = identifiers.Model
	device.Serial = identifiers.Serial

	// set device manager
	dm := &asrockrack{
		hw:     model.NewHardware(&device),
		logger: l,
		trace:  trace,
	}

	return dm, nil
}

// Returns hardware inventory for the device
func (a *asrockrack) GetInventory(ctx context.Context, options ...actions.Option) (*common.Device, error) {
	// Collect device inventory
	a.logger.Debug("Collecting inventory")

	deviceObj := common.NewDevice()
	a.hw.Device = &deviceObj

	collector := actions.NewInventoryCollectorAction(a.logger, options...)
	if err := collector.Collect(ctx, a.hw.Device); err != nil {
		return nil, err
	}

	return a.hw.Device, nil
}

func (a *asrockrack) GetModel() string {
	return a.hw.Device.Model
}

func (a *asrockrack) GetVendor() string {
	return a.hw.Device.Vendor
}

func (a *asrockrack) RebootRequired() bool {
	return a.hw.PendingReboot
}

func (a *asrockrack) UpdatesApplied() bool {
	return a.hw.UpdatesInstalled
}

// ListAvailableUpdates runs the vendor tooling (dsu) to identify updates available
func (a *asrockrack) ListAvailableUpdates(context.Context, *model.UpdateOptions) (*common.Device, error) {
	return nil, nil
}

// InstallUpdates for asrockrack based on updateOptions
func (a *asrockrack) InstallUpdates(context.Context, *model.UpdateOptions) error {
	return nil
}

// UpdateRequirements returns requirements to be met before and after a firmware install,
// the caller may use the information to determine if a powercycle, reconfiguration or other actions are required on the component.
func (a *asrockrack) UpdateRequirements(_ context.Context, _, _, _ string) (*model.UpdateRequirements, error) {
	return nil, errors.Wrap(errs.ErrUpdateReqNotImplemented, "provider: asrockrack")
}

// GetInventoryOEM collects device inventory using vendor specific tooling
// and updates the given device.OemComponents object with the OEM inventory
func (a *asrockrack) GetInventoryOEM(context.Context, *common.Device, *model.UpdateOptions) error {
	return nil
}

// ApplyUpdate is here to satisfy the actions.Updater interface
// it is to be deprecated in favor of InstallUpdates.
func (a *asrockrack) ApplyUpdate(context.Context, string, string) error {
	return nil
}
