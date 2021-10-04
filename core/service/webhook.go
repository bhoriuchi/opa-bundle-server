package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bhoriuchi/opa-bundle-server/core/bundle"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/webhook"
)

// HandleWebhook handles webhooks
func (s *Service) HandleWebhook(name string, w http.ResponseWriter, r *http.Request) {
	hook, ok := s.webhooks[name]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hook.Handle(w, r)
}

// LoadWebhooks loads webhooks
func (s *Service) LoadWebhooks(ctx context.Context) error {
	s.webhooks = map[string]webhook.Webhook{}

	// set up new webhooks
	for name, cfg := range s.config.Webhooks {
		newFunc, ok := webhook.Providers[cfg.Type]
		if !ok {
			return fmt.Errorf("invalid webhook provider type %s", cfg.Type)
		}

		hook, err := newFunc(webhook.Options{
			Name:   name,
			Logger: s.logger,
			Config: cfg.Config,
			Callback: s.HandleCallback(name, "webhook", func(b *bundle.Bundle) bool {
				return utils.StringSliceContains(b.Webhooks, name)
			}),
		})
		if err != nil {
			return fmt.Errorf("failed to initialize %s webhook %s: %s", cfg.Type, name, err)
		}

		s.webhooks[name] = hook
	}
	return nil
}
