package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/luisferreira32/stickian/server/internal/dummy"
	"github.com/luisferreira32/stickian/server/internal/game"
	"github.com/luisferreira32/stickian/server/internal/user"
)

func run(ctx context.Context, address, databaseURL, migrationsURL, secretKey string, development bool) error {
	middlewares := []func(http.HandlerFunc) http.HandlerFunc{
		panicMiddleware(), // always chain the panic middleware first to prevent panics in other middlewares from crashing the server
		authMiddleware(secretKey),
	}
	if development {
		middlewares = append(middlewares, loggingMiddleware())
	}

	err := runMigrations(migrationsURL, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	db, err := newDatabaseConnection(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(ctx); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	mux := http.NewServeMux()
	var gameDB game.GameDatabase = &game.PostgresDatabase{DB: db}
	if development {
		gameDB = game.NewInMemoryDatabase()
	}
	gameSvc := &game.GameService{Database: gameDB}
	userSvc := &user.UserService{
		SecretKey:   secretKey,
		Database:    &user.PostgresDatabase{DB: db},
		Development: development,
	}
	dummySvc := &dummy.DummyService{
		TickDuration:   time.Second, // 1s
		DummyDatabase1: &dummy.PosgresDatabase{DB: db},
		DummyDatabase2: &dummy.InMemoryDatabase{EventQueue: make(map[int64][]dummy.Event)},
	}
	go dummySvc.Run(ctx)

	// define all endpoints
	// serve static files for the client app
	mux.HandleFunc("/", chainMiddleware(http.FileServer(http.Dir("dist")).ServeHTTP, compressionMiddleware()))
	// dummy endpoints for testing purposes
	mux.HandleFunc("/api/echo", chainMiddleware(dummy.Echo, middlewares...))
	mux.HandleFunc("GET /api/hello", chainMiddleware(dummy.Hello, middlewares...))
	mux.HandleFunc("POST /api/panic", chainMiddleware(dummy.Panic, middlewares...))
	mux.HandleFunc("POST /api/foo", chainMiddleware(dummySvc.TrainFoo, middlewares...))
	mux.HandleFunc("POST /api/bar", chainMiddleware(dummySvc.BuildBar, middlewares...))
	mux.HandleFunc("GET /api/foobar", chainMiddleware(dummySvc.GetFooBar, middlewares...))
	mux.HandleFunc("GET /api/select1", chainMiddleware(dummySvc.Select1, middlewares...))
	// city endpoints
	mux.HandleFunc("GET /api/cities/{id}", chainMiddleware(gameSvc.GetCity, middlewares...))
	mux.HandleFunc("GET /api/cities", chainMiddleware(gameSvc.GetCities, middlewares...))
	// user endpoints
	mux.HandleFunc("POST /api/login", chainMiddleware(userSvc.Login, middlewares...))
	mux.HandleFunc("POST /api/signup", chainMiddleware(userSvc.Signup, middlewares...))
	// map endpoints
	mux.HandleFunc("GET /api/map", chainMiddleware(gameSvc.GetMapChunk, middlewares...))

	// run the server
	server := http.Server{Addr: address, Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen and serve err: %v", err)
		}
	}()
	log.Printf("server started on %s\n", address)

	<-ctx.Done()

	// handle shutdown with a timeout to allow in-flight requests to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}
