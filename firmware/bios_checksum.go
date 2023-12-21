//nolint:wsl // it's useless
package firmware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
)

const FirmwareDumpUtility model.CollectorUtility = "flashrom"
const UEFIParserUtility model.CollectorUtility = "uefi-firmware-parser"
const ChecksumComposedCollector model.CollectorUtility = "checksum-collector"
const hashPrefix = "SHA256"
const uefiDefaultBMPLogoGUID = "7bb28b99-61bb-11d5-9a5d-0090273fc14d"

var defaultOutputPath = "/tmp/bios_checksum"
var defaultBIOSImgName = "bios_img.bin"
var expectedLogoSuffix = fmt.Sprintf("file-%s/section0/section0.raw", uefiDefaultBMPLogoGUID)

var directoryPermissions fs.FileMode = 0o750
var errNoLogo = errors.New("no logo found")

// ChecksumCollector implements the FirmwareChecksumCollector interface
type ChecksumCollector struct {
	biosOutputPath     string
	biosOutputFilename string
	makeOutputPath     bool
	trace              bool
	biosImgFile        string // this is computed when we write out the BIOS image
	extractPath        string // this is computed when we extract the compressed BIOS image
}

type ChecksumOption func(*ChecksumCollector)

func WithOutputPath(p string) ChecksumOption {
	return func(cc *ChecksumCollector) {
		cc.biosOutputPath = p
	}
}

func WithOutputFile(n string) ChecksumOption {
	return func(cc *ChecksumCollector) {
		cc.biosOutputFilename = n
	}
}

func MakeOutputPath() ChecksumOption {
	return func(cc *ChecksumCollector) {
		cc.makeOutputPath = true
	}
}

func TraceExecution(tf bool) ChecksumOption {
	return func(cc *ChecksumCollector) {
		cc.trace = tf
	}
}

func NewChecksumCollector(opts ...ChecksumOption) *ChecksumCollector {
	cc := &ChecksumCollector{
		biosOutputPath:     defaultOutputPath,
		biosOutputFilename: defaultBIOSImgName,
	}
	for _, o := range opts {
		o(cc)
	}
	return cc
}

// Attributes implements the actions.UtilAttributeGetter interface
//
// Unlike most usages, BIOS checksums rely on several discrete executables. This function returns its own name,
// and it's incumbent on the caller to check if FirmwareDumpUtility or UEFIParserUtility are denied as well.
func (*ChecksumCollector) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	return ChecksumComposedCollector, "", nil
}

// BIOSLogoChecksum implements the FirmwareChecksumCollector interface.
func (cc *ChecksumCollector) BIOSLogoChecksum(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if cc.makeOutputPath {
		err := os.MkdirAll(cc.biosOutputPath, directoryPermissions)
		if err != nil {
			return "", errors.Wrap(err, "creating firmware extraction area")
		}
	}
	if err := cc.dumpBIOS(ctx); err != nil {
		return "", errors.Wrap(err, "reading firmware binary image")
	}
	if err := cc.extractBIOSImage(ctx); err != nil {
		return "", errors.Wrap(err, "extracting firmware binary image")
	}

	logoFileName, err := cc.findExtractedRawLogo(ctx)
	if err != nil {
		return "", errors.Wrap(err, "finding raw logo filename")
	}

	return cc.hashDiscoveredLogo(ctx, logoFileName)
}

func (cc *ChecksumCollector) hashDiscoveredLogo(ctx context.Context, logoFileName string) (string, error) {
	handle, err := os.Open(cc.extractPath + "/" + logoFileName)
	if err != nil {
		return "", errors.Wrap(err, "opening logo file")
	}
	defer handle.Close()

	hasher := sha256.New()
	if _, err = io.Copy(hasher, handle); err != nil {
		return "", errors.Wrap(err, "copying logo data to hasher")
	}

	return fmt.Sprintf("%s: %x", hashPrefix, hasher.Sum(nil)), nil
}

func (cc *ChecksumCollector) dumpBIOS(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cc.biosImgFile = fmt.Sprintf("%s/%s", cc.biosOutputPath, cc.biosOutputFilename)

	frc := utils.NewFlashromCmd(cc.trace)

	return frc.WriteBIOSImage(ctx, cc.biosImgFile)
}

func (cc *ChecksumCollector) extractBIOSImage(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cc.extractPath = fmt.Sprintf("%s/extract", cc.biosOutputPath)

	ufp := utils.NewUefiFirmwareParserCmd(cc.trace)

	return ufp.ExtractLogo(ctx, cc.extractPath, cc.biosImgFile)
}

func (cc *ChecksumCollector) findExtractedRawLogo(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	var filename string

	dirHandle := os.DirFS(cc.extractPath)
	err := fs.WalkDir(dirHandle, ".", func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if cc.trace {
			fmt.Printf("dir-walk: %s\n", path)
		}
		if strings.HasSuffix(path, expectedLogoSuffix) {
			filename = path
			return fs.SkipAll
		}
		// XXX: Check the DirEntry for a bogus size so we don't blow up trying to hash the thing!
		return nil
	})

	if err != nil {
		return "", errors.Wrap(err, "walking the extract directory")
	}

	if filename == "" {
		return "", errNoLogo
	}

	return filename, nil
}
