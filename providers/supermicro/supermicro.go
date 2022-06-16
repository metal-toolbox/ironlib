package supermicro

import (
	"context"

	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type supermicro struct {
	trace      bool
	hw         *model.Hardware
	logger     *logrus.Logger
	dmidecode  *utils.Dmidecode
	collectors *actions.Collectors
}

func New(dmidecode *utils.Dmidecode, l *logrus.Logger) (model.DeviceManager, error) {
	var trace bool
	if l.GetLevel().String() == "trace" {
		trace = true
	}

	// register inventory collectors
	collectors := &actions.Collectors{
		BMC:                utils.NewIpmicfgCmd(trace),
		BIOS:               utils.NewIpmicfgCmd(trace),
		CPLDs:              utils.NewIpmicfgCmd(trace),
		Drives:             utils.NewSmartctlCmd(trace),
		StorageControllers: utils.NewStoreCLICmd(trace),
		NICs:               utils.NewMlxupCmd(trace),
	}

	deviceVendor, err := dmidecode.Manufacturer()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("manufacturer", "", 0), err.Error())
	}

	// Supermicro's have a consistent baseboard product name
	// compared to the marketing product name which varies based on location
	deviceModel, err := dmidecode.BaseBoardProductName()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("Product name", "", 0), err.Error())
	}

	serial, err := dmidecode.SerialNumber()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("Serial", "", 0), err.Error())
	}

	device := model.NewDevice()
	device.Model = deviceModel
	device.Vendor = deviceVendor
	device.Serial = serial

	return &supermicro{
		hw:         model.NewHardware(device),
		collectors: collectors,
		logger:     l,
		dmidecode:  dmidecode,
		trace:      trace,
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
	// Collect device inventory from lshw
	s.logger.Info("Collecting hardware inventory")

	err := actions.Collect(ctx, s.hw.Device, s.collectors, s.trace, false)
	if err != nil {
		return nil, err
	}

	return s.hw.Device, nil
}

// ListUpdatesAvailable does nothing on a SMC device
func (s *supermicro) ListAvailableUpdates(ctx context.Context, options *model.UpdateOptions) (*model.Device, error) {
	return nil, nil
}

// InstallUpdates for Supermicros based on the given options
//
// errors are returned when the updater fails to apply updates
func (s *supermicro) InstallUpdates(ctx context.Context, option *model.UpdateOptions) (err error) {
	// collect device inventory if it isn't added already
	if s.hw.Device == nil || s.hw.Device.BIOS == nil {
		s.hw.Device, err = s.GetInventory(ctx)
		if err != nil {
			return err
		}
	}

	if option.Model == "" {
		option.Model = s.hw.Device.Model
	}

	err = actions.Update(ctx, s.hw.Device, []*model.UpdateOptions{option})
	if err != nil {
		return err
	}

	// this flag can be optimized further
	// BMC updates don't require a reboot
	s.hw.PendingReboot = true
	s.hw.UpdatesInstalled = true

	return nil
}

// GetInventoryOEM collects device inventory using vendor specific tooling
// and updates the given device.OemComponents object with the OEM inventory
func (s *supermicro) GetInventoryOEM(ctx context.Context, device *model.Device, options *model.UpdateOptions) error {
	return nil
}
