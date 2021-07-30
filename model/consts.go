package model

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

const (
	VendorDell               = "dell"
	VendorMicron             = "micron"
	VendorAsrockrack         = "asrockrack"
	VendorSupermicro         = "supermicro"
	VendorHPE                = "hp"
	VendorQuanta             = "quanta"
	VendorGigabyte           = "gigabyte"
	VendorIntel              = "intel"
	VendorPacket             = "packet"
	VendorMellanox           = "mellanox"
	VendorAmericanMegatrends = "ami"

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

// IdentifySlug accepts a device component object and returns its matching slug
func IdentifySlug(component interface{}) string {
	switch component.(type) {
	case *BMC:
		return SlugBMC
	case *BIOS:
		return SlugBIOS
	case []*NIC:
		return SlugNICs
	case []*PSU:
		return SlugPSUs
	case []*Drive:
		return SlugDrives
	case []*StorageController:
		return SlugStorageControllers
	default:
		return SlugUnknown
	}
}

// nolint:gocyclo // type assert is cyclomatic
// GetComponentFirmware asserts the component type and returns the component []*firmware
func GetComponentFirmware(component interface{}) ([]*Firmware, error) {
	f := []*Firmware{}

	switch c := component.(type) {
	case *BMC:
		f = append(f, c.Firmware)
	case *BIOS:
		f = append(f, c.Firmware)
	case []*NIC:
		for _, e := range c {
			f = append(f, e.Firmware)
		}
	case []*PSU:
		for _, e := range c {
			f = append(f, e.Firmware)
		}
	case []*Drive:
		for _, e := range c {
			f = append(f, e.Firmware)
		}
	case []*StorageController:
		for _, e := range c {
			f = append(f, e.Firmware)
		}
	default:
		return nil, errors.Wrap(ErrTypeComponentFirmware, reflect.TypeOf(c).String())
	}

	return f, nil
}

// IsMultipleSlug returns bool if a given slug identifies as a component
// that is found in multiples - PSUs, NICs, Drives
func IsMultipleSlug(slug string) bool {
	m := map[string]bool{
		SlugDrives:             true,
		SlugNICs:               true,
		SlugPSUs:               true,
		SlugStorageControllers: true,
	}

	_, exists := m[slug]

	return exists
}

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
