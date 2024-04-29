package utils

import (
	"context"
	"os"

	"github.com/metal-toolbox/ironlib/model"
)

const (
	EnvFlashromUtility = "IRONLIB_UTIL_FLASHROM"
)

type Flashrom struct {
	Executor Executor
}

// Return a new flashrom executor
func NewFlashromCmd(trace bool) *Flashrom {
	utility := "flashrom"

	// lookup env var for util
	if eVar := os.Getenv(EnvFlashromUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Flashrom{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (f *Flashrom) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := f.Executor.CheckExecutable()

	return "flashrom", f.Executor.CmdPath(), er
}

// ExtractBIOSImage writes the BIOS image to the given file system path.
func (f *Flashrom) WriteBIOSImage(ctx context.Context, path string) error {
	// flashrom -p internal --ifd -i bios -r /tmp/bios_region.img
	f.Executor.SetArgs("-p", "internal", "--ifd", "-i", "bios", "-r", path)

	_, err := f.Executor.ExecWithContext(ctx)
	if err != nil {
		return err
	}

	return nil
}
