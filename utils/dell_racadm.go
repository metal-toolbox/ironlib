package utils

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/bmc-toolbox/common"
	"github.com/bmc-toolbox/common/config"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const DellRacadmPath = "/opt/dell/srvadmin/bin/idracadm7"
const EnvVarRacadm7 = "IRONLIB_UTIL_RACADM7"

var ErrDellBiosCfgNil = errors.New("expected valid bios config object, got nil")
var ErrDellBiosCfgFileUndefined = errors.New("no BIOS config file defined")
var ErrDellBiosCfgFileEmpty = errors.New("BIOS config file empty or invalid")
var ErrDellBiosCfgFormatUnknown = errors.New("BIOS config file format unknown")

// DellRacadm is a dell racadm executor
type DellRacadm struct {
	Executor       Executor
	ConfigJSON     string
	BIOSCfgTmpFile string // where we dump the BIOS config to before processing it
	KeepConfigFile bool   // flag to keep the BIOS config file generated (mainly for testing)
	ShutdownType   string // Graceful, Forced or NoReboot
}

type DellRacadmOption func(r *DellRacadm)

// Return a new Dell racadm command executor
func NewDellRacadm(trace bool, options ...DellRacadmOption) *DellRacadm {
	racadmUtil := os.Getenv(EnvVarRacadm7)
	if racadmUtil == "" {
		racadmUtil = DellRacadmPath
	}

	e := NewExecutor(racadmUtil)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	r := &DellRacadm{
		Executor:       e,
		BIOSCfgTmpFile: "/tmp/bioscfg",
		ShutdownType:   "Graceful"}

	for _, opt := range options {
		opt(r)
	}

	return r
}

// Attributes implements the actions.UtilAttributeGetter interface
func (s *DellRacadm) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := s.Executor.CheckExecutable()

	return "dell-racadm", s.Executor.CmdPath(), er
}

// GetBIOSConfiguration returns a BIOS configuration object
func (s *DellRacadm) GetBIOSConfiguration(ctx context.Context, deviceModel string) (map[string]string, error) {
	var cfg map[string]string

	var err error

	// validate config we're reading from file is not empty
	if s.BIOSCfgTmpFile == "" {
		return nil, ErrDellBiosCfgFileUndefined
	}

	// older hardware return BIOS config as XML
	if strings.EqualFold(deviceModel, "c6320") {
		cfg, err = s.racadmBIOSConfigXML(ctx)
	} else {
		cfg, err = s.racadmBIOSConfigJSON(ctx)
	}

	if err != nil {
		return nil, err
	}

	if cfg == nil {
		return nil, ErrDellBiosCfgNil
	}

	return normalizeBIOSConfiguration(cfg), nil
}

// SetBIOSConfiguration takes a map of BIOS configurtation values and applies them to the host
func (s *DellRacadm) SetBIOSConfiguration(ctx context.Context, vendorOptions, cfg map[string]string) error {
	// older hardware return BIOS config as XML
	if strings.EqualFold(vendorOptions["deviceModel"], "c6320") {
		cfgFile, err := generateConfig(cfg, "json", vendorOptions["deviceModel"], vendorOptions["serviceTag"])
		if err != nil {
			return err
		}
		_, err = s.racadmSetFile(ctx, cfgFile, "json")
		if err != nil {
			return err
		}
	} else {
		cfgFile, err := generateConfig(cfg, "xml", vendorOptions["deviceModel"], vendorOptions["serviceTag"])
		if err != nil {
			return err
		}
		_, err = s.racadmSetFile(ctx, cfgFile, "xml")
		if err != nil {
			return err
		}
	}

	return nil
}

// SetBIOSConfigurationFromFile takes a raw file of BIOS configurtation values and applies them to the host
func (s *DellRacadm) SetBIOSConfigurationFromFile(ctx context.Context, deviceModel, cfg string) (err error) {
	// older hardware return BIOS config as XML
	if strings.EqualFold(deviceModel, "c6320") {
		_, err = s.racadmSetFile(ctx, cfg, "json")
	} else {
		_, err = s.racadmSetFile(ctx, cfg, "xml")
	}

	return err
}

// nolint:gocyclo // going through all bios values to standardize them is going to be high complexity
func generateConfig(cfg map[string]string, format, deviceModel, serviceTag string) (string, error) {
	vcm, err := config.NewVendorConfigManager(format, common.VendorDell, map[string]string{"model": deviceModel, "servicetag": serviceTag})
	if err != nil {
		return "", err
	}

	// TODO(jwb) These should be replaced with functions in bmc-toolbox/common/config
	for k, v := range cfg {
		switch {
		case k == "amd_sev":
			vcm.Raw("CpuMinSevAsid", v, []string{"BIOS.Setup.1-1"})
		case k == "boot_mode":
			vcm.Raw("BootMode", v, []string{"BIOS.Setup.1-1"})
		case k == "intel_txt":
			vcm.Raw("IntelTxt", v, []string{"BIOS.Setup.1-1"})
		case k == "intel_sgx":
			vcm.Raw("Software Guard Extensions (SGX)", v, []string{"BIOS.Setup.1-1"})
		case k == "secure_boot":
			vcm.Raw("SecureBoot", v, []string{"BIOS.Setup.1-1"})
		case k == "tpm":
			vcm.Raw("TpmSecurity", v, []string{"BIOS.Setup.1-1"})
		case k == "smt":
			vcm.Raw("LogicalProc", v, []string{"BIOS.Setup.1-1"})
		case k == "sr_iov":
			vcm.Raw("SriovGlobalEnable", v, []string{"BIOS.Setup.1-1"})
		case strings.HasPrefix(k, "raw:"):
			// TODO(jwb) How do we want to handle raw: for things other than BIOS.Setup
			vcm.Raw(strings.TrimPrefix(k, "raw:"), v, []string{"BIOS.Setup.1-1"})
		}
	}

	return vcm.Marshal()
}

func WithReboot() DellRacadmOption {
	return func(r *DellRacadm) {
		r.ShutdownType = "Graceful"
	}
}

func WithoutReboot() DellRacadmOption {
	return func(r *DellRacadm) {
		r.ShutdownType = "NoReboot"
	}
}

// racadmSet executes the racadm 'set' subcommand and returns &utils.Result and (nil or error)
func (s *DellRacadm) racadmSet(ctx context.Context, argInputConfigFile, argFileType string) (result *Result, err error) {
	var (
		argGracefulWait         = 300
		argPostImportPowerState = "On"
	)

	cmd := []string{"set",
		"-t", argFileType,
		"-f", argInputConfigFile,
		"-b", s.ShutdownType,
		"-w", strconv.Itoa(argGracefulWait),
		"-s", argPostImportPowerState,
	}

	s.Executor.SetArgs(cmd)

	result, err = s.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return nil, newExecError(s.Executor.GetCmd(), result)
	}

	return result, nil
}

// racadmSetJSON executes racadm to set config as JSON and returns nil or error
func (s *DellRacadm) racadmSetFile(ctx context.Context, contents, format string) (result *Result, err error) {
	// Open tmp file to hold toSet JSON
	inputConfigTmpFile, err := os.CreateTemp("", "ironlib-racadmSetFile")
	if err != nil {
		return nil, err
	}

	defer os.Remove(inputConfigTmpFile.Name())

	_, err = inputConfigTmpFile.WriteString(contents)
	if err != nil {
		return nil, err
	}

	err = inputConfigTmpFile.Close()
	if err != nil {
		return nil, err
	}

	return s.racadmSet(ctx, inputConfigTmpFile.Name(), format)
}

// racadmBIOSConfigXML executes racadm to retrieve BIOS config as XML and returns a map[string]string object
func (s *DellRacadm) racadmBIOSConfigXML(ctx context.Context) (map[string]string, error) {
	// Dump the current BIOS config to dellBiosTempFilename. The racadm
	// command won't dump the config to stdout directly, so we do this in
	// a two-step process, and read the tempfile during the parsing step.
	cmd := []string{"get", "-t", "xml", "-f", s.BIOSCfgTmpFile}
	s.Executor.SetArgs(cmd)

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return nil, newExecError(s.Executor.GetCmd(), result)
	}

	if !s.KeepConfigFile {
		defer os.Remove(s.BIOSCfgTmpFile)
	}

	return findXMLAttributes(s.BIOSCfgTmpFile, "//Component[@FQDD='BIOS.Setup.1-1']//Attribute")
}

// findXMLAttributes runs the xml query and returns a map of BIOS attributes to values
func findXMLAttributes(cfgFile, query string) (map[string]string, error) {
	xml := etree.NewDocument()

	err := xml.ReadFromFile(cfgFile)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string)

	for _, e := range xml.FindElements(query) {
		for _, a := range e.Attr {
			if strings.EqualFold(a.Key, "Name") {
				n := a.Value
				v := e.Text()

				if n != "" && v != "" {
					values[n] = v
				}
			}
		}
	}

	return values, nil
}

// racadmBIOSConfigJSON executes racadm to retrieve BIOS config as JSON and returns a map with all the settings and their value object
func (s *DellRacadm) racadmBIOSConfigJSON(ctx context.Context) (map[string]string, error) {
	// Dump the current BIOS config to dellBiosTempFilename. The racadm
	// command won't dump the config to stdout directly, so we do this in
	// a two-step process, and read the tempfile during the parsing step.
	cmd := []string{"get", "-t", "json", "-f", s.BIOSCfgTmpFile}
	s.Executor.SetArgs(cmd)

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return nil, newExecError(s.Executor.GetCmd(), result)
	}

	if !s.KeepConfigFile {
		defer os.Remove(s.BIOSCfgTmpFile)
	}

	json, err := os.ReadFile(s.BIOSCfgTmpFile)
	if err != nil {
		return nil, err
	}

	s.ConfigJSON = string(json)

	attrs := map[string]string{}

	attrJSON := gjson.Get(s.ConfigJSON, `SystemConfiguration.Components.#(FQDD=="BIOS.Setup.1-1").Attributes`)
	attrJSON.ForEach(func(key, value gjson.Result) bool {
		n := value.Get("Name").String()
		v := value.Get("Value").String()

		if n == "" || v == "" {
			return true
		}

		attrs[n] = v

		return true
	})

	return attrs, nil
}

// FakeRacadmExecute implements the utils.Executor interface for testing
type FakeRacadmExecute struct {
	Cmd    string
	Args   []string
	Env    []string
	Stdin  io.Reader
	Stdout []byte // Set this for the dummy data to be returned
	Stderr []byte // Set this for the dummy data to be returned
	Quiet  bool
	// Executor embedded in here to skip having to implement all the utils.Executor methods
	Executor
}

// NewFakeRacadmExecute returns a fake SMC sum executor for tests
func NewFakeRacadmExecutor(cmd string) Executor {
	return &FakeRacadmExecute{Cmd: cmd}
}

// NewFakeRacadm returns a fake lshw executor for testing
func NewFakeRacadm(biosCfgFile string) *DellRacadm {
	executor := NewFakeRacadmExecutor("racadm")

	return &DellRacadm{Executor: executor, BIOSCfgTmpFile: biosCfgFile, KeepConfigFile: true}
}

// ExecWithContext implements the utils.Executor interface
func (e *FakeRacadmExecute) ExecWithContext(ctx context.Context) (*Result, error) {
	return &Result{Stdout: []byte(`dummy`)}, nil
}

// SetArgs is to set cmd args to the fake execute method
func (e *FakeRacadmExecute) SetArgs(a []string) {
	e.Args = a
}
