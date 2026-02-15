package game

import (
	"encoding/json"
	"io"
	"net/http"
)

// TODO: define this in a OpenAPI spec and generate to have consistent structs with server/client.
type City struct {
	ID        string         `json:"id"`
	Name      string         `json:"cityName"`
	Buildings map[string]int `json:"buildings"`
	Resources map[string]int `json:"resources"`
}

func (g *game) GetCity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	res, err := g.db.GetCity("city_123")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error fetching city data: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(res)
}

// TODO: define this in a OpenAPI spec and generate to have consistent structs with server/client.
type UpgradeCityRequest struct {
	ID       string `json:"id"`
	Building string `json:"building"`
}

func (g *game) UpgradeCity(w http.ResponseWriter, r *http.Request) {
	var req UpgradeCityRequest
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1048576)) // limit request body to 1MB to prevent abuse
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error reading request body: " + err.Error()))
		return
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body: " + err.Error()))
		return
	}

	req.ID = "city_123" // for now we only have one test city, so we ignore the ID in the request
	err = g.db.WriteEvent(&event{Name: UpgradeCity, Data: body})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error writing event: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
