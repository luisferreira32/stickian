package game

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/luisferreira32/stickian/server/internal/utils"
)

// City defines the structure of a city in the game.
//
// TODO: define this in open api spec and generate it from there for
// both server and client, with the additional benefit of api docs.
type City struct {
	Name      string     `json:"cityName"`
	ID        string     `json:"id"`
	Buildings *Buildings `json:"buildings,omitempty"`
	Resources *Resources `json:"resources,omitempty"`
	Location  *Location  `json:"location"`
	PlayerID  string     `json:"playerID"`
}

type Buildings struct {
	CityHall    int `json:"cityHall"`
	Farm        int `json:"farm"`
	Quarry      int `json:"quarry"`
	Lumbermill  int `json:"lumbermill"`
	CrystalMine int `json:"crystalMine"`
}

type Resources struct {
	Population int `json:"population"`
	Stone      int `json:"stone"`
	Sticks     int `json:"sticks"`
	Crystal    int `json:"crystal"`
	Gold       int `json:"gold"`
}

type Location struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// GetCity gets the details of a city by its ID.
func (g *GameService) GetCity(w http.ResponseWriter, r *http.Request) {
	// TODO: validate this city can be viewed by the user making the request
	id := r.PathValue("id")

	city, err := g.Database.GetCity(r.Context(), id)
	if err != nil {
		utils.WithError(w, err)
		return
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(city); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode city: %w", err))
		return
	}
}

// GetCities gets all cities within the specified area defined by two locations (a and b).
//
// This returns a partial city object only with name, location and player.
func (g *GameService) GetCities(w http.ResponseWriter, r *http.Request) {
	// TODO: cities coordinates should follow the chunk request format too
	a := Location{}
	b := Location{}

	// NOTE: if query params are wrong, this is just zero value and won't return meaningful data
	_ = json.Unmarshal([]byte(r.URL.Query().Get("a")), &a)
	_ = json.Unmarshal([]byte(r.URL.Query().Get("b")), &b)

	partialCities, err := g.Database.GetCities(r.Context(), a, b)
	if err != nil {
		utils.WithError(w, err)
		return
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(partialCities); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode cities: %w", err))
		return
	}
}
