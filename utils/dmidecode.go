package utils

import (
	"context"
	"strings"

	"github.com/dselans/dmidecode"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
)

type Dmidecode struct {
	dmi *dmidecode.DMI
}

func NewDmidecode() (d *Dmidecode, err error) {
	var dmi = dmidecode.New()

	err = dmi.Run()
	if err != nil {
		return d, err
	}

	return &Dmidecode{dmi: dmi}, err
}

// Returns a fake dmidecode instance for tests
func NewFakeDmidecode() *Dmidecode {
	return &Dmidecode{}
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

func (d *Dmidecode) TPM(ctx context.Context) (*model.TPM, error) {
	// TPM type ID
	tpmTypeID := 43

	records, err := d.queryType(tpmTypeID)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		return &model.TPM{
			Vendor:      model.VendorFromString(r["Description"]),
			Description: r["Description"],
			Firmware:    &model.Firmware{Installed: r["Firmware Revision"]},
			Metadata: map[string]string{
				"Specification Version": r["Specification Version"],
			},
		}, nil
	}

	return nil, nil
}
