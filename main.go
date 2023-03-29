package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xsni1/quick-bin/file"
)

var addr = ":6666"

func main() {
	mux := chi.NewMux()

	fileRepo, err := file.NewFilesRepository()

	if err != nil {
		log.Panicf("failed to create file repo: %s", err)
	}

	fileHandler := file.NewHandler(fileRepo)
	fileHandler.SetupRoutes(mux)

	log.Printf("Server on up %s", addr)
	http.ListenAndServe(addr, mux)
}
