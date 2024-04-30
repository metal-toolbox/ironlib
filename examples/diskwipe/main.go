package main

import (
	"context"
	"os"
	"time"

	"github.com/bombsimon/logrusr/v4"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/sirupsen/logrus"
)

// This example invokes ironlib and wipes the disk /dev/sdZZZ with a timeout of 1 day

func main() {
	l := logrus.New()
	l.Formatter = &logrus.JSONFormatter{}
	l.Level = logrus.TraceLevel
	logger := logrusr.New(l)

	sca := actions.NewStorageControllerAction(logger)
	ctx, cancel := context.WithTimeout(context.Background(), 86400*time.Second)
	defer cancel()

	err := sca.WipeDisk(ctx, logger, "/dev/sdZZZ")
	if err != nil {
		logger.Error(err, "wiping disk")
		os.Exit(0)
	}
	logger.Info("Wiped successfully!")
}
