package utils

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"os"
	"strings"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/ironlib/model"
	"golang.org/x/net/html/charset"
)

const EnvVarSumPath = "IRONLIB_UTIL_SUM"

type biosCfg struct {
	XMLName xml.Name `xml:"BiosCfg"`
	Text    string   `xml:",chardata"`
	Menu    []*menu  `xml:"Menu,omitempty"`
}

type menu struct {
	Name    string     `xml:"name,attr"`
	Setting []*setting `xml:"Setting,omitempty"`
	Menu    []*menu    `xml:"Menu,omitempty"`
}

type setting struct {
	Name           string `xml:"name,attr"`
	Type           string `xml:"type,attr"`
	SelectedOption string `xml:"selectedOption,attr,omitempty"`
	CheckedStatus  string `xml:"checkedStatus,attr,omitempty"`
}

type SupermicroSUM struct {
	Executor Executor
}

// Return a new Supermicro sum command executor
func NewSupermicroSUM(trace bool) *SupermicroSUM {
	// TODO: rename this to smc-sum
	utility := "sum"

	// lookup env var for util
	if eVar := os.Getenv(EnvVarSumPath); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &SupermicroSUM{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (s *SupermicroSUM) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := s.Executor.CheckExecutable()

	return "smc-sum", s.Executor.CmdPath(), er
}

// Components implements the Utility interface
func (s *SupermicroSUM) Components() ([]*model.Component, error) {
	return nil, nil
}

// Collect implements the Utility interface
func (s *SupermicroSUM) Collect(_ *common.Device) error {
	return nil
}

// UpdateBIOS installs the SMC BIOS update
func (s *SupermicroSUM) UpdateBIOS(ctx context.Context, updateFile, modelNumber string) error {
	s.Executor.SetArgs("-c", "UpdateBios", "--preserve_setting", "--file", updateFile)

	// X12STH-SYS does not support the preserve_setting option
	if strings.EqualFold(modelNumber, "X12STH-SYS") {
		s.Executor.SetArgs("-c", "UpdateBios", "--file", updateFile)
	}

	result, err := s.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return newExecError(s.Executor.GetCmd(), result)
	}

	return nil
}

// UpdateBMC installs the SMC BMC update
func (s *SupermicroSUM) UpdateBMC(ctx context.Context, updateFile, _ string) error {
	s.Executor.SetArgs("-c", "UpdateBmc", "--file", updateFile)

	result, err := s.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return newExecError(s.Executor.GetCmd(), result)
	}

	return nil
}

// ApplyUpdate installs the SMC update based on the component
func (s *SupermicroSUM) ApplyUpdate(ctx context.Context, updateFile, componentSlug string) error {
	switch componentSlug {
	case common.SlugBIOS:
		s.Executor.SetArgs("-c", "UpdateBios", "--preserve_setting", "--file", updateFile)
	case common.SlugBMC:
		s.Executor.SetArgs("-c", "UpdateBmc", "--file", updateFile)
	}

	result, err := s.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return newExecError(s.Executor.GetCmd(), result)
	}

	return nil
}

// GetBIOSConfiguration implements the Getter
func (s *SupermicroSUM) GetBIOSConfiguration(ctx context.Context, _ string) (map[string]string, error) {
	return s.parseBIOSConfig(ctx)
}

// parseBIOSConfig parses the SMC sum command output BIOS config and returns a model.BIOSConfiguration object
func (s *SupermicroSUM) parseBIOSConfig(ctx context.Context) (map[string]string, error) {
	s.Executor.SetArgs("-c", "GetCurrentBiosCfg")

	result, err := s.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return nil, newExecError(s.Executor.GetCmd(), result)
	}

	cfg := &biosCfg{}

	// the xml exported by sum is ISO-8859-1 encoded
	decoder := xml.NewDecoder(bytes.NewReader(result.Stdout))
	// convert characters from non-UTF-8 to UTF-8
	decoder.CharsetReader = charset.NewReaderLabel

	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}

	settings := map[string]string{}
	s.recurseMenus(cfg.Menu, settings)

	return normalizeBIOSConfiguration(settings), nil
}

// recurseMenus recurses through SMC BIOS menu options and gathers all settings with a selected option
func (s *SupermicroSUM) recurseMenus(menus []*menu, kv map[string]string) {
	for _, menu := range menus {
		for _, s := range menu.Setting {
			s.Name = strings.TrimSpace(s.Name)

			if s.SelectedOption != "" {
				kv[s.Name] = s.SelectedOption
			}
		}

		if menu.Menu == nil {
			continue
		}

		s.recurseMenus(menu.Menu, kv)
	}
}

// FakeSMCSumExecute implements the utils.Executor interface for testing
type FakeSMCSumExecute struct {
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

// NewFakeSMCSumExecute returns a fake SMC sum executor for tests
func NewFakeSMCSumExecutor(cmd string) Executor {
	return &FakeSMCSumExecute{Cmd: cmd}
}

// NewFakeSMCSum returns a fake lshw executor for testing
func NewFakeSMCSum(stdin io.Reader) *SupermicroSUM {
	executor := NewFakeSMCSumExecutor("sum")
	executor.SetStdin(stdin)

	return &SupermicroSUM{Executor: executor}
}

// Exec implements the utils.Executor interface
func (e *FakeSMCSumExecute) Exec(_ context.Context) (*Result, error) {
	b := bytes.Buffer{}

	if e.Stdin != nil {
		_, err := b.ReadFrom(e.Stdin)
		if err != nil {
			return nil, err
		}
	}

	return &Result{Stdout: b.Bytes()}, nil
}

// SetStdin is to set input to the fake execute method
func (e *FakeSMCSumExecute) SetStdin(r io.Reader) {
	e.Stdin = r
}

// SetArgs is to set cmd args to the fake execute method
func (e *FakeSMCSumExecute) SetArgs(a ...string) {
	e.Args = a
}

// GetCmd is to retrieve the cmd args for the fake execute method
func (e *FakeSMCSumExecute) GetCmd() string {
	return strings.Join(e.Args, " ")
}
