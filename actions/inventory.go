package actions

import (
	"context"
	"log"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/r3labs/diff/v2"
	"golang.org/x/exp/slices"

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
	DriveCapabilities  []DriveCapabilityCollector
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
		Drives: []DriveCollector{
			utils.NewSmartctlCmd(trace),
			utils.NewLsblkCmd(trace),
		},
		DriveCapabilities: []DriveCapabilityCollector{
			utils.NewHdparmCmd(trace),
			utils.NewNvmeCmd(trace),
		},
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
func Collect(ctx context.Context, device *common.Device, collectors *Collectors, trace, failOnError, dynamic bool) error {
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

	// TODO (joel)
	//
	// move Drives(), NICs() and other methods under the Collectors struct
	// so that parameters can be set as fields on the Collectors struct and they don't have to be passed
	// as function parameters. This will also allow data collection to proceed even if
	// one drive/nic/storagecontroller/psu component returns an error.

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

	if dynamic {
		// Update StorageControllerCollectors
		for _, sc := range device.StorageControllers {
			collectors.StorageControllers = StorageControllerCollectorByVendor(sc.Vendor, trace)
		}
	}

	// Collect StorageController info
	err = StorageController(ctx, device.StorageControllers, collectors.StorageControllers)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving StorageController inventory")
	}

	if dynamic {
		for _, sc := range device.StorageControllers {
			if sc.SupportedRAIDTypes != "" {
				collectors.Drives = append(collectors.Drives, DriveCollectorByStorageControllerVendor(sc.Vendor, trace))

				err = Drives(ctx, device.Drives, collectors.Drives)
				if err != nil && failOnError {
					return errors.Wrap(err, "error retrieving drive inventory")
				}
			}
		}
	}

	err = DriveCapabilities(ctx, device.Drives, collectors.DriveCapabilities)
	if err != nil && failOnError {
		return errors.Wrap(err, "error retrieving DriveCapabilities")
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

// DriveCapabilities executes drive capability collectors
//
// The capability collector is identified based on the drive logical name.
func DriveCapabilities(ctx context.Context, drives []*common.Drive, collectors []DriveCapabilityCollector) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered from panic in DriveCapabilities(): ", r)
		}
	}()

	for _, drive := range drives {
		if drive.LogicalName == "" {
			continue
		}

		collector := driveCapabilityCollectorByLogicalName(drive.LogicalName, false, collectors)

		capabilities, err := collector.DriveCapabilities(ctx, drive.LogicalName)
		if err != nil {
			return err
		}

		drive.Capabilities = capabilities
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

				if len(changelog) > 0 {
					diff.Patch(changelog, e)
				}
			}
		}
	}

	return nil
}

// vetChanges looks at the diff changelog and returns an updated Changelog
// with deletions and changes that zero or unset string, int values are excluded.
func vetChanges(changes diff.Changelog) diff.Changelog {
	accepted := diff.Changelog{}

	for _, change := range changes {
		// Skip changes that delete items
		if acceptChange(&change) {
			accepted = append(accepted, change)
		}
	}

	return accepted
}

func acceptChange(change *diff.Change) bool {
	switch change.Type {
	case diff.DELETE:
		return false
	case diff.UPDATE:
		return vetUpdate(change)
	case diff.CREATE:
		return true
	}

	return false
}

// vetUpdate looks at a diff.Update change and returns true if the change is to be accepted
//
// nolint:gocyclo // validation is cyclomatic, and this logic grokable when kept in one method
func vetUpdate(change *diff.Change) bool {
	// allow vendor, model field changes only if the older value was not defined
	if slices.Contains(change.Path, "Vendor") || slices.Contains(change.Path, "Model") {
		if strings.TrimSpace(change.To.(string)) != "" && strings.TrimSpace(change.From.(string)) == "" {
			return true
		}
	}

	// accept description if its longer than the older value
	if slices.Contains(change.Path, "Description") {
		if len(change.To.(string)) > len(change.From.(string)) {
			return true
		}
	}

	// accept product name change if the older value was empty
	if slices.Contains(change.Path, "ProductName") {
		if change.From.(string) != "" {
			return false
		}
	}

	// remaining fields are type asserted,
	// if the old value is empty, accept the change
	switch newValue := change.To.(type) {
	case nil:
		return false
	case string:
		if strings.TrimSpace(newValue) != "" {
			return true
		}
	case int:
		if newValue != int(0) {
			return true
		}
	case int64:
		if newValue != int64(0) {
			return true
		}
	case int32:
		if newValue != int32(0) {
			return true
		}
	case *common.Firmware:
		if change.From == nil {
			return true
		}
	case map[string]string:
		if change.From == nil {
			return true
		}
	}

	return false
}
