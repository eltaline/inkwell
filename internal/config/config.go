package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host      string
	Port      int
	JWTSecret string
	DataDir   string
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Host:    "0.0.0.0",
		Port:    8080,
		DataDir: "data",
	}

	if v := os.Getenv("INKWELL_HOST"); v != "" {
		cfg.Host = v
	}

	if v := os.Getenv("INKWELL_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid INKWELL_PORT %q: %w", v, err)
		}
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("INKWELL_PORT %d out of range 1-65535", p)
		}
		cfg.Port = p
	}

	if v := os.Getenv("INKWELL_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}

	if v := os.Getenv("INKWELL_JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	} else {
		secret, err := generateRandomSecret(32)
		if err != nil {
			return nil, fmt.Errorf("generate jwt secret: %w", err)
		}
		cfg.JWTSecret = secret
	}

	return cfg, nil
}

func generateRandomSecret(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
