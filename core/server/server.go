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

	srvConfig := s.service.Config().Server
	if srvConfig == nil {
		srvConfig = &config.Server{
			Address: ":8085",
		}
	}

	s.service.Logger().Infof("starting bundle server on %s", srvConfig.Address)
	return http.ListenAndServe(srvConfig.Address, r)
}
