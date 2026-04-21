package migration

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"time"
)

//go:embed *.sql
var Files embed.FS

type MigrationStatus struct {
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

func Run(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("failed to create schema_migrations: %w", err)
	}

	names, err := listSQLFiles()
	if err != nil {
		return err
	}

	for _, name := range names {
		var applied bool
		if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)`, name).Scan(&applied); err != nil {
			return fmt.Errorf("failed to check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		content, err := Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", name, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("migration %s failed: %w", name, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (name) VALUES ($1)`, name); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}
	}

	return nil
}

func Status(db *sql.DB) ([]MigrationStatus, error) {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ DEFAULT NOW()
	)`); err != nil {
		return nil, fmt.Errorf("failed to create schema_migrations: %w", err)
	}

	names, err := listSQLFiles()
	if err != nil {
		return nil, err
	}

	result := make([]MigrationStatus, 0, len(names))
	for _, name := range names {
		var appliedAt *time.Time
		err := db.QueryRow(`SELECT applied_at FROM schema_migrations WHERE name = $1`, name).Scan(&appliedAt)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to query migration %s: %w", name, err)
		}
		result = append(result, MigrationStatus{
			Name:      name,
			Applied:   appliedAt != nil,
			AppliedAt: appliedAt,
		})
	}

	return result, nil
}

func listSQLFiles() ([]string, error) {
	entries, err := Files.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read migration files: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 4 && e.Name()[len(e.Name())-4:] == ".sql" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
