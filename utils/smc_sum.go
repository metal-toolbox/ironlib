package utils

import (
	"context"
)

const smcSUM = "/usr/sbin/sum"

type SupermicroSUM struct {
	Executor Executor
}

// Return a new Supermicro sum command executor
func NewSupermicroSUM(trace bool) Updater {
	e := NewExecutor(smcSUM)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &SupermicroSUM{Executor: e}
}

// Fake Supermicro SUM executor for tests
func NewFakeSupermicroSUM() *SupermicroSUM {
	return &SupermicroSUM{
		Executor: NewFakeExecutor("sum"),
	}
}

func (s *SupermicroSUM) ApplyUpdate(ctx context.Context, updateFile, componentSlug string) error {
	switch componentSlug {
	case "bios":
		s.Executor.SetArgs([]string{"-c", "UpdateBios", "--preserve_setting", "--file", updateFile})
	case "bmc":
		s.Executor.SetArgs([]string{"-c", "UpdateBmc", "--file", updateFile})
	}

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return newUtilsExecError(s.Executor.GetCmd(), result)
	}

	return nil
}
