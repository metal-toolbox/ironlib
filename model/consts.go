package model

import (
	"strings"

	"github.com/pkg/errors"
)

const (
	VendorDell                  = "dell"
	VendorMicron                = "micron"
	VendorAsrockrack            = "asrockrack"
	VendorSupermicro            = "supermicro"
	VendorHPE                   = "hp"
	VendorQuanta                = "quanta"
	VendorGigabyte              = "gigabyte"
	VendorIntel                 = "intel"
	VendorPacket                = "packet"
	VendorMellanox              = "mellanox"
	VendorAmericanMegatrends    = "ami"
	VendorBroadcom              = "broadcom"
	SystemManufacturerUndefined = "To Be Filled By O.E.M."

	// Generic component slugs
	// NOTE: when adding slugs here, if the are a multiple -
	SlugBackplaneExpander     = "Backplane Expander"
	SlugChassis               = "Chassis"
	SlugTPM                   = "TPM"
	SlugGPU                   = "GPU"
	SlugCPU                   = "CPU"
	SlugPhysicalMem           = "PhysicalMemory"
	SlugStorageController     = "StorageController"
	SlugStorageControllers    = "StorageControllers"
	SlugBMC                   = "BMC"
	SlugBIOS                  = "BIOS"
	SlugDrive                 = "Drive"
	SlugDrives                = "Drives"
	SlugDriveTypePCIeNVMEeSSD = "NVMe PCIe SSD"
	SlugDriveTypeSATASSD      = "Sata SSD"
	SlugDriveTypeSATAHDD      = "Sata HDD"
	SlugNIC                   = "NIC"
	SlugNICs                  = "NICs"
	SlugPSU                   = "Power Supply"
	SlugPSUs                  = "Power Supplies"
	SlugSASHBA330Controller   = "SAS HBA330 Controller"
	SlugCPLD                  = "CPLD"
	SlugUnknown               = "unknown"

	// Dell specific component slugs
	SlugDellSystemCPLD                  = "Dell System CPLD"
	SlugDellBossAdapter                 = "Boss Adapter"
	SlugDellIdracServiceModule          = "IDrac Service Module"
	SlugDellBossAdapterDisk0            = "Boss Adapter - Disk 0"
	SlugDellBossAdapterDisk1            = "Boss Adapter - Disk 1"
	SlugDellLifeCycleController         = "Lifecycle Controller"
	SlugDellOSCollector                 = "OS Collector"
	SlugDell64bitUefiDiagnostics        = "Dell 64 bit uEFI diagnostics"
	SlugDellBackplaneExpander           = "Backplane Expander"
	SlugDellNonExpanderStorageBackplane = "Non-Expander Storage Backplane (SEP)"

	// EnvDellDSURelease is the Dell DSU release version
	//
	// e.g: 21.11.12 from https://linux.dell.com/repo/hardware/DSU_21.11.12/
	EnvDellDSURelease = "DELL_DSU_RELEASE"
	// EnvDellDSUVersion is the Dell DSU utility package version
	//
	// e.g: 1.9.2.0-21.07.00 from https://linux.dell.com/repo/hardware/DSU_21.11.12/os_independent/x86_64/dell-system-update-1.9.2.0-21.07.00.x86_64.rpm
	EnvDellDSUVersion = "DELL_DSU_VERSION"
	// 	EnvDNFDisableGPGCheck disables GPG checks in DNF package installs
	EnvDNFDellDisableGPGCheck = "DNF_DISABLE_GPG_CHECK"
	// EnvUpdateStoreURL defines up the update store base URL prefix
	EnvUpdateBaseURL = "UPDATE_BASE_URL"
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
		// Sata HDD drives
		"HGST HUS728T8TALE6L4": SlugDriveTypeSATAHDD,
	}

	// OemComponentDell is a lookup table for dell OEM components
	// these components are specific to the OEMs - in this case Dell
	OemComponentDell = map[string]struct{}{
		SlugDellSystemCPLD:                  {},
		SlugBackplaneExpander:               {},
		SlugDellIdracServiceModule:          {},
		SlugDellBossAdapterDisk0:            {},
		SlugDellBossAdapterDisk1:            {},
		SlugDellBossAdapter:                 {},
		SlugDellLifeCycleController:         {},
		SlugDellNonExpanderStorageBackplane: {},
		SlugDellOSCollector:                 {},
		SlugDell64bitUefiDiagnostics:        {},
	}

	// DellComponentSlug is an ordered list of of dell component identifiers to component slug
	// To identify components correctly, if two components contain a similar string
	// e.g: "idrac", "dell emc idrac service module" the latter should be positioned before the former in the list.
	DellComponentSlug = [][]string{
		{"bios", SlugBIOS},
		{"ethernet", SlugNIC},
		{"dell emc idrac service module", SlugDellIdracServiceModule},
		{"idrac", SlugBMC},
		{"backplane", SlugBackplaneExpander},
		{"power supply", SlugPSU},
		{"hba330", SlugStorageController},
		{"nvmepcissd", SlugDrive},
		{"system cpld", SlugDellSystemCPLD},
		{"sep firmware", SlugDellNonExpanderStorageBackplane},
		{"lifecycle controller", SlugDellLifeCycleController},
		{"os collector", SlugDellOSCollector},
		{"disk 0 of boss adapter", SlugDellBossAdapterDisk0},
		{"disk 1 of boss adapter", SlugDellBossAdapterDisk1},
		{"boss", SlugDellBossAdapter},
		{"dell 64 bit uefi diagnostics", SlugDell64bitUefiDiagnostics},
		{"integrated dell remote access controller", SlugBMC},
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

// downcases and returns a normalized vendor name from the given string
func FormatVendorName(v string) string {
	switch v {
	case "ASRockRack":
		return VendorAsrockrack
	case "Dell Inc.":
		return VendorDell
	case "HP", "HPE":
		return VendorHPE
	case "Supermicro":
		return VendorSupermicro
	case "Quanta Cloud Technology Inc.":
		return VendorQuanta
	case "GIGABYTE":
		return VendorGigabyte
	case "Intel Corporation":
		return VendorIntel
	case "Packet":
		return VendorPacket
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
	case "PowerEdge C6320":
		return "c6320"
	case "Micron_5200_MTFDDAK480TDN":
		return "5200MAX"
	default:
		return s
	}
}

// Return the product vendor name, given a product name/model string
func VendorFromString(s string) string {
	switch {
	case strings.Contains(s, "Dell"):
		return VendorDell
	case strings.Contains(s, "LSI3008-IT"):
		return "LSI"
	case strings.Contains(s, "HGST "):
		return "HGST"
	case strings.Contains(s, "Micron_"), strings.HasPrefix(s, "MTFD"):
		return "Micron"
	case strings.Contains(s, "TOSHIBA"):
		return "Toshiba"
	case strings.Contains(s, "ConnectX4LX"):
		return "Mellanox"
	default:
		return ""
	}
}
