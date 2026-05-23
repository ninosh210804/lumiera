//go:build integration

package integration

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func newMigrate(connStr, migrationsDir string) (*migrate.Migrate, error) {
	srcURL := fmt.Sprintf("file://%s", migrationsDir)
	return migrate.New(srcURL, connStr)
}
