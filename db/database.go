package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbName = "offair.db"
)

// InitDB initializes the SQLite database
func InitDB() (*sqlx.DB, error) {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create a directory for our app if it doesn't exist
	appDir := filepath.Join(homeDir, ".offair")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create app directory: %w", err)
	}

	// Connect to the SQLite database
	dbPath := filepath.Join(appDir, dbName)
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Add airport_type column if it doesn't exist
	if err := addAirportTypeColumn(db); err != nil {
		return nil, fmt.Errorf("failed to add airport_type column: %w", err)
	}

	return db, nil
}

// createTables creates the necessary tables in the database
func createTables(db *sqlx.DB) error {
	// Create airports table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS airports (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			icao TEXT NOT NULL UNIQUE,
			country_code TEXT NOT NULL,
			iata TEXT,
			state TEXT,
			country_name TEXT,
			city TEXT,
			latitude REAL,
			longitude REAL,
			elevation REAL,
			size INTEGER,
			is_military BOOLEAN DEFAULT FALSE,
			has_lights BOOLEAN DEFAULT FALSE,
			is_basecamp BOOLEAN DEFAULT FALSE,
			map_surface_type INTEGER,
			is_in_simbrief BOOLEAN DEFAULT FALSE,
			display_name TEXT,
			has_fbo BOOLEAN DEFAULT FALSE,
			airport_type TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Add airport_type column if it doesn't exist
	_, err = db.Exec(`
		PRAGMA table_info(airports);
	`)
	if err != nil {
		return err
	}

	// SQLite doesn't have a direct way to check if a column exists, so we'll try to add it and ignore the error if it already exists
	_, _ = db.Exec(`
		ALTER TABLE airports ADD COLUMN airport_type TEXT;
	`)
	// Ignore the error if the column already exists

	// Create FBOs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS fbos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			airport_id TEXT NOT NULL,
			icao TEXT NOT NULL,
			name TEXT NOT NULL,
			latitude REAL NOT NULL,
			longitude REAL NOT NULL,
			FOREIGN KEY (airport_id) REFERENCES airports(id),
			UNIQUE(icao)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// addAirportTypeColumn adds the airport_type column to the airports table if it doesn't exist
func addAirportTypeColumn(db *sqlx.DB) error {
	// SQLite doesn't have a direct way to check if a column exists, so we'll try to add it and ignore the error if it already exists
	_, err := db.Exec(`
		ALTER TABLE airports ADD COLUMN airport_type TEXT;
	`)
	// Ignore the error if the column already exists
	// SQLite returns an error like "duplicate column name: airport_type" if the column already exists
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}
