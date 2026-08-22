package main

import (
	"context"
	"errors"
	"ertugruldasgin/shortener/internal/config"
	"ertugruldasgin/shortener/internal/httpapi"
	"ertugruldasgin/shortener/internal/link"
	"ertugruldasgin/shortener/internal/postgres"
	"ertugruldasgin/shortener/internal/slug"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var version = "dev"

const shutdownTimeout = 10 * time.Second

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

	repo := postgres.New(pool)
	gen := slug.New()
	svc := link.NewService(repo, gen)
	recorder := link.NewClickRecorder(repo, cfg.ClickBufferSize)
	h := httpapi.New(svc, recorder, version)

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}

	recorder.Close()
	log.Print("stopped")
}
