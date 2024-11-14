package utils

import (
	"cmp"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/dselans/dmidecode"
	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

const EnvDmidecodeUtility = "IRONLIB_UTIL_DMIDECODE"

type Dmidecode struct {
	dmi *dmidecode.DMI
}

func NewDmidecode() (d *Dmidecode, err error) {
	utility := cmp.Or(os.Getenv(EnvDmidecodeUtility), "dmidecode")

	dmi := dmidecode.New()
	output, err := dmi.ExecDmidecode(utility)
	if err != nil {
		return nil, err
	}

	err = dmi.ParseDmidecode(output)
	if err != nil {
		return nil, err
	}

	return &Dmidecode{dmi: dmi}, nil
}

// Attributes implements the actions.UtilAttributeGetter interface
func (d *Dmidecode) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	utility := cmp.Or(os.Getenv(EnvDmidecodeUtility), "dmidecode")
	path, err := exec.LookPath(utility)

	return "dmidecode", path, err
}

// InitFakeDmidecode returns a fake dmidecode instance loaded with the dmidecode output from testFile
func InitFakeDmidecode(testFile string) (*Dmidecode, error) {
	b, err := os.ReadFile(testFile)
	if err != nil {
		return nil, err
	}

	// setup a dmidecode instance
	d := dmidecode.New()

	err = d.ParseDmidecode(string(b))
	if err != nil {
		return nil, err
	}

	// wrap the dmidecode instance in our Dmidecode wrapper
	return &Dmidecode{dmi: d}, nil
}

func (d *Dmidecode) query(section, key string) (value string, err error) {
	var exists bool

	records, err := d.dmi.SearchByName(section)
	if err != nil {
		return value, errors.Wrap(errs.NewDmidecodeValueError(section, key, 0), err.Error())
	}

	for _, m := range records {
		if value, exists = m[key]; exists {
			return strings.TrimSpace(value), err
		}
	}

	return value, errors.Wrap(errs.NewDmidecodeValueError(section, key, 0), err.Error())
}

func (d *Dmidecode) queryType(id int) (records []dmidecode.Record, err error) {
	records, err = d.dmi.SearchByType(id)
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("", "", id), err.Error())
	}

	return records, nil
}

// Manufacturer queries dmidecode and returns server vendor
func (d *Dmidecode) Manufacturer() (string, error) {
	return d.query("System Information", "Manufacturer")
}

// ProductName queries dmidecode and returns the product name
func (d *Dmidecode) ProductName() (string, error) {
	return d.query("System Information", "Product Name")
}

// SerialNumber queries dmidecode and returns the serial number
func (d *Dmidecode) SerialNumber() (string, error) {
	return d.query("System Information", "Serial Number")
}

// BaseBoardSerialNumber queries dmidecode and returns the base board serial number
func (d *Dmidecode) BaseBoardSerialNumber() (string, error) {
	return d.query("Base Board Information", "Serial Number")
}

// BaseBoardProductName queries dmidecode and returns the base board product name
func (d *Dmidecode) BaseBoardProductName() (string, error) {
	return d.query("Base Board Information", "Product Name")
}

// BaseBoardManufacturer queries dmidecode and returns the baseboard-manufacturer
func (d *Dmidecode) BaseBoardManufacturer() (string, error) {
	return d.query("Base Board Information", "Manufacturer")
}

// ChassisSerialNumber queries dmidecode and returns the chassis serial number
func (d *Dmidecode) ChassisSerialNumber() (string, error) {
	return d.query("Chassis Information", "Serial Number")
}

// BIOSVersion queries dmidecode and returns the BIOS version
func (d *Dmidecode) BIOSVersion() (string, error) {
	return d.query("BIOS Information", "Version")
}

func (d *Dmidecode) TPMs(context.Context) ([]*common.TPM, error) {
	// TPM type ID
	tpmTypeID := 43

	records, err := d.queryType(tpmTypeID)
	if err != nil {
		return nil, err
	}

	tpms := []*common.TPM{}
	for _, r := range records {
		tpms = append(tpms, &common.TPM{
			Common: common.Common{
				Vendor:      common.VendorFromString(r["Description"]),
				Description: r["Description"],
				Firmware:    &common.Firmware{Installed: r["Firmware Revision"]},
				Metadata: map[string]string{
					"Specification Version": r["Specification Version"],
				},
			},
		})
	}

	return tpms, nil
}
