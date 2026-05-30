package server

import (
	"encoding/json"
	"net/http"

	"github.com/eltaline/inkwell/internal/config"
	"github.com/eltaline/inkwell/internal/version"
)

type Server struct {
	cfg    *config.Config
	mux    *http.ServeMux
	server *http.Server
}

func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	s := &Server{
		cfg: cfg,
		mux: mux,
		server: &http.Server{
			Addr:    cfg.Addr(),
			Handler: mux,
		},
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/version", s.handleVersion)
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version":    version.Version,
		"commit":     version.Commit,
		"build_date": version.BuildDate,
	})
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}
