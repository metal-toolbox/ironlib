package utils

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	fixtureTPMPropertiesFixed = map[string]string{
		"TPM2_PT_ACTIVE_SESSIONS_MAX": "raw:0x40",
		"TPM2_PT_CLOCK_UPDATE":        "raw:0x400000",
		"TPM2_PT_CONTEXT_GAP_MAX":     "raw:0xFFFF",
		"TPM2_PT_CONTEXT_HASH":        "raw:0xB",
		"TPM2_PT_CONTEXT_SYM":         "raw:0x6",
		"TPM2_PT_CONTEXT_SYM_SIZE":    "raw:0x80",
		"TPM2_PT_DAY_OF_YEAR":         "raw:0xA7",
		"TPM2_PT_FAMILY_INDICATOR":    "2.0",
		"TPM2_PT_FIRMWARE_VERSION_1":  "raw:0x10003",
		"TPM2_PT_FIRMWARE_VERSION_2":  "raw:0x20008",
		"TPM2_PT_HR_LOADED_MIN":       "raw:0x3",
		"TPM2_PT_HR_PERSISTENT_MIN":   "raw:0x7",
		"TPM2_PT_HR_TRANSIENT_MIN":    "raw:0x3",
		"TPM2_PT_INPUT_BUFFER":        "raw:0x400",
		"TPM2_PT_LEVEL":               "raw:0",
		"TPM2_PT_LIBRARY_COMMANDS":    "raw:0x65",
		"TPM2_PT_MANUFACTURER":        "NTC",
		"TPM2_PT_MAX_COMMAND_SIZE":    "raw:0x800",
		"TPM2_PT_MAX_DIGEST":          "raw:0x20",
		"TPM2_PT_MAX_OBJECT_CONTEXT":  "raw:0x392",
		"TPM2_PT_MAX_RESPONSE_SIZE":   "raw:0x800",
		"TPM2_PT_MAX_SESSION_CONTEXT": "raw:0xE9",
		"TPM2_PT_MEMORY":              "raw:0x6",
		"TPM2_PT_NV_BUFFER_MAX":       "raw:0x400",
		"TPM2_PT_NV_COUNTERS_MAX":     "raw:0x10",
		"TPM2_PT_NV_INDEX_MAX":        "raw:0x800",
		"TPM2_PT_ORDERLY_COUNT":       "raw:0xFF",
		"TPM2_PT_PCR_COUNT":           "raw:0x18",
		"TPM2_PT_PCR_SELECT_MIN":      "raw:0x3",
		"TPM2_PT_PS_DAY_OF_YEAR":      "raw:0x0",
		"TPM2_PT_PS_FAMILY_INDICATOR": "raw:0x1",
		"TPM2_PT_PS_LEVEL":            "raw:0x0",
		"TPM2_PT_PS_REVISION":         "raw:0x100",
		"TPM2_PT_PS_YEAR":             "raw:0x0",
		"TPM2_PT_REVISION":            "1.16",
		"TPM2_PT_SPLIT_MAX":           "raw:0x80",
		"TPM2_PT_TOTAL_COMMANDS":      "raw:0x6B",
		"TPM2_PT_VENDOR_COMMANDS":     "raw:0x6",
		"TPM2_PT_VENDOR_STRING_1":     "rls",
		"TPM2_PT_VENDOR_STRING_2":     "NPCT",
		"TPM2_PT_VENDOR_STRING_3":     "",
		"TPM2_PT_VENDOR_STRING_4":     "",
		"TPM2_PT_VENDOR_TPM_TYPE":     "raw:0x1",
		"TPM2_PT_YEAR":                "raw:0x7DF",
	}

	fixtureTPMPropertiesVariable = map[string]string{
		"TPM2_PT_ALGORITHM_SET":                "0x1",
		"TPM2_PT_AUDIT_COUNTER_0":              "0x0",
		"TPM2_PT_AUDIT_COUNTER_1":              "0x0",
		"TPM2_PT_HR_ACTIVE":                    "0x0",
		"TPM2_PT_HR_ACTIVE_AVAIL":              "0x40",
		"TPM2_PT_HR_LOADED":                    "0x0",
		"TPM2_PT_HR_LOADED_AVAIL":              "0x3",
		"TPM2_PT_HR_NV_INDEX":                  "0x8",
		"TPM2_PT_HR_PERSISTENT":                "0x2",
		"TPM2_PT_HR_PERSISTENT_AVAIL":          "0x11",
		"TPM2_PT_HR_TRANSIENT_AVAIL":           "0x3",
		"TPM2_PT_LOADED_CURVES":                "0x2",
		"TPM2_PT_LOCKOUT_COUNTER":              "0x0",
		"TPM2_PT_LOCKOUT_INTERVAL":             "0x1C20",
		"TPM2_PT_LOCKOUT_RECOVERY":             "0x15180",
		"TPM2_PT_MAX_AUTH_FAIL":                "0xA",
		"TPM2_PT_NV_COUNTERS":                  "0x0",
		"TPM2_PT_NV_COUNTERS_AVAIL":            "0x10",
		"TPM2_PT_NV_WRITE_RECOVERY":            "0x0",
		"TPM2_PT_PERMANENT.disableClear":       "0",
		"TPM2_PT_PERMANENT.endorsementAuthSet": "0",
		"TPM2_PT_PERMANENT.inLockout":          "0",
		"TPM2_PT_PERMANENT.lockoutAuthSet":     "0",
		"TPM2_PT_PERMANENT.ownerAuthSet":       "0",
		"TPM2_PT_PERMANENT.reserved1":          "0",
		"TPM2_PT_PERMANENT.reserved2":          "0",
		"TPM2_PT_PERMANENT.tpmGeneratedEPS":    "1",
		"TPM2_PT_STARTUP_CLEAR.ehEnable":       "1",
		"TPM2_PT_STARTUP_CLEAR.orderly":        "1",
		"TPM2_PT_STARTUP_CLEAR.phEnable":       "1",
		"TPM2_PT_STARTUP_CLEAR.phEnableNV":     "1",
		"TPM2_PT_STARTUP_CLEAR.reserved1":      "0",
		"TPM2_PT_STARTUP_CLEAR.shEnable":       "1",
	}
)

func Test_GetCapPropertiesFixed(t *testing.T) {
	b, err := os.ReadFile("../fixtures/utils/tpm2_getcap/properties-fixed")
	if err != nil {
		t.Error(err)
	}

	cli, err := NewFakeTpm2Utils(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	props, err := cli.getCapPropertiesFixed(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, fixtureTPMPropertiesFixed, props)
}

func Test_GetCapPropertiesVariable(t *testing.T) {
	b, err := os.ReadFile("../fixtures/utils/tpm2_getcap/properties-variable")
	if err != nil {
		t.Error(err)
	}

	cli, err := NewFakeTpm2Utils(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	props, err := cli.getCapPropertiesVariable(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, fixtureTPMPropertiesVariable, props)
}
