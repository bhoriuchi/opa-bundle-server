package service

import (
	"context"
	"fmt"

	"github.com/bhoriuchi/opa-bundle-server/plugins/lock"
)

func (s *Service) Lock(ctx context.Context) error {
	var err error

	if s.lock != nil {
		if err := s.lock.Unlock(ctx); err != nil {
			return err
		}

		if err := s.lock.Disconnect(ctx); err != nil {
			return err
		}

		s.lock = nil
	}

	if s.config.Lock == nil {
		s.logger.Warn("no lock configuration specified. extra care should be taken when using deployers to prevent duplicate deployments")
		return nil
	}

	newFunc, ok := lock.Providers[s.config.Lock.Type]
	if !ok {
		return fmt.Errorf("lock provider %q not registered", s.config.Lock.Type)
	}

	if s.lock, err = newFunc(&lock.Options{
		Config: s.config.Lock.Config,
		Logger: s.logger,
	}); err != nil {
		return err
	}

	if err := s.lock.Connect(ctx); err != nil {
		return err
	}

	go func() {
		if err := lock.Acquire(ctx, s.lock); err != nil {
			s.logger.Error("lock error: %s", err)
		}
	}()

	return nil
}
