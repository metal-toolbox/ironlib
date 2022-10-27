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

	out, err := n.list(ctx)
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

		// Collect drive capabilitiesFound
		capabilitiesFound, err := n.DriveCapabilities(ctx, d.DevicePath)
		if err != nil {
			return nil, err
		}

		for _, f := range capabilitiesFound {
			drive.Common.Metadata[f.Description] = strconv.FormatBool(f.Enabled)
		}

		drives = append(drives, drive)
	}

	return drives, nil
}

func (n *Nvme) list(ctx context.Context) ([]byte, error) {
	// nvme list --output-format=json
	n.Executor.SetArgs([]string{"list", "--output-format=json"})

	result, err := n.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

func (n *Nvme) cmdListCapabilities(ctx context.Context, logicalPath string) ([]byte, error) {
	// nvme id-ctrl -H devicepath
	n.Executor.SetArgs([]string{"id-ctrl", "-H", logicalPath})

	result, err := n.Executor.ExecWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

// DriveCapabilities returns the drive capability attributes obtained through hdparm
//
// The logicalName is the kernel/OS assigned drive name - /dev/nvmeX
//
// This method implements the actions.DriveCapabilityCollector interface.
//
// nolint:gocyclo // line parsing is cyclomatic
func (n *Nvme) DriveCapabilities(ctx context.Context, logicalName string) ([]*common.Capability, error) {
	out, err := n.cmdListCapabilities(ctx, logicalName)
	if err != nil {
		return nil, err
	}

	var capabilitiesFound []*common.Capability

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
			capability := new(common.Capability)

			var partsLen = 2

			parts := strings.Split(line, ":")
			if len(parts) != partsLen {
				continue
			}

			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			capability.Name = key

			if value != "0" {
				capability.Enabled = true
			}

			if fnaStart {
				capability.Description = "Crypto Erase Support"
			} else {
				capability.Description = "Sanitize Support"
			}

			if value != "0" {
				capability.Enabled = true
			}

			capabilitiesFound = append(capabilitiesFound, capability)

			// crypto erase
		case (fnaBool && !fnaEnd && !isBlank):
			capability := new(common.Capability)

			var partsLen = 3

			parts := strings.Split(line, ":")
			if len(parts) != partsLen {
				continue
			}

			data := strings.Split(parts[2], "\t")
			enabled := strings.TrimSpace(data[0])

			if enabled != "0" {
				capability.Enabled = true
			}

			// Generate short flag identifier
			for _, word := range strings.Fields(data[1]) {
				capability.Name += strings.ToLower(word[0:1])
			}

			capability.Description = data[1]

			if enabled != "0" {
				capability.Enabled = true
			}

			capability.Description = data[1]
			capabilitiesFound = append(capabilitiesFound, capability)
			// sanitize
		case (saniBool && !saniEnd && !isBlank):
			capability := new(common.Capability)

			var partsLen = 3

			parts := strings.Split(line, ":")
			if len(parts) != partsLen {
				continue
			}

			data := strings.Split(parts[2], "\t")
			enabled := strings.TrimSpace(data[0])

			if enabled != "0" {
				capability.Enabled = true
			}

			// Generate short flag identifier
			for _, word := range strings.Fields(data[1]) {
				capability.Name += strings.ToLower(word[0:1])
			}

			capability.Description = data[1]

			if enabled != "0" {
				capability.Enabled = true
			}

			capability.Description = data[1]
			capabilitiesFound = append(capabilitiesFound, capability)
		}
	}

	return capabilitiesFound, err
}

// NewFakeNvme returns a mock nvme collector that returns mock data for use in tests.
func NewFakeNvme() *Nvme {
	return &Nvme{
		Executor: NewFakeExecutor("nvme"),
	}
}
