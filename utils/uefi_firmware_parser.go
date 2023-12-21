package utils

import (
	"context"
	"os"

	"github.com/metal-toolbox/ironlib/model"
)

// TODO: for a future point in time
// The fiano library is in Go and could replace the code if its capable of extracting the Logo bmp image
// https://github.com/linuxboot/fiano

const (
	EnvUefiFirmwareParserUtility = "IRONLIB_UTIL_UTIL_UEFI_FIRMWARE_PARSER"
)

type UefiFirmwareParser struct {
	Executor Executor
}

// Return a new UefiFirmwareParser executor
func NewUefiFirmwareParserCmd(trace bool) *UefiFirmwareParser {
	utility := "uefi-firmware-parser"

	// lookup env var for util
	if eVar := os.Getenv(EnvUefiFirmwareParserUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &UefiFirmwareParser{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (u *UefiFirmwareParser) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := u.Executor.CheckExecutable()

	return "uefi-firmware-parser", u.Executor.CmdPath(), er
}

// ExtractLogoBMP extracts the Logo BMP image.
func (u *UefiFirmwareParser) ExtractLogoBMP(ctx context.Context, path string) error {

	// mkdir dump && uefi-firmware-parser -b bios_region.img  -o dump -e
	//
	// # list out GUIDs in firmware
	// uefi-firmware-parser -b bios_region.img  > parsed_regions
	//
	// # locate the logo
	// grep -i bmp parsed_regions
	//   File 349: 7bb28b99-61bb-11d5-9a5d-0090273fc14d (EFI_DEFAULT_BMP_LOGO_GUID) type 0x02, attr 0x00, state 0x07, size 0x13a2b (80427 bytes), (freeform)
	//
	// # find the section raw dump identified by the GUID
	// find ./ | grep 7bb28b99-61bb-11d5-9a5d-0090273fc14d | grep raw
	// ./dump/volume-23658496/file-7bb28b99-61bb-11d5-9a5d-0090273fc14d/section0/section0.raw
	//
	// mv ./dump/volume-23658496/file-7bb28b99-61bb-11d5-9a5d-0090273fc14d/section0/section0.raw /tmp/logo.bmp

	return nil
}
