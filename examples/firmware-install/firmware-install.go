package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib"
	"github.com/metal-toolbox/ironlib/model"
)

// This example invokes ironlib to install the supermicro BMC firmware

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

	options := &model.UpdateOptions{
		Vendor:     common.VendorSupermicro,
		Model:      "X11SCH-F",
		Slug:       common.SlugBMC,
		UpdateFile: "/tmp/SMT_CFLAST2500_123_07.bin",
	}

	hardware, err := device.GetInventory(context.TODO())
	if err != nil {
		logger.Error("getting inventory", "error", err)
		os.Exit(1)
	}

	fmt.Println(hardware.BMC.Firmware.Installed)

	err = device.InstallUpdates(context.TODO(), options)
	if err != nil {
		logger.Error("insatlling updates", "error", err)
		os.Exit(1)
	}
}
