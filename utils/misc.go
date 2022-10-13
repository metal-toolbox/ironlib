package utils

import (
	"io"
	"os"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/errs"
	"github.com/pkg/errors"
)

type DeviceIdentifiers struct {
	Vendor string
	Model  string
	Serial string
}

// copyFile makes a copy of the given file from src to dst
// setting the default permissions of 0644
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// nolint:gomnd // fs mode permissions are easier to read in this form
	return os.Chmod(dst, os.FileMode(0o600))
}

// IdentifyVendorModel returns the device vendor, model, serial number attributes
func IdentifyVendorModel(dmidecode *Dmidecode) (*DeviceIdentifiers, error) {
	device := &DeviceIdentifiers{}

	var err error

	device.Vendor, err = dmidecode.Manufacturer()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("manufacturer", "", 0), err.Error())
	}

	device.Model, err = dmidecode.ProductName()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("Product name", "", 0), err.Error())
	}

	// identify the vendor from the baseboard manufacturer - if the System Manufacturer attribute is unset
	if device.Vendor == "" || device.Vendor == common.SystemManufacturerUndefined {
		device.Vendor, err = dmidecode.BaseBoardManufacturer()
		if err != nil {
			return nil, errors.Wrap(errs.NewDmidecodeValueError("Baseboard Manufacturer", "", 0), err.Error())
		}
	}

	// identify the model from the baseboard - if the System Product Name is unset
	if device.Model == common.SystemManufacturerUndefined {
		device.Model, err = dmidecode.BaseBoardProductName()
		if err != nil {
			return nil, errors.Wrap(errs.NewDmidecodeValueError("Baseboard ProductName", "", 0), err.Error())
		}
	}

	device.Serial, err = dmidecode.SerialNumber()
	if err != nil {
		return nil, errors.Wrap(errs.NewDmidecodeValueError("Serial", "", 0), err.Error())
	}

	return device, nil
}
