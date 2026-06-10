package game

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func createEvent(t *testing.T, seq int64, m *Movement, playerID string) *Event {
	t.Helper()
	p, err := json.Marshal(CreateMovementPayload{Movement: m, PlayerID: playerID})
	if err != nil {
		t.Fatalf("failed to marshal create payload: %v", err)
	}
	return &Event{Seq: seq, Key: "create:" + m.ID, Type: EventCreateMovement, ProcessAfter: time.Now(), Payload: p}
}

func cancelEvent(t *testing.T, seq int64, movementID, playerID string) *Event {
	t.Helper()
	p, err := json.Marshal(CancelMovementPayload{MovementID: movementID, PlayerID: playerID})
	if err != nil {
		t.Fatalf("failed to marshal cancel payload: %v", err)
	}
	return &Event{Seq: seq, Key: "cancel:" + movementID, Type: EventCancelMovement, ProcessAfter: time.Now(), Payload: p}
}

func Test_tick_createMovement(t *testing.T) {
	m := &Movement{ID: "m1", CityFrom: "a", CityTo: "b", Troops: &Troops{}, Resources: &MaterialResources{}}
	event := createEvent(t, 7, m, "p1")

	var created *Movement
	var processedSeq int64 = -1
	mockDB := &mockDatabase{
		GetDueEventsFunc: func(_ time.Time) ([]*Event, error) { return []*Event{event}, nil },
		GetCityFunc:      func(id string) (*City, error) { return makeCity(func(c *City) { c.ID = id; c.PlayerID = "p1" }), nil },
		CreateMovementFunc: func(m *Movement) error { created = m; return nil },
		MarkEventProcessedFunc: func(seq int64) error { processedSeq = seq; return nil },
	}
	g := &GameService{Database: mockDB}

	if err := g.tick(context.Background()); err != nil {
		t.Fatalf("unexpected tick error: %v", err)
	}
	if created == nil || created.ID != "m1" {
		t.Errorf("expected movement m1 to be created, got %+v", created)
	}
	if processedSeq != 7 {
		t.Errorf("expected event 7 to be marked processed, got %v", processedSeq)
	}
}

func Test_tick_createMovement_rejectedWhenNotOwned(t *testing.T) {
	m := &Movement{ID: "m1", CityFrom: "a", CityTo: "b", Troops: &Troops{}, Resources: &MaterialResources{}}
	event := createEvent(t, 7, m, "p1")

	var processedSeq int64 = -1
	mockDB := &mockDatabase{
		GetDueEventsFunc: func(_ time.Time) ([]*Event, error) { return []*Event{event}, nil },
		GetCityFunc:      func(id string) (*City, error) { return makeCity(func(c *City) { c.ID = id; c.PlayerID = "someone-else" }), nil },
		// CreateMovementFunc is nil: a rejected event must not write
		MarkEventProcessedFunc: func(seq int64) error { processedSeq = seq; return nil },
	}
	g := &GameService{Database: mockDB}

	if err := g.tick(context.Background()); err != nil {
		t.Fatalf("unexpected tick error: %v", err)
	}
	if processedSeq != 7 {
		t.Errorf("a rejected event should still be consumed (marked processed), got seq %v", processedSeq)
	}
}

func Test_tick_cancelMovement(t *testing.T) {
	event := cancelEvent(t, 9, "m1", "p1")

	var deleted string
	var processedSeq int64 = -1
	mockDB := &mockDatabase{
		GetDueEventsFunc:       func(_ time.Time) ([]*Event, error) { return []*Event{event}, nil },
		DeleteMovementFunc:     func(id string) error { deleted = id; return nil },
		MarkEventProcessedFunc: func(seq int64) error { processedSeq = seq; return nil },
	}
	g := &GameService{Database: mockDB}

	if err := g.tick(context.Background()); err != nil {
		t.Fatalf("unexpected tick error: %v", err)
	}
	if deleted != "m1" {
		t.Errorf("expected movement m1 to be deleted, got %q", deleted)
	}
	if processedSeq != 9 {
		t.Errorf("expected event 9 to be marked processed, got %v", processedSeq)
	}
}

func Test_tick_infraErrorLeavesEventUnprocessed(t *testing.T) {
	m := &Movement{ID: "m1", CityFrom: "a", CityTo: "b", Troops: &Troops{}, Resources: &MaterialResources{}}
	event := createEvent(t, 7, m, "p1")

	marked := false
	mockDB := &mockDatabase{
		GetDueEventsFunc:       func(_ time.Time) ([]*Event, error) { return []*Event{event}, nil },
		GetCityFunc:            func(id string) (*City, error) { return makeCity(func(c *City) { c.ID = id; c.PlayerID = "p1" }), nil },
		CreateMovementFunc:     func(m *Movement) error { return errors.New("db down") },
		MarkEventProcessedFunc: func(seq int64) error { marked = true; return nil },
	}
	g := &GameService{Database: mockDB}

	if err := g.tick(context.Background()); err == nil {
		t.Fatal("expected tick to return an error on infra failure")
	}
	if marked {
		t.Error("an event must not be marked processed when its effect failed with an infra error")
	}
}
