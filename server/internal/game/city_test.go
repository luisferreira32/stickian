package game

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type mockDatabase struct {
	GetCityFunc              func(id string) (*City, error)
	GetCitiesFunc            func(q1, r1, q2, r2 int) ([]*City, error)
	CreateCityFunc           func(c *City) error
	GetMapFunc               func(minQ, maxQ, minR, maxR int) ([]*MapTile, error)
	GetNextCitySpotFunc      func() (*MapTile, error)
	GetSettleableTileFunc    func(q, r int) (*MapTile, error)
	CreateMovementFunc       func(m *Movement) error
	GetMovementFunc          func(id string) (*Movement, error)
	GetOutgoingMovementsFunc func(cityID string) ([]*Movement, error)
	GetIncomingMovementsFunc func(cityID string) ([]*Movement, error)
	DeleteMovementFunc       func(id string) error
	AddEventFunc             func(e *Event) error
	GetDueEventsFunc         func(now time.Time) ([]*Event, error)
	MarkEventProcessedFunc   func(seq int64) error
}

func (db *mockDatabase) GetCity(_ context.Context, id string) (*City, error) {
	return db.GetCityFunc(id)
}

func (db *mockDatabase) GetCities(_ context.Context, q1, r1, q2, r2 int) ([]*City, error) {
	return db.GetCitiesFunc(q1, r1, q2, r2)
}

func (db *mockDatabase) CreateCity(_ context.Context, c *City) error {
	return db.CreateCityFunc(c)
}

func (db *mockDatabase) GetMap(_ context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error) {
	return db.GetMapFunc(minQ, maxQ, minR, maxR)
}

func (db *mockDatabase) GetNextCitySpot(_ context.Context) (*MapTile, error) {
	return db.GetNextCitySpotFunc()
}

func (db *mockDatabase) GetSettleableTile(_ context.Context, q, r int) (*MapTile, error) {
	return db.GetSettleableTileFunc(q, r)
}

func (db *mockDatabase) CreateMovement(_ context.Context, m *Movement) error {
	return db.CreateMovementFunc(m)
}

func (db *mockDatabase) GetMovement(_ context.Context, id string) (*Movement, error) {
	return db.GetMovementFunc(id)
}

func (db *mockDatabase) GetOutgoingMovements(_ context.Context, cityID string) ([]*Movement, error) {
	return db.GetOutgoingMovementsFunc(cityID)
}

func (db *mockDatabase) GetIncomingMovements(_ context.Context, cityID string) ([]*Movement, error) {
	return db.GetIncomingMovementsFunc(cityID)
}

func (db *mockDatabase) DeleteMovement(_ context.Context, id string) error {
	return db.DeleteMovementFunc(id)
}

func (db *mockDatabase) AddEvent(_ context.Context, e *Event) error {
	return db.AddEventFunc(e)
}

func (db *mockDatabase) GetDueEvents(_ context.Context, now time.Time) ([]*Event, error) {
	return db.GetDueEventsFunc(now)
}

func (db *mockDatabase) MarkEventProcessed(_ context.Context, seq int64) error {
	return db.MarkEventProcessedFunc(seq)
}

func makeCity(opts ...func(*City)) *City {
	city := &City{
		Name:     "Test City",
		Q:        5,
		R:        5,
		Biome:    0,
		Points:   0,
		PlayerID: "test-user",
		Troops:   &Troops{},
	}
	for _, opt := range opts {
		opt(city)
	}
	return city
}

func unsafeToResponseBody(v any) []byte {
	b, _ := json.Marshal(v)
	// seems like http.Recorder always adds the newline, so copy it here
	return append(b, '\n')
}

func Test_GetCity(t *testing.T) {
	testcases := []struct {
		name       string
		request    *http.Request
		mockRes    *City
		mockErr    error
		wantID     string
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "success",
			request: &http.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/api/cities/123"},
			},
			mockRes:    makeCity(),
			wantStatus: 200,
			wantBody:   unsafeToResponseBody(makeCity()),
		},
		{
			name: "database error",
			request: &http.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/api/cities/123"},
			},
			mockErr:    errors.New("a database error"),
			wantStatus: 500,
			wantBody:   []byte("a database error\n"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			gotid := ""
			// given
			mockDB := &mockDatabase{
				GetCityFunc: func(id string) (*City, error) {
					gotid = id
					return testcase.mockRes, testcase.mockErr
				},
			}
			service := &GameService{Database: mockDB}

			// when
			req := testcase.request.WithContext(context.WithValue(testcase.request.Context(), "sub", "test-user"))
			service.GetCity(rec, req)

			// then
			if testcase.wantID != gotid {
				t.Errorf("unexpected id: want %v, got %v", testcase.wantID, gotid)
			}
			if testcase.wantStatus != rec.Code {
				t.Errorf("unexpected status code: want %v, got %v", testcase.wantStatus, rec.Code)
			}
			if diff := cmp.Diff(testcase.wantBody, rec.Body.Bytes()); diff != "" {
				t.Errorf("unexpected city diff (-want, +got): %v", diff)
			}
		})
	}
}

func Test_GetCities(t *testing.T) {
	city1 := makeCity(func(c *City) { c.Q = 5; c.R = 5 })
	city2 := makeCity(func(c *City) { c.Q = 8; c.R = 3 })

	testcases := []struct {
		name       string
		query      string
		mockRes    []*City
		mockErr    error
		wantQ1     int
		wantR1     int
		wantQ2     int
		wantR2     int
		wantStatus int
		wantBody   []byte
	}{
		{
			name:    "success",
			query:   "data=eyJxMSI6MCwgInIxIjowLCAicTIiOjEwLCAicjIiOjEwfQo=", // base64 for {"q1":0, "r1":0, "q2":10, "r2":10}
			mockRes: []*City{city1, city2},
			wantQ1:  0, wantR1: 0, wantQ2: 10, wantR2: 10,
			wantStatus: 200,
			wantBody:   unsafeToResponseBody([]*City{city1, city2}),
		},
		{
			name:       "empty result",
			query:      "data=eyJxMSI6MCwgInIxIjowLCAicTIiOjAsICJyMiI6MH0K", // base64 for {"q1":0, "r1":0, "q2":0, "r2":0}
			mockRes:    nil,
			wantStatus: 200,
			wantBody:   unsafeToResponseBody([]*City(nil)),
		},
		{
			name:       "missing parameter",
			query:      "data=",
			wantStatus: 400,
			wantBody:   []byte("user error: invalid request data: missing query parameter data\n"),
		},
		{
			name:       "database error",
			query:      "data=eyJxMSI6MCwgInIxIjowLCAicTIiOjAsICJyMiI6MH0K", // base64 for {"q1":0, "r1":0, "q2":0, "r2":0}
			mockErr:    errors.New("a database error"),
			wantStatus: 500,
			wantBody:   []byte("a database error\n"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			gotQ1, gotR1, gotQ2, gotR2 := 0, 0, 0, 0
			// given
			mockDB := &mockDatabase{
				GetCitiesFunc: func(q1, r1, q2, r2 int) ([]*City, error) {
					gotQ1, gotR1, gotQ2, gotR2 = q1, r1, q2, r2
					return testcase.mockRes, testcase.mockErr
				},
			}
			service := &GameService{Database: mockDB}
			req := &http.Request{
				Method: "GET",
				URL:    &url.URL{Path: "/api/cities", RawQuery: testcase.query},
			}

			// when
			service.GetCities(rec, req)

			// then
			if gotQ1 != testcase.wantQ1 || gotR1 != testcase.wantR1 || gotQ2 != testcase.wantQ2 || gotR2 != testcase.wantR2 {
				t.Errorf("unexpected coords: want (%v,%v,%v,%v), got (%v,%v,%v,%v)", testcase.wantQ1, testcase.wantR1, testcase.wantQ2, testcase.wantR2, gotQ1, gotR1, gotQ2, gotR2)
			}
			if testcase.wantStatus != rec.Code {
				t.Errorf("unexpected status code: want %v, got %v", testcase.wantStatus, rec.Code)
			}
			if diff := cmp.Diff(testcase.wantBody, rec.Body.Bytes()); diff != "" {
				t.Errorf("unexpected body diff (-want, +got): %v", diff)
			}
		})
	}
}

func Test_FoundCity(t *testing.T) {
	tile := &MapTile{Q: 3, R: 7, Biome: 2}

	testcases := []struct {
		name          string
		body          string
		tileRes       *MapTile
		tileErr       error
		createErr     error
		wantStatus    int
		wantBody      []byte
		wantCityHall  int
		wantResources *Resources
	}{
		{
			name:          "success with resources",
			body:          `{"cityName":"New Town","q":3,"r":7,"food":50,"sticks":30,"stones":20,"gems":5}`,
			tileRes:       tile,
			wantStatus:    200,
			wantCityHall:  1,
			wantResources: &Resources{Food: 50, Sticks: 30, Stones: 20, Gems: 5},
		},
		{
			name:          "success without resources",
			body:          `{"cityName":"New Town","q":3,"r":7}`,
			tileRes:       tile,
			wantStatus:    200,
			wantCityHall:  1,
			wantResources: &Resources{},
		},
		{
			name:       "missing city name",
			body:       `{"q":3,"r":7}`,
			wantStatus: 400,
			wantBody:   []byte("user error: city name is required\n"),
		},
		{
			name:       "negative resources",
			body:       `{"cityName":"New Town","q":3,"r":7,"food":-1}`,
			wantStatus: 400,
			wantBody:   []byte("user error: resources cannot be negative\n"),
		},
		{
			name:       "tile not available",
			body:       `{"cityName":"New Town","q":3,"r":7}`,
			tileErr:    errors.New("user error: tile is not available for settlement"),
			wantStatus: 500,
			wantBody:   []byte("user error: tile is not available for settlement\n"),
		},
		{
			name:       "database error on create",
			body:       `{"cityName":"New Town","q":3,"r":7}`,
			tileRes:    tile,
			createErr:  errors.New("a database error"),
			wantStatus: 500,
			wantBody:   []byte("a database error\n"),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			var gotCity *City
			mockDB := &mockDatabase{
				GetSettleableTileFunc: func(q, r int) (*MapTile, error) {
					return testcase.tileRes, testcase.tileErr
				},
				CreateCityFunc: func(c *City) error {
					gotCity = c
					return testcase.createErr
				},
			}
			service := &GameService{Database: mockDB}
			req := &http.Request{
				Method: "POST",
				URL:    &url.URL{Path: "/api/cities"},
				Body:   http.NoBody,
			}
			req.Body = http.NoBody
			if testcase.body != "" {
				req.Body = io.NopCloser(bytes.NewBufferString(testcase.body))
			}
			req = req.WithContext(context.WithValue(req.Context(), "sub", "test-user"))

			service.FoundCity(rec, req)

			if testcase.wantStatus != rec.Code {
				t.Errorf("unexpected status code: want %v, got %v", testcase.wantStatus, rec.Code)
			}
			if testcase.wantBody != nil {
				if diff := cmp.Diff(testcase.wantBody, rec.Body.Bytes()); diff != "" {
					t.Errorf("unexpected body diff (-want, +got): %v", diff)
				}
			}
			if testcase.wantCityHall != 0 && gotCity != nil {
				if gotCity.Buildings.CityHall != testcase.wantCityHall {
					t.Errorf("unexpected city hall level: want %v, got %v", testcase.wantCityHall, gotCity.Buildings.CityHall)
				}
				if diff := cmp.Diff(testcase.wantResources, gotCity.Resources); diff != "" {
					t.Errorf("unexpected resources diff (-want, +got): %v", diff)
				}
				if diff := cmp.Diff(&Troops{}, gotCity.Troops); diff != "" {
					t.Errorf("unexpected troops diff (-want, +got): %v", diff)
				}
			}
		})
	}
}
