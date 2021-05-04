package dell

import (
	"context"

	"github.com/packethost/ironlib/config"
	"github.com/packethost/ironlib/utils"
)

func (d *Dell) SetBIOSConfiguration(ctx context.Context, config *config.BIOSConfiguration) error {
	return nil
}

func (d *Dell) GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error) {
	racadm := utils.NewDellRacadm(false)
	return racadm.GetBIOSConfiguration(ctx)
}
