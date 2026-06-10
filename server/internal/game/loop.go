package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/luisferreira32/stickian/server/internal/utils"
)

// Run starts the game loop. On each tick it drains all due events from the queue
// and processes them in deterministic order. The loop is the sole writer of game
// objects; HTTP handlers only enqueue events.
func (g *GameService) Run(ctx context.Context) {
	lastProcessedTick := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// sleep until the next tick (a negative duration returns immediately)
		time.Sleep(g.TickDuration - time.Since(lastProcessedTick))

		if err := g.tick(ctx); err != nil {
			// transient error: keep lastProcessedTick so the same due events are
			// retried right away on the next iteration
			log.Printf("game tick err: %v", err)
			continue
		}
		lastProcessedTick = time.Now()
	}
}

// tick drains and processes all events due at the current time, in deterministic
// (process_after, seq) order. An event is marked processed only once its effect
// has been applied (or it has been deliberately rejected); a transient infra error
// aborts the tick so the remaining events are retried on the next tick.
func (g *GameService) tick(ctx context.Context) error {
	events, err := g.Database.GetDueEvents(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get due events: %w", err)
	}

	for _, e := range events {
		if err := g.processEvent(ctx, e); err != nil {
			// infra error: stop here, leaving this and later events unprocessed for retry
			return fmt.Errorf("failed to process event %d (%s): %w", e.Seq, e.Type, err)
		}
		if err := g.Database.MarkEventProcessed(ctx, e.Seq); err != nil {
			return fmt.Errorf("failed to mark event %d processed: %w", e.Seq, err)
		}
	}
	return nil
}

// processEvent applies a single event. It returns an error only for transient
// infrastructure failures (which should be retried). A rejected event (e.g. the
// origin city no longer exists or is not owned by the player) is logged and
// consumed by returning nil — the client reconciles via GetMovements.
func (g *GameService) processEvent(ctx context.Context, e *Event) error {
	switch e.Type {
	case EventCreateMovement:
		return g.processCreateMovement(ctx, e)
	case EventCancelMovement:
		return g.processCancelMovement(ctx, e)
	default:
		log.Printf("unknown event type %q (seq %d), skipping", e.Type, e.Seq)
		return nil
	}
}

func (g *GameService) processCreateMovement(ctx context.Context, e *Event) error {
	var payload CreateMovementPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		log.Printf("create_movement event %d: invalid payload: %v", e.Seq, err)
		return nil
	}
	m := payload.Movement
	if m == nil {
		log.Printf("create_movement event %d: nil movement, rejecting", e.Seq)
		return nil
	}

	// authoritative re-check: the origin city must still exist and be owned by the player
	from, err := g.Database.GetCity(ctx, m.CityFrom)
	if errors.Is(err, utils.ErrNotFound) {
		log.Printf("create_movement event %d: origin city %s gone, rejecting", e.Seq, m.CityFrom)
		return nil
	}
	if err != nil {
		return err
	}
	if from.PlayerID != payload.PlayerID {
		log.Printf("create_movement event %d: origin city %s not owned by %s, rejecting", e.Seq, m.CityFrom, payload.PlayerID)
		return nil
	}

	// the destination city must exist
	if _, err := g.Database.GetCity(ctx, m.CityTo); errors.Is(err, utils.ErrNotFound) {
		log.Printf("create_movement event %d: destination city %s gone, rejecting", e.Seq, m.CityTo)
		return nil
	} else if err != nil {
		return err
	}

	// commands-only: no resource/troop deduction yet. CreateMovement is idempotent
	// (ON CONFLICT (id) DO NOTHING) so re-processing the same event is safe.
	// TODO: when deduction + arrival effects land, wrap the writes in a transaction.
	return g.Database.CreateMovement(ctx, m)
}

func (g *GameService) processCancelMovement(ctx context.Context, e *Event) error {
	var payload CancelMovementPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		log.Printf("cancel_movement event %d: invalid payload: %v", e.Seq, err)
		return nil
	}

	// DeleteMovement is idempotent: removing an already-deleted (or never-created)
	// movement is a no-op, so re-processing the same event is safe.
	return g.Database.DeleteMovement(ctx, payload.MovementID)
}
