package dbx

import "github.com/jackc/pgx/v5"

func buildScanner[T any]() (func(pgx.Rows) (T, error), error) {
	return nil, nil
}
