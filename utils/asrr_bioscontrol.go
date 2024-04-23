package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/metal-toolbox/ironlib/model"
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
	EnvAsrrBiosUtility = "IRONLIB_UTIL_ASRR_BIOSCONTROL"

	// see the Dockerfile for how this bin ends up here
	asrrBiosUtility = "/usr/sbin/asrr-bioscontrol"

	// EnvAsrrKernelModule - to overide the kernel module path
	EnvAsrrKernelModule = "KERNEL_MODULE_ASRR"

	// see the Dockerfile for how this module ends up here
	asrrKernelModule = "/opt/asrr/asrdev-%s.ko"

	asrTmpBIOSConfigJSON = "/tmp/biosconfig-asrr.json"
)

var ErrASRRBIOSKernelModule = errors.New("error loading asrr bios kernel module")

// AsrrBiosControl is a asrr-bioscontrol executor
type AsrrBioscontrol struct {
	Executor    Executor
	tmpJSONFile string
}

// NewAsrrBioscontrol returns a new Asrr bios control utility executor
func NewAsrrBioscontrol(trace bool) *AsrrBioscontrol {
	utility := asrrBiosUtility

	if eVar := os.Getenv(EnvAsrrBiosUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &AsrrBioscontrol{Executor: e, tmpJSONFile: asrTmpBIOSConfigJSON}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (a *AsrrBioscontrol) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := a.Executor.CheckExecutable()

	return "asrr-bioscontrol", a.Executor.CmdPath(), er
}

// kernelVersion returns the host kernel version
func kernelVersion(ctx context.Context) (string, error) {
	unameExec := NewExecutor("uname")
	unameExec.SetArgs([]string{"-r"})
	unameExec.SetQuiet()

	r, err := unameExec.ExecWithContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "error executing uname -r")
	}

	return string(r.Stdout), nil
}

// kernelModuleLoaded returns bool if the given module name is loaded
func kernelModuleLoaded(ctx context.Context, name string) (bool, error) {
	lsmodExec := NewExecutor("lsmod")
	lsmodExec.SetQuiet()

	r, err := lsmodExec.ExecWithContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "error executing lsmod")
	}

	if bytes.Contains(r.Stdout, []byte(name)) {
		return true, nil
	}

	return false, nil
}

// loadAsrrBiosKernelModule loads the bioscontrol utility kernel module
func loadAsrrBiosKernelModule(ctx context.Context) error {
	// begin ick code
	// 1. identify kernel release
	kernelVersion, err := kernelVersion(ctx)
	if err != nil {
		return errors.Wrap(ErrASRRBIOSKernelModule, err.Error())
	}

	// 2. figure if the kernel module is loaded
	isLoaded, err := kernelModuleLoaded(ctx, "asrdev")
	if err != nil {
		return errors.Wrap(ErrASRRBIOSKernelModule, err.Error())
	}

	if isLoaded {
		return nil
	}

	// kernel module path - set from env if defined
	src := os.Getenv(EnvAsrrKernelModule)
	if src == "" {
		src = fmt.Sprintf(asrrKernelModule, kernelVersion)
	}

	// 2. copy over kernel module
	dst := fmt.Sprintf("/lib/modules/%s/kernel/asrdev.ko", kernelVersion)

	err = copyFile(src, dst)
	if err != nil {
		return errors.Wrap(err, ErrASRRBIOSKernelModule.Error())
	}

	// 3. depmod to rebuild /lib/modules/$(uname -r)/modules.dep
	depmodExec := NewExecutor("depmod")
	depmodExec.SetArgs([]string{"-a"})
	depmodExec.SetQuiet()

	_, err = depmodExec.ExecWithContext(ctx)
	if err != nil {
		return errors.Wrap(err, ErrASRRBIOSKernelModule.Error())
	}

	// 4. load module
	modprobeExec := NewExecutor("modprobe")
	modprobeExec.SetArgs([]string{"asrdev"})
	modprobeExec.SetQuiet()

	_, err = modprobeExec.ExecWithContext(ctx)
	if err != nil {
		return errors.Wrap(err, ErrASRRBIOSKernelModule.Error())
	}

	return nil
}

// GetBIOSConfiguration returns a BIOS configuration object
func (a *AsrrBioscontrol) GetBIOSConfiguration(ctx context.Context, _ string) (map[string]string, error) {
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

	bytesJSON, err := os.ReadFile(a.tmpJSONFile)
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
func asrrBiosConfigurationJSON(_ context.Context, configBytes []byte) (map[string]string, error) {
	cfg := make(map[string]string)

	params := []*asrrBiosParam{}

	err := json.Unmarshal(configBytes, &params)
	if err != nil {
		return nil, err
	}

	for _, p := range params {
		// trim garbage characters - most likely terminal colors for the parameter titles
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
	values, ok := validValues.([]interface{})
	if !ok {
		panic(fmt.Sprintf("validValues is an unexpected type...%v", validValues))
	}
	for _, m := range values {
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
				title, ok := param["Title"].(string)
				if !ok {
					panic(fmt.Sprintf("Title is not a string...%v", param["Title"]))
				}
				return title
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
