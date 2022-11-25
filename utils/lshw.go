package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
)

var (
	ErrParseLshwOutput         = errors.New("lshw output parse error")
	storageControllerInterface = map[string]string{"sas": "SAS", "sata": "SATA"}
)

// The lshw command
type Lshw struct {
	Executor   Executor
	Device     *common.Device
	nicSerials map[string]bool
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
	LogicalName   interface{}           `json:"logicalname,omitempty"`
	Dev           string                `json:"dev,omitempty"`
	Slot          string                `json:"slot,omitempty"`
	Units         string                `json:"units,omitempty"`
	Size          float64               `json:"size,omitempty"`
	Capacity      int64                 `json:"capacity,omitempty"`
	Clock         int64                 `json:"clock,omitempty"`
	Version       string                `json:"version,omitempty"`
	Serial        string                `json:"serial,omitempty"`
	Width         int                   `json:"width,omitempty"`
	ChildNodes    []*LshwNode           `json:"children,omitempty"`
	Configuration LshwNodeConfiguration `json:"configuration,omitempty"`
	Capabilities  LshwNodeCapabilities  `json:"capabilities,omitempty"`
}

const (
	EnvLshwUtility = "UTIL_LSHW"
)

// Return a new lshw executor
func NewLshwCmd(trace bool) *Lshw {
	utility := "lshw"

	// lookup env var for util
	if eVar := os.Getenv(EnvLshwUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Lshw{Executor: e, nicSerials: make(map[string]bool)}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (l *Lshw) Attributes() (utilName, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := l.Executor.CheckExecutable()

	return "lshw", l.Executor.CmdPath(), er
}

// Inventory collects and returns device hardware inventory
// based on the data parsed from lshw
//
// Implements the InventoryCollector interface
func (l *Lshw) Collect(ctx context.Context, device *common.Device) error {
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

		// overwrite vendor, model serial only if its unset
		if l.Device.Vendor == "" {
			l.Device.Vendor = parentNode.Vendor
		}

		if l.Device.Model == "" {
			l.Device.Model = parentNode.Product
		}

		if l.Device.Serial == "" {
			l.Device.Serial = parentNode.Serial
		}

		// collect components
		l.recurseNodes(parentNode.ChildNodes)
	}

	return nil
}

// ListJSON returns the lshw output as a struct
func (l *Lshw) ListJSON() (*LshwOutput, error) {
	// lshw -json -notime
	l.Executor.SetArgs([]string{"-json", "-notime", "-numeric"})

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
	case "power":
		powerSupply := l.xPSU(node)
		if powerSupply != nil {
			l.Device.PSUs = append(l.Device.PSUs, powerSupply)
			return
		}
	}
}

func (l *Lshw) xMainboard(node *LshwNode) *common.Mainboard {
	if !(node.Class == "bus" && node.ID == "core") {
		return nil
	}

	return &common.Mainboard{
		Common: common.Common{
			Description: node.Description,
			Vendor:      node.Vendor,
			Model:       node.Product,
			Serial:      node.Serial,
			ProductName: node.Product,
		},

		PhysicalID: node.Physid,
	}
}

func (l *Lshw) xBIOS(node *LshwNode) *common.BIOS {
	return &common.BIOS{
		Common: common.Common{
			Description: node.Description,
			Vendor:      node.Vendor,
			Firmware: &common.Firmware{
				Installed: node.Version,
			},
			Capabilities: l.xParseCapabilities(node.Capabilities),
		},

		SizeBytes:     int64(node.Size),
		CapacityBytes: node.Capacity,
	}
}

// Returns physical memory module struct populated with the attributes identified by lshw
func (l *Lshw) xMemoryModule(node *LshwNode) *common.Memory {
	// find all populated memory banks
	if !(strings.Contains(node.ID, "bank") &&
		node.Class == "memory" &&
		node.Vendor != "NO DIMM") {
		return nil
	}

	return &common.Memory{
		Common: common.Common{
			Description: node.Description,
			Vendor:      node.Vendor,
			Model:       node.Product,
			Serial:      node.Serial,
			ProductName: node.Product,
		},

		Slot:         node.Slot,
		SizeBytes:    int64(node.Size),
		ClockSpeedHz: node.Clock,
	}
}

// Returns CPU information struct populated with the attributes identified by lshw
func (l *Lshw) xCPU(node *LshwNode) *common.CPU {
	if !(strings.Contains(node.ID, "cpu") && node.Class == "processor") {
		return nil
	}

	// parse out cores and thread count
	var cores, threads int

	var firmware *common.Firmware

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

		microcode, defined := node.Configuration["microcode"]
		if defined {
			firmware = common.NewFirmwareObj()
			firmware.Installed = microcode
		}
	}

	return &common.CPU{
		Common: common.Common{
			Description:  node.Description,
			Vendor:       node.Vendor,
			Model:        node.Product,
			Serial:       node.Serial,
			ProductName:  node.Product,
			Firmware:     firmware,
			Capabilities: l.xParseCapabilities(node.Capabilities),
		},

		ClockSpeedHz: node.Clock,
		Slot:         node.Slot,
		Cores:        cores,
		Threads:      threads,
	}
}

// Returns NIC information struct populated with the attributes identified by lshw
func (l *Lshw) xNIC(node *LshwNode) *common.NIC {
	if !(strings.Contains(node.ID, "network") &&
		node.Class == "network" &&
		// node.Handle is set to "PCI:-"
		// bonded/virtual/usb ether interfaces have this field empty
		node.Handle != "") {
		return nil
	}

	// TODO(splaspood) We should merge on something other than serial
	if node.Serial == "" {
		log.Printf("Warn: NIC component without serial, ignored: %+v\n", node)
		return nil
	}

	serial := strings.ToLower(node.Serial)
	if l.nicSerials[serial] {
		return nil
	}

	l.nicSerials[serial] = true

	nic := &common.NIC{
		Common: common.Common{
			Description:  node.Description,
			Vendor:       node.Vendor,
			Model:        node.Product,
			Serial:       node.Serial,
			ProductName:  node.Product,
			Capabilities: l.xParseCapabilities(node.Capabilities),
		},

		Description: node.Description,
		SpeedBits:   node.Capacity,
		PhysicalID:  node.Physid,
		BusInfo:     node.Businfo,
	}

	nic.Common.PCIVendorID, nic.Common.PCIProductID, nic.Common.ProductName = lshwPciIDParse(node.Product)
	nic.Common.Model = nic.Common.ProductName

	// include additional attributes
	if node.Configuration != nil {
		nic.Metadata = map[string]string{}
		keys := []string{
			"link",
			"speed",
			"duplex",
			"firmware",
			"driver",
			"driver_version",
		}

		for _, key := range keys {
			value, exists := node.Configuration[key]
			if exists {
				if key == "firmware" {
					version := lshwNicFwStringParse(value, node.Vendor)
					nic.Firmware = &common.Firmware{Installed: version}
				}

				nic.Metadata[key] = value
			}
		}
	}

	return nic
}

// lshwPciIDParse returns PCI Vendor and Product identifiers from a given string
func lshwPciIDParse(s string) (vendor, product, sanitizied string) {
	pciIDRegex := regexp.MustCompile(` \[(\S{4}):?(\S{4})?\]$`)

	if matches := pciIDRegex.FindStringSubmatch(s); matches != nil {
		vendor = matches[1]

		const PciProductIDMatchLength = 3

		if len(matches) == PciProductIDMatchLength {
			product = matches[2]
		}
	}

	sanitizied = pciIDRegex.ReplaceAllString(s, "")

	return
}

// lshwNicFwStringParse returns the version component of the firmware string
func lshwNicFwStringParse(fw, vendor string) string {
	if fw == "" {
		return ""
	}

	vendor = strings.ToLower(vendor)

	switch {
	case strings.Contains(vendor, common.VendorIntel):
		return nicFwParseIntel(fw)
	case strings.Contains(vendor, common.VendorMellanox):
		return nicFwParseMellanox(fw)
	case strings.Contains(vendor, common.VendorBroadcom):
		return nicFwParseBroadcom(fw)
	default:
		return fw
	}
}

func nicFwParseIntel(s string) string {
	// The intel firmware version string returned when not empty is in 3 parts
	// where the last part is the actual firmware version
	// 7.10 0x800075df 19.5.12
	vParts := 3

	// unrecognized string returned as is
	if !strings.Contains(s, "0x") {
		return s
	}

	parts := strings.Split(s, " ")
	if len(parts) == vParts {
		return parts[vParts-1]
	}

	return s
}

func nicFwParseMellanox(s string) string {
	// The mellanox firmware version string returned when not empty is in 2 parts
	// where the first part is the actual firmware version
	// 14.27.1016 (MT_2420110034)
	vParts := 2

	// unrecognized string returned as is
	if !strings.Contains(s, "MT") {
		return s
	}

	parts := strings.Split(s, " ")
	if len(parts) == vParts {
		return parts[0]
	}

	return s
}

func nicFwParseBroadcom(s string) string {
	// The broadcom firmware version string returned when not empty is in 3 parts
	// where the last part is the actual firmware version
	// 5719-v1.46 NCSI v1.5.1.0
	vParts := 3

	parts := strings.Split(s, " ")
	if len(parts) == vParts {
		return parts[vParts-1]
	}

	return s
}

// Returns Drive information struct populated with the attributes identified by lshw
func (l *Lshw) xDrive(node *LshwNode) *common.Drive {
	if strings.Contains(node.Product, "Virtual") || node.Product == "" || strings.Contains(node.Description, "SATA controller") {
		return nil
	}

	drive := &common.Drive{
		Common: common.Common{
			Description: node.Description,
			Vendor:      node.Vendor,
			Model:       node.Product,
			Serial:      node.Serial,
			ProductName: node.Product,
		},

		BusInfo:       node.Businfo,
		CapacityBytes: int64(node.Size),
	}

	// type assert LogicalName since it can also be an array - in which case we
	// do not care about the data.
	logicalName, isString := node.LogicalName.(string)
	if isString {
		drive.LogicalName = logicalName
	}

	if drive.Vendor == "" {
		drive.Vendor = common.VendorFromString(node.Product)
	}

	return drive
}

// Returns Storage controller information struct populated with the attributes identified by lshw
func (l *Lshw) xStorageController(node *LshwNode) *common.StorageController {
	if node.Class != "storage" {
		return nil
	}

	intf, exists := storageControllerInterface[node.ID]
	if !exists {
		return nil
	}

	sc := &common.StorageController{
		Common: common.Common{
			Description: node.Description,
			Vendor:      node.Vendor,
			Model:       node.Product,
			Serial:      node.Serial,
			ProductName: node.Product,
		},

		SupportedDeviceProtocols: intf,
		PhysicalID:               node.Physid,
		BusInfo:                  node.Businfo,
	}

	sc.Common.PCIVendorID, sc.Common.PCIProductID, sc.ProductName = lshwPciIDParse(sc.ProductName)

	_, _, sc.Vendor = lshwPciIDParse(sc.Vendor)
	_, _, sc.Model = lshwPciIDParse(sc.Model)

	// If no serial number is present, derive the serial from the PCI vendor/product id
	if sc.Serial == "" {
		sc.Serial = sc.Common.PCIVendorID + ":" + sc.Common.PCIProductID
	}

	return sc
}

// Returns PSU information struct populated with the attributes identified by lshw
func (l *Lshw) xPSU(node *LshwNode) *common.PSU {
	if node.Class != "power" {
		return nil
	}

	return &common.PSU{
		Common: common.Common{
			Description: node.Description,
			Vendor:      node.Vendor,
			Model:       node.Product,
			Serial:      node.Serial,
			ProductName: node.Product,
		},

		ID:                 node.Physid,
		PowerCapacityWatts: node.Capacity,
	}
}

func (l *Lshw) xParseCapabilities(capabilities LshwNodeCapabilities) []*common.Capability {
	caps := make([]*common.Capability, 0)

	for key, v := range capabilities {
		switch value := v.(type) {
		default:
			continue
		case string:
			caps = append(caps, &common.Capability{
				Name:        key,
				Description: value,
				Enabled:     true,
			})

			continue

		case bool:
			caps = append(caps, &common.Capability{
				Name:    key,
				Enabled: value,
			})

			continue
		}
	}

	// sort capabilities by name - this helps in tests to have data that is consistent.
	if len(caps) > 0 {
		sort.SliceStable(caps, func(i, j int) bool {
			return caps[i].Name < caps[j].Name
		})

		return caps
	}

	return nil
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

	return &Lshw{Executor: executor, nicSerials: make(map[string]bool)}
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
