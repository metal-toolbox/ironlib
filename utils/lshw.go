package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/packethost/ironlib/model"
	"github.com/pkg/errors"
)

//

const lshw = "/usr/bin/lshw"

var (
	ErrParseLshwOutput         = errors.New("lshw output parse error")
	storageControllerInterface = map[string]string{"sas": "SAS", "sata": "SATA"}
)

// The lshw command
type Lshw struct {
	Executor Executor
	Device   *model.Device
}

// lshw JSON unmarshal data structure
type LshwOutput []*LshwNode

// fields of the ChildNodes in the lshw output
// theres some fields with non-string attributes, in these fields, which are currently ignored
type LshwNodeConfiguration map[string]string
type LshwNodeCapabilities map[string]interface{}

// lshw -json output is unmarshalled into this struct
// each ChildNode is a LshwNode with almost all of the same fields
// https://ezix.org/project/wiki/HardwareLiSter
type LshwNode struct {
	ID            string                `json:"id,omitempty"`
	Class         string                `json:"class,omitempty"`
	Claimed       bool                  `json:"claimed,omitempty"`
	Handle        string                `json:"handle,omitempty"`
	Description   string                `json:"description,omitempty"`
	Product       string                `json:"product,omitempty"`
	Vendor        string                `json:"vendor,omitempty"`
	Physid        string                `json:"physid,omitempty"`
	Businfo       string                `json:"businfo,omitempty"`
	Dev           string                `json:"dev,omitempty"`
	Slot          string                `json:"slot,omitempty"`
	Units         string                `json:"units,omitempty"`
	Size          int64                 `json:"size,omitempty"`
	Capacity      int64                 `json:"capacity,omitempty"`
	Clock         int64                 `json:"clock,omitempty"`
	Version       string                `json:"version,omitempty"`
	Serial        string                `json:"serial,omitempty"`
	Width         int                   `json:"width,omitempty"`
	ChildNodes    []*LshwNode           `json:"children,omitempty"`
	Configuration LshwNodeConfiguration `json:"configuration,omitempty"`
	Capabilities  LshwNodeCapabilities  `json:"capabilities,omitempty"`
}

// Return a new nvme executor
func NewLshwCmd(trace bool) *Lshw {
	e := NewExecutor(lshw)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Lshw{Executor: e}
}

// Inventory collects and returns device hardware inventory
// based on the data parsed from lshw
func (l *Lshw) Inventory(device *model.Device) error {
	// The device we're taking inventory of
	l.Device = device

	// lshw output
	lshwDevice, err := l.ListJSON()
	if err != nil {
		return errors.Wrap(err, ErrParseLshwOutput.Error())
	}

	// System
	if lshwDevice == nil {
		return ErrParseLshwOutput
	}

	for _, parentNode := range *lshwDevice {
		if parentNode == nil {
			continue
		}

		l.Device.Serial = parentNode.Serial
		l.Device.Vendor = parentNode.Vendor
		l.Device.Model = parentNode.Product
		// collect components
		l.recurseNodes(parentNode.ChildNodes)
	}

	return nil
}

// ListJSON returns the lshw output as a struct
func (l *Lshw) ListJSON() (*LshwOutput, error) {
	// lshw -json -notime
	l.Executor.SetArgs([]string{"-json", "-notime"})

	result, err := l.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	// since lshw vB.02.19.2, the json output is not an array
	// here we turn the data into an array if it isn't
	if !bytes.HasPrefix(result.Stdout, []byte("[")) {
		result.Stdout = append([]byte("["), result.Stdout...)
		result.Stdout = append(result.Stdout, []byte("]")...)
	}

	output := make(LshwOutput, 0)

	err = json.Unmarshal(result.Stdout, &output)
	if err != nil {
		return nil, errors.Wrap(err, ErrParseLshwOutput.Error())
	}

	return &output, nil
}

// recurse over LshwNodes and invoke parseNode to collect component data
func (l *Lshw) recurseNodes(nodes []*LshwNode) {
	for _, node := range nodes {
		if node == nil {
			continue
		}

		l.parseNode(node)
		l.recurseNodes(node.ChildNodes)
	}
}

// nolint:gocyclo // parseNode is cyclomatic
// parseNode identifies the node component type and adds them to the device
func (l *Lshw) parseNode(node *LshwNode) {
	switch node.Class {
	case "bus":
		mainboard := l.xMainboard(node)
		if mainboard != nil {
			l.Device.Mainboard = mainboard
		}
	case "memory":
		switch node.ID {
		// BIOS
		case "firmware":
			bios := l.xBIOS(node)
			if bios != nil {
				l.Device.BIOS = bios
			}
		default:
			// Memory DIMMs
			memoryModule := l.xMemoryModule(node)
			if memoryModule != nil {
				l.Device.Memory = append(l.Device.Memory, memoryModule)
			}
		}
	case "processor":
		cpu := l.xCPU(node)
		if cpu != nil {
			l.Device.CPUs = append(l.Device.CPUs, cpu)
		}
	case "network":
		nic := l.xNIC(node)
		if nic != nil {
			l.Device.NICs = append(l.Device.NICs, nic)
		}
	case "disk":
		drive := l.xDrive(node)
		if drive != nil {
			l.Device.Drives = append(l.Device.Drives, drive)
		}
	case "storage":
		sController := l.xStorageController(node)
		if sController != nil {
			l.Device.StorageControllers = append(l.Device.StorageControllers, sController)
			return
		}

		// NVMe devices show up as part of the storage class
		drive := l.xDrive(node)
		if drive != nil {
			l.Device.Drives = append(l.Device.Drives, drive)
		}
	}
}

func (l *Lshw) xMainboard(node *LshwNode) *model.Mainboard {
	if !(node.Class == "bus" && node.ID == "core") {
		return nil
	}

	return &model.Mainboard{
		Description: node.Description,
		Vendor:      node.Vendor,
		Model:       node.Product,
		Serial:      node.Serial,
		PhysicalID:  node.Physid,
	}
}

func (l *Lshw) xBIOS(node *LshwNode) *model.BIOS {
	return &model.BIOS{
		Description:   node.Description,
		Vendor:        node.Vendor,
		SizeBytes:     node.Size,
		CapacityBytes: node.Capacity,
		Firmware: &model.Firmware{
			Installed: node.Version,
			Managed:   true,
		},
	}
}

// Returns physical memory module struct populated with the attributes identified by lshw
func (l *Lshw) xMemoryModule(node *LshwNode) *model.Memory {
	// find all populated memory banks
	if !(strings.Contains(node.ID, "bank") &&
		node.Class == "memory" &&
		node.Vendor != "NO DIMM") {
		return nil
	}

	return &model.Memory{
		Description:  node.Description,
		Slot:         node.Slot,
		Serial:       node.Serial,
		SizeBytes:    node.Size,
		Model:        node.Product,
		Vendor:       node.Vendor,
		ClockSpeedHz: node.Clock,
	}
}

// Returns CPU information struct populated with the attributes identified by lshw
func (l *Lshw) xCPU(node *LshwNode) *model.CPU {
	if !(strings.Contains(node.ID, "cpu") && node.Class == "processor") {
		return nil
	}

	// parse out cores and thread count
	var cores, threads int

	if node.Configuration != nil {
		c, defined := node.Configuration["cores"]
		if defined {
			c, err := strconv.Atoi(c)
			if err == nil {
				cores = c
			}
		}

		t, defined := node.Configuration["threads"]
		if defined {
			t, err := strconv.Atoi(t)
			if err == nil {
				threads = t
			}
		}
	}

	return &model.CPU{
		ClockSpeedHz: node.Clock,
		Description:  node.Product,
		Vendor:       node.Vendor,
		Model:        node.Product,
		Slot:         node.Slot,
		Cores:        cores,
		Threads:      threads,
	}
}

// Returns NIC information struct populated with the attributes identified by lshw
func (l *Lshw) xNIC(node *LshwNode) *model.NIC {
	if !(strings.Contains(node.ID, "network") &&
		node.Class == "network" &&
		// node.Handle is set to "PCI:-"
		// bonded/virtual/usb ether interfaces have this field empty
		node.Handle != "") {
		return nil
	}

	return &model.NIC{
		Description: node.Description,
		Vendor:      node.Vendor,
		Model:       node.Product,
		Serial:      node.Serial,
		SpeedBits:   node.Capacity,
		PhysicalID:  node.Physid,
	}
}

// Returns Drive information struct populated with the attributes identified by lshw
func (l *Lshw) xDrive(node *LshwNode) *model.Drive {
	if strings.Contains(node.Product, "Virtual") || node.Product == "" {
		return nil
	}

	return &model.Drive{
		Description: node.Description,
		Model:       node.Product,
		Vendor:      node.Vendor,
		Serial:      node.Serial,
		SizeBytes:   node.Size,
	}
}

// Returns Storage controller information struct populated with the attributes identified by lshw
func (l *Lshw) xStorageController(node *LshwNode) *model.StorageController {
	if node.Class != "storage" {
		return nil
	}

	intf, exists := storageControllerInterface[node.ID]
	if !exists {
		return nil
	}

	return &model.StorageController{
		Description: node.Description,
		Vendor:      node.Vendor,
		Model:       node.Product,
		Serial:      node.Serial,
		Interface:   intf,
		PhysicalID:  node.Physid,
	}
}

// FakeLshwExecute implements the utils.Executor interface for testing
type FakeLshwExecute struct {
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

// NewFakeLshwExecutor returns a fake lshw executor for tests
func NewFakeLshwExecutor(cmd string) Executor {
	return &FakeLshwExecute{Cmd: cmd}
}

// NewFakeLshw returns a fake lshw executor for testing
func NewFakeLshw(stdin io.Reader) *Lshw {
	executor := NewFakeLshwExecutor("lshw")
	executor.SetStdin(stdin)

	return &Lshw{Executor: executor}
}

// ExecWithContext implements the utils.Executor interface
func (e *FakeLshwExecute) ExecWithContext(ctx context.Context) (*Result, error) {
	b := bytes.Buffer{}

	_, err := b.ReadFrom(e.Stdin)
	if err != nil {
		return nil, err
	}

	return &Result{Stdout: b.Bytes()}, nil
}

// SetStdin is to set input to the fake execute method
func (e *FakeLshwExecute) SetStdin(r io.Reader) {
	e.Stdin = r
}

// SetArgs is to set cmd args to the fake execute method
func (e *FakeLshwExecute) SetArgs(a []string) {
	e.Args = a
}
