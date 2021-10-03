package subscriber

import (
	"context"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
)

var (
	Providers = map[string]NewSubscriberFunc{}
)

type NewSubscriberFunc func(opts *Options) (Subscriber, error)

type Subscriber interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
	Subscribe(ctx context.Context) (err error)
	Unsubscribe(ctx context.Context) (err error)
}

type Options struct {
	Name     string
	Config   interface{}
	Logger   logger.Logger
	Callback func()
}
