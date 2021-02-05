package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/packethost/ironlib"
	"github.com/sirupsen/logrus"
)

// This example invokes ironlib and prints out the device inventory
// a sample output can be seen in the inventory.json file

func main() {
	logger := logrus.New()
	device, err := ironlib.New(logger)
	if err != nil {
		logger.Fatal(err)
	}

	inv, err := device.GetInventory(context.TODO(), false)
	if err != nil {
		logger.Fatal(err)
	}

	j, err := json.MarshalIndent(inv, " ", "  ")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println(string(j))
}
