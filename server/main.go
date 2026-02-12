package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// trigger build demo

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	run(ctx, address)
}
