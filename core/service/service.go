package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/bhoriuchi/opa-bundle-server/core/bundle"
	"github.com/bhoriuchi/opa-bundle-server/core/config"
	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/plugins/deployer"
	"github.com/bhoriuchi/opa-bundle-server/plugins/publisher"
	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
	"github.com/bhoriuchi/opa-bundle-server/plugins/subscriber"
	"github.com/bhoriuchi/opa-bundle-server/plugins/webhook"
	"github.com/open-policy-agent/opa/logging"
)

// Service implements service
type Service struct {
	mx            sync.Mutex
	serviceConfig *Config
	config        *config.Config
	stores        map[string]store.Store
	bundles       map[string]*bundle.Bundle
	webhooks      map[string]webhook.Webhook
	subscribers   map[string]subscriber.Subscriber
	publishers    map[string]publisher.Publisher
	deployers     map[string]deployer.Deployer
	logger        logger.Logger
}

type Config struct {
	Watch     bool
	File      string
	LogLevel  string
	LogFormat string
}

// NewService creates a new service
func NewService(serviceConfig *Config) (*Service, error) {
	log := logging.New()
	log.SetLevel(logger.ParseLevel(serviceConfig.LogLevel))
	log.SetFormatter(logger.ParseFormatter(serviceConfig.LogFormat))

	s := &Service{
		serviceConfig: serviceConfig,
		stores:        map[string]store.Store{},
		bundles:       map[string]*bundle.Bundle{},
		webhooks:      map[string]webhook.Webhook{},
		subscribers:   map[string]subscriber.Subscriber{},
		publishers:    map[string]publisher.Publisher{},
		deployers:     map[string]deployer.Deployer{},
		logger:        log,
	}

	// load the configuration
	if err := s.ReloadConfig(context.TODO()); err != nil {
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
func (s *Service) ReloadConfig(ctx context.Context) error {
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

	if err := s.LoadStores(ctx); err != nil {
		return err
	}

	if err := s.LoadSubscribers(ctx); err != nil {
		return err
	}

	if err := s.LoadPublishers(ctx); err != nil {
		return err
	}

	if err := s.LoadDeployers(ctx); err != nil {
		return err
	}

	if err := s.LoadWebhooks(ctx); err != nil {
		return err
	}

	if err := s.LoadBundles(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) HandleCallback(name, typ string, matcher func(b *bundle.Bundle) bool) func() {
	return func() {
		if len(s.bundles) == 0 {
			s.logger.Warn("no bundles were registered on the service")
			return
		}
		for bundleName, bundle := range s.bundles {
			s.logger.Debug("attempting to match bundle %s", bundleName)
			if matcher(bundle) {
				s.logger.Debug("%s callback handler %s matched bundle %s", typ, name, bundleName)
				if err := bundle.Rebuild(context.TODO()); err != nil {
					s.logger.Error("failed to rebuild bundle %s: %s", bundleName, err)
				}
				return
			}
		}
		s.logger.Warn("%s callback handler %s did not match and plugins", typ, name)
	}
}
