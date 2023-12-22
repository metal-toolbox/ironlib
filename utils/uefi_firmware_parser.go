// nolint: wsl,gocritic
package utils

import (
	"context"
	"io/fs"
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

var directoryPermissions fs.FileMode = 0o750

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
func (u *UefiFirmwareParser) Attributes() (model.CollectorUtility, string, error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	err := u.Executor.CheckExecutable()

	return "uefi-firmware-parser", u.Executor.CmdPath(), err
}

// ExtractLogo extracts the Logo BMP image. It creates the output directory if required.
func (u *UefiFirmwareParser) ExtractLogo(ctx context.Context, outputPath, biosImg string) error {
	if err := os.MkdirAll(outputPath, directoryPermissions); err != nil {
		return err
	}

	u.Executor.SetArgs([]string{
		"-b",
		biosImg,
		"-o",
		outputPath,
		"-e",
	})

	_, err := u.Executor.ExecWithContext(ctx)
	return err
}
