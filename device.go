package ironlib

import (
	"fmt"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/providers/asrockrack"
	"github.com/packethost/ironlib/providers/dell"
	"github.com/packethost/ironlib/providers/generic"
	"github.com/packethost/ironlib/providers/supermicro"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
)

// New returns a device Manager interface based on the hardware vendor, model attributes
// by default returns a Generic device instance that only returns the device inventory
func New(logger *logrus.Logger) (m model.DeviceManager, err error) {

	dmidecode, err := utils.NewDmidecode()
	if err != nil {
		return nil, fmt.Errorf("failed to load dmidecode to identify device: %s", err.Error())
	}

	vendor, err := dmidecode.Manufacturer()
	if err != nil {
		return nil, fmt.Errorf("unable to identify vendor: %s", err.Error())
	}

	model, err := dmidecode.ProductName()
	if err != nil {
		return nil, fmt.Errorf("unable to identify product name: %s", err.Error())
	}

	vendor = utils.FormatVendorName(vendor)
	model = utils.FormatProductName(model)

	switch vendor {
	case "dell":
		return dell.New(vendor, model, logger)
	case "supermicro":
		return supermicro.New(vendor, model, logger)
	case "packet":
		switch model {
		case "c3.small.x86":
			vendor, err = dmidecode.BaseBoardManufacturer()
			if err != nil {
				return nil, fmt.Errorf("error listing baseboard manufacturer")
			}
		}
		return asrockrack.New(vendor, model, logger)
	default:
		return generic.New(vendor, model, logger)
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
