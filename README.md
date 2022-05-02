![](https://img.shields.io/badge/Stability-Maintained-green.svg)

Ironlib wraps various opensource and various vendor tools to expose a consistent set of [interface methods](https://github.com/packethost/ironlib/blob/main/model/interface.go) to,

 - Collect inventory
 - Update firmware
 - Set/Get BIOS configuration
 - Set/Get BMC configuration

## Supported vendor hardware

- Dell
- Supermicro
- AsrockRack

## Requirements

Ironlib is expected to be executed from within the [ironlib docker image](https://quay.io/repository/packet/ironlib) running on the target hardware.

## Invoking ironlib

Ironlib identifies the hardware and executes tooling respective to the hardware/component being queried or updated,

The list of tools that ironlib wraps around are [here](https://github.com/packethost/ironlib/tree/main/utils).



#### example

More examples can be found in the [examples](examples/) directory
```
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
  // identify the hardware and get a ironlib device manager object
	device, err := ironlib.New(logger)
	if err != nil {
		logger.Fatal(err)
	}

  // retrieve device hardware inventory
	inv, err := device.GetInventory(context.TODO())
	if err != nil {
		logger.Fatal(err)
	}

	j, err := json.MarshalIndent(inv, " ", "  ")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println(string(j))
}
```

## Build ironlib docker image

`make build-image`

