package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	logger := zerolog.
		New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	mux := chi.NewMux()

	db, err := initDbConn()
	if err != nil {
		log.Fatal().
			Err(err).
			Msgf("Could not connect to db")
	}

	fileRepo := file.NewFilesRepository(db, logger)
	rand.Seed(time.Now().UnixNano())

	// pointer czy nie
	fileHandler := file.NewHandler(fileRepo, logger)
	fileHandler.SetupRoutes(mux)

	logger.Info().Msgf("Server started on: %s", addr)
	http.ListenAndServe(addr, mux)
}
