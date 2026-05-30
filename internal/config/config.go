package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host string
	Port int
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Host: "0.0.0.0",
		Port: 8080,
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

	return cfg, nil
}
