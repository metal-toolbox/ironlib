package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/metal-toolbox/ironlib"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/model"
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

	var drive *model.Drive
	for _, d := range inventory.Drives {
		if d.LogicalName == *logicalName {
			drive = d
			break
		}
	}
	if drive == nil {
		l.Fatal("unable to find disk")
	}

	// Lets see if drive knows how to wipe itself
	// If so we will *only* try the drive-reported wipers
	wipers := drive.Wipers()
	if wipers != nil {
		var wiped bool
		for _, wiper := range wipers {
			if err := wiper.Wipe(ctx, logger); err != nil {
				wiped = true
				break
			}
		}
		if !wiped {
			l.Fatal("failed to wipe drive")
		}
		os.Exit(0)
	}

	// Drive does not know how to wipe itself so lets try and figure out an appropriate Wiper based on the disk type and features supported
	var wiper actions.DriveWiper
	switch drive.Protocol {
	case "nvme":
		wiper = utils.NewNvmeCmd(*verbose)
	case "sata":
		// Lets figure out the drive capabilities in an easier format
		var sanitize bool
		var esee bool
		var trim bool
		for _, cap := range drive.Capabilities {
			switch {
			case cap.Description == "encryption supports enhanced erase":
				esee = cap.Enabled
			case cap.Description == "SANITIZE feature":
				sanitize = cap.Enabled
			case strings.HasPrefix(cap.Description, "Data Set Management TRIM supported"):
				trim = cap.Enabled
			}
		}

		switch {
		case sanitize || esee:
			// Drive supports Sanitize or Enhanced Erase, so we use hdparm
			wiper = utils.NewHdparmCmd(*verbose)
		case trim:
			// Drive supports TRIM, so we use blkdiscard
			wiper = utils.NewBlkdiscardCmd(*verbose)
		default:
			// Drive does not support any preferred wipe method so we fall back to filling it up with zero
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
