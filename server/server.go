package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/luisferreira32/stickian/server/dummy"
	"github.com/luisferreira32/stickian/server/game"
)

func run(ctx context.Context, address string) {
	middlewares := []func(http.HandlerFunc) http.HandlerFunc{
		panicMiddleware(), // always chain the panic middleware first to prevent panics in other middlewares from crashing the server
	}
	if development {
		middlewares = append(middlewares, loggingMiddleware())
	}

	mux := http.NewServeMux()
	g := game.NewGame(game.NewInMemoryDatabase())

	// define all endpoints
	mux.HandleFunc("/", chainMiddleware(http.FileServer(http.Dir("dist")).ServeHTTP, compressionMiddleware()))
	mux.HandleFunc("/echo", chainMiddleware(dummy.Echo, middlewares...))
	mux.HandleFunc("GET /hello", chainMiddleware(dummy.Hello, middlewares...))
	mux.HandleFunc("POST /panic", chainMiddleware(dummy.Panic, middlewares...))
	mux.HandleFunc("GET /api/city", chainMiddleware(g.GetCity, middlewares...))
	mux.HandleFunc("POST /api/city/building", chainMiddleware(g.UpgradeBuilding, middlewares...))

	// run the server
	server := http.Server{Addr: address, Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("listen and serve err: %v", err)
		}
	}()
	log.Printf("server started on %s\n", address)

	go g.Run(ctx)
	log.Printf("main game loop started")

	<-ctx.Done()

	// handle shutdown with a timeout to allow in-flight requests to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}
