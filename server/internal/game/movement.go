package game

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/luisferreira32/stickian/server/internal/utils"
)

const (
	MovementTypeAttack    = 0
	MovementTypeReinforce = 1
	MovementTypeTransport = 2
)

// MaterialResources are the resources that can travel with a movement.
// Population and faith are city attributes and cannot be moved.
type MaterialResources struct {
	Food   int `json:"food"`
	Sticks int `json:"sticks"`
	Stones int `json:"stones"`
	Gems   int `json:"gems"`
}

type Movement struct {
	ID          string             `json:"id"`
	CityFrom    string             `json:"cityFrom"`
	CityTo      string             `json:"cityTo"`
	Type        int                `json:"type"`
	ArrivalTime time.Time          `json:"arrivalTime"`
	Troops      *Troops            `json:"troops,omitempty"`
	Resources   *MaterialResources `json:"resources,omitempty"`
}

type CreateMovementRequest struct {
	CityFrom    string             `json:"cityFrom"`
	CityTo      string             `json:"cityTo"`
	Type        int                `json:"type"`
	ArrivalTime time.Time          `json:"arrivalTime"`
	Troops      *Troops            `json:"troops,omitempty"`
	Resources   *MaterialResources `json:"resources,omitempty"`
}

func validCreateMovementRequest(req *CreateMovementRequest) string {
	if req == nil {
		return "must provide a valid request"
	}
	if req.CityFrom == "" {
		return "cityFrom is required"
	}
	if req.CityTo == "" {
		return "cityTo is required"
	}
	if req.CityFrom == req.CityTo {
		return "cityFrom and cityTo must be different"
	}
	if req.Type != MovementTypeAttack && req.Type != MovementTypeReinforce && req.Type != MovementTypeTransport {
		return "type must be 0 (attack), 1 (reinforce), or 2 (transport)"
	}
	if req.ArrivalTime.IsZero() || req.ArrivalTime.Before(time.Now()) {
		return "arrivalTime must be in the future"
	}
	return ""
}

// CreateMovement sends troops and/or resources from one city to another.
// The authenticated user must own the origin city.
func (g *GameService) CreateMovement(w http.ResponseWriter, r *http.Request) {
	bodyReader := http.MaxBytesReader(w, r.Body, utils.MaxRead)
	defer func() {
		_ = bodyReader.Close()
	}()

	req := CreateMovementRequest{}
	if err := json.NewDecoder(bodyReader).Decode(&req); err != nil {
		utils.WithError(w, fmt.Errorf("%w: invalid request body: %w", utils.ErrUserError, err))
		return
	}
	if errReason := validCreateMovementRequest(&req); errReason != "" {
		utils.WithError(w, fmt.Errorf("%w: %s", utils.ErrUserError, errReason))
		return
	}

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, utils.ErrUnauthorized)
		return
	}

	city, err := g.Database.GetCity(r.Context(), req.CityFrom)
	if err != nil {
		utils.WithError(w, err)
		return
	}
	if city.PlayerID != userID {
		utils.WithError(w, utils.ErrForbidden)
		return
	}

	troops := req.Troops
	if troops == nil {
		troops = &Troops{}
	}
	resources := req.Resources
	if resources == nil {
		resources = &MaterialResources{}
	}

	m := &Movement{
		ID:          newCityID(),
		CityFrom:    req.CityFrom,
		CityTo:      req.CityTo,
		Type:        req.Type,
		ArrivalTime: req.ArrivalTime,
		Troops:      troops,
		Resources:   resources,
	}

	if err := g.Database.CreateMovement(r.Context(), m); err != nil {
		utils.WithError(w, err)
		return
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(m); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode movement: %w", err))
		return
	}
}

// GetMovements returns movements for a city. The authenticated user must own the city.
// Query params:
//   - cityId (required): the city to query movements for
//   - direction: "incoming", "outgoing" (default: "outgoing")
func (g *GameService) GetMovements(w http.ResponseWriter, r *http.Request) {
	cityID := r.URL.Query().Get("cityId")
	if cityID == "" {
		utils.WithError(w, fmt.Errorf("%w: cityId query parameter is required", utils.ErrUserError))
		return
	}

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, utils.ErrUnauthorized)
		return
	}

	city, err := g.Database.GetCity(r.Context(), cityID)
	if err != nil {
		utils.WithError(w, err)
		return
	}
	if city.PlayerID != userID {
		utils.WithError(w, utils.ErrForbidden)
		return
	}

	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "outgoing"
	}

	var movements []*Movement
	switch direction {
	case "incoming":
		movements, err = g.Database.GetIncomingMovements(r.Context(), cityID)
	case "outgoing":
		movements, err = g.Database.GetOutgoingMovements(r.Context(), cityID)
	default:
		utils.WithError(w, fmt.Errorf("%w: direction must be 'incoming' or 'outgoing'", utils.ErrUserError))
		return
	}
	if err != nil {
		utils.WithError(w, err)
		return
	}

	if movements == nil {
		movements = []*Movement{}
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(movements); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode movements: %w", err))
		return
	}
}
