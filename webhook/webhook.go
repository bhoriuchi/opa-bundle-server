package webhook

var (
	Providers = map[string]NewWebhookFunc{}
)

type NewWebhookFunc func(config interface{}) (Webhook, error)

type Webhook interface {
}
