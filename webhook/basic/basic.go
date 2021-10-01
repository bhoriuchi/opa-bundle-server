package basic

import (
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/webhook"
)

const (
	ProviderName = "basic"
)

func init() {
	webhook.Providers[ProviderName] = NewWebhook
}

type Webhook struct {
	Config *Config
}

type Config struct {
	Secret string `json:"secret"`
}

func NewWebhook(config interface{}) (webhook.Webhook, error) {
	h := &Webhook{
		Config: &Config{},
	}

	if err := utils.ReMarshal(config, h.Config); err != nil {
		return nil, err
	}

	return h, nil
}
