package game

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/luisferreira32/stickian/server/internal/utils"
)

func newCityID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

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
	Database     GameDatabase
	TickDuration time.Duration

	settleLock sync.Mutex
}

// Run drives the movement queue. Every TickDuration it asks the database for
// movements whose arrival_time has passed and dispatches each to the right
// completion: outbound movements run arrival effects (TBD) and returning
// movements refund troops and resources to the origin city. Errors processing
// one movement do not block the rest; the row stays in the table and the next
// tick will retry it.
func (g *GameService) Run(ctx context.Context) {
	if g.TickDuration <= 0 {
		g.TickDuration = time.Second
	}
	ticker := time.NewTicker(g.TickDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := g.tick(ctx); err != nil {
				log.Printf("game tick err: %v", err)
			}
		}
	}
}

func (g *GameService) tick(ctx context.Context) error {
	due, err := g.Database.GetDueMovements(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("get due movements: %w", err)
	}

	for _, m := range due {
		var err error
		if m.IsReturning {
			err = g.Database.CompleteReturn(ctx, m)
		} else {
			err = g.Database.CompleteArrival(ctx, m)
		}
		if err != nil {
			log.Printf("process movement %s (returning=%v): %v", m.ID, m.IsReturning, err)
			continue
		}
	}
	return nil
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
		utils.WithError(w, utils.ErrUnauthorized)
		return
	}

	// NOTE: The first city ID, which is a UUID, will be the player ID such that multiple calls to this endpoint
	// do NOT create multiple cities in the world, and instead always return the first created city.
	newCity := &City{
		ID:        userID,
		PlayerID:  userID,
		Name:      req.CityName,
		Points:    0,
		Buildings: &Buildings{},
		Resources: InitialResources,
		Troops:    &Troops{},
	}

	// HACK: to avoid collisions we can have a shared lock on our monolith server!
	var (
		citySpot *MapTile
		err      error
	)
	func() {
		g.settleLock.Lock()
		defer g.settleLock.Unlock()
		citySpot, err = g.Database.GetNextCitySpot(r.Context())
	}()
	if err != nil {
		utils.WithError(w, fmt.Errorf("failed to get next city spot: %w", err))
		return
	}
	newCity.Q = citySpot.Q
	newCity.R = citySpot.R
	newCity.Biome = citySpot.Biome

	err = g.Database.CreateCity(r.Context(), newCity)
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

type FoundCityRequest struct {
	CityName string `json:"cityName"`
	Q        int    `json:"q"`
	R        int    `json:"r"`
	Food     int    `json:"food"`
	Sticks   int    `json:"sticks"`
	Stones   int    `json:"stones"`
	Gems     int    `json:"gems"`
}

type FoundCityResponse struct {
	CityID string `json:"cityID"`
}

func validFoundCityRequest(req *FoundCityRequest) string {
	if req == nil {
		return "must provide a valid request"
	}
	if req.CityName == "" {
		return "city name is required"
	}
	if req.Food < 0 || req.Sticks < 0 || req.Stones < 0 || req.Gems < 0 {
		return "resources cannot be negative"
	}
	return ""
}

// FoundCity creates a new city at the specified hex coordinates for the authenticated player.
//
// The tile must be settleable and unoccupied. The city starts with CityHall at level 1
// and all other buildings at 0. Material resources (food, sticks, stones, gems) brought
// to the city are set as specified in the request; population and faith start at 0.
//
// TODO: support sending units with the founding party once the units system is designed.
func (g *GameService) FoundCity(w http.ResponseWriter, r *http.Request) {
	bodyReader := http.MaxBytesReader(w, r.Body, utils.MaxRead)
	defer func() {
		_ = bodyReader.Close()
	}()

	req := FoundCityRequest{}
	if err := json.NewDecoder(bodyReader).Decode(&req); err != nil {
		utils.WithError(w, fmt.Errorf("%w: invalid request body: %w", utils.ErrUserError, err))
		return
	}
	if errReason := validFoundCityRequest(&req); errReason != "" {
		utils.WithError(w, fmt.Errorf("%w: %s", utils.ErrUserError, errReason))
		return
	}

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, utils.ErrUnauthorized)
		return
	}

	var (
		tile *MapTile
		err  error
	)
	func() {
		g.settleLock.Lock()
		defer g.settleLock.Unlock()
		tile, err = g.Database.GetSettleableTile(r.Context(), req.Q, req.R)
	}()
	if err != nil {
		utils.WithError(w, err)
		return
	}

	cityID := newCityID()
	newCity := &City{
		ID:       cityID,
		PlayerID: userID,
		Name:     req.CityName,
		Q:        tile.Q,
		R:        tile.R,
		Biome:    tile.Biome,
		Points:   0,
		Buildings: &Buildings{
			CityHall: 1,
		},
		Resources: &Resources{
			Food:   req.Food,
			Sticks: req.Sticks,
			Stones: req.Stones,
			Gems:   req.Gems,
		},
		Troops: &Troops{},
	}

	if err := g.Database.CreateCity(r.Context(), newCity); err != nil {
		utils.WithError(w, err)
		return
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(FoundCityResponse{CityID: cityID}); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode response: %w", err))
		return
	}
}
