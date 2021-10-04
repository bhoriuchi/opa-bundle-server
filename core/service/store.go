package service

import (
	"context"
	"fmt"

	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
)

// LoadStores loads and connects to stores
func (s *Service) LoadStores(ctx context.Context) error {
	if s.stores != nil {
		for name, store := range s.stores {
			if err := store.Disconnect(context.Background()); err != nil {
				s.logger.Error("Failed to disconnect from store %s: %s", name, err)
			}
		}
	}

	s.stores = map[string]store.Store{}

	for name, cfg := range s.config.Stores {
		newFunc, ok := store.Providers[cfg.Type]
		if !ok {
			return fmt.Errorf("invalid store provider type %s", cfg.Type)
		}

		st, err := newFunc(&store.Options{
			Name:   name,
			Config: cfg.Config,
			Logger: s.logger,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize %s store %s: %s", cfg.Type, name, err)
		}

		if err := st.Connect(ctx); err != nil {
			return err
		}

		s.stores[name] = st
	}

	return nil
}
