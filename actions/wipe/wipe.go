package wipe

import (
	"context"

	"github.com/sirupsen/logrus"
)

// Wiper defines an interface to wipe a device clean
type Wiper interface {
	Wipe(ctx context.Context, log *logrus.Logger) error
}

// WipeFunc is an adapter to allow the use of ordinary functions as a Wiper.
// It is analogous to [pkg/net/http.HandleFunc].
func WipeFunc(wiper func(context.Context, *logrus.Logger) error) Wiper { // nolint:revive
	return wiperFunc(wiper)
}

type wiperFunc func(context.Context, *logrus.Logger) error

func (w wiperFunc) Wipe(ctx context.Context, log *logrus.Logger) error {
	return w(ctx, log)
}
