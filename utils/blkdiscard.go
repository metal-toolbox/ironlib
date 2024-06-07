package utils

import (
	"cmp"
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	EnvBlkdiscardUtility = "IRONLIB_UTIL_BLKDISCARD"
)

type Blkdiscard struct {
	Executor Executor
}

// Return a new blkdiscard executor
func NewBlkdiscardCmd(trace bool) *Blkdiscard {
	// lookup env var for util
	utility := cmp.Or(os.Getenv(EnvBlkdiscardUtility), "blkdiscard")

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Blkdiscard{Executor: e}
}

// Discard runs blkdiscard on the given device (--force is always used)
func (b *Blkdiscard) Discard(ctx context.Context, logicalName string) error {
	b.Executor.SetArgs("--force", logicalName)

	_, err := b.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// WipeDrive implements DriveWipe by calling Discard
func (b *Blkdiscard) WipeDrive(ctx context.Context, logger *logrus.Logger, logicalName string) error {
	logger.WithField("drive", logicalName).WithField("method", "blkdiscard").Info("wiping")
	return b.Discard(ctx, logicalName)
}

// NewFakeBlkdiscard returns a mock implementation of the Blkdiscard interface for use in tests.
func NewFakeBlkdiscard() *Blkdiscard {
	return &Blkdiscard{
		Executor: NewFakeExecutor("blkdiscard"),
	}
}
