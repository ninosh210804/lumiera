package bootstrap

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run with: TEST_DATABASE_URL=postgres://... go test ./internal/bootstrap -run TestMigrations -v
func TestMigrationsIncremental(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	reset := func() {
		if _, err := pool.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
			t.Fatalf("reset schema: %v", err)
		}
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	hasCoffeePunches := func() bool {
		var n int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns
			WHERE table_name='loyalty_accounts' AND column_name='coffee_punches'`).Scan(&n)
		return n == 1
	}
	ledgerCount := func() int {
		var n int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM bootstrap_migrations`).Scan(&n)
		return n
	}

	// 1) Fresh DB: everything applies.
	reset()
	if err := applyMigrationsIfNeeded(ctx, pool, logger); err != nil {
		t.Fatalf("fresh apply: %v", err)
	}
	if !hasCoffeePunches() {
		t.Fatal("fresh: coffee_punches column missing")
	}
	if c := ledgerCount(); c != 13 {
		t.Fatalf("fresh: ledger want 13, got %d", c)
	}

	// 2) Simulate a pre-ledger production DB (migrations 1-12 applied, no 13, no ledger).
	if _, err := pool.Exec(ctx, `DROP TABLE bootstrap_migrations`); err != nil {
		t.Fatalf("drop ledger: %v", err)
	}
	if _, err := pool.Exec(ctx, `ALTER TABLE loyalty_accounts DROP COLUMN coffee_punches`); err != nil {
		t.Fatalf("drop column: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM loyalty_rules WHERE code IN ('promo_discount','free_every_n')`); err != nil {
		t.Fatalf("delete rules: %v", err)
	}
	if hasCoffeePunches() {
		t.Fatal("setup: coffee_punches should be gone")
	}

	// 3) Bootstrap should adopt baseline 1-12, then apply only 000013.
	if err := applyMigrationsIfNeeded(ctx, pool, logger); err != nil {
		t.Fatalf("incremental apply: %v", err)
	}
	if !hasCoffeePunches() {
		t.Fatal("incremental: coffee_punches not re-added by migration 000013")
	}
	if c := ledgerCount(); c != 13 {
		t.Fatalf("incremental: ledger want 13, got %d", c)
	}
	var promoExists int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM loyalty_rules WHERE code='promo_discount'`).Scan(&promoExists)
	if promoExists != 1 {
		t.Fatalf("incremental: promo_discount rule not seeded, got %d", promoExists)
	}

	// 4) Idempotent: running again applies nothing new.
	if err := applyMigrationsIfNeeded(ctx, pool, logger); err != nil {
		t.Fatalf("re-run: %v", err)
	}
	if c := ledgerCount(); c != 13 {
		t.Fatalf("re-run: ledger changed, want 13, got %d", c)
	}
}
