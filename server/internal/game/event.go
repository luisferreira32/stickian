package game

import (
	"encoding/json"
	"time"
)

// EventType identifies the kind of event stored in the queue.
type EventType string

const (
	// EventCreateMovement is enqueued by the CreateMovement handler. The game loop
	// validates ownership and writes the Movement row when processing it.
	EventCreateMovement EventType = "create_movement"
	// EventCancelMovement is enqueued by the DeleteMovement handler. The game loop
	// removes the Movement row when processing it.
	EventCancelMovement EventType = "cancel_movement"
)

// Event is a single item in the game event queue.
//
// Events are submitted by HTTP handlers and processed by the game loop in a
// deterministic order: by ProcessAfter, then by Seq (the database-assigned
// monotonic sequence) to break ties. The loop is the sole writer of game
// objects; handlers only enqueue events.
type Event struct {
	Seq          int64           // database-assigned, monotonic; used as the total-order tiebreak
	Key          string          // unique idempotency key; duplicate enqueues are ignored
	Type         EventType       //
	ProcessAfter time.Time       // event becomes due once now() >= ProcessAfter
	Payload      json.RawMessage // type-specific data, see *Payload structs below
}

// CreateMovementPayload carries the full intended Movement so the loop can write
// it, plus the PlayerID so the loop can re-check ownership authoritatively.
type CreateMovementPayload struct {
	Movement *Movement `json:"movement"`
	PlayerID string    `json:"playerID"`
}

// CancelMovementPayload identifies the movement to cancel. PlayerID is retained
// for authoritative re-checks even though the handler already verified ownership.
type CancelMovementPayload struct {
	MovementID string `json:"movementID"`
	PlayerID   string `json:"playerID"`
}
