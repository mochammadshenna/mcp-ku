package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs database migrations
func RunMigrations(databaseURL string) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create pgvector extension if it doesn't exist
	if err := createPgVectorExtension(db); err != nil {
		return fmt.Errorf("failed to create pgvector extension: %w", err)
	}

	// Run migrations
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// createPgVectorExtension creates the pgvector extension
func createPgVectorExtension(db *sql.DB) error {
	query := `CREATE EXTENSION IF NOT EXISTS vector;`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}

	return nil
}
