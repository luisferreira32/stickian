package game

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/luisferreira32/stickian/server/internal/utils"
)

// City defines the structure of a city
//
// TODO: define this in open api spec and generate it from there for
// both server and client, with the additional benefit of api docs.
type City struct {
	ID        string     `json:"id"`
	PlayerID  string     `json:"playerID"`
	Name      string     `json:"cityName"`
	Q         int        `json:"q"`
	R         int        `json:"r"`
	Biome     string     `json:"biome"`
	Points    int        `json:"points"`
	Buildings *Buildings `json:"buildings,omitempty"`
	Resources *Resources `json:"resources,omitempty"`
}

type Buildings struct {
	CityHall    int `json:"cityHall"`
	Embassy     int `json:"embassy"`
	Treasury    int `json:"treasury"`
	Tavern      int `json:"tavern"`
	Farm        int `json:"farm"`
	Lumbermill  int `json:"lumbermill"`
	Quarry      int `json:"quarry"`
	CrystalMine int `json:"crystalMine"`
	Warehouse   int `json:"warehouse"`
	Market      int `json:"market"`
	Harbor      int `json:"harbor"`
	Walls       int `json:"walls"`
	Barracks    int `json:"barracks"`
	Docks       int `json:"docks"`
	SpyGuild    int `json:"spyGuild"`
	Library     int `json:"library"`
	Workshop    int `json:"workshop"`
	Observatory int `json:"observatory"`
	Temple      int `json:"temple"`
	Shrine      int `json:"shrine"`
	Cathedral   int `json:"cathedral"`
}

type Resources struct {
	Food       int `json:"food"`
	Sticks     int `json:"sticks"`
	Stones     int `json:"stones"`
	Gems       int `json:"gems"`
	Population int `json:"population"`
	Faith      int `json:"faith"`
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

// GetCities gets the city at the specified Q, R coordinate.
func (g *GameService) GetCities(w http.ResponseWriter, r *http.Request) {
	q, err := strconv.Atoi(r.URL.Query().Get("q"))
	if err != nil {
		utils.WithError(w, fmt.Errorf("invalid q parameter: %w", err))
		return
	}
	rv, err := strconv.Atoi(r.URL.Query().Get("r"))
	if err != nil {
		utils.WithError(w, fmt.Errorf("invalid r parameter: %w", err))
		return
	}

	city, err := g.Database.GetCities(r.Context(), q, rv)
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
