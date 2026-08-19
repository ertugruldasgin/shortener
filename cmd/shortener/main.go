package main

import (
	"ertugruldasgin/shortener/internal/httpapi"
	"ertugruldasgin/shortener/internal/link"
	"ertugruldasgin/shortener/internal/memstore"
	"ertugruldasgin/shortener/internal/slug"
	"log"
	"net/http"
)

func main() {
	repo := memstore.New()
	gen := slug.New()
	svc := link.NewService(repo, gen)
	h := httpapi.New(svc)

	addr := ":8080"
	log.Printf("listening on %s", addr)

	if err := http.ListenAndServe(addr, h.Routes()); err != nil {
		log.Fatal(err)
	}
}
