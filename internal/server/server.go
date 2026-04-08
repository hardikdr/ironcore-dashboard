package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
)

type Server struct {
	router   *chi.Mux
	ironcore versioned.Interface
}

func New(cs versioned.Interface) *Server {
	s := &Server{ironcore: cs}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	s.router = r
	return s
}

func (s *Server) Router() *chi.Mux { return s.router }

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
