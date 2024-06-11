package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var ErrVirtualDiskManagerUtilNotIdentified = errors.New("virtual disk management utility not identifed")

type StorageControllerAction struct {
	Logger *logrus.Logger
	trace  bool
}

func NewStorageControllerAction(logger *logrus.Logger) *StorageControllerAction {
	return &StorageControllerAction{
		Logger: logger,
		trace:  logger.Level >= logrus.TraceLevel,
	}
}

func (s *StorageControllerAction) CreateVirtualDisk(ctx context.Context, hba *common.StorageController, options *model.CreateVirtualDiskOptions) error {
	util, err := s.GetControllerUtility(hba.Vendor, hba.Model)
	if err != nil {
		return err
	}

	return util.CreateVirtualDisk(ctx, options.RaidMode, options.PhysicalDiskIDs, options.Name, options.BlockSize)
}

func (s *StorageControllerAction) DestroyVirtualDisk(ctx context.Context, hba *common.StorageController, options *model.DestroyVirtualDiskOptions) error {
	util, err := s.GetControllerUtility(hba.Vendor, hba.Model)
	if err != nil {
		return err
	}

	return util.DestroyVirtualDisk(ctx, options.VirtualDiskID)
}

func (s *StorageControllerAction) ListVirtualDisks(ctx context.Context, hba *common.StorageController) ([]*common.VirtualDisk, error) {
	util, err := s.GetControllerUtility(hba.Vendor, hba.Model)
	if err != nil {
		return nil, err
	}

	virtualDisks, err := util.VirtualDisks(ctx)
	if err != nil {
		return nil, err
	}

	cVirtualDisks := []*common.VirtualDisk{}

	for _, vd := range virtualDisks {
		cVirtualDisks = append(cVirtualDisks, &common.VirtualDisk{
			ID:       fmt.Sprintf("%d", vd.ID),
			Name:     vd.Name,
			RaidType: vd.Type,
		})
	}

	return cVirtualDisks, nil
}

// GetControllerUtility returns the utility command for the given vendor
func (s *StorageControllerAction) GetControllerUtility(vendorName, modelName string) (VirtualDiskManager, error) {
	if strings.EqualFold(vendorName, common.VendorMarvell) {
		return utils.NewMvcliCmd(s.trace), nil
	}

	return nil, errors.Wrap(ErrVirtualDiskManagerUtilNotIdentified, "vendor: "+vendorName+" model: "+modelName)
}

// GetWipeUtility returns the wipe utility based on the disk wipping features
func (s *StorageControllerAction) GetWipeUtility(drive *common.Drive) (DriveWiper, error) {
	s.Logger.Tracef("%s | Detecting wipe utility", drive.LogicalName)
	// TODO: use disk wipping features to return the best wipe utility, currently only one available

	return utils.NewFillZeroCmd(s.trace), nil
}

func (s *StorageControllerAction) WipeDrive(ctx context.Context, log *logrus.Logger, drive *common.Drive) error {
	util, err := s.GetWipeUtility(drive)
	if err != nil {
		return err
	}

	// Watermark disk
	// Before wiping the disk, we apply watermarks to later verify successful deletion
	check, err := utils.ApplyWatermarks(drive)
	if err != nil {
		return err
	}

	// Wipe the disk
	err = util.WipeDrive(ctx, log, drive)
	if err != nil {
		return err
	}

	return check()
}
