package utils

import (
	"io/ioutil"

	"github.com/packethost/ironlib/errs"
	"github.com/packethost/ironlib/model"
	"github.com/pkg/errors"
)

// StringInSlice returns a bool value if the given string was not/found in the given slice.
func StringInSlice(str string, sl []string) bool {
	for _, element := range sl {
		if element == str {
			return true
		}
	}

	return false
}

// copyFile makes a copy of the given file from src to dst
// setting the default permissions of 0644
func copyFile(src, dst string) error {
	in, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, in, 0644)
	if err != nil {
		return err
	}

	return nil
}

// IdentifyVendorModel returns the device vendor, model, serial number attributes
func IdentifyVendorModel(dmidecode *Dmidecode) (deviceVendor, deviceModel, deviceSerial string, err error) {
	deviceVendor, err = dmidecode.Manufacturer()
	if err != nil {
		return deviceVendor, deviceModel, deviceSerial, errors.Wrap(errs.NewDmidecodeValueError("manufacturer", ""), err.Error())
	}

	deviceModel, err = dmidecode.ProductName()
	if err != nil {
		return deviceVendor, deviceModel, deviceSerial, errors.Wrap(errs.NewDmidecodeValueError("Product name", ""), err.Error())
	}

	// identify the vendor from the baseboard manufacturer - if the System Manufacturer attribute is unset
	if deviceVendor == "" || deviceVendor == model.SystemManufacturerUndefined {
		deviceVendor, err = dmidecode.BaseBoardManufacturer()
		if err != nil {
			return deviceVendor, deviceModel, deviceSerial, errors.Wrap(errs.NewDmidecodeValueError("Baseboard Manufacturer", ""), err.Error())
		}
	}

	// identify the model from the baseboard - if the System Product Name is unset
	if deviceModel == model.SystemManufacturerUndefined {
		deviceModel, err = dmidecode.BaseBoardProductName()
		if err != nil {
			return deviceVendor, deviceModel, deviceSerial, errors.Wrap(errs.NewDmidecodeValueError("Baseboard ProductName", ""), err.Error())
		}
	}

	deviceSerial, err = dmidecode.SerialNumber()
	if err != nil {
		return deviceVendor, deviceModel, deviceSerial, errors.Wrap(errs.NewDmidecodeValueError("Serial", ""), err.Error())
	}

	return deviceVendor, deviceModel, deviceSerial, nil
}
