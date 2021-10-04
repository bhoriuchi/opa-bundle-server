package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bhoriuchi/opa-bundle-server/core/bundle"
	"github.com/bhoriuchi/opa-bundle-server/plugins/publisher"
)

func (s *Service) Bundles() map[string]*bundle.Bundle {
	return s.bundles
}

// LoadBundles loads bundles
func (s *Service) LoadBundles(ctx context.Context) error {
	var ok bool

	if s.bundles != nil {
		for name, b := range s.bundles {
			if err := b.Deactivate(); err != nil {
				s.logger.Error("failed to deactivate bundle %s", name)
			}
		}
	}

	for name, config := range s.config.Bundles {
		b := &bundle.Bundle{
			Name:        name,
			Logger:      s.logger,
			Webhooks:    config.Webhooks,
			Subscribers: config.Subscribers,
			Publishers:  []publisher.Publisher{},
			Config:      config,
		}

		// add the store to the bundle
		if b.Store, ok = s.stores[config.Store]; !ok {
			return fmt.Errorf("store %s for bundle %s not found", config.Store, name)
		}

		// add the publishers to the bundle
		for _, pubName := range config.Publishers {
			pub, ok := s.publishers[pubName]
			if !ok {
				return fmt.Errorf("publisher %s for bundle %s not found", pubName, name)
			}
			b.Publishers = append(b.Publishers, pub)
		}

		// activate the bundle
		if err := b.Activate(); err != nil {
			s.logger.Error("failed to activate bundle %s", name)
		}

		s.logger.Info("registered bundle %s", name)
		s.bundles[name] = b
	}
	return nil
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
		s.logger.Error("failed to write bundle request for bundle %s: %s", name, err)
	}
}
