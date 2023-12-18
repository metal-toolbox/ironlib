package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const DellRacadmPath = "/opt/dell/srvadmin/bin/idracadm7"
const EnvVarRacadm7 = "IRONLIB_UTIL_RACADM7"

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

// SetBIOSConfiguration returns a BIOS configuration object
func (s *DellRacadm) SetBIOSConfiguration(ctx context.Context, deviceModel string, cfg map[string]string) error {
	// TODO(jwb) Probably need to handle deviceModel here like we do in GetBIOSConfiguration
	err := s.racadmSetXML(ctx, cfg)
	return err
}

// racadmSet executes the racadm 'set' subcommand and returns &utils.Result and (nil or error)
func (s *DellRacadm) racadmSet(ctx context.Context, argInputConfigFile string, argFileType string) (result *Result, err error) {
	var (
		argShutdownType         string = "Graceful"
		argGracefulWait         int    = 300
		argPostImportPowerState string = "On"
	)

	cmd := []string{"set",
		"-t", argFileType,
		"-f", argInputConfigFile,
		"-b", argShutdownType,
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
func (s *DellRacadm) racadmSetJSON(ctx context.Context, toSet string) (err error) {

	// Open tmp file to hold toSet JSON
	inputConfigTmpFile, err := os.CreateTemp("", "ironlib-racadmSetJSON")
	if err != nil {
		return
	}

	defer os.Remove(inputConfigTmpFile.Name())

	// TODO(jwb) we need to process toSet and call the appropriate functions via the bmc-toolbox/common
	// config abstraction(s)

	cfg, err := dc.WriteConfig("JSON")

	if err != nil {
		return
	}

	inputConfigTmpFile.WriteString(toSet)
	inputConfigTmpFile.Close()

	_, err = s.racadmSet(ctx, inputConfigTmpFile.Name(), "json")
	if err != nil {
		return
	}

	return
}

// racadmSetJSON executes racadm to set config as JSON and returns nil or error
func (s *DellRacadm) racadmSetXML(ctx context.Context, toSet map[string]string) (err error) {
	// Open tmp file to hold toSet XML
	inputConfigTmpFile, err := os.CreateTemp("", "ironlib-racadmSetXML")
	if err != nil {
		return
	}

	defer os.Remove(inputConfigTmpFile.Name())

	dc := model.NewDellConfig()

	dc.SystemConfiguration = &model.DellSystemConfiguration{
		// Comments:   []string{"Test Comment"},
		Model:      "PowerEdge R6515", // TODO(jwb) Need to pull this and the servicetag from the actual machine
		ServiceTag: "G97ZTD3",
		TimeStamp:  "Tue Jul  7 13:24:19 2020", // TODO(jwb) Generate this
	}

	biosComponent := &model.DellComponent{FQDD: "BIOS.Setup.1-1"}
	dc.SystemConfiguration.Components = []*model.DellComponent{biosComponent}

	// TODO(jwb) Call a function here that translates toSet generic options into provider specific options

	// var attrs map[string]string = make(map[string]string)
	// attrs["SriovGlobalEnable"] = "Disabled"

	for k, v := range toSet {
		dc.SystemConfiguration.Components[0].Attributes = append(dc.SystemConfiguration.Components[0].Attributes, &model.DellComponentAttribute{Name: k, Value: v})
	}

	cfg, err := dc.WriteConfig("xml")

	if err != nil {
		return
	}

	fmt.Printf("string(cfg): %v\n", string(cfg))

	inputConfigTmpFile.WriteString(string(cfg))
	inputConfigTmpFile.Close()

	_, err = s.racadmSet(ctx, inputConfigTmpFile.Name(), "xml")
	if err != nil {
		return
	}

	return
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
