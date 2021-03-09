package errs

import "errors"

var (
	ErrNoUpdatesApplicable = errors.New("no updates applicable")
)
