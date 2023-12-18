package main

import (
	"fmt"

	"github.com/metal-toolbox/ironlib"
	"github.com/sirupsen/logrus"
)

// This example invokes ironlib and prints out the BIOS features on supported platforms
// a sample output can be seen in the biosconfig.json file

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)

	device, err := ironlib.New(logger)
	if err != nil {
		logger.Fatal(err)
	}

	// features, err := device.GetBIOSConfiguration(context.TODO())
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	// j, err := json.MarshalIndent(features, " ", "  ")
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	// fmt.Println(string(j))

	fmt.Printf("device.GetVendor(): %v\n", device.GetVendor())

	// biosSettings := config.
	// var newBiosSettings = map[string]string{"SriovGlobalEnable": "Disabled"}

	// err = device.SetBIOSConfiguration(context.TODO(), newBiosSettings)
	// if err != nil {
	// 	logger.Fatal(err)
	// }
}
