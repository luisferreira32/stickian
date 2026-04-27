//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const (
	testURL = "http://localhost:8080"
)

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	AccessToken string `json:"accessToken"`
}

type JoinWorldRequest struct {
	CityName string `json:"cityName"`
}

type JoinWorldResponse struct {
	CityID string `json:"cityID"`
}

func Test_Hotpath(t *testing.T) {
	t.Log("Starting hotpath test: ensure server and database are running locally...")

	// sign up
	signupReq := &SignupRequest{
		Username: "test user",
		Email:    "test@example.com",
		Password: "a-safe-pw",
	}
	signupReqBytes, err := json.Marshal(signupReq)
	if err != nil {
		t.Fatalf("failed to marshal signup req: %v", err)
	}
	signup, err := http.NewRequestWithContext(t.Context(), "POST", testURL+"/api/signup", bytes.NewReader(signupReqBytes))
	if err != nil {
		t.Fatalf("failed to create signup request: %v", err)
	}
	signupHTTPRsp, err := http.DefaultClient.Do(signup)
	if err != nil {
		t.Fatalf("failed to signup: %v", err)
	}
	signupRspBytes, err := io.ReadAll(signupHTTPRsp.Body)
	if err != nil {
		t.Fatalf("failed to read signup response: %v", err)
	}
	signupRsp := &SignupResponse{}
	err = json.Unmarshal(signupRspBytes, signupRsp)
	if err != nil {
		t.Fatalf("failed to unmarshal signup response: %v, %s", err, signupRspBytes)
	}

	// join world
	joinWorldReq := &JoinWorldRequest{
		CityName: "a test city",
	}
	joinWorldReqBytes, err := json.Marshal(joinWorldReq)
	if err != nil {
		t.Fatalf("failed to marshal join world req: %v", err)
	}
	joinWorld, err := http.NewRequestWithContext(t.Context(), "POST", testURL+"/api/joinworld", bytes.NewReader(joinWorldReqBytes))
	if err != nil {
		t.Fatalf("failed to create signup request: %v", err)
	}
	joinWorld.Header.Add("Authorization", "Bearer "+signupRsp.AccessToken)

	joinWorldHTTPRsp, err := http.DefaultClient.Do(joinWorld)
	if err != nil {
		t.Fatalf("failed to joinWorld: %v", err)
	}
	joinWorldRspBytes, err := io.ReadAll(joinWorldHTTPRsp.Body)
	if err != nil {
		t.Fatalf("failed to read joinWorld response: %v", err)
	}
	joinWorldRsp := &JoinWorldResponse{}
	err = json.Unmarshal(joinWorldRspBytes, joinWorldRsp)
	if err != nil {
		t.Fatalf("failed to unmarshal joinWorld response: %v, %s", err, joinWorldRspBytes)
	}

	// get city
	getCity, err := http.NewRequestWithContext(t.Context(), "GET", testURL+"/api/cities/"+joinWorldRsp.CityID, http.NoBody)
	if err != nil {
		t.Fatalf("failed to create getCity request: %v", err)
	}
	getCity.Header.Add("Authorization", "Bearer "+signupRsp.AccessToken)
	getCityHTTPRsp, err := http.DefaultClient.Do(getCity)
	if err != nil {
		t.Fatalf("failed to getCity: %v", err)
	}
	getCityRspBytes, err := io.ReadAll(getCityHTTPRsp.Body)
	if err != nil {
		t.Fatalf("failed to read getCity response: %v", err)
	}
	t.Logf("got city:\n%s", getCityRspBytes)

	t.Log("Successful hotpath test!")
}
