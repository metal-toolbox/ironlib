package dell

import (
	"context"
	"os"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
)

func (d *dell) SetBIOSConfiguration(context.Context, map[string]string) error {
	return nil
}

func (d *dell) GetBIOSConfiguration(ctx context.Context) (map[string]string, error) {
	if envRacadmUtil := os.Getenv("IRONLIB_UTIL_RACADM7"); envRacadmUtil == "" {
		err := d.pre(ctx) // ensure runtime pre-requisites are installed
		if err != nil {
			return nil, err
		}
	}

	// Make sure service that loads ipmi modules is running before attempting to collect bios config
	err := d.startSrvHelper(ctx)
	if err != nil {
		return nil, err
	}

	racadm := utils.NewDellRacadm(false)

	return racadm.GetBIOSConfiguration(ctx, model.FormatProductName(d.GetModel()))
}
