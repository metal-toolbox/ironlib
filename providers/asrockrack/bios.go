package asrockrack

import (
	"context"

	"github.com/packethost/ironlib/model"
	"github.com/packethost/ironlib/utils"
)

func (a *asrockrack) SetBIOSConfiguration(ctx context.Context, cfg map[string]string) error {
	return nil
}

func (a *asrockrack) GetBIOSConfiguration(ctx context.Context) (map[string]string, error) {
	asrr := utils.NewAsrrBioscontrol(false)

	return asrr.GetBIOSConfiguration(ctx, model.FormatProductName(a.GetModel()))
}
