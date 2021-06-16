package config

// Notes on adding new BIOS configuration pararmeters
// - Ensure each field for a BIOS parameter is as close to the actual name

// BIOSConfiguration holds BIOS configuration for each vendor
type BIOSConfiguration struct {
	Dell       *DellBIOS       `json:"dell,omitempty"`
	Supermicro *SupermicroBIOS `json:"supermicro,omitempty"`
}

// DellBIOS is an instance with Dell configuration parameters
type DellBIOS struct {
	BootMode       string `json:"boot_mode"` // UEFI/BIOS
	AMDSev         int64  `json:"cpu_min_sev_asid,omitempty"`
	Hyperthreading string `json:"logical_proc"`
	SRIOV          string `json:"sriov_global_enable,omitempty"`
	TPM            string `json:"tpm_security,omitempty"`
}

// SupermicroBIOS is SMC BIOS parameter object
type SupermicroBIOS struct {
	BootMode       string `json:"boot_mode"`       // UEFI/LEGACY/DUAL
	Hyperthreading string `json:"hyper_threading"` // enabled/disabled
	TPM            string `json:"tpm,omitempty"`   // enabled/disabled
	SecureBoot     string `json:"secure_boot"`     // enabled/disabled
	IntelSGX       string `json:"intel_sgx,omitempty"`
}
