package model

// Component is a low level device component - before its classified into a device type (BMC{}, BIOS{}, NIC{})
type Component struct {
	Serial            string            `json:"serial"`
	Vendor            string            `json:"vendor"`
	Type              string            `json:"type"`
	Model             string            `json:"model"`
	Name              string            `json:"name"`
	Slug              string            `json:"slug"`
	FirmwareInstalled string            `json:"firmware_installed"` // The firmware revision installed
	FirmwareAvailable string            `json:"firmware_available"` // The firmware revision available
	Metadata          map[string]string `json:"metadata"`           // Additional metadata if any
	Oem               bool              `json:"oem"`                // Component is an OEM component
	FirmwareManaged   bool              `json:"firmware_managed"`   // Firmware on the component is managed/unmanaged
}

// UpdateOptions sets firmware update options for a device component
type UpdateOptions struct {
	AllowDowngrade   bool // Allow firmware to be downgraded
	InstallAll       bool // Install all available updates (specific to dell DSU)
	Serial           string
	Vendor           string
	Model            string
	Name             string
	Slug             string
	UpdateFile       string
	InstallerVersion string // The all available updates installer version (specific to dell DSU)
}

// OemComponents are OEM specific device components
type OemComponents struct {
	Dell []*Component `json:"dell"`
}

// ComponentFirmware accepts a generic component object and a pointer to a firmware object
// the firmware object field values are then updated based on the component object
// values in the Firmware object are only overwritten if they are empty
func ComponentFirmware(c *Component, f *Firmware) {
	if f.Installed == "" && c.FirmwareInstalled != "" {
		f.Installed = c.FirmwareInstalled
	}

	if f.Available == "" && c.FirmwareAvailable != "" {
		f.Available = c.FirmwareAvailable
	}

	if !f.Managed && c.FirmwareManaged {
		f.Managed = c.FirmwareManaged
	}

	if len(f.Metadata) == 0 && len(c.Metadata) > 0 {
		f.Metadata = c.Metadata
	}
}

// ComponentFirmwareNICs iterates over the given components and sets the firmware information for the NICs
// there is no notion of matching the right component with the nic because in many cases, the serial/mac is unknown
// when the component firmware data is collected
// oem indicates the device firmware is managed by the OEM vendor
func ComponentFirmwareNICs(nics []*NIC, components []*Component, oem bool) {
	for _, c := range components {
		for _, n := range nics {
			n.Oem = oem
			if n.Firmware == nil {
				n.Firmware = &Firmware{}
			}

			ComponentFirmware(c, n.Firmware)
		}
	}
}

// ComponentFirmwarePSUs iterates over the given components and sets the firmware information for the PSUs
// oem indicates the device firmware is managed by the OEM vendor
func ComponentFirmwarePSUs(psus []*PSU, components []*Component, oem bool) {
	for _, c := range components {
		for _, p := range psus {
			p.Oem = oem
			if p.Firmware == nil {
				p.Firmware = &Firmware{}
			}

			ComponentFirmware(c, p.Firmware)
		}
	}
}

// ComponentFirmwareDrives iterates over the given components and sets the firmware information for the Drives
// oem indicates the device firmware is managed by the OEM vendor
func ComponentFirmwareDrives(drives []*Drive, components []*Component, oem bool) {
	for _, c := range components {
		for _, d := range drives {
			d.Oem = oem
			if d.Firmware == nil {
				d.Firmware = &Firmware{}
			}

			ComponentFirmware(c, d.Firmware)
		}
	}
}

// ComponentFirmwareStorageControllers iterates over the given components and sets the firmware information for the Drives
// oem indicates the device firmware is managed by the OEM vendor
func ComponentFirmwareStorageControllers(controllers []*StorageController, components []*Component, oem bool) {
	for _, c := range components {
		for _, s := range controllers {
			s.Oem = oem
			if s.Firmware == nil {
				s.Firmware = &Firmware{}
			}

			ComponentFirmware(c, s.Firmware)
		}
	}
}

// SetDeviceComponents populates the device with the given components
func SetDeviceComponents(device *Device, components []*Component) {
	//  multiples of components are grouped
	multiples := map[string][]*Component{
		SlugNIC:               {},
		SlugPSU:               {},
		SlugDrive:             {},
		SlugStorageController: {},
	}

	// set firmware information for device components
	for _, c := range components {
		// populate Dell specific OEM components to device
		_, isOem := OemComponentDell[c.Slug]
		if isOem {
			device.OemComponents.Dell = append(device.OemComponents.Dell, c)
		}

		// populate rest of the device components
		switch c.Slug {
		case SlugBIOS:
			device.BIOS.Firmware = &Firmware{}
			ComponentFirmware(c, device.BIOS.Firmware)
		case SlugBMC:
			device.BMC.Firmware = &Firmware{}
			ComponentFirmware(c, device.BMC.Firmware)
		case SlugNIC:
			multiples[SlugNIC] = append(multiples[SlugNIC], c)
		case SlugPSU:
			multiples[SlugPSU] = append(multiples[SlugPSU], c)
		case SlugDrive:
			multiples[SlugDrive] = append(multiples[SlugDrive], c)
		case SlugSASHBA330Controller:
			multiples[SlugSASHBA330Controller] = append(multiples[SlugSASHBA330Controller], c)
		}
	}

	// populate the firmware information for multiples components
	for slug, components := range multiples {
		switch slug {
		case SlugNIC:
			ComponentFirmwareNICs(device.NICs, components, true)
		case SlugPSU:
			ComponentFirmwarePSUs(device.PSUs, components, true)
		case SlugDrive:
			ComponentFirmwareDrives(device.Drives, components, true)
		case SlugSASHBA330Controller:
			ComponentFirmwareStorageControllers(device.StorageControllers, components, true)
		}
	}
}
