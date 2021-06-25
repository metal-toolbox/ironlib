package dell

import (
	"context"
	"os"

	"github.com/packethost/ironlib/config"
	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
)

func (d *dell) SetBIOSConfiguration(ctx context.Context, cfg *config.BIOSConfiguration) error {
	return nil
}

func (d *dell) GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error) {
	if envRacadmUtil := os.Getenv("UTIL_RACADM7"); envRacadmUtil == "" {
		err := d.pre() // ensure runtime pre-requisites are installed
		if err != nil {
			return nil, err
		}
	}

	racadm := utils.NewDellRacadm(false)

	return racadm.GetBIOSConfiguration(ctx, model.FormatProductName(d.GetModel()))
}
