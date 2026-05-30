package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eltaline/inkwell/internal/config"
	"github.com/eltaline/inkwell/internal/version"
)

func TestHandleVersion(t *testing.T) {
	version.Version = "1.2.3"
	version.Commit = "abc123"
	version.BuildDate = "2025-01-01"

	cfg := &config.Config{Host: "127.0.0.1", Port: 8080}
	s := New(cfg)

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
