package dell

import (
	"context"

	"github.com/packethost/ironlib/config"
)

func (d *Dell) SetBIOSConfiguration(ctx context.Context, config *config.BIOSConfiguration) error {
	return nil
}

func (d *Dell) GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error) {
	return &config.BIOSConfiguration{}, nil
}
