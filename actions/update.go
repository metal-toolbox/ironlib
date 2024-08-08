package actions

import (
	"context"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
)

var (
	ErrUpdaterUtilNotIdentified = errors.New("updater utility not identifed")
	ErrVendorComponentOptions   = errors.New("component vendor does not match update options vendor attribute")
)

// Updaters is a struct acting as a registry of various hardware component updaters
type Updaters struct {
	Drives             DriveUpdater
	NICs               NICUpdater
	BMC                BMCUpdater
	BIOS               BIOSUpdater
	StorageControllers StorageControllerUpdater
}

// Update runs updates based on given options
func Update(ctx context.Context, device *common.Device, options []*model.UpdateOptions) error {
	var err error

	for _, option := range options {
		switch {
		// Update BIOS
		case strings.EqualFold(common.SlugBIOS, option.Slug):
			err = UpdateBIOS(ctx, device.BIOS, option)
			if err != nil {
				return errors.Wrap(err, "error updating bios")
			}

		// Update Drive
		case strings.EqualFold(common.SlugDrive, option.Slug):
			err = UpdateDrive(ctx, device.Drives, option)
			if err != nil {
				return errors.Wrap(err, "error updating drive")
			}

		// Update NIC
		case strings.EqualFold(common.SlugNIC, option.Slug):
			err = UpdateNIC(ctx, device.NICs, option)
			if err != nil {
				return errors.Wrap(err, "error updating nic")
			}

		// Update BMC
		case strings.EqualFold(common.SlugBMC, option.Slug):
			err = UpdateBMC(ctx, device.BMC, option)
			if err != nil {
				return errors.Wrap(err, "error updating bmc")
			}
		default:
			return errors.Wrap(errs.ErrNoUpdateHandlerForComponent, "slug: "+option.Slug)
		}
	}

	return nil
}

// GetBMCUpdater returns the updater for the given vendor
func GetBMCUpdater(vendor string) (BMCUpdater, error) {
	if strings.EqualFold(vendor, common.VendorSupermicro) {
		return utils.NewSupermicroSUM(true), nil
	}

	return nil, errors.Wrap(ErrUpdaterUtilNotIdentified, "vendor: "+vendor)
}

// UpdateBMC identifies the bios eligible for update from the inventory and runs the firmware update utility based on the bmc vendor
func UpdateBMC(ctx context.Context, bmc *common.BMC, options *model.UpdateOptions) error {
	if !strings.EqualFold(options.Vendor, bmc.Vendor) {
		return ErrVendorComponentOptions
	}

	updater, err := GetBMCUpdater(bmc.Vendor)
	if err != nil {
		return err
	}

	return updater.UpdateBMC(ctx, options.UpdateFile, options.Model)
}

// GetBIOSUpdater returns the updater for the given vendor
func GetBIOSUpdater(vendor string) (BIOSUpdater, error) {
	if strings.EqualFold(vendor, common.VendorSupermicro) {
		return utils.NewSupermicroSUM(true), nil
	}

	return nil, errors.Wrap(ErrUpdaterUtilNotIdentified, "vendor: "+vendor)
}

// UpdateBIOS identifies the bios eligible for update from the inventory and runs the firmware update utility based on the bios vendor
func UpdateBIOS(ctx context.Context, bios *common.BIOS, options *model.UpdateOptions) error {
	if !strings.EqualFold(options.Vendor, bios.Vendor) {
		return ErrVendorComponentOptions
	}

	updater, err := GetBIOSUpdater(bios.Vendor)
	if err != nil {
		return err
	}

	return updater.UpdateBIOS(ctx, options.UpdateFile, options.Model)
}

// GetNICUpdater returns the updater for the given vendor
func GetNICUpdater(vendor string) (NICUpdater, error) {
	if strings.EqualFold(vendor, common.VendorMellanox) {
		return utils.NewMlxupCmd(true), nil
	}

	return nil, errors.Wrap(ErrUpdaterUtilNotIdentified, vendor)
}

// UpdateNIC identifies the nic eligible for update from the inventory and runs the firmware update utility based on the nic vendor
func UpdateNIC(ctx context.Context, nics []*common.NIC, options *model.UpdateOptions) error {
	for _, nic := range nics {
		if !strings.EqualFold(options.Vendor, nic.Vendor) {
			continue
		}

		updater, err := GetNICUpdater(nic.Vendor)
		if err != nil {
			return err
		}

		return updater.UpdateNIC(ctx, options.UpdateFile, options.Model, options.ForceInstall)
	}

	return errors.Wrap(ErrUpdaterUtilNotIdentified, options.Vendor)
}

// GetDriveUpdater returns the updater for the given vendor
func GetDriveUpdater(vendor string) (DriveUpdater, error) {
	if strings.EqualFold(vendor, common.VendorMicron) {
		return utils.NewMsecli(true), nil
	}

	return nil, errors.Wrap(ErrUpdaterUtilNotIdentified, "vendor: "+vendor)
}

// UpdateDrive identifies the drive eligible for update from the inventory and runs the firmware update utility based on the drive vendor
func UpdateDrive(ctx context.Context, drives []*common.Drive, options *model.UpdateOptions) error {
	for _, drive := range drives {
		if !strings.EqualFold(options.Vendor, drive.Vendor) {
			continue
		}

		updater, err := GetDriveUpdater(drive.Vendor)
		if err != nil {
			return err
		}

		return updater.UpdateDrive(ctx, options.UpdateFile, options.Model, options.Serial)
	}

	return errors.Wrap(ErrUpdaterUtilNotIdentified, options.Vendor)
}
