package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNoCommandOutput          = errors.New("command returned no output")
	ErrVersionStrExpectedSemver = errors.New("expected version string to follow semver format")
)

// UtilsExecError is returned when the command exits with an error or a non zero exit status
type UtilsExecError struct {
	Cmd      string
	Stderr   string
	Stdout   string
	ExitCode int
}

// Error implements the error interface
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

// DmiDecodeValueError is returned when a dmidecode value could not be retrieved
type DmiDecodeValueError struct {
	Section string
	Field   string
}

// Error implements the error interface
func (d *DmiDecodeValueError) Error() string {
	return fmt.Sprintf("unable to read section: %s, field: %s", d.Section, d.Field)
}

// newDmidevodeValueError returns a DmiDecodeValueError object
func newDmidecodeValueError(section, field string) *DmiDecodeValueError {
	return &DmiDecodeValueError{
		Section: section,
		Field:   field,
	}
}
