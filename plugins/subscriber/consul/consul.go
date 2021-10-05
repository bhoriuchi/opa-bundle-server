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
	WatchType string         `json:"watch_type" yaml:"watch_type"`
	Topic     string         `json:"topic" yaml:"topic"`
	Debounce  string         `json:"debounce" yaml:"debounce"`
	Consul    *consul.Config `json:"consul" yaml:"consul"`
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

	if s.config.Topic == "" {
		return nil, fmt.Errorf("no topic specified for consul subscriber %s", s.name)
	}

	return s, nil
}

func (s *Subscriber) Connect(ctx context.Context) (err error) {
	s.logger.Debug("connecting to consul watcher %s at %s", s.name, s.config.Consul.Address)
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
	s.client = nil
	return
}

func (s *Subscriber) Subscribe(ctx context.Context) (err error) {
	if s.wp != nil {
		err = fmt.Errorf("consul watcher already started on subscriber %s", s.name)
		return
	}

	if s.config.Topic == "" {
		err = fmt.Errorf("no topic specified for consul subscriber %s", s.name)
		return
	}

	params := map[string]interface{}{
		"type": s.config.WatchType,
	}

	// generate params
	switch s.config.WatchType {
	case "key":
		params["key"] = s.config.Topic
	case "keyprefix", "":
		params["type"] = "keyprefix"
		params["prefix"] = s.config.Topic
	case "event":
		params["name"] = s.config.Topic
	default:
		err = fmt.Errorf("unsupported watch type %q for consul subscriber %s", s.config.WatchType, s.name)
		return
	}

	if s.wp, err = watch.Parse(params); err != nil {
		return
	}

	s.wp.HybridHandler = func(bv watch.BlockingParamVal, data interface{}) {
		s.logger.Debug("consul subscriber %s received a message", s.name)
		s.debounce(s.cb)
	}

	go func() {
		s.logger.Debug("consul subscriber %s is watching prefix %s", s.name, s.config.Topic)
		if err := s.wp.RunWithClientAndHclog(s.client.Consul(), s.wp.Logger); err != nil {
			s.logger.Error("consul watcher on subscriber %s failed: %s", s.name, err)
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
