package model

import "github.com/bmc-toolbox/common"

// Hardware is a base struct that various providers inherit
type Hardware struct {
	PendingReboot    bool // set when the device requires a reboot after running an upgrade
	UpdatesInstalled bool // set when updates were installed on the device
	UpdatesAvailable int  // -1 == no update lookup as yet,  0 == no updates available, 1 == updates available
	Device           *Device
}

// NewHardware returns the base Hardware struct that various providers inherit
func NewHardware(d *Device) *Hardware {
	return &Hardware{Device: d, UpdatesAvailable: -1}
}

// NewDevice returns an ironlib Device object
func NewDevice() *Device {
	return &Device{common.NewDevice(), nil}
}

// Device is an ironlib device object which extends the common.Device
type Device struct {
	common.Device

	OemComponents *OemComponents `json:"oem_components,omitempty"`
}
