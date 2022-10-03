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
	feat_start := "Enabled"
	sec_start := "Security:"
	sec_end := "Logical Unit"

	supported := regexp.MustCompile(`(?s)^supported$`)
	seu := strings.NewReplacer("min", "")
	sfi := strings.NewReplacer("_", " ", "-", " ", "{", "", "}", "", "(", "", ")",
		"", ",", "", "|", "", "set", "", "command", "")
	feat_bool, sec_bool := false, false

	for _, line := range lines {
		line := strings.TrimSpace(line)
		parts := strings.Fields(line)
		var flag string

		// start/end match specific block delimiters
		// bools are toggled to indicate lines within a given block
		switch {
		case strings.Contains(line, feat_start):
			feat_bool = true
		case strings.Contains(line, sec_start):
			feat_bool, sec_bool = false, true
		case strings.Contains(line, sec_end):
			feat_bool, sec_bool = false, false
		}

		// Parse command features
		if feat_bool && !strings.Contains(line, feat_start) {
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

			} else if !strings.Contains(line, "*") && !strings.Contains(line, feat_start) {
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
			// Parse security features
		} else if sec_bool {
			sec_supported := supported.MatchString(line)
			if !strings.Contains(line, sec_start) {
				var feature hdparmDeviceFeatures
				switch {
				case strings.Contains(line, "65534"):
					feature.Name = "pns"
					feature.Description = "password not set"
				case sec_supported:
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
					se_time, seh_time := seu.Replace(parts[0]), seu.Replace(parts[5])
					feature.Name = "time" + se_time + "+" + seh_time
					feature.Description = "erase time: " + se_time + ", " + seh_time + " (enhanced)"
				}
				features = append(features, feature)
			}
		}

	}
	return features, err
}
