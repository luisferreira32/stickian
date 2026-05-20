package game

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/luisferreira32/stickian/server/internal/utils"
)

func Test_CancelMovement(t *testing.T) {
	originCity := makeCity(func(c *City) { c.ID = "city-1" })
	otherCity := makeCity(func(c *City) { c.ID = "city-1"; c.PlayerID = "someone-else" })

	movementID := "mov-1"
	updated := &Movement{
		ID:          movementID,
		CityFrom:    "city-1",
		CityTo:      "city-2",
		IsReturning: true,
		Troops:      &Troops{},
		Resources:   &MaterialResources{},
	}

	testcases := []struct {
		name          string
		movementRes   *Movement
		movementErr   error
		cityRes       *City
		cityErr       error
		cancelRes     *Movement
		cancelErr     error
		wantStatus    int
		wantCancelled bool
	}{
		{
			name:          "success",
			movementRes:   &Movement{ID: movementID, CityFrom: "city-1", Troops: &Troops{}, Resources: &MaterialResources{}},
			cityRes:       originCity,
			cancelRes:     updated,
			wantStatus:    200,
			wantCancelled: true,
		},
		{
			name:        "movement not found",
			movementErr: utils.ErrNotFound,
			wantStatus:  404,
		},
		{
			name:        "different owner",
			movementRes: &Movement{ID: movementID, CityFrom: "city-1", Troops: &Troops{}, Resources: &MaterialResources{}},
			cityRes:     otherCity,
			wantStatus:  403,
		},
		{
			name:          "no longer cancellable",
			movementRes:   &Movement{ID: movementID, CityFrom: "city-1", Troops: &Troops{}, Resources: &MaterialResources{}},
			cityRes:       originCity,
			cancelErr:     fmt.Errorf("%w: movement is no longer cancellable", utils.ErrUserError),
			wantStatus:    400,
			wantCancelled: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cancelCalled := false
			gotCutoff := 0.0
			mockDB := &mockDatabase{
				GetMovementFunc: func(id string) (*Movement, error) {
					return tc.movementRes, tc.movementErr
				},
				GetCityFunc: func(id string) (*City, error) {
					return tc.cityRes, tc.cityErr
				},
				CancelMovementFunc: func(id string, cutoff float64) (*Movement, error) {
					cancelCalled = true
					gotCutoff = cutoff
					return tc.cancelRes, tc.cancelErr
				},
			}
			service := &GameService{Database: mockDB}

			req := (&http.Request{
				Method: "DELETE",
				URL:    &url.URL{Path: "/api/movements/" + movementID},
			}).WithContext(context.WithValue(context.Background(), "sub", "test-user"))
			req.SetPathValue("id", movementID)

			rec := httptest.NewRecorder()
			service.CancelMovement(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d (body: %q)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if cancelCalled != tc.wantCancelled {
				t.Errorf("cancel called = %v, want %v", cancelCalled, tc.wantCancelled)
			}
			if tc.wantCancelled && gotCutoff != CancelCutoffProgress {
				t.Errorf("cutoff = %v, want %v", gotCutoff, CancelCutoffProgress)
			}
			if tc.wantStatus == 200 {
				if diff := cmp.Diff(unsafeToResponseBody(updated), rec.Body.Bytes()); diff != "" {
					t.Errorf("body diff (-want, +got): %v", diff)
				}
			}
		})
	}
}

func Test_GameService_tick(t *testing.T) {
	outbound := &Movement{ID: "out-1", IsReturning: false, Troops: &Troops{}, Resources: &MaterialResources{}}
	returning := &Movement{ID: "ret-1", IsReturning: true, Troops: &Troops{}, Resources: &MaterialResources{}}

	t.Run("dispatches outbound to CompleteArrival and returning to CompleteReturn", func(t *testing.T) {
		arrivalIDs := []string{}
		returnIDs := []string{}
		mockDB := &mockDatabase{
			GetDueMovementsFunc: func(now time.Time) ([]*Movement, error) {
				return []*Movement{outbound, returning}, nil
			},
			CompleteArrivalFunc: func(m *Movement) error {
				arrivalIDs = append(arrivalIDs, m.ID)
				return nil
			},
			CompleteReturnFunc: func(m *Movement) error {
				returnIDs = append(returnIDs, m.ID)
				return nil
			},
		}
		svc := &GameService{Database: mockDB}

		if err := svc.tick(context.Background()); err != nil {
			t.Fatalf("tick err: %v", err)
		}
		if diff := cmp.Diff([]string{"out-1"}, arrivalIDs); diff != "" {
			t.Errorf("arrival ids diff (-want, +got): %v", diff)
		}
		if diff := cmp.Diff([]string{"ret-1"}, returnIDs); diff != "" {
			t.Errorf("return ids diff (-want, +got): %v", diff)
		}
	})

	t.Run("one failing movement does not block the others", func(t *testing.T) {
		processed := []string{}
		mockDB := &mockDatabase{
			GetDueMovementsFunc: func(now time.Time) ([]*Movement, error) {
				return []*Movement{outbound, returning}, nil
			},
			CompleteArrivalFunc: func(m *Movement) error {
				return errors.New("boom")
			},
			CompleteReturnFunc: func(m *Movement) error {
				processed = append(processed, m.ID)
				return nil
			},
		}
		svc := &GameService{Database: mockDB}

		if err := svc.tick(context.Background()); err != nil {
			t.Fatalf("tick should swallow per-movement errors, got: %v", err)
		}
		if diff := cmp.Diff([]string{"ret-1"}, processed); diff != "" {
			t.Errorf("expected returning movement still processed, diff (-want, +got): %v", diff)
		}
	})

	t.Run("GetDueMovements error surfaces", func(t *testing.T) {
		mockDB := &mockDatabase{
			GetDueMovementsFunc: func(now time.Time) ([]*Movement, error) {
				return nil, errors.New("db down")
			},
		}
		svc := &GameService{Database: mockDB}

		if err := svc.tick(context.Background()); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
