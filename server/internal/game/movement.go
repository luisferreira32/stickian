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

// CreateMovementResponse is returned with 202 Accepted. The movement is created
// asynchronously by the game loop; the id lets the client track it via GetMovements.
type CreateMovementResponse struct {
	ID string `json:"id"`
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

	payload, err := json.Marshal(CreateMovementPayload{Movement: m, PlayerID: userID})
	if err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode event payload: %w", err))
		return
	}
	event := &Event{
		Key:          "create:" + m.ID,
		Type:         EventCreateMovement,
		ProcessAfter: time.Now(),
		Payload:      payload,
	}
	if err := g.Database.AddEvent(r.Context(), event); err != nil {
		utils.WithError(w, err)
		return
	}

	// the movement is written by the game loop, not here; respond 202 with the
	// pre-generated id so the client can track the pending movement via GetMovements
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(CreateMovementResponse{ID: m.ID}); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode response: %w", err))
		return
	}
}

// GetMovement returns a single movement by ID. The authenticated user must own the origin city.
func (g *GameService) GetMovement(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, utils.ErrUnauthorized)
		return
	}

	m, err := g.Database.GetMovement(r.Context(), id)
	if err != nil {
		utils.WithError(w, err)
		return
	}

	city, err := g.Database.GetCity(r.Context(), m.CityFrom)
	if err != nil {
		utils.WithError(w, err)
		return
	}
	if city.PlayerID != userID {
		utils.WithError(w, utils.ErrForbidden)
		return
	}

	utils.WithDefaultOKHeaders(w)
	if err := json.NewEncoder(w).Encode(m); err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode movement: %w", err))
		return
	}
}

// DeleteMovement recalls (cancels) a movement. The authenticated user must own the origin city.
func (g *GameService) DeleteMovement(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, utils.ErrUnauthorized)
		return
	}

	m, err := g.Database.GetMovement(r.Context(), id)
	if err != nil {
		utils.WithError(w, err)
		return
	}

	city, err := g.Database.GetCity(r.Context(), m.CityFrom)
	if err != nil {
		utils.WithError(w, err)
		return
	}
	if city.PlayerID != userID {
		utils.WithError(w, utils.ErrForbidden)
		return
	}

	payload, err := json.Marshal(CancelMovementPayload{MovementID: id, PlayerID: userID})
	if err != nil {
		utils.WithError(w, fmt.Errorf("failed to encode event payload: %w", err))
		return
	}
	event := &Event{
		Key:          "cancel:" + id,
		Type:         EventCancelMovement,
		ProcessAfter: time.Now(),
		Payload:      payload,
	}
	if err := g.Database.AddEvent(r.Context(), event); err != nil {
		utils.WithError(w, err)
		return
	}

	// the movement is removed by the game loop, not here
	w.WriteHeader(http.StatusAccepted)
}

type GetMovementsRequest struct {
	CityID    string `json:"cityId"`
	Direction string `json:"direction"` // "incoming" or "outgoing"
}

func validGetMovementsRequest(req *GetMovementsRequest) string {
	if req.CityID == "" {
		return "cityId is required"
	}
	if req.Direction != "incoming" && req.Direction != "outgoing" {
		return "direction must be 'incoming' or 'outgoing'"
	}
	return ""
}

// GetMovements returns movements for a city. The authenticated user must own the city.
func (g *GameService) GetMovements(w http.ResponseWriter, r *http.Request) {
	req, err := utils.LoadBase64QueryParam[GetMovementsRequest](r)
	if err != nil {
		utils.WithError(w, fmt.Errorf("%w: invalid request data: %w", utils.ErrUserError, err))
		return
	}
	if req.Direction == "" {
		req.Direction = "outgoing"
	}
	if errReason := validGetMovementsRequest(&req); errReason != "" {
		utils.WithError(w, fmt.Errorf("%w: %s", utils.ErrUserError, errReason))
		return
	}

	userID, ok := r.Context().Value("sub").(string)
	if !ok || userID == "" {
		utils.WithError(w, utils.ErrUnauthorized)
		return
	}

	city, err := g.Database.GetCity(r.Context(), req.CityID)
	if err != nil {
		utils.WithError(w, err)
		return
	}
	if city.PlayerID != userID {
		utils.WithError(w, utils.ErrForbidden)
		return
	}

	var movements []*Movement
	switch req.Direction {
	case "incoming":
		movements, err = g.Database.GetIncomingMovements(r.Context(), req.CityID)
	case "outgoing":
		movements, err = g.Database.GetOutgoingMovements(r.Context(), req.CityID)
	default:
		utils.WithError(w, fmt.Errorf("unexpected direction: %v", req.Direction))
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
