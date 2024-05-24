package main

import (
	"context"
	"time"

	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

var (
	device  = "/dev/sdZZZ"
	timeout = 2 * time.Minute
)

// This example invokes ironlib and runs blkdiscard on the disk /dev/sdZZZ
func main() {
	logger := logrus.New()
	logger.Formatter = new(logrus.JSONFormatter)
	logger.SetLevel(logrus.TraceLevel)

	blkdiscard := utils.NewBlkdiscardCmd()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("running blkdiscard on", device)
	err := blkdiscard.Discard(ctx, device)
	if err != nil {
		logger.WithError(err).Fatal("exiting")
	}
}
