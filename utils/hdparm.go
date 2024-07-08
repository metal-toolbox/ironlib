package utils

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/sirupsen/logrus"
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

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	// TODO break up into features and security specific blocks/parsers
	// use regex to grab features block and parse
	// use regex to grab security block and parse

	// Delimiters
	featStart := "Enabled"
	secStart := "Security:"
	secEnd := "Logical Unit"

	supported := regexp.MustCompile(`(?s)^supported$`)
	seu := strings.NewReplacer("min", "")
	sfi := strings.NewReplacer("_", " ", "-", " ", "{", "", "}", "", "(", "", ")", "",
		",", "", "|", "", "set", "", "command", "")

	var featBool, secBool bool
	var capabilities []*common.Capability
	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Fields(line)

		// start/end match specific block delimiters
		// bools are toggled to indicate lines within a given block
		switch {
		case strings.Contains(line, featStart):
			featBool, secBool = true, false
		case strings.Contains(line, secStart):
			featBool, secBool = false, true
		case strings.Contains(line, secEnd):
			featBool, secBool = false, false
		}

		// Parse command capabilities
		var flag string
		if featBool && !strings.Contains(line, featStart) {
			if strings.Contains(line, "*") {
				line = strings.TrimSpace(strings.TrimPrefix(line, "*\t"))

				// Generate short flag identifier
				line = strings.TrimSpace(sfi.Replace(line))
				for _, word := range strings.Fields(line) {
					flag += strings.ToLower(word[0:1])
				}

				capabilities = append(capabilities, &common.Capability{
					Name:        flag,
					Description: line,
					Enabled:     true,
				})
			} else if !strings.Contains(line, "*") && !strings.Contains(line, featStart) {
				// Generate short flag identifier
				line = strings.TrimSpace(sfi.Replace(line))
				for _, word := range strings.Fields(line) {
					flag += strings.ToLower(word[0:1])
				}

				capabilities = append(capabilities, &common.Capability{
					Name:        flag,
					Description: line,
					Enabled:     false,
				})
			}
		} else if secBool {
			// Parse security capabilities
			secSupported := supported.MatchString(line)
			if !strings.Contains(line, secStart) {
				var capability common.Capability
				switch {
				case strings.Contains(line, "65534"):
					capability = common.Capability{
						Name:        "pns",
						Enabled:     true,
						Description: "password not set",
					}
				case secSupported:
					capability = common.Capability{
						Name:        "es",
						Enabled:     true,
						Description: "encryption supported",
					}
				case strings.Contains(line, "not\tenabled"):
					capability = common.Capability{
						Name:        "ena",
						Enabled:     true,
						Description: "encryption not active",
					}
				case strings.Contains(line, "not\tlocked"):
					capability = common.Capability{
						Name:        "dnl",
						Enabled:     true,
						Description: "device is not locked",
					}
				case strings.Contains(line, "not\tfrozen"):
					capability = common.Capability{
						Name:        "dnf",
						Enabled:     true,
						Description: "device is not frozen",
					}
				case strings.Contains(line, "not\texpired"):
					capability = common.Capability{
						Name:        "ene",
						Enabled:     true,
						Description: "encryption not expired",
					}
				case strings.Contains(line, "supported: enhanced erase"):
					capability = common.Capability{
						Name:        "esee",
						Enabled:     true,
						Description: "encryption supports enhanced erase",
					}
				case strings.Contains(line, "SECURITY ERASE UNIT"):
					seTime, sehTime := seu.Replace(parts[0]), seu.Replace(parts[5])
					capability = common.Capability{
						Name:        "time" + seTime + ":" + sehTime,
						Description: "erase time: " + seTime + "m, " + sehTime + "m (enhanced)",
						Enabled:     false,
					}
				}
				capabilities = append(capabilities, &capability)
			}
		}
	}

	return capabilities, err
}

// WipeDrive implements DriveWipe by calling Sanitize or Erase as appropriate.
// Sanitize(CryptoErase) is preferred over Sanitize(BlockErase) which is preferred over Erase(CryptographicErase).
func (h *Hdparm) WipeDrive(ctx context.Context, logger *logrus.Logger, drive *common.Drive) error { // nolint:gocyclo
	var (
		esee     bool
		eseu     bool
		sanitize bool
		bee      bool
		cse      bool
	)
	for _, cap := range drive.Capabilities {
		switch {
		case cap.Description == "encryption supports enhanced erase":
			esee = cap.Enabled
		case cap.Description == "BLOCK ERASE EXT":
			bee = cap.Enabled
		case cap.Description == "CRYPTO SCRAMBLE EXT":
			cse = cap.Enabled
		case cap.Description == "SANITIZE feature":
			sanitize = cap.Enabled
		case strings.HasPrefix(cap.Description, "erase time:"):
			eseu = strings.Contains(cap.Description, "enhanced")
		}
	}

	l := logger.WithField("drive", drive.LogicalName)
	if sanitize && cse {
		// nolint:govet
		l := l.WithField("method", "sanitize").WithField("action", "sanitize-crypto-scramble")
		l.Info("wiping")
		err := h.Sanitize(ctx, drive, CryptoErase)
		if err == nil {
			return nil
		}
		l.WithError(err).Info("failed")
	}
	if sanitize && bee {
		// nolint:govet
		l := l.WithField("method", "sanitize").WithField("action", "sanitize-block-erase")
		l.Info("wiping")
		err := h.Sanitize(ctx, drive, BlockErase)
		if err == nil {
			return nil
		}
		l.WithError(err).Info("failed")
	}
	if esee && eseu {
		// nolint:govet
		l := l.WithField("method", "security-erase-enhanced")
		l.Info("wiping")
		err := h.Erase(ctx, drive, CryptographicErase)
		if err == nil {
			return nil
		}
		l.WithError(err).Info("failed")
	}
	return ErrIneffectiveWipe
}

// Sanitize wipes drive using `ATA Sanitize Device` via hdparm --sanitize
func (h *Hdparm) Sanitize(ctx context.Context, drive *common.Drive, sanact SanitizeAction) error {
	var sanType string
	switch sanact { // nolint:exhaustive
	case BlockErase:
		sanType = "block-erase"
	case CryptoErase:
		sanType = "crypto-scramble"
	default:
		return fmt.Errorf("%w: %v", errSanitizeInvalidAction, sanact)
	}

	verify, err := ApplyWatermarks(drive)
	if err != nil {
		return err
	}

	h.Executor.SetArgs("--yes-i-know-what-i-am-doing", "--sanitize-"+sanType, drive.LogicalName)
	_, err = h.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	// now we loop until --sanitize-status reports that sanitization is complete
	for {
		h.Executor.SetArgs("--sanitize-status", drive.LogicalName)
		result, err := h.Executor.Exec(ctx)
		if err != nil {
			return err
		}
		if h.sanitizeDone(result.Stdout) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return verify()
}

func (h *Hdparm) sanitizeDone(output []byte) bool {
	return bytes.Contains(output, []byte("Sanitize Idle"))
}

// Erase wipes drive using ATA Secure Erase via hdparm --security-erase-enhanced
func (h *Hdparm) Erase(ctx context.Context, drive *common.Drive, ses SecureEraseSetting) error {
	switch ses { // nolint:exhaustive
	case CryptographicErase:
	default:
		return fmt.Errorf("%w: %v", errFormatInvalidSetting, ses)
	}

	h.Executor.SetArgs("--user-master", "u", "--security-set-pass", "p", drive.LogicalName)
	_, err := h.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	verify, err := ApplyWatermarks(drive)
	if err != nil {
		return err
	}

	h.Executor.SetArgs("--user-master", "u", "--security-erase-enhanced", "p", drive.LogicalName)
	_, err = h.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	return verify()
}

// NewFakeHdparm returns a mock hdparm collector that returns mock data for use in tests.
func NewFakeHdparm() *Hdparm {
	return &Hdparm{
		Executor: NewFakeExecutor("hdparm"),
	}
}
