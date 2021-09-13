package utils

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_asrrBiosConfigurationJSON(t *testing.T) {
	jsonFile := "../fixtures/asrr/e3c246d4i-nl/bios.json"

	samples := map[string]string{
		"Device power-up delay": "Auto",
		"IUER Dock Enable":      "Disabled",
		"Control Logic 2":       "Disabled",
		"MachineCheck":          "Enabled",
		"AddOn ROM Display":     "Enabled",
		"Sata Port 0":           "Disabled",
	}

	b, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Error(err)
	}

	cfg, err := asrrBiosConfigurationJSON(context.TODO(), b)
	if err != nil {
		t.Error(err)
	}

	for k, v := range samples {
		assert.Equal(t, v, cfg[k])
	}
}

func Test_asrrBiosConfigValueTitle(t *testing.T) {
	type tester struct {
		value       string
		valueType   string
		expected    string
		testName    string
		validValues interface{}
	}

	tests := []*tester{
		{
			testName:  "Title for slice of bools - value disabled",
			value:     "0",
			valueType: "BOOLEAN",
			expected:  "Disabled",
			validValues: []interface{}{
				float64(0),
				float64(1),
			},
		},
		{
			testName:  "Title for slice of bools - value enabled",
			value:     "1",
			valueType: "BOOLEAN",
			expected:  "Enabled",
			validValues: []interface{}{
				float64(0),
				float64(1),
			},
		},
		{
			testName:  "Title for float64 map value disabled",
			value:     "0",
			valueType: "UINT8",
			expected:  "Disabled",
			validValues: []interface{}{
				map[string]interface{}{
					"Title": "Disabled",
					"Value": float64(0),
				},
				map[string]interface{}{
					"Title": "Enabled",
					"Value": float64(1),
				},
			},
		},
		{
			testName:  "Title for float64 map value enabled",
			value:     "1",
			valueType: "UINT8",
			expected:  "Enabled",
			validValues: []interface{}{
				map[string]interface{}{
					"Title": "Disabled",
					"Value": float64(0),
				},
				map[string]interface{}{
					"Title": "Enabled",
					"Value": float64(1),
				},
			},
		},
		{
			testName:  "Title for float64 map value other",
			value:     "3",
			valueType: "UINT8",
			expected:  "C1 and C3",
			validValues: []interface{}{
				map[string]interface{}{
					"Title": "C1",
					"Value": float64(0),
				},
				map[string]interface{}{
					"Title": "C2",
					"Value": float64(1),
				},
				map[string]interface{}{
					"Title": "C3",
					"Value": float64(2),
				},
				map[string]interface{}{
					"Title": "C1 and C3",
					"Value": float64(3),
				},
			},
		},
	}

	for _, tt := range tests {
		got := asrrBiosConfigValueTitle(tt.value, tt.valueType, tt.validValues)
		assert.Equal(t, tt.expected, got, tt.testName)
	}
}
