package dbx

import "context"

type key int

const (
	connKey = iota
)

func Attach(ctx context.Context, conn Conn) context.Context {
	return context.WithValue(ctx, connKey, conn)
}

func connFrom(ctx context.Context) Conn {
	conn, _ := ctx.Value(connKey).(Conn)
	return conn
}
