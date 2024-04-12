package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

const (
	EnvMvcliUtility = "IRONLIB_UTIL_MVCLI"
	BitsUint8       = 8
	BitsInt64       = 64
)

// RAID Modes:
// 0 - RAID-0 (Striping)
// 1 - RAID-1 (Mirroring)
// 10 - RAID-1+0 (Striped Mirrors)
// 5 - RAID-5 (Striping with Parity)
// 1e - RAID-1e (Striped Mirroring)
// hc - ?
// hs - ?
// hybrid - ?

var (
	validRaidModes  = []string{"0", "1", "10", "5", "1e", "hc", "hs", "hybrid"}
	validInfoTypes  = []string{"hba", "pd", "vd"}
	validBlockSizes = []uint{16, 32, 64, 128}
	validInitModes  = []string{"quick", "none", "intelligent"}
)

// Mvcli is a mvcli command executor object
type Mvcli struct {
	Executor Executor
}

// MvcliDevice is a marvell device object
type MvcliDevice struct {
	ID                 int
	Name               string
	Status             string
	Product            string
	SubProduct         string
	Model              string
	Serial             string
	SupportedRAIDModes string
	Firmware           string
	FirmwareRom        string
	FirmwareBios       string
	FirmwareBootLoader string
	SSDType            string
	Type               string
	CurrentSpeed       int64
	Size               int64
	PDSize             int64
	AdapterID          int
}

var (
	ErrInvalidInfoType      = errors.New("invalid info type")
	ErrInvalidRaidMode      = errors.New("invalid raid mode")
	ErrInvalidBlockSize     = errors.New("invalid block size")
	ErrInvalidInitMode      = errors.New("invalid init mode")
	ErrInvalidVirtualDiskID = errors.New("invalid virtual disk id")
	ErrDestroyVirtualDisk   = errors.New("failed to destroy virtual disk")
	ErrCreateVirtualDisk    = errors.New("failed to create virtual disk")
)

func CreateVirtualDiskError(createStdout []byte) error {
	return wrapError(ErrDestroyVirtualDisk, "stdout", string(createStdout))
}

func DestroyVirtualDiskError(destroyStdout []byte) error {
	return wrapError(ErrDestroyVirtualDisk, "stdout", string(destroyStdout))
}

func InvalidVirtualDiskIDError(virtualDiskID int) error {
	return wrapError(ErrInvalidVirtualDiskID, "virtualDiskID", virtualDiskID)
}

func InvalidInfoTypeError(infoType string) error {
	return wrapError(ErrInvalidInfoType, "infoType", infoType)
}

func InvalidRaidModeError(raidMode string) error {
	return wrapError(ErrInvalidRaidMode, "raidMode", raidMode)
}

func InvalidBlockSizeError(blockSize uint) error {
	return wrapError(ErrInvalidBlockSize, "blockSize", blockSize)
}

func InvalidInitModeError(initMode string) error {
	return wrapError(ErrInvalidInitMode, "initMode", initMode)
}

func wrapError(err error, k string, v interface{}) error {
	return errors.Wrap(err, fmt.Sprintf("%s=%s", k, v))
}

// Return a new mvcli executor
func NewMvcliCmd(trace bool) *Mvcli {
	utility := "mvcli"

	// lookup env var for util
	if eVar := os.Getenv(EnvMvcliUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Mvcli{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (m *Mvcli) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := m.Executor.CheckExecutable()

	return "mvcli", m.Executor.CmdPath(), er
}

// Return a Fake mvcli executor for tests
func NewFakeMvcli(r io.Reader) (*Mvcli, error) {
	e := NewFakeExecutor("mvcli")
	b := bytes.Buffer{}

	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	e.SetStdout(b.Bytes())

	return &Mvcli{
		Executor: e,
	}, nil
}

func (m *Mvcli) StorageControllers(ctx context.Context) ([]*common.StorageController, error) {
	devices, err := m.Info(ctx, "hba")
	if err != nil {
		return nil, err
	}

	hbas := []*common.StorageController{}

	for _, d := range devices {
		hba := &common.StorageController{
			Common: common.Common{
				Model:       d.Product,
				Vendor:      common.VendorFromString(d.Product),
				Description: d.Product + " (" + d.SubProduct + ")",
				Serial:      d.Serial,
				Metadata:    make(map[string]string),
				Firmware: &common.Firmware{
					Installed: d.Firmware,
					Metadata: map[string]string{
						"bios_version":        d.FirmwareBios,
						"rom_version":         d.FirmwareRom,
						"boot_loader_version": d.FirmwareBootLoader,
					},
				},
			},

			SupportedRAIDTypes: d.SupportedRAIDModes,
		}

		hbas = append(hbas, hba)
	}

	return hbas, nil
}

func (m *Mvcli) Drives(ctx context.Context) ([]*common.Drive, error) {
	devices, err := m.Info(ctx, "pd")
	if err != nil {
		return nil, err
	}

	drives := []*common.Drive{}

	for _, d := range devices {
		drive := &common.Drive{
			Common: common.Common{
				Model:       d.Model,
				Vendor:      common.VendorFromString(d.Model),
				Description: d.Model,
				Serial:      d.Serial,
				Firmware:    &common.Firmware{Installed: d.Firmware},
				Metadata:    make(map[string]string),
			},

			BlockSizeBytes:           d.Size,
			CapacityBytes:            d.PDSize,
			Type:                     m.processDriveType(d.Type, d.SSDType),
			NegotiatedSpeedGbps:      d.CurrentSpeed,
			StorageControllerDriveID: d.ID,
		}

		drives = append(drives, drive)
	}

	return drives, nil
}

func (m *Mvcli) Info(ctx context.Context, infoType string) ([]*MvcliDevice, error) {
	if !slices.Contains(validInfoTypes, infoType) {
		return nil, InvalidInfoTypeError(infoType)
	}

	m.Executor.SetArgs([]string{"info", "-o", infoType})

	result, err := m.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrNoCommandOutput, m.Executor.GetCmd())
	}

	return m.parseMvcliInfoOutput(infoType, result.Stdout), nil
}

func (m *Mvcli) parseMvcliInfoOutput(infoType string, b []byte) []*MvcliDevice {
	switch infoType {
	case "hba":
		return m.parseMvcliInfoHbaOutput(b)
	case "pd":
		return m.parseMvcliInfoPdOutput(b)
	case "vd":
		return m.parseMvcliInfoVdOutput(b)
	}

	return nil
}

func (m *Mvcli) parseMvcliInfoHbaOutput(b []byte) []*MvcliDevice {
	devices := []*MvcliDevice{}
	blocks := parseBytesForBlocks("Adapter ID", b)

	for _, block := range blocks {
		device := &MvcliDevice{
			Product:            block["Product"],
			SubProduct:         block["Sub Product"],
			FirmwareRom:        block["Rom version"],
			FirmwareBios:       block["BIOS version"],
			Firmware:           block["Firmware version"],
			FirmwareBootLoader: block["Boot loader version"],
			SupportedRAIDModes: block["Supported RAID mode"],
			Serial:             strings.Replace(strings.ToUpper(block["Product"]), "-", ":", 1),
		}
		devices = append(devices, device)
	}

	return devices
}

func stringToInt64(s string, b int) int64 {
	i, _ := strconv.ParseInt(s, 0, b)
	return i
}

func (m *Mvcli) parseMvcliInfoPdOutput(b []byte) []*MvcliDevice {
	const oneK = 1000

	devices := []*MvcliDevice{}

	for _, block := range parseBytesForBlocks("Adapter", b) {
		device := &MvcliDevice{
			Model:        block["model"],
			Serial:       block["Serial"],
			Firmware:     block["Firmware version"],
			Type:         block["Type"],
			SSDType:      block["SSD Type"],
			Size:         stringToInt64(strings.TrimSuffix(block["Size"], " K"), BitsInt64) * oneK,
			PDSize:       stringToInt64(strings.TrimSuffix(block["PD valid size"], " K"), BitsInt64) * oneK,
			AdapterID:    int(stringToInt64(block["Adapter"], BitsUint8)),
			ID:           int(stringToInt64(block["PD ID"], BitsUint8)),
			CurrentSpeed: stringToInt64(strings.TrimSuffix(block["Current speed"], " Gb/s"), BitsInt64),
		}

		devices = append(devices, device)
	}

	return devices
}

func (m *Mvcli) parseMvcliInfoVdOutput(b []byte) []*MvcliDevice {
	const oneM = 1000000

	devices := []*MvcliDevice{}

	for _, block := range parseBytesForBlocks("id:", b) {
		device := &MvcliDevice{
			ID:     int(stringToInt64(block["id"], BitsUint8)),
			Name:   block["name"],
			Status: block["status"],
			Size:   stringToInt64(strings.TrimSuffix(block["size"], " M"), BitsInt64) * oneM,
			Type:   block["RAID mode"],
		}

		devices = append(devices, device)
	}

	return devices
}

func (m *Mvcli) CreateVirtualDisk(ctx context.Context, raidMode string, physicalDisks []uint, name string, blockSize uint) error {
	return m.Create(ctx, physicalDisks, raidMode, name, blockSize, false, "quick")
}

func (m *Mvcli) DestroyVirtualDisk(ctx context.Context, virtualDiskID int) error {
	if vd := m.FindVdByID(ctx, virtualDiskID); vd == nil {
		return InvalidVirtualDiskIDError(virtualDiskID)
	}

	return m.Destroy(ctx, virtualDiskID)
}

func (m *Mvcli) processDriveType(pdType, ssdType string) string {
	if pdType == "SATA PD" && ssdType == "SSD" {
		return common.SlugDriveTypeSATASSD
	}

	return "Unknown"
}

func parseBytesForBlocks(blockStart string, b []byte) []map[string]string {
	blocks := []map[string]string{}

	byteSlice := bytes.Split(b, []byte("\n"))
	for idx, sl := range byteSlice {
		s := string(sl)
		if strings.Contains(s, blockStart) {
			block := parseKeyValueBlock(byteSlice[idx:])
			if block != nil {
				blocks = append(blocks, block)
			}
		}
	}

	return blocks
}

func parseKeyValueBlock(bSlice [][]byte) map[string]string {
	kv := make(map[string]string)

	for _, line := range bSlice {
		// A blank line means we've reached the end of this record
		if len(line) == 0 {
			return kv
		}

		s := string(line)
		cols := 2
		parts := strings.Split(s, ":")

		// Skip if there's no value
		if len(parts) < cols {
			continue
		}

		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		kv[key] = value
	}

	return kv
}

func (m *Mvcli) Create(ctx context.Context, physicalDiskIDs []uint, raidMode, name string, blockSize uint, _ bool, initMode string) error {
	if !slices.Contains(validRaidModes, raidMode) {
		return InvalidRaidModeError(raidMode)
	}

	if !slices.Contains(validBlockSizes, blockSize) {
		return InvalidBlockSizeError(blockSize)
	}

	if !slices.Contains(validInitModes, initMode) {
		return InvalidInitModeError(initMode)
	}

	m.Executor.SetArgs([]string{
		"create",
		"-o", "vd",
		"-r", raidMode,
		"-d", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(physicalDiskIDs)), ","), "[]"),
		"-n", name,
		"-b", fmt.Sprintf("%d", blockSize),
	})

	result, err := m.Executor.ExecWithContext(ctx)
	if err != nil {
		return err
	}

	if len(result.Stdout) == 0 {
		return errors.Wrap(ErrNoCommandOutput, m.Executor.GetCmd())
	}

	// Possible errors:
	// Specified RAID mode is not supported.
	// Gigabyte rounding scheme is not supported

	if match, _ := regexp.MatchString(`^SG driver version \S+\n$`, string(result.Stdout)); !match {
		return CreateVirtualDiskError(result.Stdout)
	}

	return nil
}

func (m *Mvcli) Destroy(ctx context.Context, virtualDiskID int) error {
	m.Executor.SetStdin(bytes.NewReader([]byte("y\n")))
	m.Executor.SetArgs([]string{
		"delete",
		"-o", "vd",
		"-i", fmt.Sprintf("%d", virtualDiskID),
	})

	result, err := m.Executor.ExecWithContext(ctx)
	if err != nil {
		return err
	}

	if len(result.Stdout) == 0 {
		return errors.Wrap(ErrNoCommandOutput, m.Executor.GetCmd())
	}

	// Possible errors:
	// Unable to get status of VD \S (error 59: Specified virtual disk doesn't exist).

	if match, _ := regexp.MatchString(`Delete VD \S successfully.`, string(result.Stdout)); !match {
		return DestroyVirtualDiskError(result.Stdout)
	}

	return nil
}

func (m *Mvcli) FindVdByName(ctx context.Context, name string) *MvcliDevice {
	return m.FindVdBy(ctx, "Name", name)
}

func (m *Mvcli) FindVdByID(ctx context.Context, virtualDiskID int) *MvcliDevice {
	return m.FindVdBy(ctx, "ID", virtualDiskID)
}

func (m *Mvcli) FindVdBy(ctx context.Context, k string, v interface{}) *MvcliDevice {
	virtualDisks, err := m.VirtualDisks(ctx)
	if err != nil {
		return nil
	}

	for _, vd := range virtualDisks {
		switch lKey := strings.ToLower(k); lKey {
		case "id":
			if vd.ID == v {
				return vd
			}
		case "name":
			if vd.Name == v {
				return vd
			}
		}
	}

	return nil
}

func (m *Mvcli) VirtualDisks(ctx context.Context) ([]*MvcliDevice, error) {
	return m.Info(ctx, "vd")
}
