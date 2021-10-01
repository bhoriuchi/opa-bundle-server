package subscriber

import "context"

var (
	Providers = map[string]NewSubscriberFunc{}
)

type NewSubscriberFunc func(config interface{}) (Subscriber, error)

type Subscriber interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
}
