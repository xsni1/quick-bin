package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xsni1/quick-bin/file"
)

var addr = ":8089"

func main() {
	mux := chi.NewMux()
	fileRepo, err := file.NewFilesRepository()

	rand.Seed(time.Now().UnixNano())

	if err != nil {
		log.Panicf("failed to create file repo: %s", err)
	}

	fileHandler := file.NewHandler(fileRepo)
	fileHandler.SetupRoutes(mux)

	log.Printf("Server on up %s", addr)
	http.ListenAndServe(addr, mux)
}
