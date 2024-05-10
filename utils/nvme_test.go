package utils

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/bmc-toolbox/common"
	tlogrus "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NvmeComponents(t *testing.T) {
	expected := []*common.Drive{
		{Common: common.Common{
			Serial: "Z9DF70I8FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}, ProductName: "NULL",
			Metadata: map[string]string{
				"Block Erase Sanitize Operation Supported":                          "false",
				"Crypto Erase Applies to All/Single Namespace(s) (t:All, f:Single)": "false",
				"Crypto Erase Sanitize Operation Supported":                         "false",
				"Crypto Erase Supported as part of Secure Erase":                    "true",
				"Format Applies to All/Single Namespace(s) (t:All, f:Single)":       "false",
				"No-Deallocate After Sanitize bit in Sanitize command Supported":    "false",
				"Overwrite Sanitize Operation Supported":                            "false",
			},
		}},
		{Common: common.Common{
			Serial: "Z9DF70I9FY3L", Vendor: "TOSHIBA", Model: "KXG60ZNV256G TOSHIBA", Description: "KXG60ZNV256G TOSHIBA", Firmware: &common.Firmware{Installed: "AGGA4104"}, ProductName: "NULL",
			Metadata: map[string]string{
				"Block Erase Sanitize Operation Supported":                          "false",
				"Crypto Erase Applies to All/Single Namespace(s) (t:All, f:Single)": "false",
				"Crypto Erase Sanitize Operation Supported":                         "false",
				"Crypto Erase Supported as part of Secure Erase":                    "true",
				"Format Applies to All/Single Namespace(s) (t:All, f:Single)":       "false",
				"No-Deallocate After Sanitize bit in Sanitize command Supported":    "false",
				"Overwrite Sanitize Operation Supported":                            "false",
			},
		}},
	}

	n := NewFakeNvme()

	drives, err := n.Drives(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, drives)
}

func Test_NvmeDriveCapabilities(t *testing.T) {
	n := NewFakeNvme()

	d := &nvmeDeviceAttributes{DevicePath: "/dev/nvme0"}

	capabilities, err := n.DriveCapabilities(context.TODO(), d.DevicePath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fixtureNvmeDeviceCapabilities, capabilities)
}

var fixtureNvmeDeviceCapabilities = []*common.Capability{
	{Name: "fmns", Description: "Format Applies to All/Single Namespace(s) (t:All, f:Single)", Enabled: false},
	{Name: "cens", Description: "Crypto Erase Applies to All/Single Namespace(s) (t:All, f:Single)", Enabled: false},
	{Name: "cese", Description: "Crypto Erase Supported as part of Secure Erase", Enabled: true},
	{Name: "cer", Description: "Crypto Erase Sanitize Operation Supported", Enabled: false},
	{Name: "ber", Description: "Block Erase Sanitize Operation Supported", Enabled: false},
	{Name: "owr", Description: "Overwrite Sanitize Operation Supported", Enabled: false},
	{Name: "ndi", Description: "No-Deallocate After Sanitize bit in Sanitize command Supported", Enabled: false},
}

func Test_NvmeParseFna(t *testing.T) {
	// These are laid out so if you squint and pretend false/true are 0/1 they match the bit pattern of the int
	// Its a map so order doesn't matter but I think it makes it easier to match a broken test to the code
	wants := []map[string]bool{
		{"cese": false, "cens": false, "fmns": false},
		{"cese": false, "cens": false, "fmns": true},
		{"cese": false, "cens": true, "fmns": false},
		{"cese": false, "cens": true, "fmns": true},
		{"cese": true, "cens": false, "fmns": false},
		{"cese": true, "cens": false, "fmns": true},
		{"cese": true, "cens": true, "fmns": false},
		{"cese": true, "cens": true, "fmns": true},
	}
	for i, want := range wants {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			caps := parseFna(uint(i))
			require.Len(t, caps, len(want))
			for _, cap := range caps {
				require.Equal(t, want[cap.Name], cap.Enabled)
			}
		})
	}
}

func Test_NvmeParseSanicap(t *testing.T) {
	// These are laid out so if you squint and pretend false/true are 0/1 they match the bit pattern of the int
	// Its a map so order doesn't matter but I think it makes it easier to match a broken test to the code

	// lower bits only
	wants := []map[string]bool{
		{"owr": false, "ber": false, "cer": false},
		{"owr": false, "ber": false, "cer": true},
		{"owr": false, "ber": true, "cer": false},
		{"owr": false, "ber": true, "cer": true},
		{"owr": true, "ber": false, "cer": false},
		{"owr": true, "ber": false, "cer": true},
		{"owr": true, "ber": true, "cer": false},
		{"owr": true, "ber": true, "cer": true},
	}
	for i, want := range wants {
		// not testing ndi yet but its being returned
		// don't want to add it above to avoid noise
		want["ndi"] = false
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			caps, err := parseSanicap(uint(i))
			require.NoError(t, err)
			require.Len(t, caps, len(want))
			for _, cap := range caps {
				require.Equal(t, want[cap.Name], cap.Enabled)
			}
		})
	}

	// higher bits only
	wants = []map[string]bool{
		{"ndi": false},
		{"ndi": true},
		{"nodmmas": false, "ndi": false},
		{"nodmmas": false, "ndi": true},
		{"nodmmas": true, "ndi": false},
		{"nodmmas": true, "ndi": true},
	}
	for i, want := range wants {
		// not testing these now but they are being returned
		// don't want to add them above to avoid noise
		want["owr"] = false
		want["ber"] = false
		want["cer"] = false
		i = (i << 29)
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			caps, err := parseSanicap(uint(i))
			require.NoError(t, err)
			require.Len(t, caps, len(want))
			for _, cap := range caps {
				require.Equal(t, want[cap.Name], cap.Enabled)
			}
		})
	}

	i := 0b11 << 30
	t.Run(strconv.Itoa(i), func(t *testing.T) {
		caps, err := parseSanicap(uint(i))
		require.Error(t, err)
		require.Nil(t, caps)
	})
}

func fakeNVMEDevice(t *testing.T) string {
	dir := t.TempDir()
	f, err := os.Create(dir + "/nvme0n1")
	require.NoError(t, err)
	require.NoError(t, f.Truncate(20*1024))
	require.NoError(t, f.Close())
	return f.Name()
}

func Test_NvmeSanitize(t *testing.T) {
	for action := range CryptoErase {
		t.Run(action.String(), func(t *testing.T) {
			n := NewFakeNvme()
			dev := fakeNVMEDevice(t)
			err := n.Sanitize(context.Background(), dev, action)

			switch action { // nolint:exhaustive
			case BlockErase, CryptoErase:
				require.NoError(t, err)
				// FakeExecute is a bad mocker since it doesn't record all calls and sanitize-log calls aren't that interesting
				// TODO: Setup better mocks
				//
				// e, ok := n.Executor.(*FakeExecute)
				// require.True(t, ok)
				// require.Equal(t, []string{"sanitize", "--sanact=2", dev}, e.Args)
			default:
				require.Error(t, err)
				require.ErrorIs(t, err, errSanitizeInvalidAction)
			}
		})
	}
}

func Test_NvmeFormat(t *testing.T) {
	for action := range Reserved {
		t.Run(action.String(), func(t *testing.T) {
			n := NewFakeNvme()
			dev := fakeNVMEDevice(t)
			err := n.Format(context.Background(), dev, action)

			switch action { // nolint:exhaustive
			case UserDataErase, CryptographicErase:
				require.NoError(t, err)
				e, ok := n.Executor.(*FakeExecute)
				require.True(t, ok)
				require.Equal(t, []string{"format", "--ses=" + strconv.Itoa(int(action)), dev}, e.Args)
			default:
				require.Error(t, err)
				require.ErrorIs(t, err, errFormatInvalidSetting)
			}
		})
	}
}

func Test_NvmeWipe(t *testing.T) {
	tests := []struct {
		caps map[string]bool
		args []string
	}{
		{caps: map[string]bool{"ber": false, "cer": false, "cese": false}, args: []string{"format", "--ses=1"}},
		{caps: map[string]bool{"ber": false, "cer": false, "cese": true}, args: []string{"format", "--ses=2"}},
		{caps: map[string]bool{"ber": false, "cer": true, "cese": false}, args: []string{"sanitize", "--sanact=4"}},
		{caps: map[string]bool{"ber": false, "cer": true, "cese": true}, args: []string{"sanitize", "--sanact=4"}},
		{caps: map[string]bool{"ber": true, "cer": false, "cese": false}, args: []string{"sanitize", "--sanact=2"}},
		{caps: map[string]bool{"ber": true, "cer": false, "cese": true}, args: []string{"sanitize", "--sanact=2"}},
		{caps: map[string]bool{"ber": true, "cer": true, "cese": false}, args: []string{"sanitize", "--sanact=4"}},
		{caps: map[string]bool{"ber": true, "cer": true, "cese": true}, args: []string{"sanitize", "--sanact=4"}},
	}
	for _, test := range tests {
		name := fmt.Sprintf("ber=%v,cer=%v,cese=%v", test.caps["ber"], test.caps["cer"], test.caps["cese"])
		t.Run(name, func(t *testing.T) {
			caps := []*common.Capability{
				{Name: "ber", Enabled: test.caps["ber"]},
				{Name: "cer", Enabled: test.caps["cer"]},
				{Name: "cese", Enabled: test.caps["cese"]},
			}
			n := NewFakeNvme()
			dev := fakeNVMEDevice(t)
			logger, hook := tlogrus.NewNullLogger()
			defer hook.Reset()

			err := n.wipe(context.Background(), logger, dev, caps)
			require.NoError(t, err)

			// FakeExecute is a bad mocker since it doesn't record all calls and sanitize-log calls aren't that interesting
			// TODO: Setup better mocks
			if test.args[0] == "format" {
				e, ok := n.Executor.(*FakeExecute)
				require.True(t, ok)
				test.args = append(test.args, dev)
				require.Equal(t, test.args, e.Args)
			}
		})
	}
}
