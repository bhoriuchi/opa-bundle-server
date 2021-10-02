package memory

import (
	"context"

	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
)

const (
	ProviderName = "memory"
)

func init() {
	store.Providers[ProviderName] = NewStore
}

// Store implements the store interface
type Store struct {
	name   string
	data   map[string]*store.Entry
	config *Config
}

type Config struct{}

// NewStore creates a new store
func NewStore(opts *store.Options) (store.Store, error) {
	s := &Store{
		name:   opts.Name,
		data:   map[string]*store.Entry{},
		config: &Config{},
	}

	if err := utils.ReMarshal(opts.Config, s.config); err != nil {
		return nil, err
	}

	return s, nil
}

// Connect is noop but required to implement store interface
func (s *Store) Connect(ctx context.Context) (err error) {
	return
}

// Disconnect is noop but required to implement store interface
func (s *Store) Disconnect(ctx context.Context) (err error) {
	return
}

// Get gets an entry from a bundle
func (s *Store) Get(ctx context.Context, key string) (out *store.Entry, err error) {
	key = store.NormalizePath(key)

	e, ok := s.data[key]
	if !ok {
		return nil, store.NotFoundError("%s not found", key)
	}

	return e, nil
}

// Sets an entry in a bundle
func (s *Store) Set(ctx context.Context, key string, entry *store.Entry) (err error) {
	key = store.NormalizePath(key)
	s.data[key] = entry
	return
}

// Del deletes an entry from a bundle
func (s *Store) Del(ctx context.Context, key string) (err error) {
	key = store.NormalizePath(key)

	if _, ok := s.data[key]; !ok {
		return store.NotFoundError("%s not found", key)
	}

	delete(s.data, key)
	return
}

// List lists bundle
func (s *Store) List(ctx context.Context) (entries store.EntryList, err error) {
	entries = []*store.Entry{}
	for _, entry := range s.data {
		entries = append(entries, entry)
	}

	return
}

func (s *Store) Archive(ctx context.Context) ([]byte, error) {
	files, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	return store.Archive(ctx, files)
}

// Bundle
func (s *Store) Bundle(ctx context.Context) ([]byte, error) {
	archive, err := s.Archive(ctx)
	if err != nil {
		return nil, err
	}

	return store.Bundle(ctx, archive, "file://")
}
