package file

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/xsni1/quick-bin/db"
)

type fileModel struct {
	id             string
	file           string
	downloads_left int
	created_on     time.Time
}

type FilesRepository struct {
	db      *sql.DB
	querier db.Querier
}

func NewFilesRepository(conn *sql.DB) FileRepository {
	r := &FilesRepository{
		db:      conn,
		querier: *db.NewQuerier(conn),
	}

	return r
}

func (r *FilesRepository) execTx(fn func(q *db.Querier) error) error {
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
	err := r.execTx(func(q *db.Querier) error {
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

	return err
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
