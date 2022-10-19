package main

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/sirupsen/logrus"
)

// This example invokes ironlib to install the supermicro BMC firmware

func main() {
	logger := logrus.New()

	device, err := ironlib.New(logger)
	if err != nil {
		logger.Fatal(err)
	}

	options := &model.UpdateOptions{
		Vendor:     common.VendorSupermicro,
		Model:      "X11SCH-F",
		Slug:       common.SlugBMC,
		UpdateFile: "/tmp/SMT_CFLAST2500_123_07.bin",
	}

	hardware, err := device.GetInventory(context.TODO(), false)
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println(hardware.BMC.Firmware.Installed)

	err = device.InstallUpdates(context.TODO(), options)
	if err != nil {
		logger.Fatal(err)
	}

}
