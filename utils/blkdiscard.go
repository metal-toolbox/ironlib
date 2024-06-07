package utils

import (
	"cmp"
	"context"
	"os"

	"github.com/bmc-toolbox/common"
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
func (b *Blkdiscard) Discard(ctx context.Context, drive *common.Drive) error {
	b.Executor.SetArgs("--force", drive.LogicalName)

	verify, err := ApplyWatermarks(drive)
	if err != nil {
		return err
	}

	_, err = b.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	return verify()
}

// WipeDrive implements DriveWipe by calling Discard
func (b *Blkdiscard) WipeDrive(ctx context.Context, logger *logrus.Logger, drive *common.Drive) error {
	logger.WithField("drive", drive.LogicalName).WithField("method", "blkdiscard").Info("wiping")
	return b.Discard(ctx, drive)
}

// NewFakeBlkdiscard returns a mock implementation of the Blkdiscard interface for use in tests.
func NewFakeBlkdiscard() *Blkdiscard {
	return &Blkdiscard{
		Executor: NewFakeExecutor("blkdiscard"),
	}
}
