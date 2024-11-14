package actions

import (
	"strings"

	common "github.com/metal-toolbox/bmc-common"

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

func driveCapabilityCollectorByLogicalName(logicalName string, trace bool, collectors []DriveCapabilityCollector) DriveCapabilityCollector {
	// when collectors are is passed in, return the collector based on the logical name
	for _, collector := range collectors {
		nvme, isNvmeCollector := collector.(*utils.Nvme)
		if isNvmeCollector && strings.Contains(logicalName, "nvme") {
			return nvme
		}

		hdparm, isHdparmCollector := collector.(*utils.Hdparm)
		if isHdparmCollector && !strings.Contains(logicalName, "nvme") {
			return hdparm
		}
	}

	// otherwise return the collector based on the logical name
	switch {
	case strings.Contains(logicalName, "nvme"):
		return utils.NewNvmeCmd(trace)
	default:
		return utils.NewHdparmCmd(trace)
	}
}
