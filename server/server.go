package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/luisferreira32/stickian/server/dummy"
	"github.com/luisferreira32/stickian/server/game"
	"github.com/luisferreira32/stickian/server/user"
)

func newDatabaseConnection(ctx context.Context, databaseURL string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return conn, nil
}

func run(ctx context.Context, address, databaseURL string, development bool) {
	middlewares := []func(http.HandlerFunc) http.HandlerFunc{
		panicMiddleware(), // always chain the panic middleware first to prevent panics in other middlewares from crashing the server
	}
	if development {
		middlewares = append(middlewares, loggingMiddleware())
	}

	db, err := newDatabaseConnection(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(ctx); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	mux := http.NewServeMux()
	gameSvc := &game.GameService{Database: &game.InMemoryDatabase{}}
	userSvc := &user.UserService{}
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
	// game endpoints
	mux.HandleFunc("GET /api/city", chainMiddleware(gameSvc.GetCity, middlewares...))
	// user endpoints
	mux.HandleFunc("POST /api/login", chainMiddleware(userSvc.Login, middlewares...))

	// run the server
	server := http.Server{Addr: address, Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("listen and serve err: %v", err)
		}
	}()
	log.Printf("server started on %s\n", address)

	<-ctx.Done()

	// handle shutdown with a timeout to allow in-flight requests to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}
