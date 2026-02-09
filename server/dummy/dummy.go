package dummy

import (
	"encoding/json"
	"io"
	"net/http"
)

// Echo is an example handler that simply echoes back the request body.
func Echo(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	_, _ = w.Write(append(b, []byte("\n")...))
}

// Panic is an example handler that will panic when called.
func Panic(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	panic("this is a panic: " + string(b))
}

// Hello is an example handler that only answers to GET requests with "hello world".
func Hello(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello world\n"))
}

func City(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(http.StatusOK)

	//Json Example sent. Mess with it to confirm the React App receives it :D

	json.NewEncoder(w).Encode(map[string]any{
		"cityName": "Stick City",
		"buildings": map[string]int{
			"city_Hall":    4,
			"farm":         2,
			"quarry":       2,
			"lumbermill":   2,
			"crystal_Mine": 3,
		},
	})
}
