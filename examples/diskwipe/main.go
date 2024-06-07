package main

import (
	"context"
	"flag"
	"time"

	"github.com/metal-toolbox/ironlib/actions"
	"github.com/sirupsen/logrus"
)

var (
	logicalName = flag.String("drive", "/dev/someN", "disk to wipe by filling with zeros")
	timeout     = flag.String("timeout", (24 * time.Hour).String(), "time to wait for command to complete")
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

	sca := actions.NewStorageControllerAction(logger)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = sca.WipeDrive(ctx, logger, *logicalName)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("Wiped successfully!")
}
