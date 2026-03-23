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

func findMapFile() string {
	if value := os.Getenv("MAP_FILE_PATH"); value != "" {
		return value
	}
	paths := []string{
		"scripts/game_init/world_data/world.csv",     // if run from root
		"../scripts/game_init/world_data/world.csv",  // if run from server/
		"./world_data/world.csv",                     // if run from game_init/
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return paths[0]
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	address := parseDefault("SERVER_ADDRESS", defaultAddress)
	development := parseDefault("DEVELOPMENT", "true") == "true"
	databaseURL := parseDefault("DATABASE_URL", testDatabaseURL)
	migrationsURL := parseDefault("MIGRATIONS_URL", deafultMigrationsURL)
	secretKey := parseDefault("SECRET_KEY", testSecretKey)
	mapFilePath := findMapFile()

	if (testDatabaseURL == databaseURL || secretKey == testSecretKey) && !development {
		log.Panicf("no.")
	}

	err := run(ctx, address, databaseURL, migrationsURL, secretKey, mapFilePath, development)
	if err != nil {
		log.Panicf("server error: %v", err)
	}
}
