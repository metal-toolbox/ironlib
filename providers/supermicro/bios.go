package supermicro

import (
	"context"

	"github.com/packethost/ironlib/config"
)

func (s *supermicro) SetBIOSConfiguration(ctx context.Context, cfg *config.BIOSConfiguration) error {
	return nil
}

func (s *supermicro) GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error) {
	return &config.BIOSConfiguration{}, nil
}
