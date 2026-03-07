package game

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockDatabase struct {
	GetCityFunc   func(id string) (*City, error)
	GetCitiesFunc func(a, b Location) ([]*City, error)
}

func (db *mockDatabase) GetCity(id string) (*City, error) {
	return db.GetCityFunc(id)
}

func (db *mockDatabase) GetCities(a, b Location) ([]*City, error) {
	return db.GetCitiesFunc(a, b)
}

func makeCity(opts ...func(*City)) *City {
	city := &City{
		Name:     "Test City",
		Location: &Location{X: 5, Y: 5},
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
				GetCityFunc: func(id string) (*City, error) {
					gotid = id
					return testcase.mockRes, testcase.mockErr
				},
			}
			service := &GameService{Database: mockDB}

			// when
			service.GetCity(rec, testcase.request)

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
