package db

import (
	"context"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool builds a pgx pool and registers the shopspring/decimal codec on every
// connection so numeric columns scan into decimal.Decimal (the type the sqlc
// layer was generated against). It pings before returning so a bad DSN fails fast.
//
// The pool size comes from the DSN, which every environment sets explicitly as
// pool_max_conns=20 (see .env.example and both compose files). Leaving it out
// is not harmless: pgxpool then defaults to max(4, NumCPU), which is 4 on a
// 2-vCPU VPS, and a bulk import holds one connection for the whole import
// rather than the milliseconds an ordinary handler needs - so a few concurrent
// imports would leave every other request blocked in Acquire, with only
// ReadHeaderTimeout set to break the wait.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
