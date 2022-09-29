package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
)

const nvmecli = "/usr/sbin/nvme"

type Nvme struct {
	Executor Executor
}

type nvmeDeviceAttributes struct {
	Namespace    int    `json:"Namespace"`
	DevicePath   string `json:"DevicePath"`
	Firmware     string `json:"Firmware"`
	Index        int    `json:"Index"`
	ModelNumber  string `json:"ModelNumber"`
	ProductName  string `json:"ProductName"`
	SerialNumber string `json:"SerialNumber"`
}

// placeholder common.Drive.Features
type nvmeDeviceFeatures struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Enabled     bool   `json:"Enabled"`
}

type nvmeList struct {
	Devices []*nvmeDeviceAttributes `json:"Devices"`
}

type nvmeFeatures struct {
	Features []*nvmeDeviceFeatures `json:Features`
}

// Return a new nvme executor
func NewNvmeCmd(trace bool) *Nvme {
	e := NewExecutor(nvmecli)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Nvme{Executor: e}
}

// Executes nvme list, parses the output and returns a slice of *common.Drive
func (n *Nvme) Drives(ctx context.Context) ([]*common.Drive, error) {
	drives := make([]*common.Drive, 0)

	out, err := n.List()
	if err != nil {
		return nil, err
	}

	list := &nvmeList{Devices: []*nvmeDeviceAttributes{}}

	err = json.Unmarshal(out, list)
	if err != nil {
		return nil, err
	}

	for _, d := range list.Devices {
		dModel := d.ModelNumber

		var vendor string

		modelTokens := strings.Split(d.ModelNumber, " ")

		if len(modelTokens) > 1 {
			vendor = modelTokens[1]
		}

		drive := &common.Drive{
			Common: common.Common{
				Serial:      d.SerialNumber,
				Vendor:      vendor,
				Model:       dModel,
				ProductName: d.ProductName,
				Description: d.ModelNumber,
				Firmware: &common.Firmware{
					Installed: d.Firmware,
				},
				Metadata: map[string]string{},
			},
		}

		// Collect drive features
		features, err := n.parseNvmeFeatures(d)
		if err != nil {
			return nil, err
		}

		for _, f := range features {
			drive.Common.Metadata[f.Description] = strconv.FormatBool(f.Enabled)
		}

		drives = append(drives, drive)
	}

	return drives, nil
}

func (n *Nvme) List() ([]byte, error) {
	// nvme list --output-format=json
	n.Executor.SetArgs([]string{"list", "--output-format=json"})

	result, err := n.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

func (n *Nvme) ListFeatures(device string) ([]byte, error) {
	// nvme id-ctrl -H devicepath
	n.Executor.SetArgs([]string{"id-ctrl", "-H", device})

	result, err := n.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

func (n *Nvme) parseNvmeFeatures(d *nvmeDeviceAttributes) ([]nvmeDeviceFeatures, error) {
	out, err := n.ListFeatures(d.DevicePath)
	if err != nil {
		return nil, err
	}

	var features []nvmeDeviceFeatures

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
	re_fna_start := regexp.MustCompile(`(?s)^fna\s`)
	re_fna_end := regexp.MustCompile(`(?s)^vwc\s`)
	re_sani_start := regexp.MustCompile(`(?s)^sanicap\s`)
	re_sani_end := regexp.MustCompile(`(?s)^hmminds\s`)
	re_blank := regexp.MustCompile(`(?m)^\s*$`)

	fna_bool, sani_bool := false, false

	for _, line := range lines {
		line := strings.TrimSpace(line)
		fna_start := re_fna_start.MatchString(line)
		fna_end := re_fna_end.MatchString(line)
		sani_start := re_sani_start.MatchString(line)
		sani_end := re_sani_end.MatchString(line)
		is_blank := re_blank.MatchString(line)

		// start/end match specific block delimiters
		// bools are toggled to indicate lines within a given block
		switch {
		case fna_start:
			fna_bool = true
		case fna_end:
			fna_bool = false
		case sani_start:
			sani_bool = true
		case sani_end:
			sani_bool = false
		}

		if fna_start || sani_start {
			var feature nvmeDeviceFeatures
			parts := strings.Split(line, ":")
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			feature.Name = key
			if fna_start {
				feature.Description = "Crypto Erase Support"
			} else {
				feature.Description = "Sanitize Support"
			}
			if value != "0" {
				feature.Enabled = true
			}
			features = append(features, feature)

			// crypto erase
		} else if fna_bool && !fna_end && !is_blank {
			var feature nvmeDeviceFeatures
			parts := strings.Split(line, ":")
			data := strings.Split(parts[2], "\t")
			enabled := strings.TrimSpace(data[0])

			// Generate short flag identifier
			for _, word := range strings.Fields(data[1]) {
				feature.Name += strings.ToLower(word[0:1])
			}

			if enabled != "0" {
				feature.Enabled = true
			}
			feature.Description = data[1]
			features = append(features, feature)
			// sanitize
		} else if sani_bool && !sani_end && !is_blank {
			var feature nvmeDeviceFeatures
			var flag string
			parts := strings.Split(line, ":")
			data := strings.Split(parts[2], "\t")
			enabled := strings.TrimSpace(data[0])

			// Generate short flag identifier
			for _, word := range strings.Fields(data[1]) {
				flag += strings.ToLower(word[0:1])
			}

			if enabled != "0" {
				feature.Enabled = true
			}
			feature.Description = data[1]
			features = append(features, feature)
		}
	}

	return features, err
}
