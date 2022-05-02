![](https://img.shields.io/badge/Stability-Maintained-green.svg)

Ironlib wraps various opensource and various vendor utilities to expose a consistent set of [interface methods](https://github.com/packethost/ironlib/blob/main/model/interface.go) to,

 - Collect inventory
 - Update firmware
 - Set/Get BIOS configuration
 - Set/Get BMC configuration

## Currently supported hardware

- Dell
- Supermicro
- AsrockRack

## Requirements

Ironlib is expected to be executed from within the [ironlib docker image](Dockerfile), on the target host,
the docker image contains all the utilities required to collect inventory, install updates, get BIOS configuration.

## Invoking ironlib

Ironlib identifies the hardware and executes tooling respective to the hardware/component being queried or updated,

The list of tools that ironlib wraps around, in no particular order,

- dell racadm
- dmidecode
- dell dsu
- lshw
- mlxup
- msecli
- nvmecli
- smartctl
- supermicro SUM
- storecli

 [For the full list see here](https://github.com/packethost/ironlib/tree/main/utils)


## Build ironlib docker image

`make build-image`


#### Examples


This example 
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


func main() {
	logger := logrus.New()
	device, err := ironlib.New(logger)
	if err != nil {
		logger.Fatal(err)
	}

	inv, err := device.GetInventory(context.TODO())
	if err != nil {
		logger.Fatal(err)
	}

	j, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println(string(j))
}

```