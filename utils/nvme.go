package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/sirupsen/logrus"
)

const EnvNvmeUtility = "IRONLIB_UTIL_NVME"

var (
	errSanicapNODMMASReserved = errors.New("sanicap nodmmas reserved bits set, not sure what to do with them")
	errSanitizeInvalidAction  = errors.New("invalid sanitize action")
	errFormatInvalidSetting   = errors.New("invalid format setting")
	errInvalidCreateNSArgs    = errors.New("invalid ns-create args")
)

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
	utility := "nvme"

	// lookup env var for util
	if eVar := os.Getenv(EnvNvmeUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Nvme{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (n *Nvme) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := n.Executor.CheckExecutable()

	return "nvme", n.Executor.CmdPath(), er
}

// Executes nvme list, parses the output and returns a slice of *common.Drive
func (n *Nvme) Drives(ctx context.Context) ([]*common.Drive, error) {
	out, err := n.list(ctx)
	if err != nil {
		return nil, err
	}

	list := &nvmeList{Devices: []*nvmeDeviceAttributes{}}

	err = json.Unmarshal(out, list)
	if err != nil {
		return nil, err
	}

	drives := make([]*common.Drive, len(list.Devices))
	for i, d := range list.Devices {
		dModel := d.ModelNumber

		var vendor string

		modelTokens := strings.Split(d.ModelNumber, " ")

		if len(modelTokens) > 1 {
			vendor = modelTokens[1]
		}

		// Collect drive capabilitiesFound
		capabilitiesFound, err := n.DriveCapabilities(ctx, d.DevicePath)
		if err != nil {
			return nil, err
		}

		metadata := map[string]string{}
		for _, f := range capabilitiesFound {
			metadata[f.Description] = strconv.FormatBool(f.Enabled)
		}

		drives[i] = &common.Drive{
			Common: common.Common{
				LogicalName:  d.DevicePath,
				Serial:       d.SerialNumber,
				Vendor:       vendor,
				Model:        dModel,
				ProductName:  d.ProductName,
				Description:  d.ModelNumber,
				Firmware:     &common.Firmware{Installed: d.Firmware},
				Capabilities: capabilitiesFound,
				Metadata:     metadata,
			},
		}
	}

	return drives, nil
}

func (n *Nvme) list(ctx context.Context) ([]byte, error) {
	// nvme list --output-format=json
	n.Executor.SetArgs("list", "--output-format=json")

	result, err := n.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

func (n *Nvme) cmdListCapabilities(ctx context.Context, logicalName string) ([]byte, error) {
	// nvme id-ctrl --output-format=json devicepath
	n.Executor.SetArgs("id-ctrl", "--output-format=json", logicalName)
	result, err := n.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}

// DriveCapabilities returns the drive capability attributes obtained through nvme
//
// The device is the kernel/OS assigned drive name - /dev/nvmeX
//
// This method implements the actions.DriveCapabilityCollector interface.
func (n *Nvme) DriveCapabilities(ctx context.Context, logicalName string) ([]*common.Capability, error) {
	out, err := n.cmdListCapabilities(ctx, logicalName)
	if err != nil {
		return nil, err
	}

	var caps struct {
		FNA     uint `json:"fna"`
		SANICAP uint `json:"sanicap"`
	}

	err = json.Unmarshal(out, &caps)
	if err != nil {
		return nil, err
	}

	var capabilitiesFound []*common.Capability
	capabilitiesFound = append(capabilitiesFound, parseFna(caps.FNA)...)

	var parsedCaps []*common.Capability
	parsedCaps, err = parseSanicap(caps.SANICAP)
	if err != nil {
		return nil, err
	}
	capabilitiesFound = append(capabilitiesFound, parsedCaps...)

	return capabilitiesFound, nil
}

func parseFna(fna uint) []*common.Capability {
	// Bit masks values came from nvme-cli repo
	// All names come from internal nvme-cli names
	// We will *not* keep in sync as these names form our API
	// https: // github.com/linux-nvme/nvme-cli/blob/v2.8/nvme-print-stdout.c#L2199-L2217
	//
	// This function uses go's binary notation + bitshifts instead of masking of hex numbers
	// I think its easier to see whats going on this way vs hex masks
	// Refresher:
	//   1<<N makes gives us all zeros except for the Nth bit which is a one
	//   fna & 1<<N is bitwise and, the result will be 1 if fna had a 1 in Nth bi

	return []*common.Capability{
		{
			Name:        "fmns",
			Description: "Format Applies to All/Single Namespace(s) (t:All, f:Single)",
			Enabled:     fna&(0b1<<0) != 0,
		},
		{
			Name:        "cens",
			Description: "Crypto Erase Applies to All/Single Namespace(s) (t:All, f:Single)",
			Enabled:     fna&(0b1<<1) != 0,
		},
		{
			Name:        "cese",
			Description: "Crypto Erase Supported as part of Secure Erase",
			Enabled:     fna&(0b1<<2) != 0,
		},
	}
}

func parseSanicap(sanicap uint) ([]*common.Capability, error) {
	// Bit masks values came from nvme-cli repo
	// All names come from internal nvme-cli names
	// We will *not* keep in sync as these names form our API
	// https://github.com/linux-nvme/nvme-cli/blob/v2.8/nvme-print-stdout.c#L2064-L2093
	//
	// This function uses go's binary notation + bitshifts instead of masking of hex numbers
	// I think its easier to see whats going on this way vs hex masks
	// Refresher:
	//   1<<N makes gives us all zeros except for the Nth bit which is a one
	//   sanicap & 1<<N is bitwise and, the result will be 1 if sanicap had a 1 in Nth bit

	caps := []*common.Capability{
		{
			Name:        "cer",
			Description: "Crypto Erase Sanitize Operation Supported",
			Enabled:     sanicap&(0b1<<0) != 0,
		},
		{
			Name:        "ber",
			Description: "Block Erase Sanitize Operation Supported",
			Enabled:     sanicap&(0b1<<1) != 0,
		},
		{
			Name:        "owr",
			Description: "Overwrite Sanitize Operation Supported",
			Enabled:     sanicap&(0b1<<2) != 0,
		},
		{
			Name:        "ndi",
			Description: "No-Deallocate After Sanitize bit in Sanitize command Supported",
			Enabled:     sanicap&(0b1<<29) != 0,
		},
	}

	switch sanicap & (0b11 << 30) >> 30 {
	case 0b00:
		// nvme prints this out for 0b00:
		//   "Additional media modification after sanitize operation completes successfully is not defined"
		// So I'm taking "not defined" literally since we can't really represent 2 bits in a bool
		// If we ever want this as a bool we could maybe call it "dmmas" maybe?
	case 0b01, 0b10:
		caps = append(caps, &common.Capability{
			Name:        "nodmmas",
			Description: "Media is additionally modified after sanitize operation completes successfully",
			Enabled:     sanicap&(0b11<<30)>>30 == 0b10,
		})
	case 0b11:
		return nil, errSanicapNODMMASReserved
	}

	return caps, nil
}

//go:generate stringer -type SanitizeAction
type SanitizeAction uint8

const (
	Invalid SanitizeAction = iota
	ExitFailureMode
	BlockErase
	Overwrite
	CryptoErase
)

//go:generate stringer -type SecureEraseSetting
type SecureEraseSetting uint8

const (
	None SecureEraseSetting = iota
	UserDataErase
	CryptographicErase
	Reserved
)

// WipeDrive implements DriveWiper by running nvme sanitize or nvme format as appropriate
func (n *Nvme) WipeDrive(ctx context.Context, logger *logrus.Logger, logicalName string) error {
	caps, err := n.DriveCapabilities(ctx, logicalName)
	if err != nil {
		return fmt.Errorf("WipeDrive: %w", err)
	}
	return n.wipe(ctx, logger, logicalName, caps)
}

func (n *Nvme) wipe(ctx context.Context, logger *logrus.Logger, logicalName string, caps []*common.Capability) error {
	var ber bool
	var cer bool
	var cese bool
	for _, cap := range caps {
		switch cap.Name {
		case "ber":
			ber = cap.Enabled
		case "cer":
			cer = cap.Enabled
		case "cese":
			cese = cap.Enabled
		}
	}

	l := logger.WithField("drive", logicalName)
	if cer {
		// nolint:govet
		l := l.WithField("method", "sanitize").WithField("action", CryptoErase)
		l.Info("wiping")
		err := n.Sanitize(ctx, logicalName, CryptoErase)
		if err == nil {
			return nil
		}
		l.WithError(err).Info("failed")
	}
	if ber {
		// nolint:govet
		l := l.WithField("method", "sanitize").WithField("action", BlockErase)
		l.Info("wiping")
		err := n.Sanitize(ctx, logicalName, BlockErase)
		if err == nil {
			return nil
		}
		l.WithError(err).Info("failed")
	}
	if cese {
		// nolint:govet
		l := l.WithField("method", "format").WithField("setting", CryptographicErase)
		l.Info("wiping")
		err := n.Format(ctx, logicalName, CryptographicErase)
		if err == nil {
			return nil
		}
		l.WithError(err).Info("failed")
	}

	l = l.WithField("method", "format").WithField("setting", UserDataErase)
	l.Info("wiping")
	err := n.Format(ctx, logicalName, UserDataErase)
	if err == nil {
		return nil
	}
	l.WithError(err).Info("failed")
	return ErrIneffectiveWipe
}

func (n *Nvme) Sanitize(ctx context.Context, logicalName string, sanact SanitizeAction) error {
	switch sanact { // nolint:exhaustive
	case BlockErase, CryptoErase:
	default:
		return fmt.Errorf("%w: %v", errSanitizeInvalidAction, sanact)
	}

	verify, err := ApplyWatermarks(logicalName)
	if err != nil {
		return err
	}

	n.Executor.SetArgs("sanitize", "--sanact="+strconv.Itoa(int(sanact)), logicalName)
	_, err = n.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	// now we loop until sanitize-log reports that sanitization is complete
	dev := path.Base(logicalName)
	var log map[string]struct {
		Progress uint16 `json:"sprog"`
	}
	for {
		n.Executor.SetArgs("sanitize-log", "--output-format=json", logicalName)
		result, err := n.Executor.Exec(ctx)
		if err != nil {
			return err
		}
		err = json.Unmarshal(result.Stdout, &log)
		if err != nil {
			return err
		}

		l, ok := log[dev]
		if !ok {
			return fmt.Errorf("%s: device not present in sanitize-log: %w: %s", dev, io.ErrUnexpectedEOF, result.Stdout)
		}

		if l.Progress == 65535 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return verify()
}

func (n *Nvme) Format(ctx context.Context, logicalName string, ses SecureEraseSetting) error {
	switch ses { // nolint:exhaustive
	case UserDataErase, CryptographicErase:
	default:
		return fmt.Errorf("%w: %v", errFormatInvalidSetting, ses)
	}

	verify, err := ApplyWatermarks(logicalName)
	if err != nil {
		return err
	}

	n.Executor.SetArgs("format", "--ses="+strconv.Itoa(int(ses)), logicalName)
	_, err = n.Executor.Exec(ctx)
	if err != nil {
		return err
	}
	return verify()
}

func (n *Nvme) listNS(ctx context.Context, logicalName string) ([]uint, error) {
	n.Executor.SetArgs("list-ns", "--output-format=json", "--all", logicalName)
	result, err := n.Executor.Exec(ctx)
	if err != nil {
		return nil, err
	}

	list := struct {
		Namespaces []struct {
			ID uint `json:"nsid"`
		} `json:"nsid_list"`
	}{}
	err = json.Unmarshal(result.Stdout, &list)
	if err != nil {
		return nil, err
	}

	ret := make([]uint, len(list.Namespaces))
	for i := range list.Namespaces {
		ret[i] = list.Namespaces[i].ID
	}
	return ret, nil
}

func (n *Nvme) createNS(ctx context.Context, logicalName string, size, blocksize uint) (uint, error) {
	if blocksize == 0 {
		return 0, fmt.Errorf("%w: blocksize(0) is zero", errInvalidCreateNSArgs)
	}
	if size <= blocksize {
		return 0, fmt.Errorf("%w: size(%d) is not larger than blocksize(%d), arguments may be swapped", errInvalidCreateNSArgs, size, blocksize)
	}
	if size%blocksize != 0 {
		return 0, fmt.Errorf("%w: size(%d) is not a multiple of blocksize(%d)", errInvalidCreateNSArgs, size, blocksize)
	}

	_size := strconv.Itoa(int(size / blocksize))
	_blocksize := strconv.Itoa(int(blocksize))
	n.Executor.SetArgs("create-ns", logicalName, "--dps=0", "--nsze="+_size, "--ncap="+_size, "--blocksize="+_blocksize)
	result, err := n.Executor.Exec(ctx)
	if err != nil {
		return 0, err
	}

	// parse namespace id from stdout which looks like: `create-ns: Success, created nsid:1`
	out := bytes.TrimSpace(result.Stdout)
	parts := bytes.Split(out, []byte(":"))
	if len(parts) != 3 {
		return 0, fmt.Errorf("unable to parse nsid: %w", io.ErrUnexpectedEOF)
	}
	nsid, err := strconv.Atoi(string(parts[2]))
	return uint(nsid), err
}

func (n *Nvme) deleteNS(ctx context.Context, logicalName string, namespaceID uint) error {
	nsid := strconv.Itoa(int(namespaceID))
	n.Executor.SetArgs("delete-ns", logicalName, "--namespace-id="+nsid)
	_, err := n.Executor.Exec(ctx)
	return err
}

func (n *Nvme) attachNS(ctx context.Context, logicalName string, controllerID, namespaceID uint) error {
	cntlid := strconv.Itoa(int(controllerID))
	nsid := strconv.Itoa(int(namespaceID))
	n.Executor.SetArgs("attach-ns", logicalName, "--controllers="+cntlid, "--namespace-id="+nsid)
	_, err := n.Executor.Exec(ctx)
	return err
}

func (n *Nvme) idNS(ctx context.Context, logicalName string, namespaceID uint) ([]byte, error) {
	nsid := strconv.Itoa(int(namespaceID))
	n.Executor.SetArgs("id-ns", "--output-format=json", logicalName, "--namespace-id="+nsid)
	result, err := n.Executor.Exec(ctx)
	return result.Stdout, err
}

func (n *Nvme) ResetNS(ctx context.Context, logicalName string) error { // nolint:gocyclo
	out, err := n.cmdListCapabilities(ctx, logicalName)
	if err != nil {
		return err
	}

	ctrl := struct {
		CNTLID  uint `json:"cntlid"`
		TNVMCAP uint `json:"tnvmcap"`
	}{}
	err = json.Unmarshal(out, &ctrl)
	if err != nil {
		return err
	}

	namespaces, err := n.listNS(ctx, logicalName)
	if err != nil {
		return err
	}

	// we need to have at least 1 namespace so we can interogate the features supported
	if len(namespaces) == 0 {
		var nsid uint
		nsid, err = n.createNS(ctx, logicalName, ctrl.TNVMCAP, 512)
		if err != nil {
			return err
		}

		err = n.attachNS(ctx, logicalName, ctrl.CNTLID, nsid)
		if err != nil {
			return err
		}

		namespaces, err = n.listNS(ctx, logicalName)
		if err != nil {
			return err
		}
		if len(namespaces) == 0 {
			err = fmt.Errorf("%s: failed to find namespaces: %w", logicalName, io.ErrUnexpectedEOF)
			return err
		}
	}

	out, err = n.idNS(ctx, logicalName, namespaces[0])
	if err != nil {
		return err
	}
	ns := struct {
		LBAFS []struct {
			DS uint `json:"ds"`
		} `json:"lbafs"`
	}{}
	err = json.Unmarshal(out, &ns)
	if err != nil {
		return err
	}

	ds := uint(0)
	for _, lbafs := range ns.LBAFS {
		// ds is specified in bit-shift count, usually 9:(512b) or 12:(4096)
		ds = max(ds, 1<<lbafs.DS)
	}

	// info gathered and looks ok, lets get dangerous

	// delete all namespaces
	for _, ns := range namespaces {
		err = n.deleteNS(ctx, logicalName, ns)
		if err != nil {
			return err
		}
	}

	// figure out nsze and ncap in terms of blocksize, we want both to be the same
	var nsid uint
	nsid, err = n.createNS(ctx, logicalName, ctrl.TNVMCAP, ds)
	if err != nil {
		return err
	}

	err = n.attachNS(ctx, logicalName, ctrl.CNTLID, nsid)
	if err != nil {
		return err
	}

	n.Executor.SetArgs("reset", logicalName)
	_, err = n.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	n.Executor.SetArgs("ns-rescan", logicalName)
	_, err = n.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// NewFakeNvme returns a mock nvme collector that returns mock data for use in tests.
func NewFakeNvme() *Nvme {
	return &Nvme{
		Executor: NewFakeExecutor("nvme"),
	}
}
