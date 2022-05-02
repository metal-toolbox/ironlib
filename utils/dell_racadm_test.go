package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	biosConfigR6515 = "../fixtures/dell/r6515/bios.json"
	biosConfigC6320 = "../fixtures/dell/c6320/bios.xml"
)

// Fake Dell Racadm executor for tests
func NewFakeDellRacadm() *DellRacadm {
	return &DellRacadm{
		Executor:       NewFakeExecutor("racadm"),
		KeepConfigFile: true,
	}
}

func Test_GetBIOSConfiguration(t *testing.T) {
	expected := map[string]string{
		"amd_sev":                        "1",
		"boot_mode":                      "BIOS",
		"secure_boot":                    "Disabled",
		"smt":                            "Enabled",
		"sr_iov":                         "Enabled",
		"tpm":                            "Enabled",
		"raw:AcPwrRcvry":                 "Last",
		"raw:AcPwrRcvryDelay":            "Immediate",
		"raw:AcPwrRcvryUserDelay":        "60",
		"raw:ApbDis":                     "Disabled",
		"raw:BiosBootSeq":                "NIC.Slot.3-1-1, HardDisk.List.1-1",
		"raw:BootSeqRetry":               "Enabled",
		"raw:CECriticalSEL":              "Enabled",
		"raw:CcdCores":                   "All",
		"raw:CcxAsNumaDomain":            "Disabled",
		"raw:ConTermType":                "Vt100Vt220",
		"raw:CorrEccSmi":                 "Enabled",
		"raw:DellAutoDiscovery":          "PlatformDefault",
		"raw:DellWyseP25BIOSAccess":      "Enabled",
		"raw:DeterminismSlider":          "PowerDeterminism",
		"raw:DramRefreshDelay":           "Minimum",
		"raw:DynamicLinkWidthManagement": "Unforced",
		"raw:EfficiencyOptimizedMode":    "Disabled",
		"raw:EmbNic1Nic2":                "DisabledOs",
		"raw:EmbSata":                    "AhciMode",
		"raw:EmbVideo":                   "Enabled",
		"raw:ErrPrompt":                  "Enabled",
		"raw:ExtSerialConnector":         "Serial1",
		"raw:FailSafeBaud":               "115200",
		"raw:ForceInt10":                 "Disabled",
		"raw:GenericUsbBoot":             "Disabled",
		"raw:HddFailover":                "Enabled",
		"raw:HddPlaceholder":             "Disabled",
		"raw:HddSeq":                     "AHCI.Slot.2-1, NonRAID.Integrated.1-1, AHCI.Slot.2-1",
		"raw:IntegratedRaid":             "Enabled",
		"raw:InternalUsb":                "Enabled",
		"raw:L1StreamHwPrefetcher":       "Enabled",
		"raw:L2StreamHwPrefetcher":       "Enabled",
		"raw:MadtCoreEnumeration":        "Linear",
		"raw:MemFrequency":               "MaxPerf",
		"raw:MemOpMode":                  "OptimizerMode",
		"raw:MemPatrolScrub":             "Standard",
		"raw:MemRefreshRate":             "1x",
		"raw:MemTest":                    "Disabled",
		"raw:MemoryInterleaving":         "Auto",
		"raw:MmioLimit":                  "8TB",
		"raw:NumLock":                    "Enabled",
		"raw:NumaNodesPerSocket":         "1",
		"raw:NvmeMode":                   "NonRaid",
		"raw:OppSrefEn":                  "Disabled",
		"raw:OsWatchdogTimer":            "Disabled",
		"raw:PasswordStatus":             "Unlocked",
		"raw:PcieAspmL1":                 "Disabled",
		"raw:PcieEnhancedPreferredIo":    "Disabled",
		"raw:PciePreferredIoBus":         "Disabled",
		"raw:ProcCStates":                "Disabled",
		"raw:ProcCcds":                   "All",
		"raw:ProcPwrPerf":                "MaxPerf",
		"raw:ProcTurboMode":              "Enabled",
		"raw:ProcVirtualization":         "Enabled",
		"raw:ProcX2Apic":                 "Enabled",
		"raw:PwrButton":                  "Enabled",
		"raw:RedirAfterBoot":             "Enabled",
		"raw:RedundantOsBoot":            "Disabled",
		"raw:RedundantOsLocation":        "None",
		"raw:RedundantOsState":           "Visible",
		"raw:SataPortA":                  "Auto",
		"raw:SataPortB":                  "Auto",
		"raw:SataPortC":                  "Auto",
		"raw:SataPortD":                  "Auto",
		"raw:SecureBootMode":             "DeployedMode",
		"raw:SecureBootPolicy":           "Standard",
		"raw:SecurityFreezeLock":         "Enabled",
		"raw:SerialComm":                 "OnConRedirCom1",
		"raw:SerialPortAddress":          "Serial1Com1Serial2Com2",
		"raw:SetBootOrderEn":             "NIC.Slot.3-1-1,HardDisk.List.1-1",
		"raw:Slot1":                      "Enabled",
		"raw:Slot2":                      "Enabled",
		"raw:Slot2Bif":                   "x16",
		"raw:Slot3":                      "Enabled",
		"raw:Slot3Bif":                   "x16",
		"raw:SysPrepClean":               "None",
		"raw:SysProfile":                 "PerfOptimized",
		"raw:Tpm2Algorithm":              "SHA1",
		"raw:Tpm2Hierarchy":              "Enabled",
		"raw:TpmPpiBypassClear":          "Disabled",
		"raw:TpmPpiBypassProvision":      "Disabled",
		"raw:UefiVariableAccess":         "Standard",
		"raw:UsbManagedPort":             "Enabled",
		"raw:UsbPorts":                   "AllOn",
		"raw:WorkloadProfile":            "NotAvailable",
		"raw:WriteCache":                 "Disabled",
		"raw:WriteDataCrc":               "Disabled",
	}

	d := NewFakeDellRacadm()
	d.BIOSCfgTmpFile = biosConfigR6515

	cfg, err := d.GetBIOSConfiguration(context.TODO(), "")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, cfg)
}

func Test_RacadmBIOSConfigJSON(t *testing.T) {
	expected := map[string]string{
		"AcPwrRcvry":                 "Last",
		"AcPwrRcvryDelay":            "Immediate",
		"AcPwrRcvryUserDelay":        "60",
		"ApbDis":                     "Disabled",
		"BiosBootSeq":                "NIC.Slot.3-1-1, HardDisk.List.1-1",
		"BootMode":                   "Bios",
		"BootSeqRetry":               "Enabled",
		"CECriticalSEL":              "Enabled",
		"CcdCores":                   "All",
		"CcxAsNumaDomain":            "Disabled",
		"ConTermType":                "Vt100Vt220",
		"CorrEccSmi":                 "Enabled",
		"CpuMinSevAsid":              "1",
		"DellAutoDiscovery":          "PlatformDefault",
		"DellWyseP25BIOSAccess":      "Enabled",
		"DeterminismSlider":          "PowerDeterminism",
		"DramRefreshDelay":           "Minimum",
		"DynamicLinkWidthManagement": "Unforced",
		"EfficiencyOptimizedMode":    "Disabled",
		"EmbNic1Nic2":                "DisabledOs",
		"EmbSata":                    "AhciMode",
		"EmbVideo":                   "Enabled",
		"ErrPrompt":                  "Enabled",
		"ExtSerialConnector":         "Serial1",
		"FailSafeBaud":               "115200",
		"ForceInt10":                 "Disabled",
		"GenericUsbBoot":             "Disabled",
		"HddFailover":                "Enabled",
		"HddPlaceholder":             "Disabled",
		"HddSeq":                     "AHCI.Slot.2-1, NonRAID.Integrated.1-1, AHCI.Slot.2-1",
		"IntegratedRaid":             "Enabled",
		"InternalUsb":                "On",
		"L1StreamHwPrefetcher":       "Enabled",
		"L2StreamHwPrefetcher":       "Enabled",
		"LogicalProc":                "Enabled",
		"MadtCoreEnumeration":        "Linear",
		"MemFrequency":               "MaxPerf",
		"MemOpMode":                  "OptimizerMode",
		"MemPatrolScrub":             "Standard",
		"MemRefreshRate":             "1x",
		"MemTest":                    "Disabled",
		"MemoryInterleaving":         "Auto",
		"MmioLimit":                  "8TB",
		"NewSetupPassword":           "***********",
		"NewSysPassword":             "*************",
		"NumLock":                    "On",
		"NumaNodesPerSocket":         "1",
		"NvmeMode":                   "NonRaid",
		"OldSetupPassword":           "**************",
		"OldSysPassword":             "************",
		"OppSrefEn":                  "Disabled",
		"OsWatchdogTimer":            "Disabled",
		"PasswordStatus":             "Unlocked",
		"PcieAspmL1":                 "Disabled",
		"PcieEnhancedPreferredIo":    "Disabled",
		"PciePreferredIoBus":         "Disabled",
		"ProcCStates":                "Disabled",
		"ProcCcds":                   "All",
		"ProcPwrPerf":                "MaxPerf",
		"ProcTurboMode":              "Enabled",
		"ProcVirtualization":         "Enabled",
		"ProcX2Apic":                 "Enabled",
		"PwrButton":                  "Enabled",
		"RedirAfterBoot":             "Enabled",
		"RedundantOsBoot":            "Disabled",
		"RedundantOsLocation":        "None",
		"RedundantOsState":           "Visible",
		"SataPortA":                  "Auto",
		"SataPortB":                  "Auto",
		"SataPortC":                  "Auto",
		"SataPortD":                  "Auto",
		"SecureBoot":                 "Disabled",
		"SecureBootMode":             "DeployedMode",
		"SecureBootPolicy":           "Standard",
		"SecurityFreezeLock":         "Enabled",
		"SerialComm":                 "OnConRedirCom1",
		"SerialPortAddress":          "Serial1Com1Serial2Com2",
		"SetBootOrderEn":             "NIC.Slot.3-1-1,HardDisk.List.1-1",
		"Slot1":                      "Enabled",
		"Slot2":                      "Enabled",
		"Slot2Bif":                   "x16",
		"Slot3":                      "Enabled",
		"Slot3Bif":                   "x16",
		"SriovGlobalEnable":          "Enabled",
		"SysPrepClean":               "None",
		"SysProfile":                 "PerfOptimized",
		"Tpm2Algorithm":              "SHA1",
		"Tpm2Hierarchy":              "Enabled",
		"TpmPpiBypassClear":          "Disabled",
		"TpmPpiBypassProvision":      "Disabled",
		"TpmSecurity":                "On",
		"UefiVariableAccess":         "Standard",
		"UsbManagedPort":             "On",
		"UsbPorts":                   "AllOn",
		"WorkloadProfile":            "NotAvailable",
		"WriteCache":                 "Disabled",
		"WriteDataCrc":               "Disabled",
	}

	// setup fake racadm, pass the bios config file
	r := NewFakeRacadm(biosConfigR6515)

	c, err := r.racadmBIOSConfigJSON(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, c)
}

func Test_RacadmBIOSConfigXML(t *testing.T) {
	expected := map[string]string{"AcPwrRcvry": "Last",
		"AcPwrRcvryDelay":         "Immediate",
		"BootMode":                "Bios",
		"BootSeqRetry":            "Enabled",
		"ConTermType":             "Vt100Vt220",
		"CorrEccSmi":              "Enabled",
		"DcuIpPrefetcher":         "Enabled",
		"DcuStreamerPrefetcher":   "Enabled",
		"DynamicCoreAllocation":   "Disabled",
		"EmbNic1Nic2":             "Enabled",
		"EmbSata":                 "AhciMode",
		"EmbVideo":                "Enabled",
		"ErrPrompt":               "Enabled",
		"ExtSerialConnector":      "Serial1",
		"FailSafeBaud":            "115200",
		"ForceInt10":              "Disabled",
		"GlobalSlotDriverDisable": "Disabled",
		"HddFailover":             "Enabled",
		"IntelTxt":                "Off",
		"IoNonPostedPrefetch":     "Enabled",
		"IoatEngine":              "Disabled",
		"LogicalProc":             "Enabled",
		"MemOpMode":               "OptimizerMode",
		"MemTest":                 "Disabled",
		"MmioAbove4Gb":            "Enabled",
		"NodeInterleave":          "Disabled",
		"NumLock":                 "On",
		"OsWatchdogTimer":         "Disabled",
		"PasswordStatus":          "Unlocked",
		"PowerSaver":              "Disabled",
		"ProcAdjCacheLine":        "Enabled",
		"ProcAts":                 "Enabled",
		"ProcConfigTdp":           "Nominal",
		"ProcCores":               "All",
		"ProcExecuteDisable":      "Enabled",
		"ProcHwPrefetcher":        "Enabled",
		"ProcVirtualization":      "Enabled",
		"ProcX2Apic":              "Disabled",
		"PwrButton":               "Enabled",
		"QpiSpeed":                "MaxDataRate",
		"RedirAfterBoot":          "Enabled",
		"RtidSetting":             "Disabled",
		"SecurityFreezeLock":      "Enabled",
		"SerialComm":              "OnConRedirAuto",
		"SerialPortAddress":       "Serial1Com1Serial2Com2",
		"Slot1":                   "Enabled",
		"Slot2":                   "Enabled",
		"SnoopMode":               "OpportunisticSnoopBroadcast",
		"SriovGlobalEnable":       "Disabled",
		"SysProfile":              "PerfPerWattOptimizedDapc",
		"Tpm2Hierarchy":           "Enabled",
		"TpmPpiBypassClear":       "Disabled",
		"TpmPpiBypassProvision":   "Disabled",
		"TpmSecurity":             "On",
		"UefiVariableAccess":      "Standard",
		"Usb3Setting":             "Enabled",
		"UsbPorts":                "AllOn",
		"WorkloadProfile":         "NotAvailable",
		"WriteCache":              "Disabled",
	}

	// setup fake racadm, pass in the read bios config
	r := NewFakeRacadm(biosConfigC6320)

	c, err := r.racadmBIOSConfigXML(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, c)
}
