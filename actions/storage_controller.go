package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
)

var (
	ErrVirtualDiskManagerUtilNotIdentified = errors.New("virtual disk management utility not identifed")
)

func CreateVirtualDisk(ctx context.Context, hba *common.StorageController, options *model.CreateVirtualDiskOptions) error {
	util, err := GetControllerUtility(hba.Vendor, hba.Model)
	if err != nil {
		return err
	}

	return util.CreateVirtualDisk(ctx, options.RaidMode, options.PhysicalDiskIDs, options.Name, options.BlockSize)
}

func DestroyVirtualDisk(ctx context.Context, hba *common.StorageController, options *model.DestroyVirtualDiskOptions) error {
	util, err := GetControllerUtility(hba.Vendor, hba.Model)
	if err != nil {
		return err
	}

	return util.DestroyVirtualDisk(ctx, options.VirtualDiskID)
}

func ListVirtualDisks(ctx context.Context, hba *common.StorageController) ([]*common.VirtualDisk, error) {
	util, err := GetControllerUtility(hba.Vendor, hba.Model)
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
func GetControllerUtility(vendorName, modelName string) (VirtualDiskManager, error) {
	if strings.EqualFold(vendorName, common.VendorMarvell) {
		return utils.NewMvcliCmd(true), nil
	}

	return nil, errors.Wrap(ErrVirtualDiskManagerUtilNotIdentified, "vendor: "+vendorName+" model: "+modelName)
}
