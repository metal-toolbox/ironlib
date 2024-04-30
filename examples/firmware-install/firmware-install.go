package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bmc-toolbox/common"
	"github.com/bombsimon/logrusr/v4"
	"github.com/metal-toolbox/ironlib"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/sirupsen/logrus"
)

// This example invokes ironlib to install the supermicro BMC firmware

func main() {
	l := logrus.New()
	l.Formatter = &logrus.JSONFormatter{}
	l.Level = logrus.TraceLevel
	logger := logrusr.New(l)

	device, err := ironlib.New(logger)
	if err != nil {
		logger.Error(err, "creating ironlib manager")
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
		logger.Error(err, "getting inventory")
		os.Exit(1)
	}

	fmt.Println(hardware.BMC.Firmware.Installed)

	err = device.InstallUpdates(context.TODO(), options)
	if err != nil {
		logger.Error(err, "insatlling updates")
		os.Exit(1)
	}
}
