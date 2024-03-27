package generic

import (
	"context"
)

func (g *Generic) SetBIOSConfiguration(_ context.Context, _ map[string]string) error {
	return nil
}

func (a *Generic) SetBIOSConfigurationFromFile(ctx context.Context, cfg string) error {
	return nil
}

func (g *Generic) GetBIOSConfiguration(_ context.Context) (map[string]string, error) {
	return nil, nil
}
