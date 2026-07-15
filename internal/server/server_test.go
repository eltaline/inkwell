package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eltaline/inkwell/internal/auth"
	"github.com/eltaline/inkwell/internal/config"
	"github.com/eltaline/inkwell/internal/user"
	"github.com/eltaline/inkwell/internal/version"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	users, err := user.NewStore(dir + "/users.json")
	if err != nil {
		t.Fatalf("user store: %v", err)
	}
	authSvc := auth.NewService(users, "test-secret-key-for-testing")
	cfg := &config.Config{Host: "127.0.0.1", Port: 8080, JWTSecret: "test-secret-key-for-testing", DataDir: dir}
	return New(cfg, users, authSvc)
}

func TestHandleVersion(t *testing.T) {
	version.Version = "1.2.3"
	version.Commit = "abc123"
	version.BuildDate = "2025-01-01"

	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body["version"] != "1.2.3" {
		t.Errorf("version = %q, want %q", body["version"], "1.2.3")
	}
	if body["commit"] != "abc123" {
		t.Errorf("commit = %q, want %q", body["commit"], "abc123")
	}
	if body["build_date"] != "2025-01-01" {
		t.Errorf("build_date = %q, want %q", body["build_date"], "2025-01-01")
	}
}

func TestRegisterAndLogin(t *testing.T) {
	s := newTestServer(t)

	// Register
	body := `{"username":"alice","email":"alice@example.com","password":"secretpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var regResp map[string]string
	json.NewDecoder(rec.Body).Decode(&regResp)
	if regResp["token"] == "" {
		t.Fatal("register returned empty token")
	}

	// Login
	body = `{"username":"alice","password":"secretpass123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var loginResp map[string]string
	json.NewDecoder(rec.Body).Decode(&loginResp)
	token := loginResp["token"]
	if token == "" {
		t.Fatal("login returned empty token")
	}

	// Me (authenticated)
	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var meResp map[string]any
	json.NewDecoder(rec.Body).Decode(&meResp)
	if meResp["username"] != "alice" {
		t.Errorf("me username = %q, want %q", meResp["username"], "alice")
	}

	// Logout
	req = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Me after logout should fail
	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("me after logout status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	s := newTestServer(t)

	// Register first
	body := `{"username":"bob","email":"bob@example.com","password":"correctpass1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	// Login with wrong password
	body = `{"username":"bob","password":"wrongpassword"}`
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	s := newTestServer(t)

	body := `{"username":"charlie","email":"charlie@example.com","password":"password123"}`

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("first register status = %d, want %d", rec.Code, http.StatusCreated)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	rec = httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate register status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestMeWithoutAuth(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	s := newTestServer(t)

	body := `{"username":"weak","email":"weak@example.com","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
