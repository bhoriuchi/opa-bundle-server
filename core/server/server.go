package server

import (
	"context"
	"net/http"

	"github.com/bhoriuchi/opa-bundle-server/core/config"
	"github.com/bhoriuchi/opa-bundle-server/core/service"
	"github.com/go-chi/chi/v5"
)

// Server implements server
type Server struct {
	service *service.Service
}

// NewServer creates a new server
func NewServer(config *service.Config) (*Server, error) {
	svc, err := service.NewService(config)
	if err != nil {
		return nil, err
	}

	s := &Server{
		service: svc,
	}

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// version the api
	r.Route("/v1", func(r chi.Router) {
		// handle webhooks
		r.Post("/webhooks/{name}", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "name")
			s.service.HandleWebhook(name, w, r)
		})

		r.Get("/bundles/{name}", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "name")
			s.service.Logger().Debug("bundle request for %s", name)
			s.service.HandleBundle(name, w, r)
		})
		r.Post("/bundles/{name}/rebuild", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "name")
			bundles := s.service.Bundles()
			b, ok := bundles[name]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			s.service.Logger().Debug("calling rebuild on bundle %s", name)
			if err := b.Rebuild(r.Context()); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
		})
	})

	srvConfig := s.service.Config().Server
	if srvConfig == nil {
		srvConfig = &config.Server{
			Address: ":8085",
		}
	}

	s.service.Logger().Info("starting bundle server on %s", srvConfig.Address)
	return http.ListenAndServe(srvConfig.Address, r)
}
