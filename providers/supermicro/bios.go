package supermicro

import (
	"context"

	"github.com/packethost/ironlib/config"
	"github.com/packethost/ironlib/utils"
	"github.com/sirupsen/logrus"
)

// SetBIOSConfiguration sets bios configuration settings
func (s *supermicro) SetBIOSConfiguration(ctx context.Context, cfg *config.BIOSConfiguration) error {
	return nil
}

// GetBIOSConfiguration returns bios configuration settings
func (s *supermicro) GetBIOSConfiguration(ctx context.Context) (*config.BIOSConfiguration, error) {
	var trace bool
	if s.logger.Level >= logrus.TraceLevel {
		trace = true
	}

	sum := utils.NewSupermicroSUM(trace)

	return sum.GetBIOSConfiguration(ctx)
}
