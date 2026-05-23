//go:build ignore

// Seed script: inserts demo data into a fresh database.
// Usage:  go run ./test/fixtures/seed.go
// Or:     make seed
//
// Idempotent: checks existing rows before inserting.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/coffeeshop?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	if err := seed(ctx, pool); err != nil {
		log.Fatalf("seed: %v", err)
	}
	fmt.Println("")
	fmt.Println("✅ Done. Credentials:")
	fmt.Println("   Admin:   admin@coffeeshop.kz  PIN: 1234")
	fmt.Println("   Manager: manager@coffeeshop.kz  PIN: 5678")
	fmt.Println("   Barista: barista@coffeeshop.kz  PIN: 0000")
	fmt.Println("")
	fmt.Println("   API: http://localhost:8080")
	fmt.Println("   Login: POST /api/v1/auth/login  {\"email\":\"admin@coffeeshop.kz\",\"pin\":\"1234\"}")
}

func seed(ctx context.Context, pool *pgxpool.Pool) error {
	// ── Location ──────────────────────────────────────────────────────────────
	var locationID string
	err := pool.QueryRow(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locationID)
	if err != nil {
		// No location yet — create one
		err = pool.QueryRow(ctx, `
			INSERT INTO locations (name, address, phone)
			VALUES ('Главная кофейня', 'ул. Абая 1, Алматы', '+7 700 000 0000')
			RETURNING id`,
		).Scan(&locationID)
		if err != nil {
			return fmt.Errorf("location insert: %w", err)
		}
		fmt.Printf("  ✓ location   created: %s\n", locationID)
	} else {
		fmt.Printf("  · location   exists: %s\n", locationID)
	}

	// ── Roles (seeded by migration, just fetch IDs) ───────────────────────────
	roleIDs := map[string]string{}
	rows, err := pool.Query(ctx, `SELECT code, id FROM roles`)
	if err != nil {
		return fmt.Errorf("roles: %w", err)
	}
	for rows.Next() {
		var code, id string
		_ = rows.Scan(&code, &id)
		roleIDs[code] = id
	}
	rows.Close()

	// ── Admin user ────────────────────────────────────────────────────────────
	var adminID string
	err = pool.QueryRow(ctx, `SELECT id FROM users WHERE email = 'admin@coffeeshop.kz'`).Scan(&adminID)
	if err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
		email := "admin@coffeeshop.kz"
		err = pool.QueryRow(ctx, `
			INSERT INTO users (default_location_id, role_id, full_name, email, pin_hash)
			VALUES ($1, $2, 'Администратор', $3, $4)
			RETURNING id`,
			locationID, roleIDs["admin"], email, string(hash),
		).Scan(&adminID)
		if err != nil {
			return fmt.Errorf("admin user: %w", err)
		}
		fmt.Printf("  ✓ admin      created: %s\n", adminID)
	} else {
		fmt.Printf("  · admin      exists: %s\n", adminID)
	}

	// ── Manager ───────────────────────────────────────────────────────────────
	var managerID string
	err = pool.QueryRow(ctx, `SELECT id FROM users WHERE email = 'manager@coffeeshop.kz'`).Scan(&managerID)
	if err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("5678"), bcrypt.DefaultCost)
		email := "manager@coffeeshop.kz"
		_ = pool.QueryRow(ctx, `
			INSERT INTO users (default_location_id, role_id, full_name, email, pin_hash)
			VALUES ($1, $2, 'Менеджер смены', $3, $4)
			RETURNING id`,
			locationID, roleIDs["manager"], email, string(hash),
		).Scan(&managerID)
		fmt.Printf("  ✓ manager    created: %s\n", managerID)
	} else {
		fmt.Printf("  · manager    exists: %s\n", managerID)
	}

	// ── Barista ───────────────────────────────────────────────────────────────
	baristaEmail := "barista@coffeeshop.kz"
	var count int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role_id = $1`, roleIDs["barista"]).Scan(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("0000"), bcrypt.DefaultCost)
		var baristaID string
		_ = pool.QueryRow(ctx, `
			INSERT INTO users (default_location_id, role_id, full_name, email, pin_hash)
			VALUES ($1, $2, 'Айдана Б.', $3, $4)
			RETURNING id`,
			locationID, roleIDs["barista"], baristaEmail, string(hash),
		).Scan(&baristaID)
		fmt.Printf("  ✓ barista    created: %s\n", baristaID)
	} else {
		// Ensure existing barista has an email (patch for installs seeded before this fix)
		_, _ = pool.Exec(ctx, `
			UPDATE users SET email = $1
			WHERE role_id = $2 AND full_name = 'Айдана Б.' AND (email IS NULL OR email = '')`,
			baristaEmail, roleIDs["barista"],
		)
		fmt.Printf("  · baristas   exist: %d\n", count)
	}

	// ── Categories ────────────────────────────────────────────────────────────
	type catDef struct {
		name      string
		icon      string
		sortOrder int
	}
	cats := []catDef{
		{"Кофе", "☕", 1},
		{"Чай", "🍵", 2},
		{"Еда", "🥐", 3},
		{"Холодные напитки", "🧊", 4},
	}
	catIDs := map[string]string{}
	var catCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM categories WHERE location_id = $1`, locationID).Scan(&catCount)
	if catCount == 0 {
		for _, c := range cats {
			var id string
			err := pool.QueryRow(ctx, `
				INSERT INTO categories (location_id, name, icon, sort_order, created_by)
				VALUES ($1, $2, $3, $4, $5) RETURNING id`,
				locationID, c.name, c.icon, c.sortOrder, adminID,
			).Scan(&id)
			if err == nil {
				catIDs[c.name] = id
			}
		}
		fmt.Printf("  ✓ categories created: %d\n", len(catIDs))
	} else {
		rows, _ := pool.Query(ctx, `SELECT name, id FROM categories WHERE location_id = $1`, locationID)
		for rows.Next() {
			var name, id string
			_ = rows.Scan(&name, &id)
			catIDs[name] = id
		}
		rows.Close()
		fmt.Printf("  · categories exist: %d\n", catCount)
	}

	// ── Products ──────────────────────────────────────────────────────────────
	type prodDef struct {
		cat   string
		name  string
		price float64
		sort  int
	}
	prods := []prodDef{
		{"Кофе", "Эспрессо", 800, 1},
		{"Кофе", "Американо", 900, 2},
		{"Кофе", "Капучино", 1200, 3},
		{"Кофе", "Латте", 1300, 4},
		{"Кофе", "Флэт уайт", 1400, 5},
		{"Кофе", "Раф", 1500, 6},
		{"Чай", "Чёрный чай", 700, 1},
		{"Чай", "Зелёный чай", 700, 2},
		{"Чай", "Матча латте", 1400, 3},
		{"Еда", "Круассан", 600, 1},
		{"Еда", "Сырник", 800, 2},
		{"Холодные напитки", "Лимонад", 1000, 1},
		{"Холодные напитки", "Холодный кофе", 1200, 2},
	}
	var prodCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE location_id = $1`, locationID).Scan(&prodCount)
	if prodCount == 0 {
		inserted := 0
		for _, p := range prods {
			catID := catIDs[p.cat]
			if catID == "" {
				continue
			}
			_, err := pool.Exec(ctx, `
				INSERT INTO products (location_id, category_id, name, base_price, sort_order, created_by)
				VALUES ($1, $2, $3, $4, $5, $6)`,
				locationID, catID, p.name, p.price, p.sort, adminID,
			)
			if err == nil {
				inserted++
			}
		}
		fmt.Printf("  ✓ products   created: %d\n", inserted)
	} else {
		fmt.Printf("  · products   exist: %d\n", prodCount)
	}

	// ── Ingredients ───────────────────────────────────────────────────────────
	type ingDef struct {
		name         string
		unit         string
		minAlert     float64
		isPerishable bool
	}
	ings := []ingDef{
		{"Кофе зерно", "g", 500, false},
		{"Молоко цельное", "ml", 2000, true},
		{"Сливки 33%", "ml", 500, true},
		{"Сахар", "g", 1000, false},
		{"Стакан 350мл", "pcs", 50, false},
		{"Стакан 450мл", "pcs", 50, false},
		{"Крышка", "pcs", 100, false},
		{"Мука пшеничная", "g", 2000, false},
		{"Яйца", "pcs", 20, true},
	}
	var ingCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM ingredients WHERE location_id = $1`, locationID).Scan(&ingCount)
	if ingCount == 0 {
		inserted := 0
		for _, ing := range ings {
			_, err := pool.Exec(ctx, `
				INSERT INTO ingredients (location_id, name, unit, min_stock_alert, is_perishable, created_by)
				VALUES ($1, $2, $3, $4, $5, $6)`,
				locationID, ing.name, ing.unit, ing.minAlert, ing.isPerishable, adminID,
			)
			if err == nil {
				inserted++
			}
		}
		fmt.Printf("  ✓ ingredients created: %d\n", inserted)
	} else {
		fmt.Printf("  · ingredients exist: %d\n", ingCount)
	}

	return nil
}
