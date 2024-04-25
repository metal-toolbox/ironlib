package actions

import (
	"context"
	"fmt"
	"log"
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
}

func NewStorageControllerAction(logger *logrus.Logger) *StorageControllerAction {
	return &StorageControllerAction{logger}
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
	var trace bool

	if s.Logger.GetLevel().String() == "trace" {
		trace = true
	}

	if strings.EqualFold(vendorName, common.VendorMarvell) {
		return utils.NewMvcliCmd(trace), nil
	}

	return nil, errors.Wrap(ErrVirtualDiskManagerUtilNotIdentified, "vendor: "+vendorName+" model: "+modelName)
}

// GetWipeUtility returns the wipe utility based on the disk wipping features
func (s *StorageControllerAction) GetWipeUtility(logicalName string) (DiskWiper, error) {
	var trace bool

	if s.Logger.GetLevel().String() == "trace" {
		trace = true
	}
	// TODO: use disk wipping features to return the best wipe utility, currently only one available
	if trace {
		log.Printf("%s | Detecting wipe utility", logicalName)
	}

	return utils.NewFillZeroCmd(trace), nil
}

func (s *StorageControllerAction) WipeDisk(ctx context.Context, logicalName string) error {
	util, err := s.GetWipeUtility(logicalName)
	if err != nil {
		return err
	}
	// Watermark disk
	log.Printf("%s | Initiating watermarking process", logicalName)
	check := utils.ApplyWatermarks(logicalName)

	err = util.WipeDisk(ctx, logicalName)
	if err != nil {
		return err
	}
	log.Printf("%s | Checking if the watermark has been removed", logicalName)
	err = check()
	if err != nil {
		return err
	}
	log.Printf("%s | Watermarks has been removed", logicalName)
	return nil
}
