package publisher

import (
	"context"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
)

var (
	Providers = map[string]NewPublisherFunc{}
)

type NewPublisherFunc func(opts *Options) (Publisher, error)

type Options struct {
	Name   string
	Config interface{}
	Logger logger.Logger
}

type Publisher interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
	Publish(ctx context.Context, payload []byte) (err error)
}
