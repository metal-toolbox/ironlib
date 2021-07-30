package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNoCommandOutput          = errors.New("command returned no output")
	ErrVersionStrExpectedSemver = errors.New("expected version string to follow semver format")
	ErrFakeExecutorInvalidArgs  = errors.New("invalid number of args passed to fake executor")
)

// ExecError is returned when the command exits with an error or a non zero exit status
type ExecError struct {
	Cmd      string
	Stderr   string
	Stdout   string
	ExitCode int
}

// Error implements the error interface
func (u *ExecError) Error() string {
	return fmt.Sprintf("cmd %s exited with error: %s\n\t exitCode: %d\n\t stdout: %s", u.Cmd, u.Stderr, u.ExitCode, u.Stdout)
}

func newExecError(cmd string, r *Result) *ExecError {
	return &ExecError{
		Cmd:      cmd,
		Stderr:   string(r.Stderr),
		Stdout:   string(r.Stdout),
		ExitCode: r.ExitCode,
	}
}
