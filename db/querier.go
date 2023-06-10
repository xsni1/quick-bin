package db

import "database/sql"

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
