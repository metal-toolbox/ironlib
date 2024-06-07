package main

import (
	"context"
	"flag"
	"time"

	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

var (
	device  = flag.String("device", "/dev/sdZZZ", "disk to wipe using blkdiscard")
	timeout = flag.String("timeout", (2 * time.Minute).String(), "time to wait for command to complete")
)

// This example invokes ironlib and runs blkdiscard on the disk /dev/sdZZZ
func main() {
	flag.Parse()

	logger := logrus.New()
	logger.Formatter = new(logrus.TextFormatter)
	logger.SetLevel(logrus.TraceLevel)

	timeout, err := time.ParseDuration(*timeout)
	if err != nil {
		logger.WithError(err).Fatal("failed to parse timeout duration")
	}

	var blkdiscard actions.DiskWiper = utils.NewBlkdiscardCmd()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("running blkdiscard on ", *device)
	err = blkdiscard.WipeDisk(ctx, logger, *device)
	if err != nil {
		logger.WithError(err).Fatal("exiting")
	}
}
