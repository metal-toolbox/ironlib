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

type nvmeDeviceFeatures struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Enabled     bool   `json:"Enabled"`
}

type nvmeList struct {
	Devices []*nvmeDeviceAttributes `json:"Devices"`
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
		features, err := n.parseNvmeFeatures(d.DevicePath)
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

func (n *Nvme) ListFeatures(devicePath string) ([]byte, error) {
	// nvme id-ctrl -H devicepath
	n.Executor.SetArgs([]string{"id-ctrl", "-H", devicePath})

	result, err := n.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

// nolint:gocyclo // line parsing is cyclomatic
func (n *Nvme) parseNvmeFeatures(devicePath string) ([]nvmeDeviceFeatures, error) {
	out, err := n.ListFeatures(devicePath)
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
	reFnaStart := regexp.MustCompile(`(?s)^fna\s`)
	reFnaEnd := regexp.MustCompile(`(?s)^vwc\s`)
	reSaniStart := regexp.MustCompile(`(?s)^sanicap\s`)
	reSaniEnd := regexp.MustCompile(`(?s)^hmminds\s`)
	reBlank := regexp.MustCompile(`(?m)^\s*$`)

	var fnaBool, saniBool bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		fnaStart := reFnaStart.MatchString(line)
		fnaEnd := reFnaEnd.MatchString(line)
		saniStart := reSaniStart.MatchString(line)
		saniEnd := reSaniEnd.MatchString(line)
		isBlank := reBlank.MatchString(line)

		// start/end match specific block delimiters
		// bools are toggled to indicate lines within a given block
		switch {
		case fnaStart:
			fnaBool = true
		case fnaEnd:
			fnaBool = false
		case saniStart:
			saniBool = true
		case saniEnd:
			saniBool = false
		}

		switch {
		case (fnaStart || saniStart):
			var feature nvmeDeviceFeatures

			var partsLen = 2

			parts := strings.Split(line, ":")
			if len(parts) == partsLen {
				key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
				feature.Name = key

				if value != "0" {
					feature.Enabled = true
				}
			}

			if fnaStart {
				feature.Description = "Crypto Erase Support"
			} else {
				feature.Description = "Sanitize Support"
			}

			features = append(features, feature)

			// crypto erase
		case (fnaBool && !fnaEnd && !isBlank):
			var feature nvmeDeviceFeatures

			var partsLen = 3

			parts := strings.Split(line, ":")
			if len(parts) == partsLen {
				data := strings.Split(parts[2], "\t")
				enabled := strings.TrimSpace(data[0])

				if enabled != "0" {
					feature.Enabled = true
				}

				// Generate short flag identifier
				for _, word := range strings.Fields(data[1]) {
					feature.Name += strings.ToLower(word[0:1])
				}

				feature.Description = data[1]
			}

			features = append(features, feature)

			// sanitize
		case (saniBool && !saniEnd && !isBlank):
			var feature nvmeDeviceFeatures

			var partsLen = 3

			parts := strings.Split(line, ":")
			if len(parts) == partsLen {
				data := strings.Split(parts[2], "\t")
				enabled := strings.TrimSpace(data[0])

				if enabled != "0" {
					feature.Enabled = true
				}

				// Generate short flag identifier
				for _, word := range strings.Fields(data[1]) {
					feature.Name += strings.ToLower(word[0:1])
				}

				feature.Description = data[1]
			}

			features = append(features, feature)
		}
	}

	return features, err
}
