package utils

import (
	"context"
	"os"
)

const (
	EnvBlkdiscardUtility = "IRONLIB_UTIL_BLKDISCARD"
)

type Blkdiscard struct {
	Executor Executor
}

// Return a new blkdiscard executor
func NewBlkdiscardCmd() *Blkdiscard {
	utility := "blkdiscard"

	// lookup env var for util
	if eVar := os.Getenv(EnvBlkdiscardUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	return &Blkdiscard{Executor: e}
}

// Discard runs blkdiscard on the given device (--force is always used)
func (b *Blkdiscard) Discard(ctx context.Context, device string) error {
	b.Executor.SetArgs("--force", device)

	_, err := b.Executor.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// NewFakeBlkdiscard returns a mock implementation of the Blkdiscard interface for use in tests.
func NewFakeBlkdiscard() *Blkdiscard {
	return &Blkdiscard{
		Executor: NewFakeExecutor("blkdiscard"),
	}
}
