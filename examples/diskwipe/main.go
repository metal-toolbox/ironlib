package main

import (
	"context"
	"time"

	"github.com/metal-toolbox/ironlib/actions"
	"github.com/sirupsen/logrus"
)

// This example invokes ironlib and wipes the disk /dev/sdZZZ with a timeout of 1 day

func main() {
	logger := logrus.New()
	logger.Formatter = new(logrus.JSONFormatter)
	logger.SetLevel(logrus.TraceLevel)
	sca := actions.NewStorageControllerAction(logger)
	ctx, cancel := context.WithTimeout(context.Background(), 86400*time.Second)
	defer cancel()
	err := sca.WipeDisk(ctx, "/dev/sdZZZ")
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("Wiped successfully!")
}
