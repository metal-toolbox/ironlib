package model

import (
	"fmt"

	"github.com/bmc-toolbox/common"
)

type Device struct {
	common.Device
	Drives []*Drive `json:"drives,omitempty"`
}

type WithField func(*Device)

func NewDevice(device *common.Device, setters ...WithField) *Device {
	d := &Device{
		Device: *device,
	}
	for _, setter := range setters {
		fmt.Println("calling setter:", setter)
		setter(d)
	}
	return d
}

func WithDrives(drives []*Drive) WithField {
	return func(d *Device) {
		d.SetDrives(drives)
	}
}

func (d *Device) SetDrives(drives []*Drive) {
	d.Drives = drives
	d.Device.Drives = make([]*common.Drive, len(drives))
	for i := range drives {
		d.Device.Drives[i] = &drives[i].Drive
	}
}

func (d *Device) AddDrive(drive *Drive) {
	d.Drives = append(d.Drives, drive)
	d.Device.Drives = append(d.Device.Drives, &drive.Drive)
}
