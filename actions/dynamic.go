package actions

import (
	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/utils"
)

func StorageControllerCollectorByVendor(vendor string, trace bool) StorageControllerCollector {
	if common.FormatVendorName(vendor) == common.VendorMarvell {
		return utils.NewMvcliCmd(trace)
	}

	return nil
}

func DriveCollectorByStorageControllerVendor(vendor string, trace bool) DriveCollector {
	if common.FormatVendorName(vendor) == common.VendorMarvell {
		return utils.NewMvcliCmd(trace)
	}

	return nil
}
