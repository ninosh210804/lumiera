// Package bootstrap makes the server self-provision its database on startup:
// it applies the SQL migrations if the schema is missing, then ensures a
// default admin user exists. This lets a fresh deploy (e.g. Render) work
// without a separate migrate/seed step.
//
// The migration files in ./sql are copied from db/migrations/postgres. If you
// add a migration there, copy its *.up.sql here as well.
package bootstrap

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

//go:embed sql/*.up.sql
var migrationFS embed.FS

// Default admin credentials, mirroring test/fixtures/seed.go.
const (
	adminEmail = "admin@coffeeshop.kz"
	adminPIN   = "1234"
	adminName  = "Администратор"
)

// Run applies migrations (only when the schema is absent) and ensures the
// default admin user. It is idempotent and safe to call on every boot.
func Run(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	if err := applyMigrationsIfNeeded(ctx, pool, logger); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	if err := ensureAdmin(ctx, pool, logger); err != nil {
		return fmt.Errorf("admin bootstrap: %w", err)
	}
	return nil
}

// legacyBaseline is the highest migration version that existed before this
// incremental ledger was introduced. On a DB that was already provisioned by
// the old all-or-nothing bootstrap, migrations up to and including this version
// are assumed applied and are recorded (not re-run). Migrations after it must
// therefore be idempotent, since they may run on a long-lived production DB.
const legacyBaseline = "000012"

// applyMigrationsIfNeeded applies any embedded *.up.sql migrations that have not
// yet been recorded in the bootstrap_migrations ledger. It handles three cases:
//   - fresh DB (no schema): apply every migration in order
//   - pre-ledger DB (schema present, empty ledger): record the baseline as
//     applied, then apply anything newer
//   - normal: apply only unrecorded migrations
func applyMigrationsIfNeeded(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS bootstrap_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}

	var usersTable *string
	if err := pool.QueryRow(ctx, `SELECT to_regclass('public.users')`).Scan(&usersTable); err != nil {
		return fmt.Errorf("schema probe: %w", err)
	}
	schemaExists := usersTable != nil

	applied := map[string]bool{}
	rows, err := pool.Query(ctx, `SELECT version FROM bootstrap_migrations`)
	if err != nil {
		return fmt.Errorf("read ledger: %w", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return fmt.Errorf("scan ledger: %w", err)
		}
		applied[v] = true
	}
	rows.Close()

	entries, err := migrationFS.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	// Pre-ledger DB: adopt the legacy baseline as already-applied.
	if schemaExists && len(applied) == 0 {
		for _, name := range names {
			v := migrationVersion(name)
			if v <= legacyBaseline {
				if _, err := pool.Exec(ctx,
					`INSERT INTO bootstrap_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`, v); err != nil {
					return fmt.Errorf("seed baseline %s: %w", v, err)
				}
				applied[v] = true
			}
		}
		logger.Info("bootstrap: adopted legacy migration baseline", "through", legacyBaseline)
	}

	for _, name := range names {
		v := migrationVersion(name)
		if applied[v] {
			continue
		}
		sqlBytes, err := migrationFS.ReadFile("sql/" + name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO bootstrap_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`, v); err != nil {
			return fmt.Errorf("record %s: %w", v, err)
		}
		logger.Info("bootstrap: migration applied", "file", name)
	}
	return nil
}

// migrationVersion extracts the numeric prefix, e.g. "000013_loyalty.up.sql" -> "000013".
func migrationVersion(name string) string {
	if i := strings.IndexByte(name, '_'); i > 0 {
		return name[:i]
	}
	return name
}

// ensureAdmin creates the default admin user if no account with its email
// exists, attaching it to a location (creating one if the DB has none).
func ensureAdmin(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	var existing string
	err := pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, adminEmail).Scan(&existing)
	if err == nil {
		logger.Info("bootstrap: admin user already exists", "email", adminEmail)
		return nil
	}

	// Reuse any existing location, otherwise create the default one.
	var locationID string
	if err := pool.QueryRow(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locationID); err != nil {
		if err := pool.QueryRow(ctx, `
			INSERT INTO locations (name, address, phone)
			VALUES ('Главная кофейня', 'ул. Абая 1, Алматы', '+7 700 000 0000')
			RETURNING id`,
		).Scan(&locationID); err != nil {
			return fmt.Errorf("create location: %w", err)
		}
		logger.Info("bootstrap: location created", "id", locationID)
	}

	var adminRoleID string
	if err := pool.QueryRow(ctx, `SELECT id FROM roles WHERE code = 'admin'`).Scan(&adminRoleID); err != nil {
		return fmt.Errorf("admin role lookup (did migrations seed roles?): %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPIN), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash pin: %w", err)
	}

	var newID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (default_location_id, role_id, full_name, email, pin_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		locationID, adminRoleID, adminName, adminEmail, string(hash),
	).Scan(&newID); err != nil {
		return fmt.Errorf("insert admin: %w", err)
	}
	logger.Info("bootstrap: admin user created", "email", adminEmail, "id", newID)
	return nil
}
