package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xsni1/quick-bin/file"
)

var addr = ":8089"

var host = "localhost"
var dbname = "quick-fix"
var user = "postgres"
var password = "password"
var port = "5444"

func initDbConn() (*sql.DB, error) {
	connString := fmt.Sprintf(
		"port=%s host=%s dbname=%s user=%s password=%s sslmode=disable",
		port,
		host,
		dbname,
		user,
		password,
	)

	db, err := sql.Open("postgres", connString)

	if err != nil {
		return nil, fmt.Errorf("failed to create db value: %w", err)
	}

	err = db.Ping()

	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return db, nil
}

func main() {
	mux := chi.NewMux()
	db, err := initDbConn()
	if err != nil {
		log.Panicf("failed to create file repo: %s", err)
	}

	fileRepo, err := file.NewFilesRepository(db)

	rand.Seed(time.Now().UnixNano())

	if err != nil {
		log.Panicf("failed to create file repo: %s", err)
	}

	fileHandler := file.NewHandler(fileRepo)
	fileHandler.SetupRoutes(mux)

	log.Printf("Server on up %s", addr)
	http.ListenAndServe(addr, mux)
}
