package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// The asrr bioscontrol utility depends on a custom vendor provided asrdev kernel module
// this module then depends on the host kernel version, which is a major kludge :|
//
// The module was built on Ubuntu 20.04.2 LTS 5.4.0-73-generic,
// the sources for this module were included in the BIOSControl_v1.0.3.zip provided by asrr.
//
// The s3/https endpoint to retrieve the sources for this utility and the kernel module
// can be found in the Dockerfile in the root of this repository.
//
// 🤞 heres to hoping ASRR releases a bioscontrol utility that does not depend on a custom kernel module.

const (
	// EnvAsrrBiosUtility - to override the utility path
	EnvAsrrBiosUtility = "UTIL_ASRR_BIOSCONTROL"

	// see the Dockerfile for how this bin ends up here
	asrrBiosUtility = "/usr/sbin/asrr-bioscontrol"

	// EnvAsrrKernelModule - to overide the kernel module path
	EnvAsrrKernelModule = "KERNEL_MODULE_ASRR"

	// see the Dockerfile for how this module ends up here
	asrrKernelModule = "/opt/asrr/asrdev-##VERSION##.ko"

	asrTmpBIOSConfigJSON = "/tmp/biosconfig-asrr.json"
)

// AsrrBiosControl is a asrr-bioscontrol executor
type AsrrBioscontrol struct {
	Executor    Executor
	tmpJSONFile string
}

// Return a new Asrr bios control utility executor
func NewAsrrBioscontrol(trace bool) *AsrrBioscontrol {
	utility := asrrBiosUtility

	envVar := os.Getenv(EnvAsrrBiosUtility)
	if envVar != "" {
		utility = envVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &AsrrBioscontrol{Executor: e, tmpJSONFile: asrTmpBIOSConfigJSON}
}

// loadAsrrBiosKernelModule loads the bioscontrol utility kernel module
func loadAsrrBiosKernelModule(ctx context.Context) error {
	// begin ick code

	// 1. identify kernel release
	unameExec := NewExecutor("uname")
	unameExec.SetArgs([]string{"-r"})

	r, err := unameExec.ExecWithContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error setting up asrr bios kernel module")
	}

	kernelVersion := bytes.TrimSpace(r.Stdout)

	// kernel module path - set from env if defined
	src := strings.Replace(asrrKernelModule, "##VERSION##", string(kernelVersion), 1)

	envVar := os.Getenv(EnvAsrrKernelModule)
	if envVar != "" {
		src = envVar
	}

	// 2. copy over kernel module
	dst := fmt.Sprintf("/lib/modules/%s/kernel/asrdev.ko", kernelVersion)

	err = copyFile(src, dst)
	if err != nil {
		return errors.Wrap(err, "error setting up asrr bios kernel module")
	}

	// 3. depmod to rebuild /lib/modules/$(uname -r)/modules.dep
	depmodExec := NewExecutor("depmod")
	depmodExec.SetArgs([]string{"-a"})

	_, err = depmodExec.ExecWithContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error setting up asrr bios kernel module")
	}

	// 4. load module
	modprobeExec := NewExecutor("modprobe")
	modprobeExec.SetArgs([]string{"asrdev"})

	_, err = modprobeExec.ExecWithContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error setting up asrr bios kernel module")
	}

	return nil
}

// GetBIOSConfiguration returns a BIOS configuration object
func (a *AsrrBioscontrol) GetBIOSConfiguration(ctx context.Context, deviceModel string) (map[string]string, error) {
	var cfg map[string]string

	// load kernel module
	err := loadAsrrBiosKernelModule(ctx)
	if err != nil {
		return nil, err
	}

	a.Executor.SetArgs([]string{"/g", a.tmpJSONFile})

	_, err = a.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	bytesJSON, err := ioutil.ReadFile(a.tmpJSONFile)
	if err != nil {
		return nil, err
	}

	cfg, err = asrrBiosConfigurationJSON(ctx, bytesJSON)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing bios config JSON")
	}

	return normalizeBIOSConfiguration(cfg), nil
}

type asrrBiosParam struct {
	Title       string      `json:"Title"`
	Value       uint64      `json:"Value"`
	ValidValues interface{} `json:"Valid Value"`
	ValueType   string      `json:"Value Type"`
}

// asrrBiosConfigurationJSON returns a map of BIOS settings and values
func asrrBiosConfigurationJSON(ctx context.Context, configBytes []byte) (map[string]string, error) {
	cfg := make(map[string]string)

	params := []*asrrBiosParam{}

	err := json.Unmarshal(configBytes, &params)
	if err != nil {
		return nil, err
	}

	for _, p := range params {
		// trim garbage characters
		key := strings.Replace(p.Title, "\x1b{a1#\x1b{f4#\x1b{w1125#", "", -1)
		key = strings.Replace(key, "\x1b{a1#", "", -1)
		key = strings.Trim(key, " ")

		base10 := 10
		value := strconv.FormatUint(p.Value, base10)

		validValue := value
		if p.ValidValues != nil {
			validValue = asrrBiosConfigValueTitle(value, p.ValueType, p.ValidValues)
		}

		_, exists := cfg[key]
		if exists {
			cfg["[dup]"+key] = validValue
			continue
		}

		cfg[key] = validValue
	}

	return cfg, nil
}

// asrrBiosConfigValueTitle returns the Title name of the BIOS attribute value
func asrrBiosConfigValueTitle(value, valueType string, validValues interface{}) string {
	for _, m := range validValues.([]interface{}) {
		// UINT8, UINT16 fields
		switch param := m.(type) {
		// slice item is map
		case map[string]interface{}:
			var v string

			switch value := param["Value"].(type) {
			case float64:
				bitSize := 32
				v = strconv.FormatFloat(value, 'f', -1, bitSize)
			case int:
				v = strconv.Itoa(value)
			default:
			}

			if v == value {
				return param["Title"].(string)
			}
		// BOOLEAN fields
		// slice item is a float64
		case float64:
			if valueType != "BOOLEAN" {
				continue
			}

			if value == "0" {
				return "Disabled"
			}

			return "Enabled"
		default:
		}
	}

	return value
}
