package game

import (
	"encoding/json"
	"net/http"

	"github.com/luisferreira32/stickian/server/httputils"
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

	city, err := g.Database.GetCity(id)
	if err != nil {
		// TODO: if city is not found return a 404
		// TODO: do a better translation of internal errors to codes and do logging instead of returning to client
		http.Error(w, "failed to get city: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httputils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(city); err != nil {
		http.Error(w, "failed to encode city", http.StatusInternalServerError)
		return
	}
}

// GetCities gets all cities within the specified area defined by two locations (a and b).
//
// This returns a partial city object only with name, location and player.
func (g *GameService) GetCities(w http.ResponseWriter, r *http.Request) {
	a := Location{}
	b := Location{}

	// NOTE: if query params are wrong, this is just zero value and won't return meaningful data
	_ = json.Unmarshal([]byte(r.URL.Query().Get("a")), &a)
	_ = json.Unmarshal([]byte(r.URL.Query().Get("b")), &b)

	partialCities, err := g.Database.GetCities(a, b)
	if err != nil {
		http.Error(w, "failed to get cities", http.StatusInternalServerError)
		return
	}

	httputils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(partialCities); err != nil {
		http.Error(w, "failed to encode cities", http.StatusInternalServerError)
		return
	}
}
