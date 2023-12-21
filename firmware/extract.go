package firmware

import (
	"context"

	"github.com/metal-toolbox/ironlib/model"
)

// ChecksumCollector implements the
type ChecksumCollector struct {
	// logger (?)
}

func NewChecksumCollector(trace bool) *ChecksumCollector {
	return &ChecksumCollector{}
}

// Attributes implements the actions.UtilAttributeGetter interface
//
// this is implemented to verify both executables are available
// but can be excluded if considered its not required, since the executables should be available through the image.
func (c *ChecksumCollector) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	return
}

func (f *ChecksumCollector) BIOSLogoChecksum(ctx context.Context) (sha256 [32]byte, err error) {
	// dump bios with flashrom util
	// extract logo with uefi_firmware_parser util
	// return sha256sum of logo

	return [32]byte{}, nil
}
