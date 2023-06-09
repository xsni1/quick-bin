package file

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type fileModel struct {
	id             string
	file           string
	downloads_left int
	created_on     time.Time
}

type Connection interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
}

type Querier struct {
	Connection
}

func NewQuerier(db Connection) *Querier {
	return &Querier{
		db,
	}
}

func (q *Querier) WithTx(tx *sql.Tx) (Querier, error) {
	nr := &Querier{
		tx,
	}

	return *nr, nil
}

type FilesRepository struct {
	db      *sql.DB
	querier Querier
}

func NewFilesRepository(db *sql.DB) (FileRepository, error) {
	r := &FilesRepository{
		db:      db,
		querier: *NewQuerier(db),
	}

	return r, nil
}

func (r *FilesRepository) execTx(fn func(q *Querier) error) error {
	tx, err := r.db.BeginTx(context.TODO(), nil)
	if err != nil {
		fmt.Println(err)
	}

	withTx, err := r.querier.WithTx(tx)
	if err != nil {
		fmt.Println(err)
	}

	err = fn(&withTx)
	fmt.Errorf(err.Error())
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (r *FilesRepository) GetIfDownloadsLeft(id string) error {
	r.execTx(func(q *Querier) error {
		rows, err := q.Query(
			"SELECT id, file, created_on FROM files WHERE id = $1",
			id,
		)
		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Println("row", rows)

		rows, err = q.Query(
			"SELECT id, file, created_on FROM files WHERE id = $1",
			id,
		)

		fmt.Println("row", rows)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return nil
	})
	return nil
}

func (r *FilesRepository) Insert(file File) error {
	_, err := r.querier.Exec(
		"INSERT INTO files(id, file, downloads_left, created_on) VALUES ($1, $2, $3, $4)",
		file.Id, file.Name, file.DownloadsLeft, time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *FilesRepository) Get(id string) (*File, error) {
	file := fileModel{}
	rows, err := r.querier.Query(
		"SELECT id, file, created_on FROM files WHERE id = $1",
		id,
	)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	rows.Next()
	err = rows.Scan(&file.id, &file.file, &file.created_on)

	if err != nil {
		return nil, err
	}

	domainFile := &File{
		Name: file.file,
		Id:   file.id,
	}
	return domainFile, nil
}
