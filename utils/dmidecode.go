package utils

import (
	"fmt"
	"strings"

	"github.com/dselans/dmidecode"
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

func (d *Dmidecode) query(section string, key string) (value string, err error) {
	var exists bool

	records, err := d.dmi.SearchByName(section)
	if err != nil {
		return value, fmt.Errorf("unable to read '%s'. error: %v", section, err)
	}

	for _, m := range records {
		if value, exists = m[key]; exists {
			return strings.TrimSpace(value), err
		}
	}

	return value, fmt.Errorf("unable to read '%s[%s]'. error: %v", section, key, err)
}

// Manufacturer queries dmidecode and returns server vendor
func (d *Dmidecode) Manufacturer() (vendor string, err error) {
	return d.query("System Information", "Manufacturer")
}

// ProductName queries dmidecode and returns the product name
func (d *Dmidecode) ProductName() (serial string, err error) {
	return d.query("System Information", "Product Name")
}

// SerialNumber queries dmidecode and returns the serial number
func (d *Dmidecode) SerialNumber() (serial string, err error) {
	return d.query("System Information", "Serial Number")
}

// BaseBoardSerialNumber queries dmidecode and returns the base board serial number
func (d *Dmidecode) BaseBoardSerialNumber() (serial string, err error) {
	return d.query("Base Board Information", "Serial Number")
}

// BaseBoardProductName queries dmidecode and returns the base board product name
func (d *Dmidecode) BaseBoardProductName() (serial string, err error) {
	return d.query("Base Board Information", "Product Name")
}

// ChassisSerialNumber queries dmidecode and returns the chassis serial number
func (d *Dmidecode) ChassisSerialNumber() (serial string, err error) {
	return d.query("Chassis Information", "Serial Number")
}

// BIOSVersion queries dmidecode and returns the BIOS version
func (d *Dmidecode) BIOSVersion() (serial string, err error) {
	return d.query("BIOS Information", "Version")
}

func FormatVendorStr(v string) string {

	switch v {
	case "Dell Inc.":
		return "dell"
	case "HP", "HPE":
		return "hp"
	case "Supermicro":
		return "supermicro"
	case "Quanta Cloud Technology Inc.":
		return "quanta"
	case "GIGABYTE":
		return "gigabyte"
	case "Intel Corporation":
		return "intel"
	default:
		return v
	}

}
