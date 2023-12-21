// The inventory package ties together the various collectors under utils through its InventoryCollectorAction type.
package actions

import (
	"context"
	"runtime/debug"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
	"github.com/r3labs/diff/v2"
	"golang.org/x/exp/slices"

	"github.com/metal-toolbox/ironlib/firmware"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/metal-toolbox/ironlib/utils"
)

var (
	ErrInventoryDeviceObjNil = errors.New("method Inventory() expects a valid device object, got nil")
	ErrPanic                 = errors.New("recovered from panic")
)

// InventoryCollectorAction provides methods to collect hardware, firmware inventory.
type InventoryCollectorAction struct {
	// collectors registered for inventory collection.
	collectors Collectors

	// device is the model in which the collected inventory is recorded.
	device *common.Device

	// Enable trace logging on the collectors.
	trace bool

	// a.failOnError when enabled causes ironlib to return an error
	// if a collector returns an error
	failOnError bool

	// dynamicCollection when enabled will cause ironlib to
	// enable collectors based on the detected component vendor.
	dynamicCollection bool

	// disabledCollectorUtilities is the list of collector utilities
	// to be disabled, this is the name of collector utility
	// which is returned by its Attributes() method.
	disabledCollectorUtilities []model.CollectorUtility
}

// Collectors is a struct acting as a registry of various inventory collectors
type Collectors struct {
	InventoryCollector
	NICCollector
	BMCCollector
	CPLDCollector
	BIOSCollector
	TPMCollector
	FirmwareChecksumCollector
	UEFIVarsCollector
	StorageControllerCollectors []StorageControllerCollector
	DriveCollectors             []DriveCollector
	DriveCapabilitiesCollectors []DriveCapabilityCollector
}

// Empty returns a bool value
func (c *Collectors) Empty() bool {
	if c.InventoryCollector == nil &&
		c.NICCollector == nil &&
		c.BMCCollector == nil &&
		c.CPLDCollector == nil &&
		c.BIOSCollector == nil &&
		c.TPMCollector == nil &&
		len(c.StorageControllerCollectors) == 0 &&
		len(c.DriveCollectors) == 0 &&
		len(c.DriveCapabilitiesCollectors) == 0 &&
		c.UEFIVarsCollector == nil &&
		c.FirmwareChecksumCollector == nil {
		return true
	}

	return false
}

// Option returns a function that sets a InventoryCollectorAction parameter
type Option func(*InventoryCollectorAction)

// WithTraceLevel sets trace level logging on the action runner.
func WithTraceLevel() Option {
	return func(a *InventoryCollectorAction) {
		a.trace = true
	}
}

// WithFailOnError sets the InventoryCollectorAction to return on any error
// that may occur when collecting inventory.
//
// By default the InventoryCollectorAction continues to collect inventory
// even if one or more collectors fail.
func WithFailOnError() Option {
	return func(a *InventoryCollectorAction) {
		a.failOnError = true
	}
}

// DynamicCollection when enabled will cause ironlib to
// identify collectors based on the detected component vendor.
func WithDynamicCollection() Option {
	return func(a *InventoryCollectorAction) {
		a.dynamicCollection = true
	}
}

// WithCollectors sets collectors to the ones passed in as a parameter.
func WithCollectors(collectors *Collectors) Option {
	return func(a *InventoryCollectorAction) {
		a.collectors = *collectors
	}
}

// WithDisabledCollectorUtilities disables the given collector utilities.
func WithDisabledCollectorUtilities(utilityNames []model.CollectorUtility) Option {
	return func(a *InventoryCollectorAction) {
		a.disabledCollectorUtilities = utilityNames
	}
}

// NewActionrunner returns an Actions runner that is capable of collecting inventory.
func NewInventoryCollectorAction(options ...Option) *InventoryCollectorAction {
	a := &InventoryCollectorAction{}

	// set options to override
	for _, opt := range options {
		opt(a)
	}

	// set default collectors when none have been set through options.
	if a.collectors.Empty() {
		a.collectors = Collectors{
			InventoryCollector: utils.NewLshwCmd(a.trace),
			DriveCollectors: []DriveCollector{
				utils.NewSmartctlCmd(a.trace),
				utils.NewLsblkCmd(a.trace),
			},
			DriveCapabilitiesCollectors: []DriveCapabilityCollector{
				utils.NewHdparmCmd(a.trace),
				utils.NewNvmeCmd(a.trace),
			},
			FirmwareChecksumCollector: firmware.NewChecksumCollector(
				firmware.MakeOutputPath(),
				firmware.TraceExecution(a.trace),
			),
			// implement uefi vars collector and plug in here
			// UEFIVarsCollector: ,
		}
	}

	// the lshw collector cannot be disabled, since its the primary inventory collector.
	if a.collectors.InventoryCollector == nil {
		a.collectors.InventoryCollector = utils.NewLshwCmd(a.trace)
	}

	return a
}

// Collect collects device inventory data based on the given collectors,
// the device object fields are updated based on the collected inventory
//
// Inventory expects a device object and optionally a collectors object,
// when the collectors object provided is nil, the default collectors are added - lshw, smartctl
//
// The lshw collector always executes first and is included by default.
// nolint:gocyclo //since we're collecting inventory for each type, this is cyclomatic
func (a *InventoryCollectorAction) Collect(ctx context.Context, device *common.Device) error {
	// initialize a new device object - when a device isn't already provided
	if device == nil {
		deviceObj := common.NewDevice()
		device = &deviceObj
	}

	a.device = device

	// register a TPM inventory collector
	if a.collectors.TPMCollector == nil && !slices.Contains(a.disabledCollectorUtilities, model.CollectorUtility("dmidecode")) {
		var err error

		a.collectors.TPMCollector, err = utils.NewDmidecode()
		if err != nil && a.failOnError {
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
	err := a.collectors.InventoryCollector.Collect(ctx, a.device)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving device inventory")
	}

	// Collect drive smart data
	err = a.CollectDrives(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving drive inventory")
	}

	// Collect NIC info
	err = a.CollectNICs(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving NIC inventory")
	}

	// Collect BIOS info
	err = a.CollectBIOS(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving BIOS inventory")
	}

	// Collect CPLD info
	err = a.CollectCPLDs(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving CPLD inventory")
	}

	// Collect BMC info
	err = a.CollectBMC(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving BMC inventory")
	}

	// Collect TPM info
	err = a.CollectTPMs(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving TPM inventory")
	}

	// Collect Firmware checksums
	err = a.CollectFirmwareChecksums(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving Firmware checksums")
	}

	// Collect UEFI variables
	err = a.CollectUEFIVariables(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving UEFI variables")
	}

	// Update StorageControllerCollectors based on controller vendor attributes
	if a.dynamicCollection {
		for _, sc := range a.device.StorageControllers {
			if c := StorageControllerCollectorByVendor(sc.Vendor, a.trace); c != nil {
				a.collectors.StorageControllerCollectors = append(a.collectors.StorageControllerCollectors, c)
			}
		}
	}

	// Collect StorageController info
	err = a.CollectStorageControllers(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving StorageController inventory")
	}

	// Update DriveCollectors based on drive vendor attributes
	if a.dynamicCollection {
		for _, sc := range a.device.StorageControllers {
			if c := DriveCollectorByStorageControllerVendor(sc.Vendor, a.trace); c != nil {
				a.collectors.DriveCollectors = append(a.collectors.DriveCollectors, c)
			}
		}

		if len(a.collectors.DriveCollectors) > 0 {
			err = a.CollectDrives(ctx)

			if err != nil && a.failOnError {
				return errors.Wrap(err, "error retrieving drive inventory")
			}
		}
	}

	// CollectDriveCapabilities is to be invoked after Drives()
	err = a.CollectDriveCapabilities(ctx)
	if err != nil && a.failOnError {
		return errors.Wrap(err, "error retrieving DriveCapabilities")
	}

	a.setDefaultAttributes()

	return nil
}

// setDefaultAttributes sets device default attributes
func (a *InventoryCollectorAction) setDefaultAttributes() {
	// default set model numbers to device model
	if a.device.BMC != nil && a.device.BMC.Model == "" {
		a.device.BMC.Model = a.device.Model
	}

	if a.device.BIOS != nil && a.device.BIOS.Model == "" {
		a.device.BIOS.Model = a.device.Model
	}

	for _, cpld := range a.device.CPLDs {
		if cpld != nil {
			cpld.Model = a.device.Model
		}
	}
}

// CollectDrives executes drive collectors and merges the data into device.[]*Drive
// nolint:gocyclo // TODO(joel) if theres more conditionals to be added in here, the method is to be split up.
func (a *InventoryCollectorAction) CollectDrives(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if a.collectors.DriveCollectors == nil {
		return nil
	}

	for _, collector := range a.collectors.DriveCollectors {
		// skip collector if its been disabled
		collectorKind, _, _ := collector.Attributes()
		if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
			continue
		}

		ndrives, err := collector.Drives(ctx)
		if err != nil {
			return err
		}

		if len(ndrives) == 0 {
			return nil
		}

		for _, existing := range a.device.Drives {
			// match existing drives by serial, and patch with changes
			found := a.findDriveBySerial(existing.Serial, ndrives)
			if found != nil {
				// diff existing drive fields with the one found by the collector
				changelog, err := diff.Diff(existing, found)
				if err != nil {
					return err
				}

				changelog = a.vetChanges(changelog)
				diff.Patch(changelog, existing)

				continue
			}

			// as a fallback for the ndrives data that might not include a serial number,
			// match existing drives by logical name and patch with changes
			found = a.findDriveByLogicalName(existing.LogicalName, ndrives)
			if found != nil {
				// diff existing drive fields with the one found by the collector
				changelog, err := diff.Diff(existing, found)
				if err != nil {
					return err
				}

				changelog = a.vetChanges(changelog)
				diff.Patch(changelog, existing)
			}
		}

		// add drive if it isn't part of the drives slice based on its serial
		for _, new := range ndrives {
			found := a.findDriveBySerial(new.Serial, a.device.Drives)
			if found != nil && found.Serial != "" {
				continue
			}

			a.device.Drives = append(a.device.Drives, new)
		}
	}

	return nil
}

func (a *InventoryCollectorAction) findDriveBySerial(serial string, drives []*common.Drive) *common.Drive {
	for _, drive := range drives {
		if strings.EqualFold(serial, drive.Serial) {
			return drive
		}
	}

	return nil
}

func (a *InventoryCollectorAction) findDriveByLogicalName(logicalName string, drives []*common.Drive) *common.Drive {
	for _, drive := range drives {
		if strings.EqualFold(logicalName, drive.LogicalName) {
			return drive
		}
	}

	return nil
}

// CollectDriveCapabilities executes drive capability collectors
//
// The capability collector is identified based on the drive logical name.
func (a *InventoryCollectorAction) CollectDriveCapabilities(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	for _, drive := range a.device.Drives {
		// check capabilities on drives that are either SATA or NVME,
		//
		// if theres others to be supported, the driveCapabilityCollectorByLogicalName() method
		// is to be updated to include the required support for SAS/USB/SCSI or other kinds of transports.
		if !slices.Contains([]string{"sata", "nvme"}, drive.Protocol) {
			continue
		}

		// drive logical name is required
		if drive.LogicalName == "" {
			continue
		}

		collector := driveCapabilityCollectorByLogicalName(drive.LogicalName, false, a.collectors.DriveCapabilitiesCollectors)

		// skip collector if its been disabled
		collectorKind, _, _ := collector.Attributes()
		if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
			continue
		}

		capabilities, err := collector.DriveCapabilities(ctx, drive.LogicalName)
		if err != nil {
			return err
		}

		drive.Capabilities = capabilities
	}

	return nil
}

// CollectNICs executes nic collectors and merges the nic data into device.[]*NIC
//
// nolint:gocyclo // this is fine for now.
func (a *InventoryCollectorAction) CollectNICs(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if a.collectors.NICCollector == nil {
		return nil
	}

	// skip collector if its been disabled
	collectorKind, _, _ := a.collectors.NICCollector.Attributes()
	if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
		return nil
	}

	found, err := a.collectors.NICs(ctx)
	if err != nil {
		return err
	}

	if len(found) == 0 {
		return nil
	}

	// TODO: handle case where the object may not already be present in device.NICs and needs to be added
	for _, e := range a.device.NICs {
		for _, n := range found {
			// object is matched by serial identifier and patched
			if strings.EqualFold(e.Serial, n.Serial) {
				changelog, err := diff.Diff(e, n)
				if err != nil {
					return err
				}

				changelog = a.vetChanges(changelog)
				diff.Patch(changelog, e)
			}
		}
	}

	return nil
}

// CollectBMC executes the bmc collector and updates device bmc information
func (a *InventoryCollectorAction) CollectBMC(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if a.collectors.BMCCollector == nil {
		return nil
	}

	// skip collector if its been disabled
	collectorKind, _, _ := a.collectors.BMCCollector.Attributes()
	if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
		return nil
	}

	found, err := a.collectors.BMC(ctx)
	if err != nil {
		return err
	}

	changelog, err := diff.Diff(a.device.BMC, found)
	if err != nil {
		return err
	}

	changelog = a.vetChanges(changelog)
	diff.Patch(changelog, a.device.BMC)

	return nil
}

// CollectCPLDs executes the bmc collector and updates device cpld information
func (a *InventoryCollectorAction) CollectCPLDs(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if a.collectors.CPLDCollector == nil {
		return nil
	}

	// skip collector if its been disabled
	collectorKind, _, _ := a.collectors.CPLDCollector.Attributes()
	if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
		return nil
	}

	found, err := a.collectors.CPLDs(ctx)
	if err != nil {
		return err
	}

	// no new cplds identified
	if len(found) == 0 {
		return nil
	}

	if len(a.device.CPLDs) > 0 {
		changelog, err := diff.Diff(a.device.CPLDs, found)
		if err != nil {
			return err
		}

		changelog = a.vetChanges(changelog)
		diff.Patch(changelog, a.device.CPLDs)
	} else {
		a.device.CPLDs = append(a.device.CPLDs, found...)
	}

	return nil
}

// CollectBIOS executes the bios collector and updates device bios information
func (a *InventoryCollectorAction) CollectBIOS(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if a.collectors.BIOSCollector == nil {
		return nil
	}

	// skip collector if its been disabled
	collectorKind, _, _ := a.collectors.BIOSCollector.Attributes()
	if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
		return nil
	}

	found, err := a.collectors.BIOS(ctx)
	if err != nil {
		return err
	}

	if found == nil {
		return nil
	}

	changelog, err := diff.Diff(a.device.BIOS, found)
	if err != nil {
		return err
	}

	changelog = a.vetChanges(changelog)
	diff.Patch(changelog, a.device.BIOS)

	return nil
}

// CollectTPMs executes the TPM collector and updates device TPM information
func (a *InventoryCollectorAction) CollectTPMs(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if a.collectors.TPMCollector == nil {
		return nil
	}

	// skip collector if its been disabled
	collectorKind, _, _ := a.collectors.TPMCollector.Attributes()
	if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
		return nil
	}

	found, err := a.collectors.TPMs(ctx)
	if err != nil {
		return err
	}

	if found == nil {
		return nil
	}

	if len(a.device.TPMs) > 0 {
		changelog, err := diff.Diff(a.device.TPMs, found)
		if err != nil {
			return err
		}

		changelog = a.vetChanges(changelog)
		diff.Patch(changelog, a.device.TPMs)
	} else {
		a.device.TPMs = append(a.device.TPMs, found...)
	}

	return nil
}

// CollectFirmwareChecksums executes the Firmware checksum collector and updates the component metadata.
func (a *InventoryCollectorAction) CollectFirmwareChecksums(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if a.collectors.FirmwareChecksumCollector == nil {
		return nil
	}

	// skip collector if we explicitly disable anything related to firmware checksumming.
	collectorKind, _, _ := a.collectors.FirmwareChecksumCollector.Attributes()
	if slices.Contains(a.disabledCollectorUtilities, collectorKind) ||
		slices.Contains(a.disabledCollectorUtilities, firmware.FirmwareDumpUtility) ||
		slices.Contains(a.disabledCollectorUtilities, firmware.UEFIParserUtility) {
		return nil
	}

	sumStr, err := a.collectors.FirmwareChecksumCollector.BIOSLogoChecksum(ctx)
	if err != nil {
		return err
	}

	if a.device.BIOS == nil {
		// XXX: how did we get here?
		return nil
	}

	a.device.BIOS.Metadata["bios-logo-checksum"] = sumStr

	return nil
}

// CollectUEFIVariables executes the UEFI variable collector and stores them on the device object
func (a *InventoryCollectorAction) CollectUEFIVariables(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if a.collectors.UEFIVarsCollector == nil {
		return nil
	}

	// skip collector if its been disabled
	collectorKind, _, _ := a.collectors.UEFIVarsCollector.Attributes()
	if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
		return nil
	}

	keyValues, err := a.collectors.UEFIVariables(ctx)
	if err != nil {
		return err
	}

	if len(keyValues) == 0 || a.device.BIOS == nil {
		return nil
	}

	for k, v := range keyValues {
		// do we want a prefix?
		a.device.Metadata["EFI_VAR-"+k] = v
	}

	return nil
}

// CollectStorageControllers executes the StorageControllers collectors and updates device storage controller data
//
// nolint:gocyclo // this is fine for now
func (a *InventoryCollectorAction) CollectStorageControllers(ctx context.Context) error {
	// nolint:errcheck // deferred method catches a panic, error check not required.
	defer func() error {
		if r := recover(); r != nil && a.failOnError {
			return errors.Wrap(ErrPanic, string(debug.Stack()))
		}

		return nil
	}()

	if len(a.collectors.StorageControllerCollectors) == 0 {
		return nil
	}

	for _, collector := range a.collectors.StorageControllerCollectors {
		// skip collector if its been disabled
		collectorKind, _, _ := collector.Attributes()
		if slices.Contains(a.disabledCollectorUtilities, collectorKind) {
			continue
		}

		found, err := collector.StorageControllers(ctx)
		if err != nil {
			return err
		}

		if len(found) == 0 {
			return nil
		}

		for _, existing := range a.device.StorageControllers {
			a.findStorageControllerBySerial(existing.Serial, found)

			if found != nil {
				// diff existing fields with the one found
				changelog, err := diff.Diff(existing, found)
				if err != nil {
					return err
				}

				changelog = a.vetChanges(changelog)
				diff.Patch(changelog, existing)

				continue
			}

			// add storage controller if it isn't part of existing controllers on the device
			for _, new := range found {
				found := a.findStorageControllerBySerial(new.Serial, a.device.StorageControllers)
				if found != nil && found.Serial != "" {
					continue
				}

				a.device.StorageControllers = append(a.device.StorageControllers, new)
			}
		}
	}

	return nil
}

func (a *InventoryCollectorAction) findStorageControllerBySerial(serial string, controllers []*common.StorageController) *common.StorageController {
	for _, controller := range controllers {
		if strings.EqualFold(serial, controller.Serial) {
			return controller
		}
	}

	return nil
}

// vetChanges looks at the diff changelog and returns an updated Changelog
// with deletions and changes that zero or unset string, int values are excluded.
func (a *InventoryCollectorAction) vetChanges(changes diff.Changelog) diff.Changelog {
	accepted := diff.Changelog{}

	for _, change := range changes {
		// Skip changes that delete items
		if a.acceptChange(&change) {
			accepted = append(accepted, change)
		}
	}

	return accepted
}

func (a *InventoryCollectorAction) acceptChange(change *diff.Change) bool {
	switch change.Type {
	case diff.DELETE:
		return false
	case diff.UPDATE:
		return a.vetUpdate(change)
	case diff.CREATE:
		return true
	}

	return false
}

// vetUpdate looks at a diff.Update change and returns true if the change is to be accepted
//
// nolint:gocyclo // validation is cyclomatic, and this logic grokable when kept in one method
func (a *InventoryCollectorAction) vetUpdate(change *diff.Change) bool {
	// allow vendor, model field changes only if the older value was not defined
	if slices.Contains(change.Path, "Vendor") || slices.Contains(change.Path, "Model") {
		if strings.TrimSpace(change.To.(string)) != "" && strings.TrimSpace(change.From.(string)) == "" {
			return true
		}
	}

	// accept description if its longer than the older value
	if slices.Contains(change.Path, "Description") {
		if len(strings.TrimSpace(change.To.(string))) > len(strings.TrimSpace(change.From.(string))) {
			return true
		}
	}

	// keep product name if not empty
	if slices.Contains(change.Path, "ProductName") {
		if strings.TrimSpace(change.From.(string)) != "" {
			return false
		}
	}

	// keep existing serial value if not empty
	if slices.Contains(change.Path, "Serial") {
		if strings.TrimSpace(change.From.(string)) != "" {
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
		if newValue >= -1 {
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
