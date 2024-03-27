package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/metal-toolbox/ironlib"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/sirupsen/logrus"
)

// This example takes one or more arguments, they are:
//
// ./biosconfig read # print local system's bios config out to STDOUT
// ./biosconfig write <filename.json> # reads filename.json, converts that to vendor specific format and writes that to the BIOS
// ./biosconfig raw <filename.json|xml> # reads filename.json|xml and writes that *directly* to the BIOS

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)

	args := os.Args[1:]

	if len(args) == 0 {
		logger.Fatal("usage: biosconfig <read|write|raw> [<filename>]")
	}
	switch strings.ToLower(args[0]) {
	case "read":
		device, err := ironlib.New(logger)
		if err != nil {
			logger.Fatal(err)
		}

		readBiosConfig(logger, device)
	case "write":
		device, err := ironlib.New(logger)
		if err != nil {
			logger.Fatal(err)
		}

		writeBiosConfig(logger, device, args[1])
	case "rawwrite", "raw":
		device, err := ironlib.New(logger)
		if err != nil {
			logger.Fatal(err)
		}

		writeRawBiosConfig(logger, device, args[1])
	default:
		logger.Fatal("Unknown mode " + args[0])
	}
}

func readBiosConfig(logger *logrus.Logger, device actions.DeviceManager) {
	features, err := device.GetBIOSConfiguration(context.TODO())
	if err != nil {
		logger.Fatal(err)
	}

	formatJson(logger, features)
}

func writeBiosConfig(logger *logrus.Logger, device actions.DeviceManager, filename string) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		logger.Fatal(err)
	}

	defer jsonFile.Close()

	jsonData, _ := io.ReadAll(jsonFile)
	newBios := make(map[string]string)
	err = json.Unmarshal(jsonData, &newBios)
	if err != nil {
		fmt.Println(err)
	}

	formatJson(logger, newBios)

	err = device.SetBIOSConfiguration(context.TODO(), newBios)
	if err != nil {
		fmt.Println(err)
	}
}

func writeRawBiosConfig(logger *logrus.Logger, device actions.DeviceManager, filename string) {
	cfg, err := os.Open(filename)
	if err != nil {
		logger.Fatal(err)
	}

	defer cfg.Close()

	cfgData, _ := io.ReadAll(cfg)

	fmt.Println(cfgData)

	err = device.SetBIOSConfigurationFromFile(context.TODO(), string(cfgData))
	if err != nil {
		fmt.Println(err)
	}
}

func formatJson(logger *logrus.Logger, settings map[string]string) {
	j, err := json.MarshalIndent(settings, " ", "  ")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println(string(j))
}
