package utils

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/packethost/ironlib/config"
)

const DellRacadmPath = "/opt/dell/srvadmin/bin/idracadm7"
const EnvVarRacadm7 = "UTIL_RACADM7"

var ErrDellBiosCfgNil = errors.New("expected valid bios config object, got nil")
var ErrDellBiosCfgFileUndefined = errors.New("no BIOS config file defined")
var ErrDellBiosCfgFileEmpty = errors.New("BIOS config file empty or invalid")

// DellRacadm is a dell racadm executor
type DellRacadm struct {
	Executor       Executor
	ConfigJSON     string
	BIOSCfgTmpFile string // where we dump the BIOS config to before processing it
	KeepConfigFile bool   // flag to keep the BIOS config file generated (mainly for testing)
}

// Return a new Dell racadm command executor
func NewDellRacadm(trace bool) *DellRacadm {
	racadmUtil := os.Getenv(EnvVarRacadm7)
	if racadmUtil == "" {
		racadmUtil = DellRacadmPath
	}

	e := NewExecutor(racadmUtil)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &DellRacadm{Executor: e, BIOSCfgTmpFile: "/tmp/bioscfg"}
}

// GetBIOSConfiguration returns a BIOS configuration object
func (s *DellRacadm) GetBIOSConfiguration(ctx context.Context, deviceModel string) (*config.BIOSConfiguration, error) {
	var cfg *config.DellBIOS

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

	return &config.BIOSConfiguration{Dell: cfg}, nil
}

// racadmBIOSConfigXML executes racadm to retrieve BIOS config as XML and returns a config.DellBIOS object
func (s *DellRacadm) racadmBIOSConfigXML(ctx context.Context) (*config.DellBIOS, error) {
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

	// map of attribute name to xpath query
	// when updating these attributes, ensure the DellBIOS struct below is updated as well
	queries := map[string]string{
		"HyperThreading": "//Component[@FQDD='BIOS.Setup.1-1']//Attribute[@Name='LogicalProc']",
		"BootMode":       "//Component[@FQDD='BIOS.Setup.1-1']//Attribute[@Name='BootMode']",
		"SRIOV":          "//Component[@FQDD='BIOS.Setup.1-1']//Attribute[@Name='SriovGlobalEnable']",
		"TPM":            "//Component[@FQDD='BIOS.Setup.1-1']//Attribute[@Name='TpmSecurity']",
		"IntelTXT":       "//Component[@FQDD='BIOS.Setup.1-1']//Attribute[@Name='IntelTxt']",
	}

	settings, err := findXMLAttributes(s.BIOSCfgTmpFile, queries)
	if err != nil {
		return nil, err
	}

	return &config.DellBIOS{
		BootMode:       settings["BootMode"],
		Hyperthreading: settings["HyperThreading"],
		SRIOV:          settings["SRIOV"],
		TPM:            settings["TPM"],
	}, nil
}

// findXMLAttributes runs the xml queries and returns a map of BIOS attributes to values
func findXMLAttributes(cfgFile string, queries map[string]string) (map[string]string, error) {
	xml := etree.NewDocument()

	err := xml.ReadFromFile(cfgFile)
	if err != nil {
		return nil, err
	}

	values := make(map[string]string)

	for key, xpath := range queries {
		e := xml.FindElement(xpath)
		if e == nil {
			continue
		}

		values[key] = e.Text()
	}

	return values, nil
}

// racadmBIOSConfigJSON executes racadm to retrieve BIOS config as JSON and returns a config.DellBIOS object
func (s *DellRacadm) racadmBIOSConfigJSON(ctx context.Context) (*config.DellBIOS, error) {
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

	queries := map[string]string{
		"BootMode":       "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"BootMode\").Value",
		"AMDSEV":         "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"CpuMinSevAsid\").Value",
		"Hyperthreading": "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"LogicalProc\").Value",
		"SRIOV":          "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"SriovGlobalEnable\").Value",
		"TPM":            "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"TpmSecurity\").Value",
	}

	json, err := ioutil.ReadFile(s.BIOSCfgTmpFile)
	if err != nil {
		return nil, err
	}

	s.ConfigJSON = string(json)

	return &config.DellBIOS{
		AMDSev:         gjson.Get(s.ConfigJSON, queries["AMDSEV"]).Int(),
		BootMode:       gjson.Get(s.ConfigJSON, queries["BootMode"]).String(),
		Hyperthreading: gjson.Get(s.ConfigJSON, queries["Hyperthreading"]).String(),
		SRIOV:          gjson.Get(s.ConfigJSON, queries["SRIOV"]).String(),
		TPM:            gjson.Get(s.ConfigJSON, queries["TPM"]).String(),
	}, nil
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
