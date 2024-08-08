package dbx

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var errNoConn = errors.New("no connection found in ctx, did you call Attach?")

type Conn interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

func Tx(ctx context.Context, body func(ctx context.Context) error) error {
	conn := connFrom(ctx)
	if conn == nil {
		return errNoConn
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	err = body(Attach(ctx, tx))
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func Query[T any](ctx context.Context, sql string, args ...any) (T, error) {
	conn := connFrom(ctx)
	if conn == nil {
		return *new(T), errNoConn
	}

	scanner, err := buildScanner[T]()
	if err != nil {
		return *new(T), fmt.Errorf("building scanner: %w", err)
	}

	sql, args, err = RewriteQuery(ctx, sql, args)
	if err != nil {
		return *new(T), fmt.Errorf("parameterizing: %w", err)
	}

	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		return *new(T), fmt.Errorf("running: %w", err)
	}

	return scanner(rows)
}

func Exec(ctx context.Context, sql string, args ...any) error {
	conn := connFrom(ctx)
	if conn == nil {
		return errNoConn
	}

	sql, args, err := RewriteQuery(ctx, sql, args)
	if err != nil {
		return fmt.Errorf("parameterizing: %w", err)
	}

	_, err = conn.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("running: %w", err)
	}

	return err
}
