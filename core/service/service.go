package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/bhoriuchi/opa-bundle-server/core/bundle"
	"github.com/bhoriuchi/opa-bundle-server/core/config"
	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
	"github.com/bhoriuchi/opa-bundle-server/plugins/webhook"
	"github.com/sirupsen/logrus"
)

// Service implements service
type Service struct {
	mx            sync.Mutex
	serviceConfig *Config
	config        *config.Config
	stores        map[string]store.Store
	bundles       map[string]*bundle.Bundle
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
	fmt.Println("level", level)
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

func (s *Service) Bundles() map[string]*bundle.Bundle {
	return s.bundles
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

	if err := s.LoadBundles(); err != nil {
		return err
	}

	return nil
}

// LoadBundles loads bundles
func (s *Service) LoadBundles() error {
	var ok bool

	if s.bundles != nil {
		for name, b := range s.bundles {
			if err := b.Deactivate(); err != nil {
				s.logger.Errorf("failed to deactivate bundle %s", name)
			}
		}
	}

	s.bundles = map[string]*bundle.Bundle{}

	for name, config := range s.config.Bundles {
		b := &bundle.Bundle{
			Name:    name,
			Logger:  s.logger,
			Webhook: config.Webhook,
			Config:  config,
		}

		if b.Store, ok = s.stores[config.Store]; !ok {
			return fmt.Errorf("store %s for bundle %s not found", config.Store, name)
		}

		if err := b.Activate(); err != nil {
			s.logger.Errorf("failed to activate bundle %s", name)
		}

		s.bundles[name] = b
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

		st, err := newFunc(&store.Options{
			Name:   name,
			Config: cfg.Config,
			Logger: s.logger,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize %s store %s: %s", cfg.Type, name, err)
		}

		if err := st.Connect(context.Background()); err != nil {
			return err
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

		hook, err := newFunc(webhook.Options{
			Name:     name,
			Logger:   s.logger,
			Config:   cfg.Config,
			Callback: s.HandleWebhookAction(name),
		})
		if err != nil {
			return fmt.Errorf("failed to initialize %s webhook %s: %s", cfg.Type, name, err)
		}

		s.webhooks[name] = hook
	}
	return nil
}

// HandleWebhook handles webhooks
func (s *Service) HandleWebhook(name string, w http.ResponseWriter, r *http.Request) {
	hook, ok := s.webhooks[name]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hook.Handle(w, r)
}

// Handles the webhook action
func (s *Service) HandleWebhookAction(hookName string) func() {
	return func() {
		for bundleName, bundle := range s.bundles {
			if bundle.Webhook == hookName {
				if err := bundle.Rebuild(context.TODO()); err != nil {
					s.logger.Errorf("failed to rebuild bundle %s: %s", bundleName, err)
				}
			}
		}
	}
}

// HandleBundle handles bundle requests
func (s *Service) HandleBundle(name string, w http.ResponseWriter, r *http.Request) {
	b, ok := s.bundles[name]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/tar+gzip")
	w.Header().Set("ETag", b.Etag())

	etag := r.Header.Get("If-None-Match")
	if etag != "" && etag == b.Etag() {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if _, err := w.Write(b.Data()); err != nil {
		s.logger.Errorf("failed to write bundle request for bundle %s: %s", name, err)
	}
}
