package asrockrack

import (
	"context"

	"github.com/packethost/ironlib/config"
)

func (a *ASRockRack) SetBIOSConfiguration(ctx context.Context, cfg *config.BIOSConfiguration) error {
	return nil
}

func (a *ASRockRack) GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error) {
	return &config.BIOSConfiguration{}, nil
}
