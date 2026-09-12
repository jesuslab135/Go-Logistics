// Package dbctx carries a database transaction on the request context, so
// every query built on DB, and every Begin, joins it.
//
// A bulk import wraps many Store.Create calls in one transaction. The stores
// hold a *gen.Queries built at startup, and a dozen open their own
// transaction; threading a tx through every constructor would be a rewrite.
// With the transaction on the context the same stores become transactional
// unchanged: gen.Queries runs on DB, which uses the context's transaction when
// there is one, and Begin opens a savepoint inside it instead of a second
// transaction.
package dbctx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

// WithTx returns a context whose queries and Begin calls join tx.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func txFrom(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok && tx != nil
}

// DB is a gen.DBTX that runs on the context's transaction when there is one,
// and on the pool otherwise.
type DB struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *DB { return &DB{pool: pool} }

func (d *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if tx, ok := txFrom(ctx); ok {
		return tx.Exec(ctx, sql, args...)
	}
	return d.pool.Exec(ctx, sql, args...)
}

func (d *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx, ok := txFrom(ctx); ok {
		return tx.Query(ctx, sql, args...)
	}
	return d.pool.Query(ctx, sql, args...)
}

func (d *DB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx, ok := txFrom(ctx); ok {
		return tx.QueryRow(ctx, sql, args...)
	}
	return d.pool.QueryRow(ctx, sql, args...)
}

// Begin starts a transaction, or a savepoint inside the one ctx carries.
func Begin(ctx context.Context, pool *pgxpool.Pool) (pgx.Tx, error) {
	if tx, ok := txFrom(ctx); ok {
		return tx.Begin(ctx)
	}
	return pool.Begin(ctx)
}
