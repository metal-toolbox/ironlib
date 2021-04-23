package utils

import (
	"errors"
	"fmt"
)

var (
	ErrUtilsNoCommandOutput = errors.New("command returned no output")
)

// UtilsExecError is returned when the command exits with an error or a non zero exit status
type UtilsExecError struct {
	Cmd      string
	Stderr   string
	Stdout   string
	ExitCode int
}

func (u *UtilsExecError) Error() string {
	return fmt.Sprintf("cmd %s exited with error: %s\n\t exitCode: %d\n\t stdout: %s", u.Cmd, u.Stderr, u.ExitCode, u.Stdout)
}

func newUtilsExecError(cmd string, r *Result) *UtilsExecError {
	return &UtilsExecError{
		Cmd:      cmd,
		Stderr:   string(r.Stderr),
		Stdout:   string(r.Stdout),
		ExitCode: r.ExitCode,
	}
}
