package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

var (
	defaultTimeout = time.Minute

	logicalName = flag.String("drive", "/dev/someN", "disk to wipe by filling with zeros")
	timeout     = flag.String("timeout", defaultTimeout.String(), "time to wait for command to complete")
	verbose     = flag.Bool("verbose", false, "show command runs and output")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	logger.Formatter = new(logrus.TextFormatter)
	if *verbose {
		logger.SetLevel(logrus.TraceLevel)
	}
	l := logger.WithField("drive", *logicalName)

	timeout, err := time.ParseDuration(*timeout)
	if err != nil {
		l.WithError(err).Fatal("failed to parse timeout duration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	collector, err := ironlib.New(logger)
	if err != nil {
		l.WithError(err).Fatal("exiting")
	}

	inventory, err := collector.GetInventory(ctx, actions.WithDynamicCollection())
	if err != nil {
		l.WithError(err).Fatal("exiting")
	}

	var drive *common.Drive
	for _, d := range inventory.Drives {
		if d.LogicalName == *logicalName {
			drive = d
			break
		}
	}
	if drive == nil {
		l.Fatal("unable to find disk")
	}

	// Pick the most appropriate wipe based on the disk type and features supported
	var wiper actions.DriveWiper
	switch drive.Protocol {
	case "nvme":
		wiper = utils.NewNvmeCmd(*verbose)
	case "sata":
		// Lets see if drive supports TRIM, if so we'll use blkdiscard
		for _, cap := range drive.Capabilities {
			if strings.HasPrefix(cap.Description, "Data Set Management TRIM supported") {
				if cap.Enabled {
					wiper = utils.NewBlkdiscardCmd(*verbose)
				}
				break
			}
		}

		// drive does not support TRIM so we fall back to filling it up with zero
		if wiper == nil {
			wiper = utils.NewFillZeroCmd(*verbose)

			// If the user supplied a non-default timeout then we'll honor it, otherwise we just go with a huge timeout.
			// If this were *real* code and not an example some work could be done to guesstimate a timeout based on disk size.
			if timeout == defaultTimeout {
				l.WithField("timeout", timeout.String()).Info("increasing timeout")
				timeout = 24 * time.Hour
				ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), timeout)
				defer cancel()
			}
		}
	}

	if wiper == nil {
		l.Fatal("failed find appropriate wiper drive")
	}

	err = wiper.WipeDrive(ctx, logger, drive)
	if err != nil {
		l.Fatal("failed to wipe drive")
	}
}
