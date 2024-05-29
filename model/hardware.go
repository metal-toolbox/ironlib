package model

import "github.com/bmc-toolbox/common"

// Hardware is a base struct that various providers inherit
type Hardware struct {
	PendingReboot    bool // set when the device requires a reboot after running an upgrade
	UpdatesInstalled bool // set when updates were installed on the device
	UpdatesAvailable int  // -1 == no update lookup as yet,  0 == no updates available, 1 == updates available
	Device           *common.Device
	OEMComponents    []*Component // OEMComponents hold OEM specific components that may not show up in dmidecode/lshw and other collectors.
}

// NewHardware returns the base Hardware struct that various providers inherit
func NewHardware(d *common.Device) *Hardware {
	return &Hardware{Device: d, UpdatesAvailable: -1}
}
