package errs

import "errors"

var (
	ErrNoUpdatesApplicable   = errors.New("no updates applicable")
	ErrDmiDecodeRun          = errors.New("error running dmidecode")
	ErrComponentListExpected = errors.New("expected a list of components to apply updates")
	ErrDeviceInventory       = errors.New("failed to collect device inventory")
)
