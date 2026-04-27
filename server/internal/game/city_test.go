package game

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockDatabase struct {
	GetCityFunc   func(id string, userID string) (*City, error)
	GetCitiesFunc func(q1, r1, q2, r2 int) ([]*City, error)
	GetMapFunc    func(minQ, maxQ, minR, maxR int) ([]*MapTile, error)
}

func (db *mockDatabase) GetCity(_ context.Context, id string, userID string) (*City, error) {
	return db.GetCityFunc(id, userID)
}

func (db *mockDatabase) GetCities(_ context.Context, q1, r1, q2, r2 int) ([]*City, error) {
	return db.GetCitiesFunc(q1, r1, q2, r2)
}

func (db *mockDatabase) GetMap(_ context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error) {
	return db.GetMapFunc(minQ, maxQ, minR, maxR)
}

func makeCity(opts ...func(*City)) *City {
	city := &City{
		Name:     "Test City",
		Q:        5,
		R:        5,
		Biome:    "plains",
		Points:   0,
		PlayerID: "Test Player",
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
				GetCityFunc: func(id string, userID string) (*City, error) {
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
			query:   "q1=0&r1=0&q2=10&r2=10",
			mockRes: []*City{city1, city2},
			wantQ1:  0, wantR1: 0, wantQ2: 10, wantR2: 10,
			wantStatus: 200,
			wantBody:   unsafeToResponseBody([]*City{city1, city2}),
		},
		{
			name:       "empty result",
			query:      "q1=0&r1=0&q2=1&r2=1",
			mockRes:    nil,
			wantStatus: 200,
			wantBody:   unsafeToResponseBody([]*City(nil)),
		},
		{
			name:       "missing parameter",
			query:      "q1=0&r1=0&q2=10",
			wantStatus: 400,
			wantBody:   []byte("user error: invalid r2 parameter: missing required parameter: r2\n"),
		},
		{
			name:       "database error",
			query:      "q1=0&r1=0&q2=10&r2=10",
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
			if testcase.mockErr == nil && testcase.query != "q1=0&r1=0&q2=1&r2=1" {
				if gotQ1 != testcase.wantQ1 || gotR1 != testcase.wantR1 || gotQ2 != testcase.wantQ2 || gotR2 != testcase.wantR2 {
					t.Errorf("unexpected coords: want (%v,%v,%v,%v), got (%v,%v,%v,%v)",
						testcase.wantQ1, testcase.wantR1, testcase.wantQ2, testcase.wantR2,
						gotQ1, gotR1, gotQ2, gotR2)
				}
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
