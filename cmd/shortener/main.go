package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"ertugruldasgin/shortener/internal/config"
	"ertugruldasgin/shortener/internal/httpapi"
	"ertugruldasgin/shortener/internal/link"
	"ertugruldasgin/shortener/internal/postgres"
	"ertugruldasgin/shortener/internal/rediscache"
	"ertugruldasgin/shortener/internal/slug"
)

var version = "dev"

func main() {
	log.Printf("shortener %s starting", version)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("pinging database: %v", err)
	}

	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrating: %v", err)
	}

	cache, err := rediscache.New(cfg.RedisURL)
	if err != nil {
		log.Fatalf("configuring cache: %v", err)
	}
	defer cache.Close()

	if err := cache.Ping(context.Background()); err != nil {
		log.Printf("warning: redis unreachable, serving from database: %v", err)
	}

	repo := postgres.New(pool)
	gen := slug.New()
	svc := link.NewService(repo, gen, cache)
	recorder := link.NewClickRecorder(repo, cfg.ClickBufferSize)
	recorder.OnDrop(httpapi.ClicksDropped())
	h := httpapi.New(httpapi.Config{
		Service:  svc,
		Recorder: recorder,
		Limiter:  cache,
		RateLimits: httpapi.RateLimits{
			Create:   cfg.RateLimitCreate,
			Redirect: cfg.RateLimitRedirect,
			Window:   cfg.RateLimitWindow,
		},
		Version:    version,
		APIToken:   cfg.APIToken,
		AdminToken: cfg.AdminToken,
	})
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: h.Routes(),
	}

	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serving: %v", err)
		}
	}()

	<-ctx.Done()
	log.Print("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}

	recorder.Close()
	log.Print("stopped")
}
