package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

func runMigrations(migrationsURL, databaseURL string) error {
	m, err := migrate.New(migrationsURL, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		err1, err2 := m.Close()
		if err1 != nil || err2 != nil {
			fmt.Printf("failed to close migrate instance: %v, %v", err1, err2)
		}
	}()
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration up: %w", err)
	}
	return nil
}

func newDatabaseConnection(ctx context.Context, databaseURL string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return conn, nil
}
