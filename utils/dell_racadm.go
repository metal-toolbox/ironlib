package utils

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/packethost/ironlib/config"
	"github.com/tidwall/gjson"
)

const DellRacadmPath = "/opt/dell/srvadmin/bin/idracadm7"
const EnvVarRacadm7 = "UTIL_RACADM7"
const DellBiosTempFilename = "/tmp/bios.json" // where we dump the BIOS config to before processing it

type DellRacadm struct {
	Executor   Executor
	ConfigJSON string
}

// Return a new Dell racadm command executor
func NewDellRacadm(trace bool) BIOSConfiguror {
	racadmUtil := os.Getenv(EnvVarRacadm7)
	if racadmUtil == "" {
		racadmUtil = DellRacadmPath
	}

	e := NewExecutor(racadmUtil)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	if !trace {
		e.SetQuiet()
	}

	return &DellRacadm{Executor: e}
}

func (s *DellRacadm) GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error) {
	// Dump the current BIOS config to dellBiosTempFilename. The racadm
	// command won't dump the config to stdout directly, so we do this in
	// a two-step process, and read the tempfile during the parsing step.
	s.Executor.SetArgs([]string{"get", "-t", "json", "-f", DellBiosTempFilename})
	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return nil, newUtilsExecError(s.Executor.GetCmd(), result)
	}

	cfg, err := s.parseRacadmBIOSConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &config.BIOSConfiguration{Dell: cfg}, nil
}

func (s *DellRacadm) parseRacadmBIOSConfig(ctx context.Context) (*config.DellBIOS, error) {
	queries := map[string]string{
		"BootMode":       "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"BootMode\").Value",
		"AMDSEV":         "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"CpuMinSevAsid\").Value",
		"Hyperthreading": "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"LogicalProc\").Value",
		"SRIOV":          "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"SriovGlobalEnable\").Value",
		"TPM":            "SystemConfiguration.Components.#(FQDD==\"BIOS.Setup.1-1\").Attributes.#(Name==\"TpmSecurity\").Value",
	}
	json, err := ioutil.ReadFile(DellBiosTempFilename)
	defer func() { os.Remove(DellBiosTempFilename) }()
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
