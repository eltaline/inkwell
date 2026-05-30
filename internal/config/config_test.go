package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("INKWELL_HOST")
	os.Unsetenv("INKWELL_PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Addr() != "0.0.0.0:8080" {
		t.Errorf("Addr() = %q, want %q", cfg.Addr(), "0.0.0.0:8080")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("INKWELL_HOST", "127.0.0.1")
	t.Setenv("INKWELL_PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %q, want %q", cfg.Host, "127.0.0.1")
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want %d", cfg.Port, 9090)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("INKWELL_PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestLoadPortOutOfRange(t *testing.T) {
	t.Setenv("INKWELL_PORT", "70000")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for port out of range")
	}
}
