package utils

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
)

const EnvTPM2Utility = "IRONLIB_UTIL_TPM2_GETCAP"

type Tpm2Util struct {
	Executor Executor
}

// Return a new tpm2_getcap executor
func NewTpm2GetCapCmd(trace bool) *Tpm2Util {
	utility := "tpm2_getcap"

	// lookup env var for util
	if eVar := os.Getenv(EnvTPM2Utility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Tpm2Util{Executor: e}
}

// Executes nvme list, parses the output and returns a slice of *common.TPM
//func (t *Tpm2Util) TPMs(ctx context.Context) ([]*common.TPM, error) {
//	drives := make([]*common.TPM, 0)
//
//	out, err := t.getCapPropertiesFixed(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	list := &nvmeList{Devices: []*nvmeDeviceAttributes{}}
//
//	err = json.Unmarshal(out, list)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, d := range list.Devices {
//		dModel := d.ModelNumber
//
//		var vendor string
//
//		modelTokens := strings.Split(d.ModelNumber, " ")
//
//		if len(modelTokens) > 1 {
//			vendor = modelTokens[1]
//		}
//
//		drive := &common.Drive{
//			Common: common.Common{
//				Serial:      d.SerialNumber,
//				Vendor:      vendor,
//				Model:       dModel,
//				ProductName: d.ProductName,
//				Description: d.ModelNumber,
//				Firmware: &common.Firmware{
//					Installed: d.Firmware,
//				},
//				Metadata: map[string]string{},
//			},
//		}
//
//		// Collect drive capabilitiesFound
//		capabilitiesFound, err := n.DriveCapabilities(ctx, d.DevicePath)
//		if err != nil {
//			return nil, err
//		}
//
//		for _, f := range capabilitiesFound {
//			drive.Common.Metadata[f.Description] = strconv.FormatBool(f.Enabled)
//		}
//
//		drives = append(drives, drive)
//	}
//
//	return drives, nil
//}

type tpm2PropertiesFixed struct {
	Firmware     string
	Revision     string
	Manufacturer string
	VendorName   string
	SerialNumber string
}

func (t *Tpm2Util) getCapPropertiesFixed(ctx context.Context) (map[string]string, error) {
	// tpm2_getcap properties-fixed
	t.Executor.SetArgs([]string{"tpm2_getcap", "properties-fixed"})

	result, err := t.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	var key string
	properties := map[string]string{}
	lines := bytes.Split(result.Stdout, []byte("\n"))
	for _, line := range lines {
		str := string(line)

		if strings.Contains(str, "TPM2_PT") {
			key = strings.Replace(str, ":", "", -1)

			continue
		}

		switch {
		case strings.Contains(str, "value:"):
			parts := strings.Split(string(line), ":")
			// nolint:gomnd // split into 2 parts - value: "NTC"
			if len(parts) < 2 {
				continue
			}

			value := strings.Replace(parts[1], "\"", "", -1)
			properties[key] = strings.TrimSpace(value)

		case strings.Contains(str, "raw:"):
			parts := strings.Split(str, ":")
			// nolint:gomnd // raw: 0x20000000
			if len(parts) < 2 {
				continue
			}

			properties[key] = "raw:" + strings.TrimSpace(parts[1])
		}

	}

	return properties, nil
}

func (t *Tpm2Util) getCapPropertiesVariable(ctx context.Context) (map[string]string, error) {
	// tpm2_getcap properties-fixed
	t.Executor.SetArgs([]string{"tpm2_getcap", "properties-variable"})

	result, err := t.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	const propPermanant = "TPM2_PT_PERMANENT"

	const propStartupClear = "TPM2_PT_STARTUP_CLEAR"

	var key string

	properties := map[string]string{}

	lines := bytes.Split(result.Stdout, []byte("\n"))
	for _, line := range lines {
		str := string(line)
		if strings.Contains(str, propPermanant) {
			key = propPermanant
			continue
		}

		if strings.Contains(str, propStartupClear) {
			key = propStartupClear
			continue
		}

		// other TPM2_PT_* attributes
		if strings.HasPrefix(str, "TPM2_PT") {
			parts := strings.Split(str, ":")
			if len(parts) < 2 {
				continue
			}

			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])

			properties[k] = v

			continue
		}

		switch key {
		case propPermanant:
			parts := strings.Split(str, ":")
			if len(parts) < 2 {
				continue
			}

			subkey := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			properties[propPermanant+"."+subkey] = value

		case propStartupClear:
			parts := strings.Split(str, ":")
			if len(parts) < 2 {
				continue
			}

			subkey := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			properties[propStartupClear+"."+subkey] = value

		}

	}

	return properties, nil
}

// Return a Fake tpm2 utils executor for tests
func NewFakeTpm2Utils(r io.Reader) (*Tpm2Util, error) {
	e := NewFakeExecutor("tpm2_getcap")
	b := bytes.Buffer{}

	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	e.SetStdout(b.Bytes())

	return &Tpm2Util{
		Executor: e,
	}, nil
}
