package utils

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
)

const (
	EnvHdparmUtility = "IRONLIB_UTIL_HDPARM"
)

type Hdparm struct {
	Executor Executor
}

// Return a new hdparm executor
func NewHdparmCmd(trace bool) *Hdparm {
	utility := "hdparm"

	// lookup env var for util
	if eVar := os.Getenv(EnvHdparmUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Hdparm{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (h *Hdparm) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := h.Executor.CheckExecutable()

	return "hdparm", h.Executor.CmdPath(), er
}

func (h *Hdparm) cmdListCapabilities(ctx context.Context, logicalName string) ([]byte, error) {
	// hdparm -I devicepath
	h.Executor.SetArgs("-I", logicalName)

	result, err := h.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

// DriveCapabilities returns the capability attributes obtained through hdparm
//
// The logicalName is the kernel/OS assigned drive name - /dev/sdX
//
// This method implements the actions.DriveCapabilityCollector interface.
//
// nolint:gocyclo // line parsing is cyclomatic
func (h *Hdparm) DriveCapabilities(ctx context.Context, logicalName string) ([]*common.Capability, error) {
	out, err := h.cmdListCapabilities(ctx, logicalName)
	if err != nil {
		return nil, err
	}

	var capabilities []*common.Capability

	var lines []string

	s := string(out)

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	// Delimiters
	featStart := "Enabled"
	secStart := "Security:"
	secEnd := "Logical Unit"

	supported := regexp.MustCompile(`(?s)^supported$`)
	seu := strings.NewReplacer("min", "")
	sfi := strings.NewReplacer("_", " ", "-", " ", "{", "", "}", "", "(", "", ")", "",
		",", "", "|", "", "set", "", "command", "")

	var featBool, secBool bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)

		var flag string

		// start/end match specific block delimiters
		// bools are toggled to indicate lines within a given block
		switch {
		case strings.Contains(line, featStart):
			featBool = true
		case strings.Contains(line, secStart):
			featBool, secBool = false, true
		case strings.Contains(line, secEnd):
			featBool, secBool = false, false
		}

		// Parse command capabilities
		if featBool && !strings.Contains(line, featStart) {
			if strings.Contains(line, "*") {
				line = strings.TrimSpace(strings.TrimPrefix(line, "*\t"))

				// Generate short flag identifier
				line = strings.TrimSpace(sfi.Replace(line))
				for _, word := range strings.Fields(line) {
					flag += strings.ToLower(word[0:1])
				}

				capability := new(common.Capability)
				capability.Name = flag
				capability.Description = line
				capability.Enabled = true
				capabilities = append(capabilities, capability)
			} else if !strings.Contains(line, "*") && !strings.Contains(line, featStart) {
				// Generate short flag identifier
				line = strings.TrimSpace(sfi.Replace(line))
				for _, word := range strings.Fields(line) {
					flag += strings.ToLower(word[0:1])
				}

				capability := new(common.Capability)
				capability.Name = flag
				capability.Description = line
				capability.Enabled = false
				capabilities = append(capabilities, capability)
			}
		} else if secBool {
			// Parse security capabilities
			secSupported := supported.MatchString(line)
			if !strings.Contains(line, secStart) {
				capability := new(common.Capability)
				switch {
				case strings.Contains(line, "65534"):
					capability.Name, capability.Enabled = "pns", true
					capability.Enabled = true
					capability.Description = "password not set"
				case secSupported:
					capability.Name = "es"
					capability.Enabled = true
					capability.Description = "encryption supported"
				case strings.Contains(line, "not\tenabled"):
					capability.Name = "ena"
					capability.Enabled = true
					capability.Description = "encryption not active"
				case strings.Contains(line, "not\tlocked"):
					capability.Name = "dnl"
					capability.Enabled = true
					capability.Description = "device is not locked"
				case strings.Contains(line, "not\tfrozen"):
					capability.Name = "dnf"
					capability.Enabled = true
					capability.Description = "device is not frozen"
				case strings.Contains(line, "not\texpired"):
					capability.Name = "ene"
					capability.Enabled = true
					capability.Description = "encryption not expired"
				case strings.Contains(line, "supported: enhanced erase"):
					capability.Name = "esee"
					capability.Enabled = true
					capability.Description = "encryption supports enhanced erase"
				case strings.Contains(line, "SECURITY ERASE UNIT"):
					seTime, sehTime := seu.Replace(parts[0]), seu.Replace(parts[5])
					capability.Name = "time" + seTime + ":" + sehTime
					capability.Description = "erase time: " + seTime + "m, " + sehTime + "m (enhanced)"
				}
				capabilities = append(capabilities, capability)
			}
		}
	}

	return capabilities, err
}

// NewFakeHdparm returns a mock hdparm collector that returns mock data for use in tests.
func NewFakeHdparm() *Hdparm {
	return &Hdparm{
		Executor: NewFakeExecutor("hdparm"),
	}
}
