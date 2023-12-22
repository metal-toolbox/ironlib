package supermicro

import (
	"context"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/firmware"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type supermicro struct {
	trace     bool
	hw        *model.Hardware
	logger    *logrus.Logger
	dmidecode *utils.Dmidecode
}

func New(dmidecode *utils.Dmidecode, l *logrus.Logger) (actions.DeviceManager, error) {
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

	device := common.NewDevice()
	device.Model = deviceModel
	device.Vendor = deviceVendor
	device.Serial = serial

	return &supermicro{
		hw:        model.NewHardware(&device),
		logger:    l,
		dmidecode: dmidecode,
		trace:     l.GetLevel().String() == "trace",
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
func (s *supermicro) GetInventory(ctx context.Context, options ...actions.Option) (*common.Device, error) {
	// Collect device inventory
	s.logger.Info("Collecting hardware inventory")

	var trace bool
	if s.logger.GetLevel().String() == "trace" {
		trace = true
	}

	// define collectors for supermicro hardware
	collectors := &actions.Collectors{
		BMCCollector:  utils.NewIpmicfgCmd(trace),
		BIOSCollector: utils.NewIpmicfgCmd(trace),
		CPLDCollector: utils.NewIpmicfgCmd(trace),
		DriveCollectors: []actions.DriveCollector{
			utils.NewSmartctlCmd(trace),
			utils.NewLsblkCmd(trace),
		},
		DriveCapabilitiesCollectors: []actions.DriveCapabilityCollector{
			utils.NewHdparmCmd(trace),
			utils.NewNvmeCmd(trace),
		},
		StorageControllerCollectors: []actions.StorageControllerCollector{
			utils.NewStoreCLICmd(trace),
		},
		NICCollector: utils.NewMlxupCmd(trace),
		FirmwareChecksumCollector: firmware.NewChecksumCollector(
			firmware.MakeOutputPath(),
			firmware.TraceExecution(trace),
		),
	}

	options = append(options, actions.WithCollectors(collectors))

	collector := actions.NewInventoryCollectorAction(options...)
	if err := collector.Collect(ctx, s.hw.Device); err != nil {
		return nil, err
	}

	return s.hw.Device, nil
}

// ListUpdatesAvailable does nothing on a SMC device
func (s *supermicro) ListAvailableUpdates(ctx context.Context, options *model.UpdateOptions) (*common.Device, error) {
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
func (s *supermicro) GetInventoryOEM(ctx context.Context, device *common.Device, options *model.UpdateOptions) error {
	return nil
}

// ApplyUpdate is here to satisfy the actions.Updater interface
// it is to be deprecated in favor of InstallUpdates.
func (s *supermicro) ApplyUpdate(ctx context.Context, updateFile, component string) error {
	return nil
}
