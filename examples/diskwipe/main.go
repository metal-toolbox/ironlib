package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/metal-toolbox/ironlib/actions"
)

// This example invokes ironlib and wipes the disk /dev/sdZZZ with a timeout of 1 day

func main() {
	trace := &slog.LevelVar{}
	trace.Set(-5)
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: trace})
	logger := slog.New(h)

	sca := actions.NewStorageControllerAction(logger)
	ctx, cancel := context.WithTimeout(context.Background(), 86400*time.Second)
	defer cancel()

	err := sca.WipeDisk(ctx, logger, "/dev/sdZZZ")
	if err != nil {
		logger.Error("wiping disk", "error", err)
		os.Exit(0)
	}
	logger.Info("Wiped successfully!")
}
