package dell

import (
	"context"
	"os"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
)

func (d *dell) SetBIOSConfiguration(ctx context.Context, cfg map[string]string) error {
	err := d.prepareRacadm()
	if err != nil {
		return err
	}

	vendorOptions := map[string]string{
		"deviceModel": d.GetModel(),
		"serviceTag":  d.GetSerial(),
	}

	racadm := utils.NewDellRacadm(true, utils.WithReboot())
	return racadm.SetBIOSConfiguration(ctx, vendorOptions, cfg)
}

func (d *dell) SetBIOSConfigurationFromFile(ctx context.Context, cfg string) error {
	err := d.prepareRacadm()
	if err != nil {
		return err
	}

	racadm := utils.NewDellRacadm(true, utils.WithReboot())
	return racadm.SetBIOSConfigurationFromFile(ctx, model.FormatProductName(d.GetModel()), cfg)
}

func (d *dell) GetBIOSConfiguration(ctx context.Context) (map[string]string, error) {
	err := d.prepareRacadm()
	if err != nil {
		return nil, err
	}

	racadm := utils.NewDellRacadm(false, utils.WithoutReboot())
	return racadm.GetBIOSConfiguration(ctx, model.FormatProductName(d.GetModel()))
}

func (d *dell) prepareRacadm() error {
	if envRacadmUtil := os.Getenv("IRONLIB_UTIL_RACADM7"); envRacadmUtil == "" {
		err := d.pre() // ensure runtime pre-requisites are installed
		if err != nil {
			return err
		}
	}

	// Make sure service that loads ipmi modules is running before attempting to collect bios config
	err := d.startSrvHelper()
	if err != nil {
		return err
	}

	return nil
}
