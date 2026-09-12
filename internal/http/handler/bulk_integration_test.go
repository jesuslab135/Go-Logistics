//go:build integration

package handler

import (
	"context"
	"testing"

	"fleet/internal/db/gen"
	"fleet/internal/platform/dbctx"
)

// Queries built on dbctx.DB join the transaction the context carries, and
// dbctx.Begin inside it is a savepoint that can roll back alone.
func TestContextTransactionJoinsQueriesAndBegin(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	q := gen.New(dbctx.New(pool))

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	txCtx := dbctx.WithTx(ctx, tx)

	if _, err := q.CreateAccount(txCtx, "outer write"); err != nil {
		t.Fatalf("write through the context transaction: %v", err)
	}

	inner, err := dbctx.Begin(txCtx, pool)
	if err != nil {
		t.Fatalf("savepoint: %v", err)
	}
	if _, err := gen.New(inner).CreateAccount(txCtx, "inner write"); err != nil {
		t.Fatalf("write inside the savepoint: %v", err)
	}
	if err := inner.Rollback(txCtx); err != nil {
		t.Fatalf("roll back the savepoint: %v", err)
	}

	count := func(ctx context.Context, db gen.DBTX) int {
		var n int
		if err := db.QueryRow(ctx,
			"SELECT count(*) FROM account WHERE name IN ('outer write','inner write')").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n := count(txCtx, tx); n != 1 {
		t.Fatalf("inside the transaction: %d rows, want 1 (the savepoint's write undone, the outer kept)", n)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if n := count(ctx, pool); n != 0 {
		t.Fatalf("after rolling back: %d rows, want 0", n)
	}
}
