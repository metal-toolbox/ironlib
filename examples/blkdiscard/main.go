package main

import (
	"context"
	"flag"
	"time"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

var (
	drive   = flag.String("drive", "/dev/someN", "disk to wipe using blkdiscard")
	timeout = flag.String("timeout", (2 * time.Minute).String(), "time to wait for command to complete")
	verbose = flag.Bool("verbose", false, "show command runs and output")
)

// This example invokes ironlib and runs blkdiscard on the disk /dev/sdZZZ
func main() {
	flag.Parse()

	logger := logrus.New()
	logger.Formatter = new(logrus.TextFormatter)
	if *verbose {
		logger.SetLevel(logrus.TraceLevel)
	}

	timeout, err := time.ParseDuration(*timeout)
	if err != nil {
		logger.WithError(err).Fatal("failed to parse timeout duration")
	}

	var blkdiscard actions.DriveWiper = utils.NewBlkdiscardCmd(*verbose)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	drive := &common.Drive{Common: common.Common{LogicalName: *drive}}
	logger.Info("running blkdiscard on ", drive.LogicalName)
	err = blkdiscard.WipeDrive(ctx, logger, drive)
	if err != nil {
		logger.WithError(err).Fatal("exiting")
	}
}
