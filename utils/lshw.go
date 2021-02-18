package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/packethost/ironlib/model"
	"github.com/pkg/errors"
)

const lshw = "/usr/bin/lshw"

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
type LshwNodeCapabilities map[string]string

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
func (l *Lshw) Inventory() (*model.Device, error) {

	// The device we're taking inventory of
	l.Device = &model.Device{}

	// lshw output
	lshwDevice, err := l.ListJSON()
	if err != nil {
		return nil, errors.Wrap(err, "error parsing lshw output")
	}

	// System
	if lshwDevice == nil {
		return nil, fmt.Errorf("invalid data parsed from lshw")
	}

	for _, parentNode := range *lshwDevice {
		if parentNode == nil {
			continue
		}
		// collect components
		l.recurseNodes(parentNode.ChildNodes)
	}

	return l.Device, nil
}

// ListJSON returns the lshw output as a struct
func (l *Lshw) ListJSON() (*LshwOutput, error) {
	// lshw -json
	l.Executor.SetArgs([]string{"-json"})
	result, err := l.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	output := make(LshwOutput, 0)
	json.Unmarshal(result.Stdout, &output)

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

// nolint: gocyclo
// parseNode identifies the node component type and adds them to the device
func (l *Lshw) parseNode(node *LshwNode) {
	//fmt.Println(node.Class)
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
			l.Device.StorageController = append(l.Device.StorageController, sController)
		}
	}

}

func (l *Lshw) xMainboard(node *LshwNode) *model.Mainboard {

	if !(node.Class == "bus" && node.ID == "core") {
		return nil
	}

	//spew.Dump(node)
	return &model.Mainboard{
		Description:     node.Description,
		Vendor:          node.Vendor,
		Model:           node.Product,
		Serial:          node.Serial,
		PhysicalID:      node.Physid,
		FirmwareManaged: true,
	}
}

func (l *Lshw) xBIOS(node *LshwNode) *model.BIOS {
	return &model.BIOS{
		Description:       node.Description,
		Vendor:            node.Vendor,
		SizeBytes:         node.Size,
		CapacityBytes:     node.Capacity,
		FirmwareInstalled: node.Version,
		FirmwareManaged:   true,
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

	if !(strings.Contains(node.ID, "disk") &&
		node.Class == "disk" &&
		// skip Virtual Floppy/HDisk disks
		!strings.Contains(node.Product, "Virtual")) {
		return nil
	}

	return &model.Drive{
		Description: node.Description,
		Model:       node.Product,
		Serial:      node.Serial,
		SizeBytes:   node.Size,
	}
}

// Returns Storage controller information struct populated with the attributes identified by lshw
func (l *Lshw) xStorageController(node *LshwNode) *model.StorageController {

	if !(strings.Contains(node.ID, "sata") && node.Class == "storage") {
		return nil
	}

	return &model.StorageController{
		Description: node.Description,
		Vendor:      node.Vendor,
		Model:       node.Product,
		Serial:      node.Serial,
		PhysicalID:  node.Physid,
	}

}
