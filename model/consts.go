package model

import (
	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
)

// CollectorUtility is the name of a utility defined in utils/
type CollectorUtility string

var (

	// ModelDriveTypeSlug is a map of drive models number to slug
	// Until we figure a better way to differentiate drive information
	// into SATA vs PCI NVMe or others, this map is going to be annoying to keep updated
	// As of now - neither lshw or smartctl clearly points out the difference in the controller
	modelDriveTypeSlug = map[string]string{
		// Sata SSD drives
		"Micron_5200_MTFDDAK480TDN": common.SlugDriveTypeSATASSD,
		"Micron_5200_MTFDDAK960TDN": common.SlugDriveTypeSATASSD,
		"MTFDDAV240TDU":             common.SlugDriveTypeSATASSD,
		// PCI NVMe SSD drives
		"KXG60ZNV256G TOSHIBA":      common.SlugDriveTypePCIeNVMEeSSD,
		"Micron_9300_MTFDHAL3T8TDP": common.SlugDriveTypePCIeNVMEeSSD,
		// Sata HDD drives
		"HGST HUS728T8TALE6L4": common.SlugDriveTypeSATAHDD,
	}

	ErrTypeComponentFirmware = errors.New("ironlib.GetComponentFirmware() was passed an object type which is not handled")
)

func DriveTypeSlug(m string) string {
	t, exists := modelDriveTypeSlug[m]
	if !exists {
		return "Unknown"
	}

	return t
}

// Return a normalized product name given a product name
func FormatProductName(s string) string {
	switch s {
	case "SSG-6029P-E1CR12L-PH004":
		return "SSG-6029P-E1CR12L"
	case "SYS-5019C-MR-PH004", "PIO-519C-MR-PH004":
		return "SYS-5019C-MR"
	case "PowerEdge R640":
		return "r640"
	case "PowerEdge C6320":
		return "c6320"
	case "Micron_5200_MTFDDAK480TDN":
		return "5200MAX"
	default:
		return s
	}
}
