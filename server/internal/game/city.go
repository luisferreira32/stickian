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
	id := r.PathValue("id")

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, fmt.Errorf("unauthorized"))
		return
	}

	city, err := g.Database.GetCity(r.Context(), id, userID)
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

// GetCities returns the city table rows for all cities whose coordinates lie
// within the bounding box defined by vertices (q1, r1) and (q2, r2).
// Buildings and Resources are not included in the response.
func (g *GameService) GetCities(w http.ResponseWriter, r *http.Request) {
	parseIntParam := func(name string) (int, error) {
		v := r.URL.Query().Get(name)
		if v == "" {
			return 0, fmt.Errorf("missing required parameter: %s", name)
		}
		return strconv.Atoi(v)
	}

	q1, err := parseIntParam("q1")
	if err != nil {
		utils.WithError(w, fmt.Errorf("invalid q1 parameter: %w", err))
		return
	}
	r1, err := parseIntParam("r1")
	if err != nil {
		utils.WithError(w, fmt.Errorf("invalid r1 parameter: %w", err))
		return
	}
	q2, err := parseIntParam("q2")
	if err != nil {
		utils.WithError(w, fmt.Errorf("invalid q2 parameter: %w", err))
		return
	}
	r2, err := parseIntParam("r2")
	if err != nil {
		utils.WithError(w, fmt.Errorf("invalid r2 parameter: %w", err))
		return
	}

	cities, err := g.Database.GetCities(r.Context(), q1, r1, q2, r2)
	if err != nil {
		utils.WithError(w, err)
		return
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(cities); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode cities: %w", err))
		return
	}
}
