package service

import (
	"context"
	"fmt"

	"github.com/bhoriuchi/opa-bundle-server/core/bundle"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/subscriber"
)

// LoadSubscribers loads and connects subscribers
func (s *Service) LoadSubscribers(ctx context.Context) error {
	if s.subscribers != nil {
		for name, sub := range s.subscribers {
			if err := sub.Disconnect(context.Background()); err != nil {
				s.logger.Error("Failed to disconnect subscriber %s: %s", name, err)
			}
		}
	}

	s.subscribers = map[string]subscriber.Subscriber{}

	// set up new subscribers
	for name, cfg := range s.config.Subscribers {
		newFunc, ok := subscriber.Providers[cfg.Type]
		if !ok {
			return fmt.Errorf("invalid subscriber provider type %s", cfg.Type)
		}

		sub, err := newFunc(&subscriber.Options{
			Name:   name,
			Logger: s.logger,
			Config: cfg.Config,
			Callback: s.HandleCallback(name, "subscriber", func(b *bundle.Bundle) bool {
				return utils.StringSliceContains(b.Subscribers, name)
			}),
		})
		if err != nil {
			return fmt.Errorf("failed to initialize %s subscriber %s: %s", cfg.Type, name, err)
		}

		if err := sub.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect %s subscriber %s: %s", cfg.Type, name, err)
		}

		if err := sub.Subscribe(ctx); err != nil {
			return fmt.Errorf("failed to subscribe %s subscriber %s: %s", cfg.Type, name, err)
		}

		s.logger.Info("registering subscriber %s", name)
		s.subscribers[name] = sub
	}

	return nil
}
