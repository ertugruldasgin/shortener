// Package config loads application settings from the env.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every setting the application needs at startup.
type Config struct {
	DatabaseURL     string
	Addr            string
	BaseURL         string
	ShutdownTimeout time.Duration
	ClickBufferSize int
}

// Load reads the env and returns a validated Config.
func Load() (*Config, error) {
	c := &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		Addr:            envOr("ADDR", ":8080"),
		BaseURL:         envOr("BASE_URL", "http://localhost:8080"),
		ShutdownTimeout: 10 * time.Second,
		ClickBufferSize: 256,
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
		}
		c.ShutdownTimeout = d
	}

	if v := os.Getenv("CLICK_BUFFER_SIZE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("CLICK_BUFFER_SIZE must be a positive integer")
		}
		c.ClickBufferSize = n
	}

	return c, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
