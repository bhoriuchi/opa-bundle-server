package bundle

import (
	"github.com/bhoriuchi/opa-bundle-server/core/webhook"
	"github.com/bhoriuchi/opa-bundle-server/publisher"
	"github.com/bhoriuchi/opa-bundle-server/store"
	"github.com/bhoriuchi/opa-bundle-server/subscriber"
)

type Bundle struct {
	Store      store.Store
	Webhook    *webhook.Webhook
	Subscriber subscriber.Subscriber
	Publisher  publisher.Publisher
}
