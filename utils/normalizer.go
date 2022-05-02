package utils

import "strings"

const (
	enabledValue  = "Enabled"
	disabledValue = "Disabled"
)

// nolint:gocyclo // going through all bios values to standardize them is going to be high complexity
func normalizeBIOSConfiguration(cfg map[string]string) map[string]string {
	normalizedCfg := make(map[string]string)

	for k, v := range cfg {
		nV := normalizeValue(v)

		switch k {
		case "CpuMinSevAsid":
			normalizedCfg["amd_sev"] = v
		case "BootMode":
			normalizedCfg["boot_mode"] = normalizeBootMode(v)
		case "Boot mode select":
			normalizedCfg["boot_mode"] = normalizeBootMode(v)
		case "IntelTxt":
			normalizedCfg["intel_txt"] = nV
		case "Software Guard Extensions (SGX)":
			normalizedCfg["intel_sgx"] = nV
		case "SecureBoot":
			normalizedCfg["secure_boot"] = nV
		case "Secure Boot":
			normalizedCfg["secure_boot"] = nV
		case "Hyper-Threading":
			normalizedCfg["smt"] = nV
		case "Hyper-Threading [ALL]":
			normalizedCfg["smt"] = nV
		case "LogicalProc":
			normalizedCfg["smt"] = nV
		case "SriovGlobalEnable":
			normalizedCfg["sr_iov"] = nV
		case "TpmSecurity":
			normalizedCfg["tpm"] = nV
		case "Security Device Support":
			normalizedCfg["tpm"] = nV
		// we want to drop these values
		case "NewSetupPassword":
		case "NewSysPassword":
		case "OldSetupPassword":
		case "OldSysPassword":
		default:
			// When we don't normalize the value append "raw:" to the value
			normalizedCfg["raw:"+k] = nV
		}
	}

	return normalizedCfg
}

func normalizeValue(v string) string {
	switch strings.ToLower(v) {
	case "disable":
		return disabledValue
	case "disabled":
		return disabledValue
	case "enable":
		return enabledValue
	case "enabled":
		return enabledValue
	case "off":
		return disabledValue
	case "on":
		return enabledValue
	default:
		return v
	}
}

func normalizeBootMode(v string) string {
	switch strings.ToLower(v) {
	case "legacy":
		return "BIOS"
	default:
		return strings.ToUpper(v)
	}
}
