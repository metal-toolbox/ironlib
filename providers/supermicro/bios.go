package supermicro

import (
	"context"

	"github.com/packethost/ironlib/config"
)

func (s *Supermicro) SetBIOSConfiguration(ctx context.Context, cfg *config.BIOSConfiguration) error {
	return nil
}

func (s *Supermicro) GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error) {
	return &config.BIOSConfiguration{}, nil
}
