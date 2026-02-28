package main

const (
	defaultServerPort    = "8080"
	defaultAddress       = "0.0.0.0:" + defaultServerPort
	deafultMigrationsURL = "file://./migrations"
	// YES, this is in plain text. But it is literally just in your local environment.
	testDatabaseURL = "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"
)
