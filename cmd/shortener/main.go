package main

import (
	"context"
	"ertugruldasgin/shortener/internal/httpapi"
	"ertugruldasgin/shortener/internal/link"
	"ertugruldasgin/shortener/internal/postgres"
	"ertugruldasgin/shortener/internal/slug"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var version = "dev"

func main() {
	log.Printf("shortener %s starting", version)
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
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
	recorder := link.NewClickRecorder(repo)
	defer recorder.Close()

	h := httpapi.New(svc, recorder, version)

	addr := ":8080"
	log.Printf("listening on %s", addr)

	if err := http.ListenAndServe(addr, h.Routes()); err != nil {
		log.Fatal(err)
	}
}
