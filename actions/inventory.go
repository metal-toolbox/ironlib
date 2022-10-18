package actions

import (
	"context"
	"log"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/r3labs/diff/v2"

	"github.com/metal-toolbox/ironlib/utils"
	"github.com/pkg/errors"
)

// The inventory package ties together the various collectors under utils
// using the interfaces defined utils.interface.

var (
	ErrInventoryDeviceObjNil = errors.New("method Inventory() expects a valid device object, got nil")
)

// Collectors is a struct acting as a registry of various inventory collectors
type Collectors struct {
	Inventory          InventoryCollector
	Drives             []DriveCollector
	NICs               NICCollector
	BMC                BMCCollector
	CPLDs              CPLDCollector
	BIOS               BIOSCollector
	TPMs               TPMCollector
	StorageControllers StorageControllerCollector
}

// InitCollectors constructs the Collectors object
// with lshw and smartctl collectors by default
//
// param trace causes the collectors to log command outputs
func InitCollectors(trace bool) *Collectors {
	return &Collectors{
		Inventory: utils.NewLshwCmd(trace),
		Drives:    []DriveCollector{utils.NewSmartctlCmd(trace)},
	}
}

// Collect collects device inventory data based on the given collectors,
// the device object fields are updated based on the collected inventory
//
// Inventory expects a device object and optionally a collectors object,
// when the collectors object provided is nil, the default collectors are added - lshw, smartctl
//
// The lshw collector always executes first and is included by default.
// nolint:gocyclo //since we're collecting inventory for each type, this is cyclomatic
func Collect(ctx context.Context, device *common.Device, collectors *Collectors, trace, failOnError bool) error {
	// register default collectors
	if collectors == nil {
		collectors = InitCollectors(trace)
	}

	// initialize a new device object - when a device isn't already provided
	if device == nil {
		deviceObj := common.NewDevice()
		device = &deviceObj
	}

	// register inventory collector
	if collectors.Inventory == nil {
		collectors.Inventory = utils.NewLshwCmd(trace)
	}

	// register a TPM inventory collector
	if collectors.TPMs == nil {
		var err error

		collectors.TPMs, err = utils.NewDmidecode()
		if err != nil && failOnError {
			return errors.Wrap(err, "error in dmidecode inventory collector")
		}
	}

	// Collect initial device inventory
	err := collectors.Inventory.Collect(ctx, device)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving device inventory")
	}

	// Collect drive smart data
	err = Drives(ctx, device.Drives, collectors.Drives)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving drive inventory")
	}

	// Collect NIC info
	err = NICs(ctx, device.NICs, collectors.NICs)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving NIC inventory")
	}

	// Collect BIOS info
	err = BIOS(ctx, device.BIOS, collectors.BIOS)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving BIOS inventory")
	}

	// Collect CPLD info
	err = CPLDs(ctx, &device.CPLDs, collectors.CPLDs)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving CPLD inventory")
	}

	// Collect BMC info
	err = BMC(ctx, device.BMC, collectors.BMC)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving BMC inventory")
	}

	// Collect TPM info
	err = TPMs(ctx, &device.TPMs, collectors.TPMs)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving TPM inventory")
	}

	// Collect StorageController info
	err = StorageController(ctx, device.StorageControllers, collectors.StorageControllers)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving StorageController inventory")
	}

	// default set model numbers to device model
	if device.BMC != nil && device.BMC.Model == "" {
		device.BMC.Model = device.Model
	}

	if device.BIOS != nil && device.BIOS.Model == "" {
		device.BIOS.Model = device.Model
	}

	for _, cpld := range device.CPLDs {
		if cpld != nil {
			cpld.Model = device.Model
		}
	}

	return nil
}

// Drives executes drive collectors and merges the data into device.[]*Drive
// nolint:gocyclo // TODO(joel) if theres more conditionals to be added in here, the method is to be split up.
func Drives(ctx context.Context, drives []*common.Drive, collectors []DriveCollector) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in Drives(): ", r)
		}
	}()

	if collectors == nil {
		return nil
	}

	for _, collector := range collectors {
		ndrives, err := collector.Drives(ctx)
		if err != nil {
			return err
		}

		if len(ndrives) == 0 {
			return nil
		}

		// TODO: handle case where the object may not already be present in device.Drives and needs to be added
		for _, e := range drives {
			for _, i := range ndrives {
				// object is matched by serial identifier and patched
				if strings.EqualFold(e.Serial, i.Serial) {
					changelog, err := diff.Diff(e, i)
					if err != nil {
						return err
					}

					changelog = vetChanges(changelog)
					diff.Patch(changelog, e)
				}
			}
		}
	}

	return nil
}

// NICs executes nic collectors and merges the nic data into device.[]*NIC
func NICs(ctx context.Context, nics []*common.NIC, c NICCollector) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in NICs(): ", r)
		}
	}()

	if c == nil {
		return nil
	}

	nnics, err := c.NICs(ctx)
	if err != nil {
		return err
	}

	if len(nnics) == 0 {
		return nil
	}

	// TODO: handle case where the object may not already be present in device.NICs and needs to be added
	for _, e := range nics {
		for _, i := range nnics {
			// object is matched by serial identifier and patched
			if strings.EqualFold(e.Serial, i.Serial) {
				changelog, err := diff.Diff(e, i)
				if err != nil {
					return err
				}

				changelog = vetChanges(changelog)
				diff.Patch(changelog, e)
			}
		}
	}

	return nil
}

// BMC executes the bmc collector and updates device bmc information
func BMC(ctx context.Context, bmc *common.BMC, c BMCCollector) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in BMC(): ", r)
		}
	}()

	if c == nil {
		return nil
	}

	nbmc, err := c.BMC(ctx)
	if err != nil {
		return err
	}

	changelog, err := diff.Diff(bmc, nbmc)
	if err != nil {
		return err
	}

	changelog = vetChanges(changelog)
	diff.Patch(changelog, bmc)

	return nil
}

// CPLDs executes the bmc collector and updates device cpld information
func CPLDs(ctx context.Context, cplds *[]*common.CPLD, c CPLDCollector) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in CPLDs(): ", r)
		}
	}()

	if c == nil {
		return nil
	}

	ncplds, err := c.CPLDs(ctx)
	if err != nil {
		return err
	}

	// no new cplds identified
	if len(ncplds) == 0 {
		return nil
	}

	// no existing cplds were passed in
	if len(*cplds) > 0 {
		changelog, err := diff.Diff(cplds, ncplds)
		if err != nil {
			return err
		}

		changelog = vetChanges(changelog)
		diff.Patch(changelog, cplds)
	} else {
		*cplds = append(*cplds, ncplds...)
	}

	return nil
}

// BIOS executes the bios collector and updates device bios information
func BIOS(ctx context.Context, bios *common.BIOS, c BIOSCollector) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in BIOS(): ", r)
		}
	}()

	if c == nil {
		return nil
	}

	nbios, err := c.BIOS(ctx)
	if err != nil {
		return err
	}

	if nbios == nil {
		return nil
	}

	changelog, err := diff.Diff(bios, nbios)
	if err != nil {
		return err
	}

	changelog = vetChanges(changelog)
	diff.Patch(changelog, bios)

	return nil
}

// TPMs executes the TPM collector and updates device TPM information
func TPMs(ctx context.Context, tpms *[]*common.TPM, c TPMCollector) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in TPM(): ", r)
		}
	}()

	if c == nil {
		return nil
	}

	ntpms, err := c.TPMs(ctx)
	if err != nil {
		return err
	}

	if ntpms == nil {
		return nil
	}

	// no tpms identified
	if len(ntpms) == 0 {
		return nil
	}

	// no existing tpms were passed in
	if len(*tpms) > 0 {
		changelog, err := diff.Diff(tpms, ntpms)
		if err != nil {
			return err
		}

		changelog = vetChanges(changelog)
		diff.Patch(changelog, tpms)
	} else {
		*tpms = append(*tpms, ntpms...)
	}

	return nil
}

// StorageControllers executes the StorageControllers collector and updates device storage controller data
func StorageController(ctx context.Context, controllers []*common.StorageController, c StorageControllerCollector) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in StorageController(): ", r)
		}
	}()

	if c == nil {
		return nil
	}

	ncontrollers, err := c.StorageControllers(ctx)
	if err != nil {
		return err
	}

	if len(ncontrollers) == 0 {
		return nil
	}

	// TODO: handle case where the object may not already be present in device.Controllers and needs to be added
	for _, e := range controllers {
		for _, i := range ncontrollers {
			// object is matched by serial identifier and patched
			if strings.EqualFold(e.Serial, i.Serial) {
				changelog, err := diff.Diff(e, i)
				if err != nil {
					return err
				}

				changelog = vetChanges(changelog)
				diff.Patch(changelog, e)
			}
		}
	}

	return nil
}

// vetChanges looks at the diff changelog and returns an updated Changelog
// with deletions and changes that zero or unset string, int values are excluded.
func vetChanges(changes diff.Changelog) diff.Changelog {
	accepted := diff.Changelog{}

	for _, c := range changes {
		// Skip changes that delete items
		if c.Type == diff.DELETE {
			continue
		}

		if c.Type == diff.UPDATE {
			if structFieldNotEmpty(c.From) {
				// Allow changes in the Vendor, Model fields
				if !utils.StringInSlice("Vendor", c.Path) &&
					!utils.StringInSlice("Model", c.Path) {
					continue
				}
			}
		}

		accepted = append(accepted, c)
	}

	return accepted
}

// returns true with the given field is empty or zero
func structFieldNotEmpty(field interface{}) bool {
	switch changeFrom := field.(type) {
	case string:
		if changeFrom != "" {
			return true
		}
	case int:
		if changeFrom != int(0) {
			return true
		}
	case int64:
		if changeFrom != int64(0) {
			return true
		}
	case int32:
		if changeFrom != int32(0) {
			return true
		}
	}

	return false
}
