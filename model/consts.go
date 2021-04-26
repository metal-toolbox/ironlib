package model

const (
	// Vendors
	VendorDell       = "dell"
	VendorMicron     = "micron"
	VendorAsrockrack = "asrockrack"
	VendorSupermicro = "supermicro"

	// Generic component slugs
	SlugBackplaneExpander           = "Backplane Expander"
	SlugBMC                         = "BMC"
	SlugBIOS                        = "BIOS"
	SlugDisk                        = "disk"
	SlugDiskPciNvmeSsd              = "Disk - NVME PCI SSD"
	SlugDiskSataSsd                 = "Disk - Sata SSD"
	SlugNIC                         = "NIC"
	SlugPowerSupply                 = "Power Supply"
	SlugSasHba330Controller         = "SAS HBA330 Controller"
	SlugCPLD                        = "CPLD"
	SlugNonExpanderStorageBackplane = "Non-Expander Storage Backplane (SEP)"

	// Dell component slugs
	SlugDellIdracServiceModel    = "IDrac Service Module"
	SlugDellBossAdapter          = "Boss Adapter"
	SlugDellLifeCycleController  = "Lifecycle Controller"
	SlugDellOSCollector          = "OS Collector"
	SlugDell64bitUefiDiagnostics = "Dell 64 bit uEFI diagnostics"
)
