package utils

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/packethost/ironlib/model"
)

const nvmecli = "/usr/sbin/nvme"

type Nvme struct {
	Executor Executor
}

type nvmeDeviceAttributes struct {
	Namespace    int    `json:"Namespace"`
	DevicePath   string `json:"DevicePath"`
	Firmware     string `json:"Firmware"`
	Index        int    `json:"Index"`
	ModelNumber  string `json:"ModelNumber"`
	ProductName  string `json:"ProductName"`
	SerialNumber string `json:"SerialNumber"`
}

type nvmeList struct {
	Devices []*nvmeDeviceAttributes `json:"Devices"`
}

// Return a new nvme executor
func NewNvmeCmd(trace bool) Collector {

	e := NewExecutor(nvmecli)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	if !trace {
		e.SetQuiet()
	}

	return &Nvme{Executor: e}
}

// Executes nvme list, parses the output and returns a slice of model.Component's
func (n *Nvme) Components() ([]*model.Component, error) {

	inv := make([]*model.Component, 0)

	out, err := n.List()
	if err != nil {
		return nil, err
	}

	list := &nvmeList{Devices: []*nvmeDeviceAttributes{}}
	err = json.Unmarshal(out, list)
	if err != nil {
		return nil, err
	}

	for _, d := range list.Devices {
		dModel := d.ModelNumber

		var vendor string
		modelTokens := strings.Split(d.ModelNumber, " ")

		if len(modelTokens) > 1 {
			vendor = modelTokens[1]
		}

		item := &model.Component{
			Serial:            d.SerialNumber,
			Vendor:            vendor,
			Model:             dModel,
			FirmwareInstalled: d.Firmware,
			Slug:              "NVME drive",
			Name:              "NVME drive",
		}

		inv = append(inv, item)
	}

	return inv, nil
}

func (n *Nvme) List() ([]byte, error) {
	// nvme list --output-format=json
	n.Executor.SetArgs([]string{"list", "--output-format=json"})

	result, err := n.Executor.ExecWithContext(context.Background())
	if err != nil {
		return nil, err
	}

	return result.Stdout, nil
}
