package dell

import (
	"context"
	"os"

	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
)

func (d *dell) SetBIOSConfiguration(ctx context.Context, cfg map[string]string) error {
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

	racadm := utils.NewDellRacadm(false)

	return racadm.SetBIOSConfiguration(ctx, model.FormatProductName(d.GetModel()), cfg)
}

func (d *dell) SetBIOSConfigurationFromFile(ctx context.Context, cfg string) error {
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

	racadm := utils.NewDellRacadm(false)

	return racadm.SetBIOSConfigurationFromFile(ctx, model.FormatProductName(d.GetModel()), cfg)
}

func (d *dell) GetBIOSConfiguration(ctx context.Context) (map[string]string, error) {
	if envRacadmUtil := os.Getenv("IRONLIB_UTIL_RACADM7"); envRacadmUtil == "" {
		err := d.pre() // ensure runtime pre-requisites are installed
		if err != nil {
			return nil, err
		}
	}

	// Make sure service that loads ipmi modules is running before attempting to collect bios config
	err := d.startSrvHelper()
	if err != nil {
		return nil, err
	}

	racadm := utils.NewDellRacadm(false)

	return racadm.GetBIOSConfiguration(ctx, model.FormatProductName(d.GetModel()))
}

// func processSlice(toSet map[string]string, configFormat string) (err error, output string) {
// 	dc, err := config.NewVendorConfigManager(configFormat, common.VendorDell)
// 	if err != nil {
// 		return
// 	}

// 	// TODO(jwb) This limits us to only BIOS related values, we'll want to fix that.
// 	for k, v := range toSet {
// 		if strings.HasPrefix(k, "raw:") {
// 			dc.Raw(strings.TrimPrefix(k, "raw:"), v, []string{"BIOS.Setup.1-1"})
// 		} else {
// 			// call func by name from k, passing v
// 		}
// 	}

// 	output, err = dc.Marshal()

// 	if err != nil {
// 		return
// 	}

// 	return
// }
