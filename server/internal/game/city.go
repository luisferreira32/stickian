package game

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	Biome     int        `json:"biome"`
	Points    int        `json:"points"`
	Buildings *Buildings `json:"buildings,omitempty"`
	Resources *Resources `json:"resources,omitempty"`
	Troops    *Troops    `json:"troops,omitempty"`
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

type Troops struct {
	Swordsmen int `json:"swordsmen"`
	Archers   int `json:"archers"`
	Cavalry   int `json:"cavalry"`
	Ships     int `json:"ships"`
	Spies     int `json:"spies"`
}

// GetCity gets the details of a city by its ID.
func (g *GameService) GetCity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, utils.ErrUnauthorized)
		return
	}

	city, err := g.Database.GetCity(r.Context(), id)
	if err != nil {
		utils.WithError(w, err)
		return
	}

	if city.PlayerID != userID {
		utils.WithError(w, utils.ErrForbidden)
		return
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(city); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode city: %w", err))
		return
	}
}

type GetCitiesRequest struct {
	Q1 int `json:"q1"`
	R1 int `json:"r1"`
	Q2 int `json:"q2"`
	R2 int `json:"r2"`
}

// GetCities returns the city table rows for all cities whose coordinates lie
// within the bounding box defined by vertices (q1, r1) and (q2, r2).
// Buildings and Resources are not included in the response.
func (g *GameService) GetCities(w http.ResponseWriter, r *http.Request) {
	req, err := utils.LoadBase64QueryParam[GetCitiesRequest](r)
	if err != nil {
		utils.WithError(w, fmt.Errorf("%w: invalid request data: %w", utils.ErrUserError, err))
		return
	}

	cities, err := g.Database.GetCities(r.Context(), req.Q1, req.R1, req.Q2, req.R2)
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
