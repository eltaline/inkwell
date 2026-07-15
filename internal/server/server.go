package server

import (
	"encoding/json"
	"net/http"

	"github.com/eltaline/inkwell/internal/auth"
	"github.com/eltaline/inkwell/internal/config"
	"github.com/eltaline/inkwell/internal/user"
	"github.com/eltaline/inkwell/internal/version"
)

type Server struct {
	cfg    *config.Config
	mux    *http.ServeMux
	server *http.Server
	auth   *auth.Service
	users  *user.Store
}

func New(cfg *config.Config, users *user.Store, authSvc *auth.Service) *Server {
	mux := http.NewServeMux()
	s := &Server{
		cfg:   cfg,
		mux:   mux,
		auth:  authSvc,
		users: users,
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
	s.mux.HandleFunc("POST /api/auth/register", s.handleRegister)
	s.mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	s.mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	s.mux.Handle("GET /api/auth/me", s.auth.Middleware(http.HandlerFunc(s.handleMe)))
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version":    version.Version,
		"commit":     version.Commit,
		"build_date": version.BuildDate,
	})
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	token, err := s.auth.Register(req.Username, req.Email, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case user.ErrUserExists:
			status = http.StatusConflict
		case user.ErrInvalidEmail, user.ErrWeakPassword, user.ErrEmptyUsername:
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"token": token})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	token, err := s.auth.Login(req.Username, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	tokenStr := auth.ExtractBearerToken(r)
	if tokenStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing authorization header"})
		return
	}

	if err := s.auth.Logout(tokenStr); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	u, err := s.users.GetByID(claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":         u.ID,
		"username":   u.Username,
		"email":      u.Email,
		"created_at": u.CreatedAt,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}
