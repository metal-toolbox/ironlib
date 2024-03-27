package supermicro

import (
	"context"

	"github.com/metal-toolbox/ironlib/utils"
	"github.com/sirupsen/logrus"
)

// SetBIOSConfiguration sets bios configuration settings
func (s *supermicro) SetBIOSConfiguration(_ context.Context, _ map[string]string) error {
	return nil
}

func (a *supermicro) SetBIOSConfigurationFromFile(ctx context.Context, cfg string) error {
	return nil
}

// GetBIOSConfiguration returns bios configuration settings
func (s *supermicro) GetBIOSConfiguration(ctx context.Context) (map[string]string, error) {
	var trace bool
	if s.logger.Level >= logrus.TraceLevel {
		trace = true
	}

	sum := utils.NewSupermicroSUM(trace)

	return sum.GetBIOSConfiguration(ctx, "")
}
