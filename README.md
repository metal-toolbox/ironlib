[![Go Reference](https://pkg.go.dev/badge/github.com/metal-toolbox/ironlib.svg)](https://pkg.go.dev/github.com/metal-toolbox/ironlib)
![](https://img.shields.io/badge/Stability-Maintained-green.svg)

Ironlib wraps various opensource and various vendor utilities to expose a consistent set of [interface methods](https://github.com/metal-toolbox/ironlib/blob/main/actions/interface.go) to,

 - Collect inventory
 - Update firmware
 - Set/Get BIOS configuration
 - Set/Get BMC configuration

For the available methods,

- The supported actions interface and method docs can be found [here](https://pkg.go.dev/github.com/metal-toolbox/ironlib/actions)
- The supported utilities and its methods can be found [here](https://pkg.go.dev/github.com/metal-toolbox/ironlib/utils)

## Currently supported hardware

- Dell
- Supermicro
- AsrockRack

## Requirements

Ironlib expects various vendor utilities to be made available.

TODO: define a list of utility path env vars a user can set.

## Invoking ironlib

Ironlib identifies the hardware and executes tooling respective to the hardware/component being queried or updated,

The list of tools that ironlib wraps around, in no particular order,

- blkdiscard
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

 [For the full list see here](https://github.com/metal-toolbox/ironlib/tree/main/utils)


#### Examples

More examples can be found in the [examples](examples/) directory
```
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/metal-toolbox/ironlib"
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

#### Executable path environment variables.

By default ironlib will lookup the executable path, if required the path can be overriden by
exporting one or more of these environment variables

```
IRONLIB_UTIL_ASRR_BIOSCONTROL
IRONLIB_UTIL_RACADM7
IRONLIB_UTIL_DNF
IRONLIB_UTIL_DSU
IRONLIB_UTIL_BLKDISCARD
IRONLIB_UTIL_HDPARM
IRONLIB_UTIL_LSBLK
IRONLIB_UTIL_LSHW
IRONLIB_UTIL_MLXUP
IRONLIB_UTIL_MSECLI
IRONLIB_UTIL_MVCLI
IRONLIB_UTIL_NVME
IRONLIB_UTIL_SMARTCTL
IRONLIB_UTIL_SMC_IPMICFG
IRONLIB_UTIL_SUM
IRONLIB_UTIL_STORECLI
```

Check out this [snippet](examples/dependencies/main.go) to determine if all required dependencies are available to ironlib.

### Build image without the non-distributable files.

```sh
docker build -f Dockerfile -t ironlib:devel .
```

### Build image with non-distributable files

ironlib will attempt to execute proprietary vendor executables based on the hardware its run on,
to build docker image with executables like `mvcli`, `mlxup`, SMCI `sum` etc follow the steps below.

- Ensure the executable files are made available in an s3 bucket, update the `S3_BUCKET_ALIAS` and `S3_PATH` in the snippet below.
- The list of files expected can be found within the [install-non-distributable.sh](scripts/install-non-distributable.sh) script.
- Build image with the `ACCESS_KEY`, `SECRET_KEY` values defined.

```sh
docker build -f Dockerfile \
  --build-arg INSTALL_NON_DISTRIBUTABLE=true
  --build-arg S3_BUCKET_ALIAS="tools"
  --build-arg S3_PATH="tools/bucket-name/path/non-dist"
  --build-arg ACCESS_KEY=<>
  --build-arg SECRET_KEY=<>
  -t ironlib:devel-non-dist .
```
