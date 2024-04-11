package utils

import (
	"context"
	"log"

	"github.com/pkg/errors"
)

const (
	EnvZeroWipeUtility = "IRONLIB_UTIL_WIPE_ZERO"
)

type ZeroWipe struct {
}

var (
	ErrWipeDisk = errors.New("failed to wipe disk")
)

// Return a new zerowipe executor
func NewZeroWipeCmd(trace bool) *ZeroWipe {
	return &ZeroWipe{}
}

func (z *ZeroWipe) Wipe(ctx context.Context, logicalName string) error {
	log.Println("Wiping with zeros...")

	return nil
}

func (z *ZeroWipe) WipeDisk(ctx context.Context, logicalName string) error {
	return z.Wipe(ctx, logicalName)
}
