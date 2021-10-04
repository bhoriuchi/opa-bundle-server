package consul

import (
	"context"
	"fmt"

	"github.com/bhoriuchi/opa-bundle-server/core/clients/consul"
	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/publisher"
	"github.com/google/uuid"
	consulapi "github.com/hashicorp/consul/api"
)

const (
	ProviderName = "consul"
)

func init() {
	publisher.Providers[ProviderName] = NewPublisher
}

type Publisher struct {
	name   string
	client *consul.Client
	config *Config
	logger logger.Logger
}

type Config struct {
	Topic         string         `json:"topic" yaml:"topic"`
	NodeFilter    string         `json:"node_filter" yaml:"node_filter"`
	ServiceFilter string         `json:"service_filter" yaml:"service_filter"`
	TagFilter     string         `json:"tag_filter" yaml:"tag_filter"`
	Version       int            `json:"version" yaml:"version"`
	Consul        *consul.Config `json:"consul" yaml:"consul"`
}

// NewPublisher creates a new publisher
func NewPublisher(opts *publisher.Options) (publisher.Publisher, error) {
	p := &Publisher{
		name:   opts.Name,
		config: &Config{},
		logger: opts.Logger,
	}

	if opts.Config == nil {
		return nil, fmt.Errorf("invalid configuration for publisher %s", opts.Name)
	}

	if err := utils.ReMarshal(opts.Config, p.config); err != nil {
		return nil, err
	}

	if p.config.Consul == nil {
		return nil, fmt.Errorf("no consul configuration provided for publisher %s", opts.Name)
	}

	if p.config.Topic == "" {
		return nil, fmt.Errorf("no topic specified for consul publisher %s", p.name)
	}

	return p, nil
}

func (p *Publisher) Connect(ctx context.Context) (err error) {
	p.logger.Debug("connecting to consul publisher %s at %s", p.name, p.config.Consul.Address)
	if p.client != nil {
		return fmt.Errorf("already connected")
	}

	p.client, err = consul.NewClient(p.config.Consul)
	if err != nil {
		return
	}

	return
}
func (p *Publisher) Disconnect(ctx context.Context) (err error) {
	if p.client == nil {
		err = fmt.Errorf("not connected")
		return
	}

	p.client = nil
	return
}

// Publish publishes a message to a topic
func (p *Publisher) Publish(ctx context.Context, payload []byte) (err error) {
	evt := &consulapi.UserEvent{
		ID:            uuid.NewString(),
		Name:          p.config.Topic,
		Payload:       payload,
		NodeFilter:    p.config.NodeFilter,
		ServiceFilter: p.config.ServiceFilter,
		TagFilter:     p.config.TagFilter,
		Version:       p.config.Version,
	}

	p.logger.Debug("publishing event %s to topic %s", evt.ID, evt.Name)
	if _, _, err = p.client.Consul().Event().Fire(evt, &consulapi.WriteOptions{}); err != nil {
		p.logger.Error("failed to publish event %s to topic %s on publisher %s: %s", evt.ID, evt.Name, p.name, err)
		return
	}

	return
}
