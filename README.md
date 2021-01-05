### ironlib

A vendor neutral library to interact with server hardware.

Ironlib has similar goals as [bmclib](https://github.com/bmc-toolbox/bmclib/) which works Out Of Band,
while Ironlib works In Band - its intended to run from a docker image like [image-firmware-update](https://github.com/packethost/image-firmware-update)
which provides the various utilities required to collect hardware inventory, upgrade firmware, configure the BIOS or the BMC.

Hardware Support

Hardware      | Firmware update | Inventory   | BIOS configuration | BMC configuration |
:-----------  | :-------------: | :---------: | :----------------: | :---------------: |
Dell          | :heavy_check_mark: | :heavy_check_mark: | | |
Supermicro    |                    | :heavy_check_mark: | | |
