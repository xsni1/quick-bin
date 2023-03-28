package file

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type FilesRepository struct {
	db *sql.DB
}

var host = "localhost"
var dbname = "quick-fix"
var user = "postgres"
var password = "password"

func NewFilesRepository() (FileRepository, error) {
	connString := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s sslmode=disable",
		host,
		dbname,
		user,
		password,
	)

	db, err := sql.Open("postgres", connString)

	if err != nil {
		return nil, fmt.Errorf("failed to create db value: %w", err)
	}

	r := &FilesRepository{
		db: db,
	}

	err = db.Ping()

	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return r, nil
}
