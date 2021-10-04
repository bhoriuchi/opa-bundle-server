package service

import (
	"context"
	"fmt"

	"github.com/bhoriuchi/opa-bundle-server/plugins/publisher"
)

// LoadPublishers loads and connects publishers
func (s *Service) LoadPublishers(ctx context.Context) error {
	if s.publishers != nil {
		for name, pub := range s.publishers {
			if err := pub.Disconnect(ctx); err != nil {
				s.logger.Error("Failed to disconnect publisher %s: %s", name, err)
			}
		}
	}

	s.publishers = map[string]publisher.Publisher{}

	// set up new publishers
	for name, cfg := range s.config.Publishers {
		newFunc, ok := publisher.Providers[cfg.Type]
		if !ok {
			return fmt.Errorf("invalid publisher provider type %s", cfg.Type)
		}

		pub, err := newFunc(&publisher.Options{
			Name:   name,
			Logger: s.logger,
			Config: cfg.Config,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize %s publisher %s: %s", cfg.Type, name, err)
		}

		if err := pub.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect %s publisher %s: %s", cfg.Type, name, err)
		}

		s.logger.Info("registering publisher %s", name)
		s.publishers[name] = pub
	}

	return nil
}
