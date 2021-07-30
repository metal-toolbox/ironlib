package utils

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"strings"

	"github.com/packethost/ironlib/config"
	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/model/supermicro"
	"golang.org/x/net/html/charset"
)

const smcSUM = "/usr/sbin/sum"

type SupermicroSUM struct {
	Executor Executor
}

// Return a new Supermicro sum command executor
func NewSupermicroSUM(trace bool) *SupermicroSUM {
	e := NewExecutor(smcSUM)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &SupermicroSUM{Executor: e}
}

// Components implements the Utility interface
func (s *SupermicroSUM) Components() ([]*model.Component, error) {
	return nil, nil
}

// Collect implements the Utility interface
func (s *SupermicroSUM) Collect(device *model.Device) error {
	return nil
}

// UpdateBIOS installs the SMC BIOS update
func (s *SupermicroSUM) UpdateBIOS(ctx context.Context, updateFile, modelNumber string) error {
	s.Executor.SetArgs([]string{"-c", "UpdateBios", "--preserve_setting", "--file", updateFile})

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return newExecError(s.Executor.GetCmd(), result)
	}

	return nil
}

// UpdateBMC installs the SMC BMC update
func (s *SupermicroSUM) UpdateBMC(ctx context.Context, updateFile, modelNumber string) error {
	s.Executor.SetArgs([]string{"-c", "UpdateBmc", "--file", updateFile})

	result, err := s.Executor.ExecWithContext(ctx)
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
	case model.SlugBIOS:
		s.Executor.SetArgs([]string{"-c", "UpdateBios", "--preserve_setting", "--file", updateFile})
	case model.SlugBMC:
		s.Executor.SetArgs([]string{"-c", "UpdateBmc", "--file", updateFile})
	}

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return newExecError(s.Executor.GetCmd(), result)
	}

	return nil
}

// GetBIOSConfiguration implements the Getter
func (s *SupermicroSUM) GetBIOSConfiguration(ctx context.Context, deviceModel string) (*config.BIOSConfiguration, error) {
	c, err := s.parseBIOSConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &config.BIOSConfiguration{Supermicro: c}, nil
}

// parseBIOSConfig parses the SMC sum command output BIOS config and returns a model.SupermicroBIOS object
func (s *SupermicroSUM) parseBIOSConfig(ctx context.Context) (*config.SupermicroBIOS, error) {
	s.Executor.SetArgs([]string{"-c", "GetCurrentBiosCfg"})

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return nil, newExecError(s.Executor.GetCmd(), result)
	}

	cfg := &supermicro.BiosCfg{}

	// the xml exported by sum is ISO-8859-1 encoded
	decoder := xml.NewDecoder(bytes.NewReader(result.Stdout))
	// convert characters from non-UTF-8 to UTF-8
	decoder.CharsetReader = charset.NewReaderLabel

	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}

	// BIOS settings to collect
	// The map key should be the exact BIOS setting name as it appears in the config
	// Note: when updating this map, ensure the config attributes returned are updated as well.
	settings := map[string]string{
		"Boot mode select":                "",
		"Hyper-Threading":                 "", // X11SCHF-F
		"Hyper-Threading [ALL]":           "", // X11DPH-T
		"Secure Boot":                     "",
		"Security Device Support":         "",
		"Software Guard Extensions (SGX)": "",
	}

	// recurse through the BIOS config, find params we care about
	s.recurseMenus(cfg.Menu, settings)

	bios := &config.SupermicroBIOS{
		BootMode:       settings["Boot mode select"],
		Hyperthreading: settings["Hyper-Threading"],
		SecureBoot:     settings["Secure Boot"],
		TPM:            settings["Security Device Support"],
		IntelSGX:       settings["Software Guard Extensions (SGX)"],
	}

	if settings["Hyper-Threading [ALL]"] != "" {
		bios.Hyperthreading = settings["Hyper-Threading [ALL]"]
	}

	return bios, nil
}

// recurseMenus recurses through SMC BIOS menu options
func (s *SupermicroSUM) recurseMenus(menus []*supermicro.Menu, kv map[string]string) {
	for _, menu := range menus {
		for _, s := range menu.Setting {
			s.Name = strings.TrimSpace(s.Name)

			_, exists := kv[s.Name]
			if exists {
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

// ExecWithContext implements the utils.Executor interface
func (e *FakeSMCSumExecute) ExecWithContext(ctx context.Context) (*Result, error) {
	b := bytes.Buffer{}

	_, err := b.ReadFrom(e.Stdin)
	if err != nil {
		return nil, err
	}

	return &Result{Stdout: b.Bytes()}, nil
}

// SetStdin is to set input to the fake execute method
func (e *FakeSMCSumExecute) SetStdin(r io.Reader) {
	e.Stdin = r
}

// SetArgs is to set cmd args to the fake execute method
func (e *FakeSMCSumExecute) SetArgs(a []string) {
	e.Args = a
}
