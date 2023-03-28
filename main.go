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

	file.NewFilesRepository()

	fileHandler := file.NewHandler()
	fileHandler.SetupRoutes(mux)

	log.Printf("Server on up %s", addr)
	http.ListenAndServe(addr, mux)
}
