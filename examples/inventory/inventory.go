package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/metal-toolbox/ironlib"
)

// This example invokes ironlib and prints out the device inventory
// a sample output can be seen in the inventory.json file

func main() {
	trace := &slog.LevelVar{}
	trace.Set(-5)
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: trace})
	logger := slog.New(h)

	device, err := ironlib.New(logger)
	if err != nil {
		logger.Error("creating ironlib manager", "error", err)
		os.Exit(1)
	}

	inv, err := device.GetInventory(context.TODO())
	if err != nil {
		logger.Error("getting inventory", "error", err)
		os.Exit(1)
	}

	j, err := json.MarshalIndent(inv, " ", "  ")
	if err != nil {
		logger.Error("formatting json", "error", err)
		os.Exit(1)
	}

	fmt.Println(string(j))
}
