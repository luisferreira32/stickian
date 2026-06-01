package game

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func Test_GetMovements(t *testing.T) {
	// TODO!
}

func Test_CreateMovement(t *testing.T) {
	arrival := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)

	t.Run("enqueues create event and returns 202", func(t *testing.T) {
		var gotEvent *Event
		mockDB := &mockDatabase{
			GetCityFunc: func(id string) (*City, error) {
				return makeCity(func(c *City) { c.ID = id }), nil // owned by "test-user"
			},
			AddEventFunc: func(e *Event) error { gotEvent = e; return nil },
			// CreateMovementFunc is intentionally nil: the handler must NOT write directly
		}
		service := &GameService{Database: mockDB}

		body := fmt.Sprintf(`{"cityFrom":"city-a","cityTo":"city-b","type":0,"arrivalTime":%q}`, arrival)
		req := (&http.Request{
			Method: "POST",
			URL:    &url.URL{Path: "/api/movements"},
			Body:   io.NopCloser(bytes.NewBufferString(body)),
		}).WithContext(context.WithValue(context.Background(), "sub", "test-user"))

		rec := httptest.NewRecorder()
		service.CreateMovement(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("unexpected status: want 202, got %v (%s)", rec.Code, rec.Body.String())
		}
		var resp CreateMovementResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID == "" {
			t.Fatal("expected a movement id in the response")
		}
		if gotEvent == nil {
			t.Fatal("expected an event to be enqueued")
		}
		if gotEvent.Type != EventCreateMovement {
			t.Errorf("unexpected event type: want %v, got %v", EventCreateMovement, gotEvent.Type)
		}
		if gotEvent.Key != "create:"+resp.ID {
			t.Errorf("unexpected event key: want %v, got %v", "create:"+resp.ID, gotEvent.Key)
		}
		var payload CreateMovementPayload
		if err := json.Unmarshal(gotEvent.Payload, &payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		if payload.PlayerID != "test-user" {
			t.Errorf("unexpected player id: want test-user, got %v", payload.PlayerID)
		}
		if payload.Movement == nil || payload.Movement.ID != resp.ID || payload.Movement.CityFrom != "city-a" || payload.Movement.CityTo != "city-b" {
			t.Errorf("unexpected movement in payload: %+v", payload.Movement)
		}
	})

	t.Run("forbidden when origin city not owned", func(t *testing.T) {
		enqueued := false
		mockDB := &mockDatabase{
			GetCityFunc: func(id string) (*City, error) {
				return makeCity(func(c *City) { c.PlayerID = "someone-else" }), nil
			},
			AddEventFunc: func(e *Event) error { enqueued = true; return nil },
		}
		service := &GameService{Database: mockDB}

		body := fmt.Sprintf(`{"cityFrom":"city-a","cityTo":"city-b","type":0,"arrivalTime":%q}`, arrival)
		req := (&http.Request{
			Method: "POST",
			URL:    &url.URL{Path: "/api/movements"},
			Body:   io.NopCloser(bytes.NewBufferString(body)),
		}).WithContext(context.WithValue(context.Background(), "sub", "test-user"))

		rec := httptest.NewRecorder()
		service.CreateMovement(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("unexpected status: want 403, got %v", rec.Code)
		}
		if enqueued {
			t.Error("no event should be enqueued when the request is forbidden")
		}
	})

	t.Run("bad request on invalid body", func(t *testing.T) {
		mockDB := &mockDatabase{
			AddEventFunc: func(e *Event) error { t.Fatal("should not enqueue on invalid body"); return nil },
		}
		service := &GameService{Database: mockDB}

		req := (&http.Request{
			Method: "POST",
			URL:    &url.URL{Path: "/api/movements"},
			Body:   io.NopCloser(bytes.NewBufferString(`{"cityFrom":"city-a"}`)), // missing fields
		}).WithContext(context.WithValue(context.Background(), "sub", "test-user"))

		rec := httptest.NewRecorder()
		service.CreateMovement(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("unexpected status: want 400, got %v", rec.Code)
		}
	})
}

func Test_DeleteMovement(t *testing.T) {
	t.Run("enqueues cancel event and returns 202", func(t *testing.T) {
		var gotEvent *Event
		mockDB := &mockDatabase{
			GetMovementFunc: func(id string) (*Movement, error) {
				return &Movement{ID: id, CityFrom: "city-a", CityTo: "city-b"}, nil
			},
			GetCityFunc: func(id string) (*City, error) {
				return makeCity(func(c *City) { c.ID = id }), nil // owned by "test-user"
			},
			AddEventFunc: func(e *Event) error { gotEvent = e; return nil },
			// DeleteMovementFunc is intentionally nil: the handler must NOT delete directly
		}
		service := &GameService{Database: mockDB}

		req := (&http.Request{
			Method: "DELETE",
			URL:    &url.URL{Path: "/api/movements/mov-1"},
		}).WithContext(context.WithValue(context.Background(), "sub", "test-user"))
		req.SetPathValue("id", "mov-1")

		rec := httptest.NewRecorder()
		service.DeleteMovement(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("unexpected status: want 202, got %v (%s)", rec.Code, rec.Body.String())
		}
		if gotEvent == nil {
			t.Fatal("expected a cancel event to be enqueued")
		}
		if gotEvent.Type != EventCancelMovement {
			t.Errorf("unexpected event type: want %v, got %v", EventCancelMovement, gotEvent.Type)
		}
		if gotEvent.Key != "cancel:mov-1" {
			t.Errorf("unexpected event key: want cancel:mov-1, got %v", gotEvent.Key)
		}
		var payload CancelMovementPayload
		if err := json.Unmarshal(gotEvent.Payload, &payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		if payload.MovementID != "mov-1" || payload.PlayerID != "test-user" {
			t.Errorf("unexpected cancel payload: %+v", payload)
		}
	})

	t.Run("forbidden when movement not owned", func(t *testing.T) {
		enqueued := false
		mockDB := &mockDatabase{
			GetMovementFunc: func(id string) (*Movement, error) {
				return &Movement{ID: id, CityFrom: "city-a", CityTo: "city-b"}, nil
			},
			GetCityFunc: func(id string) (*City, error) {
				return makeCity(func(c *City) { c.PlayerID = "someone-else" }), nil
			},
			AddEventFunc: func(e *Event) error { enqueued = true; return nil },
		}
		service := &GameService{Database: mockDB}

		req := (&http.Request{
			Method: "DELETE",
			URL:    &url.URL{Path: "/api/movements/mov-1"},
		}).WithContext(context.WithValue(context.Background(), "sub", "test-user"))
		req.SetPathValue("id", "mov-1")

		rec := httptest.NewRecorder()
		service.DeleteMovement(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("unexpected status: want 403, got %v", rec.Code)
		}
		if enqueued {
			t.Error("no event should be enqueued when the request is forbidden")
		}
	})
}
