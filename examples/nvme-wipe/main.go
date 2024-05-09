package main

import (
	"context"
	"flag"
	"time"

	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

var (
	device  = flag.String("device", "/dev/nvmeX", "nvme disk to wipe")
	verbose = flag.Bool("verbose", false, "show command runs and output")
	dur     = flag.String("timeout", (1 * time.Minute).String(), "time to wait for command to complete")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	logger.Formatter = new(logrus.TextFormatter)
	logger.SetLevel(logrus.TraceLevel)

	timeout, err := time.ParseDuration(*dur)
	if err != nil {
		logger.WithError(err).Fatal("failed to parse timeout duration")
	}

	nvme := utils.NewNvmeCmd(*verbose)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("resetting namespaces")
	err = nvme.ResetNS(ctx, *device)
	if err != nil {
		logger.WithError(err).Fatal("exiting")
	}

	logger.Info("wiping")
	err = nvme.WipeDisk(ctx, logger, *device+"n1")
	if err != nil {
		logger.WithError(err).Fatal("exiting")
	}
}
