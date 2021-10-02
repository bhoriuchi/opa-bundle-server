package webhook

import (
	"net/http"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
)

var (
	Providers = map[string]NewWebhookFunc{}
)

type NewWebhookFunc func(options Options) (Webhook, error)

type Webhook interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

type Options struct {
	Name     string
	Logger   logger.Logger
	Config   interface{}
	Callback func()
}
