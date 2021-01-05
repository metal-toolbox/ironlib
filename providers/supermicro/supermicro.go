package supermicro

import (
	"context"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
)

//import (
//	"encoding/json"
//	"fmt"
//
//	"github.com/packethost/fup/apiclient"
//	"github.com/packethost/fup/internal/device/utils"
//	"github.com/packethost/fup/internal/model"
//	"github.com/sirupsen/logrus"
//)
//
//// For the future Joel, it may be worth parsing the inventory catalog
//// /usr/libexec/dell_dup/inv.xml
//
type Supermicro struct {
	requiresReboot   bool // set when the device requires a reboot after update
	UpdatesAvailable int
	Vendor           string
	Model            string
	Serial           string
	Components       []*model.Component
	Collectors       map[string]utils.Collector
	Logger           *logrus.Logger
	Dmidecode        *utils.Dmidecode
}

func (s *Supermicro) GetInventory(ctx context.Context) (*model.Device, error) {
	return nil, nil
}

//
//type Component struct {
//	Serial   string
//	Type     string
//	Model    string
//	Firmware string
//}
//
//func New(dmi *utils.Dmidecode, vendor, model string, api *apiclient.Client, l *logrus.Logger) *Supermicro {
//
//	var trace bool
//
//	if l.GetLevel().String() == "trace" {
//		trace = true
//	}
//
//	collectors := map[string]utils.Collector{
//		"ipmi":     utils.NewIpmicfgCmd(trace),
//		"smartctl": utils.NewSmartctlCmd(trace),
//		"storecli": utils.NewStoreCLICmd(trace),
//		"mlxup":    utils.NewMlxupCmd(trace),
//	}
//
//	return &Supermicro{
//		api:        api,
//		dmi:        dmi,
//		collectors: collectors,
//		logger:     l,
//		Model:      model,
//		Vendor:     "supermicro",
//	}
//}
//
//func (s *Supermicro) ValidateUpdatesInstalled() (err error) {
//
//	err = s.CollectFirmwareInventory()
//	if err != nil {
//		return err
//	}
//
//	if s.UpdatesAvailable > 0 {
//		s.logger.WithField("updates", s.UpdatesAvailable).Fatal("All Updates for device were not applied, or newer updates available ", model.ClientFirmareInventoryLog)
//	}
//
//	return nil
//}
//
//// Collect firmware inventory and push to fup API
//func (s *Supermicro) CollectFirmwareInventory() error {
//
//	inventory := make([]*model.FirmwareInventory, 0)
//	// collect current component firmware versions
//	s.logger.Info("Identifying component firmware versions...")
//	for cmd, collector := range s.collectors {
//		inv, err := collector.DeviceAttributes()
//		if err != nil {
//			s.logger.WithFields(logrus.Fields{"cmd": cmd, "err": err}).Error("inventory collector error")
//		}
//
//		if len(inv) == 0 {
//			s.logger.WithFields(logrus.Fields{"cmd": cmd}).Trace("inventory collector returned no items")
//			continue
//		}
//
//		inventory = append(inventory, inv...)
//	}
//
//	s.FirmwareInventory = inventory
//	err := s.api.InsertFirmwareInventory(inventory)
//	if err != nil {
//		return err
//	}
//	s.logger.Info("Component firmware inventory sent to fup")
//
//	return err
//}
//
//func (s *Supermicro) ListFirmwareInventory() (err error) {
//
//	inventory := make([]*model.FirmwareInventory, 0)
//	// collect current component firmware versions
//	s.logger.Info("Identifying component firmware versions...")
//	for cmd, collector := range s.collectors {
//		inv, err := collector.DeviceAttributes()
//		if err != nil {
//			s.logger.WithFields(logrus.Fields{"cmd": cmd, "err": err}).Error("inventory collector error")
//		}
//
//		if len(inv) == 0 {
//			s.logger.WithFields(logrus.Fields{"cmd": cmd}).Trace("inventory collector returned no items")
//			continue
//		}
//		inventory = append(inventory, inv...)
//	}
//
//	b, _ := json.MarshalIndent(inventory, "", " ")
//	fmt.Print(string(b))
//	return nil
//}
//
//func (s *Supermicro) ApplyUpdate() (err error) {
//	return err
//}
//
//func (s *Supermicro) BIOS() (err error) {
//	return err
//}
//
//func (s *Supermicro) BMC() (err error) {
//	return err
//}
//
//func (s *Supermicro) SetDeviceID(id string) {
//	s.DeviceID = id
//}
//
//func (s *Supermicro) SetLocation(l string) {
//	s.Location = l
//}
//
//func (s *Supermicro) SetFirmwareManifest(m *model.FirmwareConfig) {
//	s.FirmwareManifest = m
//}
//
//func (s *Supermicro) SetDeployManifest(m *model.Deployment) {
//	s.DeployManifest = m
//}
//
//func (s *Supermicro) GetDeviceID() string {
//	return s.DeviceID
//}
//
//func (s *Supermicro) GetLocation() string {
//	return s.Location
//}
//
//func (s *Supermicro) GetModel() string {
//	return s.Model
//}
//
//func (s *Supermicro) GetVendor() string {
//	return s.Vendor
//}
//
//func (s *Supermicro) RequiresReboot() bool {
//	return s.requiresReboot
//}
//
