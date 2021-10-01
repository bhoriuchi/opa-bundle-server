package memory

import (
	"context"
	"path"
	"strings"

	"github.com/bhoriuchi/opa-bundle-server/store"
)

const (
	ProviderName = "memory"
)

func init() {
	store.Providers[ProviderName] = NewStore
}

// Store implements the store interface
type Store struct {
	bundles map[string]map[string]*store.Entry
}

// NewStore creates a new store
func NewStore(config interface{}) (store.Store, error) {
	s := &Store{
		bundles: map[string]map[string]*store.Entry{},
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
func (s *Store) Get(ctx context.Context, bundle, key string) (out *store.Entry, err error) {
	key = store.NormalizePath(key)
	b, ok := s.bundles[bundle]
	if !ok {
		return nil, store.NotFoundError("bundle %s was not found", bundle)
	}

	e, ok := b[key]
	if !ok {
		return nil, store.NotFoundError("%s not found", path.Join(bundle, key))
	}

	return e, nil
}

// Sets an entry in a bundle
func (s *Store) Set(ctx context.Context, bundle, key string, entry *store.Entry) (err error) {
	key = store.NormalizePath(key)
	b, ok := s.bundles[bundle]
	if !ok {
		b = map[string]*store.Entry{}
	}

	s.bundles[bundle] = b

	b[key] = entry
	return
}

// Del deletes an entry from a bundle
func (s *Store) Del(ctx context.Context, bundle, key string) (err error) {
	key = store.NormalizePath(key)
	b, ok := s.bundles[bundle]
	if !ok {
		return store.NotFoundError("bundle %s was not found", bundle)
	}

	if _, ok := b[key]; !ok {
		return store.NotFoundError("%s not found", path.Join(bundle, key))
	}

	delete(b, key)
	return
}

// List lists bundle
func (s *Store) List(ctx context.Context, bundle, prefix string) (entries []*store.Entry, err error) {
	prefix = store.NormalizePath(prefix)

	b, ok := s.bundles[bundle]
	if !ok {
		return nil, store.NotFoundError("bundle %s was not found", bundle)
	}

	entries = []*store.Entry{}
	for k, entry := range b {
		if prefix == "" || strings.HasPrefix(k, prefix) {
			entries = append(entries, entry)
		}
	}

	return
}
