package file

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xsni1/quick-bin/db"
	"github.com/xsni1/quick-bin/file/domain"
)

var NoDownloadsLeftErr = errors.New("File has no downloads left")

type fileModel struct {
	id             string
	file           string
	downloads_left int
	created_on     time.Time
}

type FilesRepository struct {
	db      *sql.DB
	querier db.Querier
	logger  zerolog.Logger
}

func NewFilesRepository(conn *sql.DB, logger zerolog.Logger) FileRepository {
	r := &FilesRepository{
		db:      conn,
		querier: *db.NewQuerier(conn),
		logger:  logger,
	}

	return r
}

// alternatively callback could receive Querier
// to make it work this way every db method should be defined on querier
// current implementation that takes repository as a paremeter could even be used in lower layers
// by creating some exposed method like this
func (r *FilesRepository) inTransaction(callback func(r *FilesRepository) (any, error)) (any, error) {
	log.Trace().Msg("Executing execTx")

	tx, err := r.db.BeginTx(context.TODO(), nil)
	if err != nil {
		fmt.Println(err)
	}

	querierWithTx, err := r.querier.WithTx(tx)
	if err != nil {
		fmt.Println(err)
	}

	repoWithTx := r.withTx(querierWithTx)
	result, err := callback(repoWithTx)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return result, nil
}

func (r *FilesRepository) withTx(tx db.Querier) *FilesRepository {
	return &FilesRepository{
		db:      r.db,
		querier: tx,
		logger:  r.logger,
	}
}

func (r *FilesRepository) GetIfDownloadsLeft(id string) (*domain.File, error) {
	result, err := r.inTransaction(func(r *FilesRepository) (any, error) {
		file, err := r.Get(id)
		if err != nil {
			return file, err
		}

		if file.DownloadsLeft == 0 {
			return nil, NoDownloadsLeftErr
		}

		file.DownloadsLeft -= 1

		err = r.Update(*file, file.Id)
		if err != nil {
			return nil, err
		}

		return file, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*domain.File), nil
}

func (r *FilesRepository) Insert(file domain.File) error {
	_, err := r.querier.Exec(
		"INSERT INTO files(id, file, downloads_left, created_on) VALUES ($1, $2, $3, $4)",
		file.Id, file.Name, file.DownloadsLeft, time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *FilesRepository) Update(file domain.File, whereId string) error {
	_, err := r.querier.Exec(
		"UPDATE files SET id = $1, file = $2, downloads_left = $3 WHERE id = $4",
		file.Id, file.Name, file.DownloadsLeft, whereId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *FilesRepository) Get(id string) (*domain.File, error) {
	file := fileModel{}
	rows, err := r.querier.Query(
		"SELECT id, file, created_on, downloads_left FROM files WHERE id = $1",
		id,
	)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	rows.Next()
	err = rows.Scan(&file.id, &file.file, &file.created_on, &file.downloads_left)
	if err != nil {
		return nil, err
	}

	domainFile := &domain.File{
		Name:          file.file,
		Id:            file.id,
		DownloadsLeft: file.downloads_left,
	}
	return domainFile, nil
}
