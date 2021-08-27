package generic

import (
	"context"
)

func (g *Generic) SetBIOSConfiguration(ctx context.Context, cfg map[string]string) error {
	return nil
}

func (g *Generic) GetBIOSConfiguration(ctx context.Context) (map[string]string, error) {
	return nil, nil
}
