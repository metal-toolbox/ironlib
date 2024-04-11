package main

import (
	"context"
	"fmt"

	"github.com/metal-toolbox/ironlib/actions"
	"github.com/sirupsen/logrus"
)

// This example invokes ironlib and prints out the device inventory
// a sample output can be seen in the inventory.json file

func main() {
	logger := logrus.New()
	logger.Formatter = new(logrus.JSONFormatter)
	logger.SetLevel(logrus.TraceLevel)
	sca := actions.NewStorageControllerAction(logger)
	ctx := context.Background()
	err := sca.WipeDisk(ctx, "/dev/sdZZZ")
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("Wiped")
}
