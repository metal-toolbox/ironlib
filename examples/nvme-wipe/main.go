package main

import (
	"context"
	"flag"
	"time"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

var (
	logicalName = flag.String("drive", "/dev/nvmeX", "nvme disk to wipe")
	timeout     = flag.String("timeout", (1 * time.Minute).String(), "time to wait for command to complete")
	verbose     = flag.Bool("verbose", false, "show command runs and output")
)

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

	nvme := utils.NewNvmeCmd(*verbose)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("resetting namespaces")
	err = nvme.ResetNS(ctx, *logicalName)
	if err != nil {
		logger.WithError(err).Fatal("exiting")
	}

	drives, err := nvme.Drives(ctx)
	if err != nil {
		logger.WithError(err).Fatal("exiting")
	}

	var drive *common.Drive
	for _, d := range drives {
		if d.LogicalName == *logicalName+"n1" {
			drive = d
		}
	}

	if drive == nil {
		logger.Fatal("unable to find drive after reset")
	}

	logger.Info("wiping")
	err = nvme.WipeDrive(ctx, logger, drive)
	if err != nil {
		logger.WithError(err).Fatal("exiting")
	}
}
