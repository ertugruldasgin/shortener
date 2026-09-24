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
	DatabaseURL       string
	Addr              string
	BaseURL           string
	ShutdownTimeout   time.Duration
	ClickBufferSize   int
	APIToken          string
	AdminToken        string
	RedisURL          string
	RateLimitCreate   int // requests per window for POST /api/links
	RateLimitRedirect int // requests per window for GET /{slug}
	RateLimitWindow   time.Duration
}

// Load reads the env and returns a validated Config.
func Load() (*Config, error) {
	c := &Config{
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		Addr:              envOr("ADDR", ":8080"),
		BaseURL:           envOr("BASE_URL", "http://localhost:8080"),
		ShutdownTimeout:   10 * time.Second,
		ClickBufferSize:   2048,
		APIToken:          os.Getenv("API_TOKEN"),
		AdminToken:        os.Getenv("ADMIN_TOKEN"),
		RedisURL:          os.Getenv("REDIS_URL"),
		RateLimitCreate:   60,
		RateLimitRedirect: 300,
		RateLimitWindow:   time.Minute,
	}

	var err error

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

	if c.ClickBufferSize, err = envInt("CLICK_BUFFER_SIZE", 2048); err != nil {
		return nil, err
	}

	if c.APIToken == "" {
		return nil, fmt.Errorf("API_TOKEN is required.")
	}

	if c.AdminToken == "" {
		return nil, fmt.Errorf("ADMIN_TOKEN is required.")
	}

	if c.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	if c.RateLimitCreate, err = envInt("RATE_LIMIT_CREATE", 60); err != nil {
		return nil, err
	}
	if c.RateLimitRedirect, err = envInt("RATE_LIMIT_REDIRECT", 300); err != nil {
		return nil, err
	}
	return c, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// envInt reads key as a positive integer, falling back to def when unset.
func envInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}

	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}

	return n, nil
}
