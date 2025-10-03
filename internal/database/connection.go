package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func InitDB() (*sql.DB, *Queries, error) {
	return InitDBWithSchema("")
}

func InitDBWithSchema(schemaSQL string) (*sql.DB, *Queries, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	// Try new location first: ~/.config/newsgoat/
	newDir := filepath.Join(homeDir, ".config", "newsgoat")
	newPath := filepath.Join(newDir, "newsgoat.db")

	// Check if database exists in old location
	oldDir := filepath.Join(homeDir, ".newsgoat")
	oldPath := filepath.Join(oldDir, "newsgoat.db")

	var dbPath string
	if _, err := os.Stat(oldPath); err == nil {
		// Use old location if it exists
		dbPath = oldPath
	} else {
		// Use new location (create directory if needed)
		if err := os.MkdirAll(newDir, 0755); err != nil {
			return nil, nil, err
		}
		dbPath = newPath
	}

	// Add SQLite connection parameters for better concurrency
	// dsn := dbPath + "?_busy_timeout=60000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=2000&_locking_mode=NORMAL&_foreign_keys=ON"
	db, err := sql.Open("sqlite3", dbPath) // dsn)
	if err != nil {
		return nil, nil, err
	}

	// Configure connection pool - limit connections to reduce contention
	// SQLite with WAL mode works best with fewer concurrent writers
	// db.SetMaxOpenConns(3)    // Reduce from 10 to 3
	// db.SetMaxIdleConns(2)    // Keep some idle connections
	// db.SetConnMaxLifetime(0) // No limit

	if schemaSQL != "" {
		if err := createTables(db, schemaSQL); err != nil {
			_ = db.Close()
			return nil, nil, err
		}
	}

	queries := New(db)
	return db, queries, nil
}

func createTables(db *sql.DB, schemaSQL string) error {
	_, err := db.Exec(schemaSQL)
	return err
}

