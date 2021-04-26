package config

// Notes on adding new BIOS configuration pararmeters
// - Ensure each field for a BIOS parameter is as close to the actual name

// BIOSConfiguration holds BIOS configuration for each vendor
type BIOSConfiguration struct {
	Dell *DellBIOS `json:"dell"`
}

// DellBIOS is an instance with Dell configuration parameters
type DellBIOS struct {
	BootMode    string `json:"boot_mode"`
	TPMSecurity string `json:"tpm_security"`
}
