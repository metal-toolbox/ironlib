package ironlib

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/packethost/ironlib/providers/dell"
	"github.com/packethost/ironlib/providers/supermicro"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
)

func NewDell(model string, dmidecode *utils.Dmidecode, l *logrus.Logger) Manager {

	var trace bool

	if l.GetLevel().String() == "trace" {
		trace = true
	}

	uid, _ := uuid.NewRandom()
	return &dell.Dell{
		ID:        uid.String(),
		Vendor:    "dell",
		Model:     model,
		Dmidecode: dmidecode,
		Dnf:       utils.NewDnf(trace),
		Dsu:       utils.NewDsu(trace),
		Logger:    l,
	}
}

func NewSupermicro(model string, dmidecode *utils.Dmidecode, l *logrus.Logger) Manager {

	var trace bool

	if l.GetLevel().String() == "trace" {
		trace = true
	}

	// register inventory collectors
	collectors := map[string]utils.Collector{
		"ipmi":     utils.NewIpmicfgCmd(trace),
		"smartctl": utils.NewSmartctlCmd(trace),
		"storecli": utils.NewStoreCLICmd(trace),
		"mlxup":    utils.NewMlxupCmd(trace),
	}

	uid, _ := uuid.NewRandom()
	return &supermicro.Supermicro{
		ID:         uid.String(),
		Vendor:     "supermicro",
		Model:      model,
		Dmidecode:  dmidecode,
		Collectors: collectors,
		Logger:     l,
	}
}

// New returns a device based on the hardware vendor, model attributes
// the apiclient is passed to the device instance to submit inventory information
func New(logger *logrus.Logger) (m Manager, err error) {

	if os.Getenv("TEST_DELL") != "" {
		dmidecode := utils.NewFakeDmidecode()
		return NewDell("PowerEdge R640 fake", dmidecode, logger), err
	}

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

	serial, err := dmidecode.SerialNumber()
	if err != nil {
		return nil, fmt.Errorf("unable to identify product name: %s", err.Error())
	}

	vendor = utils.FormatVendorStr(vendor)

	// update the log formatter
	logger.SetFormatter(&LogFormat{
		Fields: logrus.Fields{
			"vendor": vendor,
			"model":  model,
			"serial": serial,
		},
		Formatter: logger.Formatter,
	})

	logger.Info("hi there")
	switch vendor {
	case "dell":
		return NewDell(model, dmidecode, logger), nil
		//	case "HP", "HPE":
		//		return hpe.New(d), err
	case "supermicro":
		return NewSupermicro(model, dmidecode, logger), nil
		//	case "Quanta Cloud Technology Inc.":
		//		return quanta.New(d), err
		//	case "GIGABYTE":
		//		return gigabyte.New(d), err
		//	case "Intel Corporation":
		//		return intel.New(d), err
	default:
		return nil, fmt.Errorf("Unidentified vendor: " + vendor)
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
