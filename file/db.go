package file

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type fileModel struct {
	id         string
	file       string
	created_on time.Time
}

type FilesRepository struct {
	db *sql.DB
}

var host = "localhost"
var dbname = "quick-fix"
var user = "postgres"
var password = "password"
var port = "5444"

func NewFilesRepository() (FileRepository, error) {
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

	r := &FilesRepository{
		db: db,
	}

	err = db.Ping()

	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return r, nil
}

func (r *FilesRepository) Insert(file File) error {
	_, err := r.db.Exec(
		"INSERT INTO files(id, file, created_on) VALUES ($1, $2, $3)",
		file.Id, file.Name, time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *FilesRepository) Get(id string) (*File, error) {
	file := fileModel{}
	rows, err := r.db.Query(
		"SELECT id, file, created_on FROM files WHERE id = $1",
		id,
	)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	err = rows.Scan(file)

	if err != nil {
		return nil, err
	}

	domainFile := &File{
		Name: file.file,
		Id:   file.id,
	}
	return domainFile, nil
}
