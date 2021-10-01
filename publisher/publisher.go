package publisher

import "context"

var (
	Providers = map[string]NewPublisherFunc{}
)

type NewPublisherFunc func(config interface{}) (Publisher, error)

type Publisher interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
}
