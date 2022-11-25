package ironlib

import (
	"fmt"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/actions"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/providers/asrockrack"
	"github.com/metal-toolbox/ironlib/providers/dell"
	"github.com/metal-toolbox/ironlib/providers/generic"
	"github.com/metal-toolbox/ironlib/providers/supermicro"
	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// New returns a device Manager interface based on the hardware deviceVendor, model attributes
// by default returns a Generic device instance that only returns the device inventory
func New(logger *logrus.Logger) (m model.DeviceManager, err error) {
	dmidecode, err := utils.NewDmidecode()
	if err != nil {
		return nil, errors.Wrap(errs.ErrDmiDecodeRun, err.Error())
	}

	deviceVendor, err := dmidecode.Manufacturer()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("system manufacturer", "", 0), err.Error())
	}

	if deviceVendor == "" || deviceVendor == common.SystemManufacturerUndefined {
		deviceVendor, err = dmidecode.BaseBoardManufacturer()
		if err != nil {
			return nil, errors.Wrap(errs.NewDmidecodeValueError("baseboard manufacturer", "", 0), err.Error())
		}
	}

	deviceVendor = common.FormatVendorName(deviceVendor)

	switch deviceVendor {
	case common.VendorDell:
		return dell.New(dmidecode, logger)
	case common.VendorSupermicro:
		return supermicro.New(dmidecode, logger)
	case common.VendorPacket, common.VendorAsrockrack:
		return asrockrack.New(dmidecode, logger)
	default:
		return generic.New(dmidecode, logger)
	}
}

// CheckDependencies checks and lists available utilities.
func CheckDependencies() {
	utilities := []actions.UtilAttributeGetter{
		utils.NewAsrrBioscontrol(false),
		utils.NewDellRacadm(false),
		utils.NewDnf(false),
		utils.NewDsu(false),
		utils.NewHdparmCmd(false),
		utils.NewLshwCmd(false),
		utils.NewLsblkCmd(false),
		utils.NewMlxupCmd(false),
		utils.NewMsecli(false),
		utils.NewMvcliCmd(false),
		utils.NewNvmeCmd(false),
		utils.NewSmartctlCmd(false),
		utils.NewIpmicfgCmd(false),
		utils.NewSupermicroSUM(false),
		utils.NewStoreCLICmd(false),
	}

	red := "\033[31m"
	green := "\033[32m"
	reset := "\033[0m"

	dmi, err := utils.NewDmidecode()

	dmiName, dmiPath, dmiErr := dmi.Attributes()
	if err != nil || dmiErr != nil {
		fmt.Printf("util: %s, path: %s %s[err]%s - %s\n", dmiName, dmiPath, red, reset, dmiErr.Error())
	} else {
		fmt.Printf("util: %s, path: %s %s[ok]%s\n", dmiName, dmiPath, green, reset)
	}

	for _, utility := range utilities {
		name, uPath, err := utility.Attributes()
		if err != nil {
			fmt.Printf("util: %s, path: %s %s[err]%s - %s\n", name, uPath, red, reset, err.Error())
			continue
		}

		fmt.Printf("util: %s, path: %s %s[ok]%s\n", name, uPath, green, reset)
	}
}

// Logformat adds default fields to each log entry.
type LogFormat struct {
	Fields    logrus.Fields
	Formatter logrus.Formatter
}

// Format satisfies the logrus.Formatter interface.
func (f *LogFormat) Format(e *logrus.Entry) ([]byte, error) {
	for k, v := range f.Fields {
		e.Data[k] = v
	}

	return f.Formatter.Format(e)
}
