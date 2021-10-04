package directory

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
	"github.com/open-policy-agent/opa/bundle"
)

const (
	ProviderName = "directory"
)

func init() {
	store.Providers[ProviderName] = NewStore
}

type Store struct {
	name   string
	config *Config
	logger logger.Logger
}

type Config struct {
	Directory string `json:"directory" yaml:"directory"`
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

	return s, nil
}

// Connect is noop but required to implement store interface
func (s *Store) Connect(ctx context.Context) (err error) {
	s.logger.Debug("connecting to directory store %s", s.name)
	return
}

// Disconnect is noop but required to implement store interface
func (s *Store) Disconnect(ctx context.Context) (err error) {
	return
}

// Bundle
func (s *Store) Bundle(ctx context.Context) ([]byte, error) {
	dir, err := filepath.Abs(s.config.Directory)
	if err != nil {
		return nil, err
	}

	loader := bundle.NewDirectoryLoader(dir)
	return store.Bundle(ctx, loader)
}
