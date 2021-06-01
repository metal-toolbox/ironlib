package errs

import (
	"errors"
	"fmt"
)

var (
	ErrNoUpdatesApplicable   = errors.New("no updates applicable")
	ErrDmiDecodeRun          = errors.New("error running dmidecode")
	ErrComponentListExpected = errors.New("expected a list of components to apply updates")
	ErrDeviceInventory       = errors.New("failed to collect device inventory")
)

// DmiDecodeValueError is returned when a dmidecode value could not be retrieved
type DmiDecodeValueError struct {
	Section string
	Field   string
}

// Error implements the error interface
func (d *DmiDecodeValueError) Error() string {
	return fmt.Sprintf("unable to read section: %s, field: %s", d.Section, d.Field)
}

// NewDmidevodeValueError returns a DmiDecodeValueError object
func NewDmidecodeValueError(section, field string) *DmiDecodeValueError {
	return &DmiDecodeValueError{
		Section: section,
		Field:   field,
	}
}
