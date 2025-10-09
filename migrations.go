package main

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed sql/migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending migrations to the database
func RunMigrations(db *sql.DB) error {
	// Get all migration files
	entries, err := migrationsFS.ReadDir("sql/migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Parse and sort migration files by version
	type migration struct {
		version int
		name    string
		file    string
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		// Parse version from filename (format: XXXXXX_name.sql)
		parts := strings.SplitN(fileName, "_", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid migration filename format: %s", fileName)
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid migration version in %s: %w", fileName, err)
		}

		migrations = append(migrations, migration{
			version: version,
			name:    strings.TrimSuffix(parts[1], ".sql"),
			file:    fileName,
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Get applied migrations
	appliedVersions, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, m := range migrations {
		if appliedVersions[m.version] {
			continue
		}

		if err := applyMigration(db, m.version, m.file); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", m.file, err)
		}
	}

	return nil
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	// Check if schema_migrations table exists
	var tableExists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM sqlite_master
		WHERE type='table' AND name='schema_migrations'
	`).Scan(&tableExists)
	if err != nil {
		return nil, err
	}

	applied := make(map[int]bool)
	if !tableExists {
		return applied, nil
	}

	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration
func applyMigration(db *sql.DB, version int, file string) error {
	// Read migration file
	content, err := migrationsFS.ReadFile("sql/migrations/" + file)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version) VALUES (?)",
		version,
	); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
