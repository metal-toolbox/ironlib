package model

import "strings"

const (
	// Vendors
	VendorDell       = "dell"
	VendorMicron     = "micron"
	VendorAsrockrack = "asrockrack"
	VendorSupermicro = "supermicro"

	// Generic component slugs
	SlugBackplaneExpander     = "Backplane Expander"
	SlugChassis               = "Chassis"
	SlugTPM                   = "TPM"
	SlugGPU                   = "GPU"
	SlugCPU                   = "CPU"
	SlugPhysicalMem           = "PhysicalMemory"
	SlugStorageController     = "StorageController"
	SlugBMC                   = "BMC"
	SlugBIOS                  = "BIOS"
	SlugDrive                 = "Drive"
	SlugDriveTypePCIeNVMEeSSD = "NVMe PCIe SSD"
	SlugDriveTypeSATASSD      = "Sata SSD"
	SlugNIC                   = "NIC"
	SlugPSU                   = "Power Supply"
	SlugSASHBA330Controller   = "SAS HBA330 Controller"
	SlugCPLD                  = "CPLD"

	// Dell specific component slugs
	SlugDellSystemCPLD                  = "Dell System CPLD"
	SlugDellIdracServiceModel           = "IDrac Service Module"
	SlugDellBossAdapter                 = "Boss Adapter"
	SlugDellBossAdapterDisk0            = "Boss Adapter - Disk 0"
	SlugDellLifeCycleController         = "Lifecycle Controller"
	SlugDellOSCollector                 = "OS Collector"
	SlugDell64bitUefiDiagnostics        = "Dell 64 bit uEFI diagnostics"
	SlugDellBackplaneExpander           = "Backplane Expander"
	SlugDellNonExpanderStorageBackplane = "Non-Expander Storage Backplane (SEP)"

	EnvDnfPackageRepository = "DNF_REPO_ENVIRONMENT"
)

// UpdateReleaseEnvironments is the list of update environments
// this is related to the fup tooling
func UpdateReleaseEnvironments() []string {
	return []string{"production", "canary", "vanguard"}
}

var (
	// ModelDriveTypeSlug is a map of drive models number to slug
	// Until we figure a better way to differentiate drive information
	// into SATA vs PCI NVMe or others, this map is going to be annoying to keep updated
	// As of now - neither lshwn or smartctl clearly points out the difference in the controller
	modelDriveTypeSlug = map[string]string{
		// Sata SSD drives
		"Micron_5200_MTFDDAK480TDN": SlugDriveTypeSATASSD,
		"Micron_5200_MTFDDAK960TDN": SlugDriveTypeSATASSD,
		"MTFDDAV240TDU":             SlugDriveTypeSATASSD,
		// PCI NVMe SSD drives
		"KXG60ZNV256G TOSHIBA":      SlugDriveTypePCIeNVMEeSSD,
		"Micron_9300_MTFDHAL3T8TDP": SlugDriveTypePCIeNVMEeSSD,
	}

	// OemComponentDell is a lookup table for dell OEM components
	OemComponentDell = map[string]struct{}{
		SlugDellSystemCPLD:                  {},
		SlugBackplaneExpander:               {},
		SlugDellIdracServiceModel:           {},
		SlugDellBossAdapterDisk0:            {},
		SlugDellBossAdapter:                 {},
		SlugDellLifeCycleController:         {},
		SlugDellNonExpanderStorageBackplane: {},
		SlugDellOSCollector:                 {},
		SlugDell64bitUefiDiagnostics:        {},
	}

	// DellComponentSlug map of dell component names to component slug
	DellComponentSlug = map[string]string{
		"bios":                         SlugBIOS,
		"ethernet":                     SlugNIC,
		"idrac":                        SlugBMC,
		"backplane":                    SlugBackplaneExpander,
		"power supply":                 SlugPSU,
		"hba330":                       SlugSASHBA330Controller,
		"nvmepcissd":                   SlugDrive,
		"system cpld":                  SlugDellSystemCPLD,
		"sep firmware":                 SlugDellNonExpanderStorageBackplane,
		"lifecycle controller":         SlugDellLifeCycleController,
		"os collector":                 SlugDellOSCollector,
		"idrac service module":         SlugDellIdracServiceModel,
		"boss":                         SlugDellBossAdapter,
		"boss adapter":                 SlugDellBossAdapter,
		"disk 0 of boss adapter":       SlugDellBossAdapterDisk0,
		"dell 64 bit uefi diagnostics": SlugDell64bitUefiDiagnostics,
		"integrated dell remote access controller": SlugBMC,
	}
)

func DriveTypeSlug(m string) string {
	t, exists := modelDriveTypeSlug[m]
	if !exists {
		return "Unknown"
	}

	return t
}

// downcases and returns a normalized vendor name from the given string
func FormatVendorName(v string) string {
	switch v {
	case "Dell Inc.":
		return "dell"
	case "HP", "HPE":
		return "hp"
	case "Supermicro":
		return "supermicro"
	case "Quanta Cloud Technology Inc.":
		return "quanta"
	case "GIGABYTE":
		return "gigabyte"
	case "Intel Corporation":
		return "intel"
	case "Packet":
		return "packet"
	default:
		return v
	}
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
	case "Micron_5200_MTFDDAK480TDN":
		return "5200MAX"
	default:
		return s
	}
}

// Return the product vendor name, given a product name/model string
func VendorFromString(s string) string {
	switch {
	case strings.Contains(s, "LSI3008-IT"):
		return "LSI"
	case strings.Contains(s, "HGST "):
		return "HGST"
	case strings.Contains(s, "Micron_"):
		return "Micron"
	case strings.Contains(s, "TOSHIBA"):
		return "Toshiba"
	case strings.Contains(s, "ConnectX4LX"):
		return "Mellanox"
	default:
		return "unknown"
	}
}
