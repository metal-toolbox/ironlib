Utilities implement methods to interface with hardware.

The interfaces implemented by utilities are defined in [actions/interface.go](https://github.com/metal-toolbox/ironlib/blob/main/actions/interface.go).

When adding a new Utility, make sure to implement the `UtilAttributeGetter` and
register the utility in [device.go](https://github.com/metal-toolbox/ironlib/blob/main/device.go) - `CheckDependencies()`.