package model

import (
	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/actions/wipe"
)

type Drive struct {
	common.Drive
	wipersGetter WipersGetter
}

type WipersGetter interface {
	Wipers(*Drive) []wipe.Wiper
}

func NewDrive(d *common.Drive, w WipersGetter) *Drive {
	if d == nil {
		return &Drive{}
	}
	return &Drive{
		Drive:        *d,
		wipersGetter: w,
	}
}

func (d *Drive) Wipers() []wipe.Wiper {
	if d.wipersGetter == nil {
		return nil
	}
	return d.wipersGetter.Wipers(d)
}
