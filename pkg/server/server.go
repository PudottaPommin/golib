package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	ctx context.Context
	e   *chi.Mux
	srv *http.Server
}

func New(ctx context.Context, e *chi.Mux) *Server {
	return &Server{ctx: ctx, e: e}
}

func (s *Server) E() *chi.Mux {
	return s.e
}

func (s *Server) Run(addr string) (err error) {
	s.srv = &http.Server{Addr: addr, Handler: s.e}
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) (err error) {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}
