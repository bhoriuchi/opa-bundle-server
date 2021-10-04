package service

import (
	"context"
	"fmt"

	"github.com/bhoriuchi/opa-bundle-server/plugins/deployer"
)

// LoadDeployers loads and connects deployers
func (s *Service) LoadDeployers(ctx context.Context) error {
	if s.deployers != nil {
		for name, deployer := range s.deployers {
			if err := deployer.Disconnect(ctx); err != nil {
				s.logger.Error("Failed to disconnect deployer %s: %s", name, err)
			}
		}
	}

	s.deployers = map[string]deployer.Deployer{}

	// set up new deployer
	for name, cfg := range s.config.Deployers {
		newFunc, ok := deployer.Providers[cfg.Type]
		if !ok {
			return fmt.Errorf("invalid deployer provider type %s", cfg.Type)
		}

		deployer, err := newFunc(&deployer.Options{
			Name:   name,
			Logger: s.logger,
			Config: cfg.Config,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize %s deployer %s: %s", cfg.Type, name, err)
		}

		if err := deployer.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect %s deployer %s: %s", cfg.Type, name, err)
		}

		s.logger.Info("registering deployer %s", name)
		s.deployers[name] = deployer
	}

	return nil
}
