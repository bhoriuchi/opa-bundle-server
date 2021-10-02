package consul

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/bhoriuchi/opa-bundle-server/core/clients/consul"
	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
	consulapi "github.com/hashicorp/consul/api"
)

const (
	ProviderName = "consul"
)

func init() {
	store.Providers[ProviderName] = NewStore
}

type Store struct {
	name   string
	client *consul.Client
	config *Config
	logger logger.Logger
}

type Config struct {
	Prefix string         `json:"prefix" yaml:"prefix"`
	Consul *consul.Config `json:"consul" yaml:"consul"`
}

// NewStore creates a new store
func NewStore(opts *store.Options) (store.Store, error) {
	s := &Store{
		name:   opts.Name,
		config: &Config{},
		logger: opts.Logger,
	}

	if opts.Config == nil {
		return nil, fmt.Errorf("invalid configuration for store %s", opts.Name)
	}

	if err := utils.ReMarshal(opts.Config, s.config); err != nil {
		return nil, err
	}

	if s.config.Consul == nil {
		return nil, fmt.Errorf("no consul configuration provided for store %s", opts.Name)
	}

	if s.config.Prefix == "" {
		s.config.Prefix = path.Join("bundles", s.name)
	}

	return s, nil
}

// Connect is noop but required to implement store interface
func (s *Store) Connect(ctx context.Context) (err error) {
	s.logger.Debugf("connecting to consul store %s", s.name)
	if s.client != nil {
		return fmt.Errorf("already connected")
	}

	s.client, err = consul.NewClient(s.config.Consul)
	return
}

// Disconnect is noop but required to implement store interface
func (s *Store) Disconnect(ctx context.Context) (err error) {
	if s.client == nil {
		return fmt.Errorf("not connected")
	}

	return
}

// Bundle
func (s *Store) Bundle(ctx context.Context) ([]byte, error) {
	s.logger.Tracef("listing prefix %s", s.config.Prefix)
	pairs, _, err := s.client.List(s.config.Prefix, &consulapi.QueryOptions{})
	if err != nil {
		s.logger.Errorf("failed to list consul store %s: %s", s.name, err)
		return nil, err
	}

	list := store.EntryList{}
	for _, pair := range pairs {
		key := strings.TrimLeft(pair.Key, "/")
		key = strings.TrimLeft(key, s.config.Prefix)
		key = strings.TrimLeft(key, "/")

		list = append(list, &store.Entry{
			Key:   key,
			Value: pair.Value,
		})
	}

	archive, err := store.Archive(ctx, list)
	if err != nil {
		return nil, err
	}

	return store.Bundle(ctx, archive, s.config.Consul.Address)
}
