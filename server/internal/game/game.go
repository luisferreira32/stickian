package game

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/luisferreira32/stickian/server/internal/utils"
)

var (
	// InitialResources for a new city
	//
	// The objective of these initial resources is that it is enough for the creation
	// of the resource producing buildings to a certain extent
	//
	// TODO: change these resources to match building costs once we define a cost table for building upgrades
	InitialResources = &Resources{
		Food:       100,
		Sticks:     100,
		Stones:     100,
		Gems:       0,
		Population: 10,
		Faith:      0,
	}
)

type GameService struct {
	Database GameDatabase
}

type JoinWorldRequest struct {
	CityName string `json:"cityName"`
}

type JoinWorldResponse struct {
	CityID string `json:"cityID"`
}

func validJoinWorldRequest(req *JoinWorldRequest) string {
	if req == nil {
		return "must provide a valid request"
	}
	if req.CityName == "" {
		return "city name is required"
	}
	return ""
}

// JoinWorld is the first endpoint called once a logged in player wants to start a game.
//
// The endpoint returns the city ID of the newly generated city, or, if the endpoint was called
// multiple times (e.g., network issues), it returns the city ID of the first city created for
// this player.
//
// The first city ID, which is a UUID, will be the player ID such that multiple calls to this endpoint
// do NOT create multiple cities in the world, and instead always return the first created city.
func (g *GameService) JoinWorld(w http.ResponseWriter, r *http.Request) {
	bodyReader := http.MaxBytesReader(w, r.Body, utils.MaxRead)
	defer func() {
		_ = bodyReader.Close()
	}()

	req := JoinWorldRequest{}
	if err := json.NewDecoder(bodyReader).Decode(&req); err != nil {
		utils.WithError(w, fmt.Errorf("%w: invalid request body: %w", utils.ErrUserError, err))
		return
	}
	if errReason := validJoinWorldRequest(&req); errReason != "" {
		utils.WithError(w, fmt.Errorf("%w: %s", utils.ErrUserError, errReason))
		return
	}

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, utils.ErrUserUnauthorized)
		return
	}

	// NOTE: The first city ID, which is a UUID, will be the player ID such that multiple calls to this endpoint
	// do NOT create multiple cities in the world, and instead always return the first created city.
	newCity := &City{
		ID:        userID,
		PlayerID:  userID,
		Name:      req.CityName,
		Q:         0,          // TODO: figure out once we know which spots of the map are buildable where the first city will land
		R:         0,          // TODO: figure out once we know which spots of the map are buildable where the first city will land
		Biome:     "mountain", // TODO: figure out once we know which spots of the map are buildable where the first city will land
		Points:    0,
		Buildings: &Buildings{},
		Resources: InitialResources,
	}

	err := g.Database.CreateCity(r.Context(), newCity)
	if err != nil {
		utils.WithError(w, err)
		return
	}

	rsp := JoinWorldResponse{CityID: userID}
	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(rsp); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode response: %w", err))
		return
	}
}
