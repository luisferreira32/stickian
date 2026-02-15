package game

import (
	"context"
	"time"
)

// NewGame returns a new instance of the game struct with the provided database.
func NewGame(db database) *game {
	return &game{db: db}
}

// game is the singleton struct that will hold all the ephemeral game state, logic
// and needed clients (e.g., database).
type game struct {
	db database
}

type eventName string

const (
	UpgradeCity eventName = "UpgradeCity"
)

func (g *game) Run(ctx context.Context) {
	clock := time.NewTicker(1 * time.Second)
	defer clock.Stop()

	for {
		select {
		case <-clock.C:
			// 1. get events from the database
			// 2. get current state
			// 3. calculate new state based on events
			// 4. write new state in a transaction with a clean-up of processed events
			// if anything fails, next tick will just re-process the unprocessed events
		case <-ctx.Done():
			return
		}
	}
}
