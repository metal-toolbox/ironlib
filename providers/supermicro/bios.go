package supermicro

import (
	"context"

	"github.com/metal-toolbox/ironlib/utils"
)

// SetBIOSConfiguration sets bios configuration settings
func (s *supermicro) SetBIOSConfiguration(context.Context, map[string]string) error {
	return nil
}

// GetBIOSConfiguration returns bios configuration settings
func (s *supermicro) GetBIOSConfiguration(ctx context.Context) (map[string]string, error) {
	sum := utils.NewSupermicroSUM(s.trace)

	return sum.GetBIOSConfiguration(ctx, "")
}
