package consul

import (
	"context"
	"fmt"
	"time"

	"github.com/bep/debounce"
	"github.com/bhoriuchi/opa-bundle-server/core/clients/consul"
	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/subscriber"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

const (
	ProviderName    = "consul"
	DefaultDebounce = "200ms"
)

func init() {
	subscriber.Providers[ProviderName] = NewSubscriber
}

type Subscriber struct {
	name     string
	wp       *watch.Plan
	cb       func()
	client   *consul.Client
	config   *Config
	logger   logger.Logger
	debounce func(f func())
}

type Config struct {
	Prefix   string         `json:"prefix" yaml:"prefix"`
	Debounce string         `json:"debounce" yaml:"debounce"`
	Consul   *consul.Config `json:"consul" yaml:"consul"`
}

// NewSubscriber creates a new subscriber
func NewSubscriber(opts *subscriber.Options) (subscriber.Subscriber, error) {
	s := &Subscriber{
		name:   opts.Name,
		config: &Config{},
		cb:     opts.Callback,
		logger: opts.Logger,
	}

	if opts.Config == nil {
		return nil, fmt.Errorf("invalid configuration for subscriber %s", opts.Name)
	}

	if err := utils.ReMarshal(opts.Config, s.config); err != nil {
		return nil, err
	}

	if s.config.Debounce == "" {
		s.config.Debounce = DefaultDebounce
	}

	duration, err := time.ParseDuration(s.config.Debounce)
	if err != nil {
		return nil, fmt.Errorf("invalid debounce duration for consul subscriber %s: %s", s.name, err)
	}

	s.debounce = debounce.New(duration)

	if s.config.Consul == nil {
		return nil, fmt.Errorf("no consul configuration provided for subscriber %s", opts.Name)
	}

	if s.config.Prefix == "" {
		s.config.Prefix = "/"
	}

	return s, nil
}

func (s *Subscriber) Connect(ctx context.Context) (err error) {
	s.logger.Debugf("connecting to consul watcher %s", s.name)
	if s.client != nil {
		return fmt.Errorf("already connected")
	}

	s.client, err = consul.NewClient(s.config.Consul)
	if err != nil {
		return
	}

	return
}
func (s *Subscriber) Disconnect(ctx context.Context) (err error) {
	if s.client == nil {
		err = fmt.Errorf("not connected")
		return
	}

	s.Unsubscribe(ctx)
	return
}

func (s *Subscriber) Subscribe(ctx context.Context) (err error) {
	if s.wp != nil {
		err = fmt.Errorf("consul watcher already started on subscriber %s", s.name)
		return
	}

	if s.wp, err = watch.Parse(map[string]interface{}{
		"type":   "keyprefix",
		"prefix": s.config.Prefix,
	}); err != nil {
		return
	}

	s.wp.HybridHandler = func(bv watch.BlockingParamVal, data interface{}) {
		s.logger.Tracef("consul subscriber %s received a message", s.name)
		switch data.(type) {
		case api.KVPairs:
			s.debounce(s.cb)
		}
	}

	go func() {
		s.logger.Debugf("consul subscriber %s is watching prefix %s", s.name, s.config.Prefix)
		if err := s.wp.RunWithClientAndHclog(s.client.Consul(), s.wp.Logger); err != nil {
			s.logger.Errorf("consul watcher on subscriber %s failed: %s", s.name, err)
		}
	}()

	return
}

func (s *Subscriber) Unsubscribe(ctx context.Context) (err error) {
	if s.wp == nil || s.wp.IsStopped() {
		err = fmt.Errorf("consul watcher on subscriber %s is already stopped", s.name)
		return
	}

	s.wp.Stop()
	s.wp = nil
	return
}
