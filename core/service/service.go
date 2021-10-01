package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/bhoriuchi/opa-bundle-server/core/config"
	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/store"
	"github.com/bhoriuchi/opa-bundle-server/webhook"
	"github.com/sirupsen/logrus"
)

// Service implements service
type Service struct {
	mx            sync.Mutex
	serviceConfig *Config
	config        *config.Config
	stores        map[string]store.Store
	webhooks      map[string]webhook.Webhook
	logger        logger.Logger
}

type Config struct {
	Watch    bool
	File     string
	LogLevel string
}

// NewService creates a new service
func NewService(serviceConfig *Config) (*Service, error) {
	log := logrus.New()
	levelStr := serviceConfig.LogLevel
	if levelStr == "" {
		levelStr = "info"
	}

	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return nil, err
	}

	log.SetLevel(level)

	s := &Service{
		serviceConfig: serviceConfig,
		logger:        log,
	}

	// load the configuration
	if err := s.ReloadConfig(); err != nil {
		return nil, err
	}

	// TODO: handle file watching

	return s, nil
}

func (s *Service) Logger() logger.Logger {
	return s.logger
}

func (s *Service) Config() *config.Config {
	return s.config
}

// RelodConfig reloads the configuration file
func (s *Service) ReloadConfig() error {
	s.mx.Lock()
	defer s.mx.Unlock()

	content, err := ioutil.ReadFile(s.serviceConfig.File)
	if err != nil {
		return fmt.Errorf("failed to read configuration file %s: %s", s.serviceConfig.File, err)
	}

	cfg, err := config.NewConfig(content)
	if err != nil {
		return fmt.Errorf("failed to parse configuration file %s: %s", s.serviceConfig.File, err)
	}

	s.config = cfg

	if err := s.LoadStores(); err != nil {
		return err
	}

	if err := s.LoadSubscribers(); err != nil {
		return err
	}

	if err := s.LoadPublishers(); err != nil {
		return err
	}

	if err := s.LoadWebhooks(); err != nil {
		return err
	}

	return nil
}

// LoadStores loads and connects to stores
func (s *Service) LoadStores() error {
	if s.stores != nil {
		for name, store := range s.stores {
			if err := store.Disconnect(context.Background()); err != nil {
				s.logger.Errorf("Failed to disconnect from store %s: %s", name, err)
			}
		}
	}

	s.stores = map[string]store.Store{}

	for name, cfg := range s.config.Stores {
		newFunc, ok := store.Providers[cfg.Type]
		if !ok {
			return fmt.Errorf("invalid store provider type %s", cfg.Type)
		}

		st, err := newFunc(cfg.Config)
		if err != nil {
			return fmt.Errorf("failed to initialize %s store %s: %s", cfg.Type, name, err)
		}

		s.stores[name] = st
	}

	return nil
}

// LoadSubscribers loads and connects subscribers
func (s *Service) LoadSubscribers() error {
	return nil
}

// LoadPublishers loads and connects publishers
func (s *Service) LoadPublishers() error {
	return nil
}

// LoadWebhooks loads webhooks
func (s *Service) LoadWebhooks() error {

	s.webhooks = map[string]webhook.Webhook{}

	// set up new webhooks
	for name, cfg := range s.config.Webhooks {
		newFunc, ok := webhook.Providers[cfg.Type]
		if !ok {
			return fmt.Errorf("invalid webhook provider type %s", cfg.Type)
		}

		hook, err := newFunc(cfg.Config)
		if err != nil {
			return fmt.Errorf("failed to initialize %s webhook %s: %s", cfg.Type, name, err)
		}

		s.webhooks[name] = hook
	}
	return nil
}
