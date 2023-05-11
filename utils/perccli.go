package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/metal-toolbox/ironlib/model"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

const (
	EnvPerccliUtility = "IRONLIB_UTIL_PERCCLI"
	// TODO(jwb) These should probably be moved out of mvcli.go...
	// BitsUint8         = 8
	// BitsInt64         = 64
)

var (
	validPercRaidModes = []string{"0", "1", "5", "6", "10", "50", "60"}
	// TODO(jwb) These should probably be moved out of mvcli.go...
	// validInfoTypes  = []string{"hba", "pd", "vd"}
	// validBlockSizes = []uint{16, 32, 64, 128}
	// validInitModes  = []string{"quick", "none", "intelligent"}
)

// Perccli is a perccli command executor object
type Perccli struct {
	Executor Executor
}

type PerccliControllers struct {
	Controllers []struct {
		CommandStatus struct {
			CLIVersion      string `json:"CLI Version"`
			OperatingSystem string `json:"Operating system"`
			Controller      int    `json:"Controller"`
			Status          string `json:"Status"`
			Description     string `json:"Description"`
		} `json:"Command Status"`
		ResponseData struct {
			Basics struct {
				Controller                int    `json:"Controller"`
				Model                     string `json:"Model"`
				SerialNumber              string `json:"Serial Number"`
				CurrentControllerDateTime string `json:"Current Controller Date/Time"`
				CurrentSystemDateTime     string `json:"Current System Date/time"`
				SASAddress                string `json:"SAS Address"`
				PCIAddress                string `json:"PCI Address"`
				MfgDate                   string `json:"Mfg Date"`
				ReworkDate                string `json:"Rework Date"`
				RevisionNo                string `json:"Revision No"`
			} `json:"Basics"`
			Version struct {
				FirmwarePackageBuild string `json:"Firmware Package Build"`
				FirmwareVersion      string `json:"Firmware Version"`
				SBRVersion           string `json:"SBR Version"`
				BootBlockVersion     string `json:"Boot Block Version"`
				BiosVersion          string `json:"Bios Version"`
				HIIVersion           string `json:"HII Version"`
				NVDATAVersion        string `json:"NVDATA Version"`
				DriverName           string `json:"Driver Name"`
				DriverVersion        string `json:"Driver Version"`
			} `json:"Version"`
			Bus struct {
				VendorID        int    `json:"Vendor Id"`
				DeviceID        int    `json:"Device Id"`
				SubVendorID     int    `json:"SubVendor Id"`
				SubDeviceID     int    `json:"SubDevice Id"`
				HostInterface   string `json:"Host Interface"`
				DeviceInterface string `json:"Device Interface"`
				BusNumber       int    `json:"Bus Number"`
				DeviceNumber    int    `json:"Device Number"`
				FunctionNumber  int    `json:"Function Number"`
			} `json:"Bus"`
			PendingImagesInFlash struct {
				ImageName string `json:"Image name"`
			} `json:"Pending Images in Flash"`
			Status struct {
				ControllerStatus                                    string `json:"Controller Status"`
				MemoryCorrectableErrors                             int    `json:"Memory Correctable Errors"`
				MemoryUncorrectableErrors                           int    `json:"Memory Uncorrectable Errors"`
				ECCBucketCount                                      int    `json:"ECC Bucket Count"`
				AnyOfflineVDCachePreserved                          string `json:"Any Offline VD Cache Preserved"`
				BBUStatus                                           string `json:"BBU Status"`
				PDFirmwareDownloadInProgress                        string `json:"PD Firmware Download in progress"`
				SupportPDFirmwareDownload                           string `json:"Support PD Firmware Download"`
				LockKeyAssigned                                     string `json:"Lock Key Assigned"`
				FailedToGetLockKeyOnBootup                          string `json:"Failed to get lock key on bootup"`
				LockKeyHasNotBeenBackedUp                           string `json:"Lock key has not been backed up"`
				BiosWasNotDetectedDuringBoot                        string `json:"Bios was not detected during boot"`
				ControllerMustBeRebootedToCompleteSecurityOperation string `json:"Controller must be rebooted to complete security operation"`
				ARollbackOperationIsInProgress                      string `json:"A rollback operation is in progress"`
				AtLeastOnePFKExistsInNVRAM                          string `json:"At least one PFK exists in NVRAM"`
				SSCPolicyIsWB                                       string `json:"SSC Policy is WB"`
				ControllerHasBootedIntoSafeMode                     string `json:"Controller has booted into safe mode"`
				ControllerShutdownRequired                          string `json:"Controller shutdown required"`
				CurrentPersonality                                  string `json:"Current Personality"`
			} `json:"Status"`
			SupportedAdapterOperations struct {
				RebuildRate                       string `json:"Rebuild Rate"`
				CCRate                            string `json:"CC Rate"`
				BGIRate                           string `json:"BGI Rate "`
				ReconstructionRate                string `json:"Reconstruction Rate"`
				PatrolReadRate                    string `json:"Patrol Read Rate"`
				AlarmControl                      string `json:"Alarm Control"`
				ClusterSupport                    string `json:"Cluster Support"`
				Bbu                               string `json:"BBU"`
				Spanning                          string `json:"Spanning"`
				DedicatedHotSpare                 string `json:"Dedicated Hot Spare"`
				RevertibleHotSpares               string `json:"Revertible Hot Spares"`
				ForeignConfigImport               string `json:"Foreign Config Import"`
				SelfDiagnostic                    string `json:"Self Diagnostic"`
				AllowMixedRedundancyOnArray       string `json:"Allow Mixed Redundancy on Array"`
				GlobalHotSpares                   string `json:"Global Hot Spares"`
				DenySCSIPassthrough               string `json:"Deny SCSI Passthrough"`
				DenySMPPassthrough                string `json:"Deny SMP Passthrough"`
				DenySTPPassthrough                string `json:"Deny STP Passthrough"`
				SupportMoreThan8Phys              string `json:"Support more than 8 Phys"`
				FWAndEventTimeInGMT               string `json:"FW and Event Time in GMT"`
				SupportEnhancedForeignImport      string `json:"Support Enhanced Foreign Import"`
				SupportEnclosureEnumeration       string `json:"Support Enclosure Enumeration"`
				SupportAllowedOperations          string `json:"Support Allowed Operations"`
				AbortCCOnError                    string `json:"Abort CC on Error"`
				SupportMultipath                  string `json:"Support Multipath"`
				SupportOddEvenDriveCountInRAID1E  string `json:"Support Odd & Even Drive count in RAID1E"`
				SupportSecurity                   string `json:"Support Security"`
				SupportConfigPageModel            string `json:"Support Config Page Model"`
				SupportTheOCEWithoutAddingDrives  string `json:"Support the OCE without adding drives"`
				SupportEKM                        string `json:"Support EKM"`
				SnapshotEnabled                   string `json:"Snapshot Enabled"`
				SupportPFK                        string `json:"Support PFK"`
				SupportPI                         string `json:"Support PI"`
				SupportLdBBMInfo                  string `json:"Support Ld BBM Info"`
				SupportShieldState                string `json:"Support Shield State"`
				BlockSSDWriteDiskCacheChange      string `json:"Block SSD Write Disk Cache Change"`
				SupportSuspendResumeBGOps         string `json:"Support Suspend Resume BG ops"`
				SupportEmergencySpares            string `json:"Support Emergency Spares"`
				SupportSetLinkSpeed               string `json:"Support Set Link Speed"`
				SupportBootTimePFKChange          string `json:"Support Boot Time PFK Change"`
				SupportSystemPD                   string `json:"Support SystemPD"`
				DisableOnlinePFKChange            string `json:"Disable Online PFK Change"`
				SupportPerfTuning                 string `json:"Support Perf Tuning"`
				SupportSSDPatrolRead              string `json:"Support SSD PatrolRead"`
				RealTimeScheduler                 string `json:"Real Time Scheduler"`
				SupportResetNow                   string `json:"Support Reset Now"`
				SupportEmulatedDrives             string `json:"Support Emulated Drives"`
				HeadlessMode                      string `json:"Headless Mode"`
				DedicatedHotSparesLimited         string `json:"Dedicated HotSpares Limited"`
				PointInTimeProgress               string `json:"Point In Time Progress"`
				ExtendedLD                        string `json:"Extended LD"`
				SupportUnevenSpan                 string `json:"Support Uneven span "`
				SupportConfigAutoBalance          string `json:"Support Config Auto Balance"`
				SupportMaintenanceMode            string `json:"Support Maintenance Mode"`
				SupportDiagnosticResults          string `json:"Support Diagnostic results"`
				SupportExtEnclosure               string `json:"Support Ext Enclosure"`
				SupportSesmonitoring              string `json:"Support Sesmonitoring"`
				SupportSecurityonJBOD             string `json:"Support SecurityonJBOD"`
				SupportForceFlash                 string `json:"Support ForceFlash"`
				SupportDisableImmediateIO         string `json:"Support DisableImmediateIO"`
				SupportLargeIOSupport             string `json:"Support LargeIOSupport"`
				SupportDrvActivityLEDSetting      string `json:"Support DrvActivityLEDSetting"`
				SupportFlushWriteVerify           string `json:"Support FlushWriteVerify"`
				SupportCPLDUpdate                 string `json:"Support CPLDUpdate"`
				SupportForceTo512E                string `json:"Support ForceTo512e"`
				SupportDiscardCacheDuringLDDelete string `json:"Support discardCacheDuringLDDelete"`
				SupportJBODWriteCache             string `json:"Support JBOD Write cache"`
				SupportLargeQDSupport             string `json:"Support Large QD Support"`
				SupportCtrlInfoExtended           string `json:"Support Ctrl Info Extended"`
				SupportIButtonLess                string `json:"Support IButton less"`
				SupportAESEncryptionAlgorithm     string `json:"Support AES Encryption Algorithm"`
				SupportEncryptedMFC               string `json:"Support Encrypted MFC"`
				SupportSnapdump                   string `json:"Support Snapdump"`
				SupportForcePersonalityChange     string `json:"Support Force Personality Change"`
				SupportDualFwImage                string `json:"Support Dual Fw Image"`
				SupportPSOCUpdate                 string `json:"Support PSOC Update"`
				SupportSecureBoot                 string `json:"Support Secure Boot"`
				SupportClearSnapdump              string `json:"Support Clear Snapdump"`
				SupportDebugQueue                 string `json:"Support Debug Queue"`
				SupportLeastLatencyMode           string `json:"Support Least Latency Mode"`
				SupportOnDemandSnapdump           string `json:"Support OnDemand Snapdump"`
				SupportFWTriggeredSnapdump        string `json:"Support FW Triggered Snapdump"`
			} `json:"Supported Adapter Operations"`
			SupportedPDOperations struct {
				ForceOnline             string `json:"Force Online"`
				ForceOffline            string `json:"Force Offline"`
				ForceRebuild            string `json:"Force Rebuild"`
				DenyForceFailed         string `json:"Deny Force Failed"`
				DenyForceGoodBad        string `json:"Deny Force Good/Bad"`
				DenyMissingReplace      string `json:"Deny Missing Replace"`
				DenyClear               string `json:"Deny Clear"`
				DenyLocate              string `json:"Deny Locate"`
				SupportPowerState       string `json:"Support Power State"`
				SetPowerStateForCfg     string `json:"Set Power State For Cfg"`
				SupportT10PowerState    string `json:"Support T10 Power State"`
				SupportTemperature      string `json:"Support Temperature"`
				Ncq                     string `json:"NCQ"`
				SupportMaxRateSATA      string `json:"Support Max Rate SATA"`
				SupportDegradedMedia    string `json:"Support Degraded Media"`
				SupportParallelFWUpdate string `json:"Support Parallel FW Update"`
				SupportDriveCryptoErase string `json:"Support Drive Crypto Erase"`
				SupportSSDWearGauge     string `json:"Support SSD Wear Gauge"`
			} `json:"Supported PD Operations"`
			SupportedVDOperations struct {
				ReadPolicy                          string `json:"Read Policy"`
				WritePolicy                         string `json:"Write Policy"`
				IOPolicy                            string `json:"IO Policy"`
				AccessPolicy                        string `json:"Access Policy"`
				DiskCachePolicy                     string `json:"Disk Cache Policy"`
				Reconstruction                      string `json:"Reconstruction"`
				DenyLocate                          string `json:"Deny Locate"`
				DenyCC                              string `json:"Deny CC"`
				AllowCtrlEncryption                 string `json:"Allow Ctrl Encryption"`
				EnableLDBBM                         string `json:"Enable LDBBM"`
				SupportFastPath                     string `json:"Support FastPath"`
				PerformanceMetrics                  string `json:"Performance Metrics"`
				PowerSavings                        string `json:"Power Savings"`
				SupportPowersaveMaxWithCache        string `json:"Support Powersave Max With Cache"`
				SupportBreakmirror                  string `json:"Support Breakmirror"`
				SupportSSCWriteBack                 string `json:"Support SSC WriteBack"`
				SupportSSCAssociation               string `json:"Support SSC Association"`
				SupportVDHide                       string `json:"Support VD Hide"`
				SupportVDCachebypass                string `json:"Support VD Cachebypass"`
				SupportVDDiscardCacheDuringLDDelete string `json:"Support VD discardCacheDuringLDDelete"`
				SupportVDScsiUnmap                  string `json:"Support VD Scsi Unmap"`
			} `json:"Supported VD Operations"`
			HwCfg struct {
				ChipRevision                   string `json:"ChipRevision"`
				BatteryFRU                     string `json:"BatteryFRU"`
				FrontEndPortCount              int    `json:"Front End Port Count"`
				BackendPortCount               int    `json:"Backend Port Count"`
				Bbu                            string `json:"BBU"`
				Alarm                          string `json:"Alarm"`
				SerialDebugger                 string `json:"Serial Debugger"`
				NVRAMSize                      string `json:"NVRAM Size"`
				FlashSize                      string `json:"Flash Size"`
				OnBoardMemorySize              string `json:"On Board Memory Size"`
				CacheVaultFlashSize            string `json:"CacheVault Flash Size"`
				Tpm                            string `json:"TPM"`
				UpgradeKey                     string `json:"Upgrade Key"`
				OnBoardExpander                string `json:"On Board Expander"`
				TemperatureSensorForROC        string `json:"Temperature Sensor for ROC"`
				TemperatureSensorForController string `json:"Temperature Sensor for Controller"`
				UpgradableCPLD                 string `json:"Upgradable CPLD"`
				UpgradablePSOC                 string `json:"Upgradable PSOC"`
				CurrentSizeOfCacheCadeGB       int    `json:"Current Size of CacheCade (GB)"`
				CurrentSizeOfFWCacheMB         int    `json:"Current Size of FW Cache (MB)"`
				ROCTemperatureDegreeCelsius    int    `json:"ROC temperature(Degree Celsius)"`
				CtrlTemperatureDegreeCelsius   int    `json:"Ctrl temperature(Degree Celsius)"`
			} `json:"HwCfg"`
			Policies struct {
				PoliciesTable []struct {
					Policy  string `json:"Policy"`
					Current string `json:"Current"`
					Default string `json:"Default"`
				} `json:"Policies Table"`
				FlushTimeDefault             string `json:"Flush Time(Default)"`
				DriveCoercionMode            string `json:"Drive Coercion Mode"`
				AutoRebuild                  string `json:"Auto Rebuild"`
				BatteryWarning               string `json:"Battery Warning"`
				ECCBucketSize                int    `json:"ECC Bucket Size"`
				ECCBucketLeakRateHrs         int    `json:"ECC Bucket Leak Rate (hrs)"`
				RestoreHotSpareOnInsertion   string `json:"Restore Hot Spare on Insertion"`
				ExposeEnclosureDevices       string `json:"Expose Enclosure Devices"`
				MaintainPDFailHistory        string `json:"Maintain PD Fail History"`
				ReorderHostRequests          string `json:"Reorder Host Requests"`
				AutoDetectBackPlane          string `json:"Auto detect BackPlane"`
				LoadBalanceMode              string `json:"Load Balance Mode"`
				SecurityKeyAssigned          string `json:"Security Key Assigned"`
				DisableOnlineControllerReset string `json:"Disable Online Controller Reset"`
				UseDriveActivityForLocate    string `json:"Use drive activity for locate"`
			} `json:"Policies"`
			Boot struct {
				BIOSEnumerateVDs                                  int    `json:"BIOS Enumerate VDs"`
				StopBIOSOnError                                   string `json:"Stop BIOS on Error"`
				DelayDuringPOST                                   int    `json:"Delay during POST"`
				SpinDownMode                                      string `json:"Spin Down Mode"`
				EnableCtrlR                                       string `json:"Enable Ctrl-R"`
				EnableWebBIOS                                     string `json:"Enable Web BIOS"`
				EnablePreBootCLI                                  string `json:"Enable PreBoot CLI"`
				EnableBIOS                                        string `json:"Enable BIOS"`
				MaxDrivesToSpinupAtOneTime                        int    `json:"Max Drives to Spinup at One Time"`
				MaximumNumberOfDirectAttachedDrivesToSpinUpIn1Min int    `json:"Maximum number of direct attached drives to spin up in 1 min"`
				DelayAmongSpinupGroupsSec                         int    `json:"Delay Among Spinup Groups (sec)"`
				AllowBootWithPreservedCache                       string `json:"Allow Boot with Preserved Cache"`
			} `json:"Boot"`
			HighAvailability struct {
				TopologyType     string `json:"Topology Type"`
				ClusterPermitted string `json:"Cluster Permitted"`
				ClusterActive    string `json:"Cluster Active"`
			} `json:"High Availability"`
			Defaults struct {
				PhyPolarity                   int    `json:"Phy Polarity"`
				PhyPolaritySplit              int    `json:"Phy PolaritySplit"`
				StripSize                     string `json:"Strip Size"`
				WritePolicy                   string `json:"Write Policy"`
				ReadPolicy                    string `json:"Read Policy"`
				CacheWhenBBUBad               string `json:"Cache When BBU Bad"`
				CachedIO                      string `json:"Cached IO"`
				VDPowerSavePolicy             string `json:"VD PowerSave Policy"`
				DefaultSpinDownTimeMins       int    `json:"Default spin down time (mins)"`
				CoercionMode                  string `json:"Coercion Mode"`
				ZCRConfig                     string `json:"ZCR Config"`
				MaxChainedEnclosures          int    `json:"Max Chained Enclosures"`
				DirectPDMapping               string `json:"Direct PD Mapping"`
				RestoreHotSpareOnInsertion    string `json:"Restore Hot Spare on Insertion"`
				ExposeEnclosureDevices        string `json:"Expose Enclosure Devices"`
				MaintainPDFailHistory         string `json:"Maintain PD Fail History"`
				ZeroBasedEnclosureEnumeration string `json:"Zero Based Enclosure Enumeration"`
				DisablePuncturing             string `json:"Disable Puncturing"`
				EnableLDBBM                   string `json:"EnableLDBBM"`
				DisableHII                    string `json:"DisableHII"`
				UnCertifiedHardDiskDrives     string `json:"Un-Certified Hard Disk Drives"`
				SMARTMode                     string `json:"SMART Mode"`
				EnableLEDHeader               string `json:"Enable LED Header"`
				LEDShowDriveActivity          string `json:"LED Show Drive Activity"`
				DirtyLEDShowsDriveActivity    string `json:"Dirty LED Shows Drive Activity"`
				EnableCrashDump               string `json:"EnableCrashDump"`
				DisableOnlineControllerReset  string `json:"Disable Online Controller Reset"`
				TreatSingleSpanR1EAsR10       string `json:"Treat Single span R1E as R10"`
				PowerSavingOption             string `json:"Power Saving option"`
				TTYLogInFlash                 string `json:"TTY Log In Flash"`
				AutoEnhancedImport            string `json:"Auto Enhanced Import"`
				BreakMirrorRAIDSupport        string `json:"BreakMirror RAID Support"`
				DisableJoinMirror             string `json:"Disable Join Mirror"`
				EnableShieldState             string `json:"Enable Shield State"`
				TimeTakenToDetectCME          string `json:"Time taken to detect CME"`
			} `json:"Defaults"`
			Capabilities struct {
				SupportedDrives                string `json:"Supported Drives"`
				RAIDLevelSupported             string `json:"RAID Level Supported"`
				EnableSystemPD                 string `json:"Enable SystemPD"`
				MixInEnclosure                 string `json:"Mix in Enclosure"`
				MixOfSASSATAOfHDDTypeInVD      string `json:"Mix of SAS/SATA of HDD type in VD"`
				MixOfSASSATAOfSSDTypeInVD      string `json:"Mix of SAS/SATA of SSD type in VD"`
				MixOfSSDHDDInVD                string `json:"Mix of SSD/HDD in VD"`
				SASDisable                     string `json:"SAS Disable"`
				MaxArmsPerVD                   int    `json:"Max Arms Per VD"`
				MaxSpansPerVD                  int    `json:"Max Spans Per VD"`
				MaxArrays                      int    `json:"Max Arrays"`
				MaxVDPerArray                  int    `json:"Max VD per array"`
				MaxNumberOfVDs                 int    `json:"Max Number of VDs"`
				MaxParallelCommands            int    `json:"Max Parallel Commands"`
				MaxSGECount                    int    `json:"Max SGE Count"`
				MaxDataTransferSize            string `json:"Max Data Transfer Size"`
				MaxStripsPerIO                 int    `json:"Max Strips PerIO"`
				MaxConfigurableCacheCadeSizeGB int    `json:"Max Configurable CacheCade Size(GB)"`
				MaxTransportableDGs            int    `json:"Max Transportable DGs"`
				EnableSnapdump                 string `json:"Enable Snapdump"`
				EnableSCSIUnmap                string `json:"Enable SCSI Unmap"`
				FDEDriveMixSupport             string `json:"FDE Drive Mix Support"`
				MinStripSize                   string `json:"Min Strip Size"`
				MaxStripSize                   string `json:"Max Strip Size"`
			} `json:"Capabilities"`
			ScheduledTasks struct {
				ConsistencyCheckReoccurrence string `json:"Consistency Check Reoccurrence"`
				NextConsistencyCheckLaunch   string `json:"Next Consistency check launch"`
				PatrolReadReoccurrence       string `json:"Patrol Read Reoccurrence"`
				NextPatrolReadLaunch         string `json:"Next Patrol Read launch"`
				BatteryLearnReoccurrence     string `json:"Battery learn Reoccurrence"`
				Oemid                        string `json:"OEMID"`
			} `json:"Scheduled Tasks"`
		} `json:"Response Data"`
	} `json:"Controllers"`
}

var (
	ErrParsePerccliOutput = errors.New("perccli output parse error")
)

// TODO(jwb) These should probably be moved out of mvcli.go...
// var (
// 	ErrInvalidInfoType      = errors.New("invalid info type")
// 	ErrInvalidRaidMode      = errors.New("invalid raid mode")
// 	ErrInvalidBlockSize     = errors.New("invalid block size")
// 	ErrInvalidInitMode      = errors.New("invalid init mode")
// 	ErrInvalidVirtualDiskID = errors.New("invalid virtual disk id")
// 	ErrDestroyVirtualDisk   = errors.New("failed to destroy virtual disk")
// 	ErrCreateVirtualDisk    = errors.New("failed to create virtual disk")
// )

// Return a new perccli executor
func NewPerccliCmd(trace bool) *Perccli {
	utility := "perccli"

	// lookup env var for util
	if eVar := os.Getenv(EnvPerccliUtility); eVar != "" {
		utility = eVar
	}

	e := NewExecutor(utility)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})

	if !trace {
		e.SetQuiet()
	}

	return &Perccli{Executor: e}
}

// Attributes implements the actions.UtilAttributeGetter interface
func (m *Perccli) Attributes() (utilName model.CollectorUtility, absolutePath string, err error) {
	// Call CheckExecutable first so that the Executable CmdPath is resolved.
	er := m.Executor.CheckExecutable()

	return "perccli", m.Executor.CmdPath(), er
}

// Return a Fake perccli executor for tests
func NewFakePerccli(r io.Reader) (*Perccli, error) {
	e := NewFakeExecutor("perccli")
	b := bytes.Buffer{}

	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	e.SetStdout(b.Bytes())

	return &Perccli{
		Executor: e,
	}, nil
}

func (m *Perccli) StorageControllers(ctx context.Context) ([]*common.StorageController, error) {
	devices, err := m.ShowAll(ctx)
	if err != nil {
		return nil, err
	}

	hbas := []*common.StorageController{}

	for _, d := range devices.Controllers {
		hba := &common.StorageController{
			Common: common.Common{
				Model:       d.ResponseData.Basics.Model,
				Vendor:      common.VendorFromString(d.ResponseData.Basics.Model),
				Description: d.ResponseData.Basics.Model,
				Serial:      d.ResponseData.Basics.SerialNumber,
				Metadata:    make(map[string]string),
				Firmware: &common.Firmware{
					Installed: d.ResponseData.Version.FirmwareVersion,
					Metadata:  map[string]string{
						// TODO(jwb) Include all the additional ResponseData.Version fields
					},
				},
			},

			SupportedRAIDTypes: d.ResponseData.Capabilities.RAIDLevelSupported,
		}

		hbas = append(hbas, hba)
	}

	return hbas, nil
}

func (m *Perccli) Drives(ctx context.Context) ([]*common.Drive, error) {
	devices, err := m.ShowAll(ctx)
	devices, err := m.Info(ctx, "pd")
	if err != nil {
		return nil, err
	}

	drives := []*common.Drive{}

	for _, d := range devices {
		drive := &common.Drive{
			Common: common.Common{
				Model:       d.Model,
				Vendor:      common.VendorFromString(d.Model),
				Description: d.Model,
				Serial:      d.Serial,
				Firmware:    &common.Firmware{Installed: d.Firmware},
				Metadata:    make(map[string]string),
			},

			BlockSizeBytes:           d.Size,
			CapacityBytes:            d.PDSize,
			Type:                     m.processDriveType(d.Type, d.SSDType),
			NegotiatedSpeedGbps:      d.CurrentSpeed,
			StorageControllerDriveID: d.ID,
		}

		drives = append(drives, drive)
	}

	return drives, nil
}

func (c *Perccli) ShowAll(ctx context.Context) (output *PerccliControllers, err error) {
	c.Executor.SetArgs([]string{"/call", "show", "all", "J"})

	result, err := c.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	if len(result.Stdout) == 0 {
		return nil, errors.Wrap(ErrNoCommandOutput, c.Executor.GetCmd())
	}

	err = json.Unmarshal(result.Stdout, &output)

	if err != nil {
		return nil, errors.Wrap(err, ErrParsePerccliOutput.Error())
	}

	return output, nil
}

func (m *Perccli) parsePerccliInfoOutput(infoType string, b []byte) []*PerccliDevice {
	switch infoType {
	case "hba":
		return m.parsePerccliInfoHbaOutput(b)
	case "pd":
		return m.parsePerccliInfoPdOutput(b)
	case "vd":
		return m.parsePerccliInfoVdOutput(b)
	}

	return nil
}

func (m *Perccli) parsePerccliInfoHbaOutput(b []byte) []*PerccliDevice {
	devices := []*PerccliDevice{}
	blocks := parseBytesForBlocks("Adapter ID", b)

	for _, block := range blocks {
		device := &PerccliDevice{
			Product:            block["Product"],
			SubProduct:         block["Sub Product"],
			FirmwareRom:        block["Rom version"],
			FirmwareBios:       block["BIOS version"],
			Firmware:           block["Firmware version"],
			FirmwareBootLoader: block["Boot loader version"],
			SupportedRAIDModes: block["Supported RAID mode"],
			Serial:             strings.Replace(strings.ToUpper(block["Product"]), "-", ":", 1),
		}
		devices = append(devices, device)
	}

	return devices
}

func stringToInt64(s string, b int) int64 {
	i, _ := strconv.ParseInt(s, 0, b)
	return i
}

func (m *Perccli) parsePerccliInfoPdOutput(b []byte) []*PerccliDevice {
	const oneK = 1000

	devices := []*PerccliDevice{}

	for _, block := range parseBytesForBlocks("Adapter", b) {
		device := &PerccliDevice{
			Model:        block["model"],
			Serial:       block["Serial"],
			Firmware:     block["Firmware version"],
			Type:         block["Type"],
			SSDType:      block["SSD Type"],
			Size:         stringToInt64(strings.TrimSuffix(block["Size"], " K"), BitsInt64) * oneK,
			PDSize:       stringToInt64(strings.TrimSuffix(block["PD valid size"], " K"), BitsInt64) * oneK,
			AdapterID:    int(stringToInt64(block["Adapter"], BitsUint8)),
			ID:           int(stringToInt64(block["PD ID"], BitsUint8)),
			CurrentSpeed: stringToInt64(strings.TrimSuffix(block["Current speed"], " Gb/s"), BitsInt64),
		}

		devices = append(devices, device)
	}

	return devices
}

func (m *Perccli) parsePerccliInfoVdOutput(b []byte) []*PerccliDevice {
	const oneM = 1000000

	devices := []*PerccliDevice{}

	for _, block := range parseBytesForBlocks("id:", b) {
		device := &PerccliDevice{
			ID:     int(stringToInt64(block["id"], BitsUint8)),
			Name:   block["name"],
			Status: block["status"],
			Size:   stringToInt64(strings.TrimSuffix(block["size"], " M"), BitsInt64) * oneM,
			Type:   block["RAID mode"],
		}

		devices = append(devices, device)
	}

	return devices
}

func (m *Perccli) CreateVirtualDisk(ctx context.Context, raidMode string, physicalDisks []uint, name string, blockSize uint) error {
	return m.Create(ctx, physicalDisks, raidMode, name, blockSize, false, "quick")
}

func (m *Perccli) DestroyVirtualDisk(ctx context.Context, virtualDiskID int) error {
	if vd := m.FindVdByID(ctx, virtualDiskID); vd == nil {
		return InvalidVirtualDiskIDError(virtualDiskID)
	}

	return m.Destroy(ctx, virtualDiskID)
}

func (m *Perccli) processDriveType(pdType, ssdType string) string {
	if pdType == "SATA PD" && ssdType == "SSD" {
		return common.SlugDriveTypeSATASSD
	}

	return "Unknown"
}

func parseBytesForBlocks(blockStart string, b []byte) []map[string]string {
	blocks := []map[string]string{}

	byteSlice := bytes.Split(b, []byte("\n"))
	for idx, sl := range byteSlice {
		s := string(sl)
		if strings.Contains(s, blockStart) {
			block := parseKeyValueBlock(byteSlice[idx:])
			if block != nil {
				blocks = append(blocks, block)
			}
		}
	}

	return blocks
}

func parseKeyValueBlock(bSlice [][]byte) map[string]string {
	kv := make(map[string]string)

	for _, line := range bSlice {
		// A blank line means we've reached the end of this record
		if len(line) == 0 {
			return kv
		}

		s := string(line)
		cols := 2
		parts := strings.Split(s, ":")

		// Skip if there's no value
		if len(parts) < cols {
			continue
		}

		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		kv[key] = value
	}

	return kv
}

func (m *Perccli) Create(ctx context.Context, physicalDiskIDs []uint, raidMode, name string, blockSize uint, cacheMode bool, initMode string) error {
	if !slices.Contains(validRaidModes, raidMode) {
		return InvalidRaidModeError(raidMode)
	}

	if !slices.Contains(validBlockSizes, blockSize) {
		return InvalidBlockSizeError(blockSize)
	}

	if !slices.Contains(validInitModes, initMode) {
		return InvalidInitModeError(initMode)
	}

	m.Executor.SetArgs([]string{"create",
		"-o", "vd",
		"-r", raidMode,
		"-d", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(physicalDiskIDs)), ","), "[]"),
		"-n", name,
		"-b", fmt.Sprintf("%d", blockSize),
	})

	result, err := m.Executor.ExecWithContext(ctx)
	if err != nil {
		return err
	}

	if len(result.Stdout) == 0 {
		return errors.Wrap(ErrNoCommandOutput, m.Executor.GetCmd())
	}

	// Possible errors:
	// Specified RAID mode is not supported.
	// Gigabyte rounding scheme is not supported

	if match, _ := regexp.MatchString(`^SG driver version \S+\n$`, string(result.Stdout)); !match {
		return CreateVirtualDiskError(result.Stdout)
	}

	return nil
}

func (m *Perccli) Destroy(ctx context.Context, virtualDiskID int) error {
	m.Executor.SetStdin(bytes.NewReader([]byte("y\n")))
	m.Executor.SetArgs([]string{"delete",
		"-o", "vd",
		"-i", fmt.Sprintf("%d", virtualDiskID),
	})

	result, err := m.Executor.ExecWithContext(ctx)
	if err != nil {
		return err
	}

	if len(result.Stdout) == 0 {
		return errors.Wrap(ErrNoCommandOutput, m.Executor.GetCmd())
	}

	// Possible errors:
	// Unable to get status of VD \S (error 59: Specified virtual disk doesn't exist).

	if match, _ := regexp.MatchString(`Delete VD \S successfully.`, string(result.Stdout)); !match {
		return DestroyVirtualDiskError(result.Stdout)
	}

	return nil
}

func (m *Perccli) FindVdByName(ctx context.Context, name string) *PerccliDevice {
	return m.FindVdBy(ctx, "Name", name)
}

func (m *Perccli) FindVdByID(ctx context.Context, virtualDiskID int) *PerccliDevice {
	return m.FindVdBy(ctx, "ID", virtualDiskID)
}

func (m *Perccli) FindVdBy(ctx context.Context, k string, v interface{}) *PerccliDevice {
	virtualDisks, err := m.VirtualDisks(ctx)

	if err != nil {
		return nil
	}

	for _, vd := range virtualDisks {
		switch lKey := strings.ToLower(k); lKey {
		case "id":
			if vd.ID == v {
				return vd
			}
		case "name":
			if vd.Name == v {
				return vd
			}
		}
	}

	return nil
}

func (m *Perccli) VirtualDisks(ctx context.Context) ([]*PerccliDevice, error) {
	return m.Info(ctx, "vd")
}
