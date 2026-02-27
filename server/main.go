package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func parseDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	address := parseDefault("SERVER_ADDRESS", defaultAddress)
	development := parseDefault("DEVELOPMENT", "true") == "true"
	databaseURL := parseDefault("DATABASE_URL", "postgres://user:password@localhost:5432/dbname")
	migrationsURL := parseDefault("MIGRATIONS_URL", deafultMigrationsURL)

	if testDatabaseURL == databaseURL && !development {
		log.Panicf("no.")
	}

	err := run(ctx, address, databaseURL, migrationsURL, development)
	if err != nil {
		log.Panicf("server error: %v", err)
	}
}
