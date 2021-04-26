package config

// Notes on adding new BIOS configuration pararmeters
// - Ensure each field for a BIOS parameter is as close to the actual name

// BIOSConfiguration holds BIOS configuration for each vendor
type BIOSConfiguration struct {
	Dell *DellBIOS `json:"dell"`
}

// DellBIOS is an instance with Dell configuration parameters
type DellBIOS struct {
	BootMode       string `json:"boot_mode"`
	AMDSev         int64  `json:"cpu_min_sev_asid"`
	Hyperthreading string `json:"logical_proc"`
	SRIOV          string `json:"sriov_global_enable"`
	TPM            string `json:"tpm_security"`
}
