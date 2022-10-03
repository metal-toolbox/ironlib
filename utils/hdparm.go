package utils

import (
	"bufio"
	"context"
	"regexp"
	"strings"
)

const hdparm = "/usr/sbin/hdparm"

type Hdparm struct {
	Executor Executor
}

type hdparmDeviceFeatures struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Enabled     bool   `json:"Enabled"`
}

// Return a new hdparm executor
func NewHdparmCmd(trace bool) *Hdparm {
	e := NewExecutor(hdparm)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Hdparm{Executor: e}
}

func (h *Hdparm) ListFeatures(device string) ([]byte, error) {
	// hdparm -I devicepath
	h.Executor.SetArgs([]string{"-I", device})

	result, err := h.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

// nolint:gocyclo // line parsing is cyclomatic
func (h *Hdparm) parseHdparmFeatures(device string) ([]hdparmDeviceFeatures, error) {
	out, err := h.ListFeatures(device)
	if err != nil {
		return nil, err
	}

	var features []hdparmDeviceFeatures

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
	sfi := strings.NewReplacer("_", " ", "-", " ", "{", "", "}", "", "(", "", ")",
		"", ",", "", "|", "", "set", "", "command", "")
	featBool, secBool := false, false

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

		// Parse command features
		if featBool && !strings.Contains(line, featStart) {
			if strings.Contains(line, "*") {
				line = strings.TrimSpace(strings.TrimPrefix(line, "*\t"))

				// Generate short flag identifier
				line = sfi.Replace(line)
				for _, word := range strings.Fields(line) {
					flag += strings.ToLower(word[0:1])
				}

				var feature hdparmDeviceFeatures
				feature.Name = flag
				feature.Description = line
				feature.Enabled = true
				features = append(features, feature)
			} else if !strings.Contains(line, "*") && !strings.Contains(line, featStart) {
				// Generate short flag identifier
				line = strings.TrimSpace(sfi.Replace(line))
				for _, word := range strings.Fields(line) {
					flag += strings.ToLower(word[0:1])
				}

				var feature hdparmDeviceFeatures
				feature.Name = flag
				feature.Description = line
				feature.Enabled = false
				features = append(features, feature)
			}
		} else if secBool {
			// Parse security features
			secSupported := supported.MatchString(line)
			if !strings.Contains(line, secStart) {
				var feature hdparmDeviceFeatures
				switch {
				case strings.Contains(line, "65534"):
					feature.Name = "pns"
					feature.Description = "password not set"
				case secSupported:
					feature.Name = "es"
					feature.Enabled = true
					feature.Description = "encryption supported"
				case strings.Contains(line, "not\tenabled"):
					feature.Name = "ena"
					feature.Description = "encryption not active"
				case strings.Contains(line, "not\tlocked"):
					feature.Name = "dnl"
					feature.Description = "device is not locked"
				case strings.Contains(line, "not\texpired"):
					feature.Name = "ene"
					feature.Description = "encryption not expired"
				case strings.Contains(line, "supported: enhanced erase"):
					feature.Name = "esee"
					feature.Enabled = true
					feature.Description = "encryption supports enhanced erase"
				case strings.Contains(line, "SECURITY ERASE UNIT"):
					seTime, sehTime := seu.Replace(parts[0]), seu.Replace(parts[5])
					feature.Name = "time" + seTime + "+" + sehTime
					feature.Description = "erase time: " + seTime + ", " + sehTime + " (enhanced)"
				}
				features = append(features, feature)
			}
		}
	}

	return features, err
}
