package game

import (
	"encoding/json"
	"net/http"
)

// City defines the structure of a city in the game.
//
// TODO: define this in open api spec and generate it from there for
// both server and client, with the additional benefit of api docs.
type City struct {
	Name      string    `json:"cityName"`
	Buildings Buildings `json:"buildings"`
	Resources Resources `json:"resources"`
}

type Buildings struct {
	CityHall    int `json:"city_Hall"`
	Farm        int `json:"farm"`
	Quarry      int `json:"quarry"`
	Lumbermill  int `json:"lumbermill"`
	CrystalMine int `json:"crystal_Mine"`
}

type Resources struct {
	Population int `json:"population"`
	Stone      int `json:"stone"`
	Sticks     int `json:"sticks"`
	Crystal    int `json:"crystal"`
	Gold       int `json:"gold"`
}

func (g *GameService) GetCity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(http.StatusOK)
	city, err := g.Database.GetCity()
	if err != nil {
		http.Error(w, "failed to get city", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(city); err != nil {
		http.Error(w, "failed to encode city", http.StatusInternalServerError)
		return
	}
}
