package game

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TODO: define this in a OpenAPI spec and generate to have consistent structs with server/client.
type City struct {
	ID             string              `json:"id"`
	Name           string              `json:"cityName"`
	Buildings      map[string]int      `json:"buildings"`
	Resources      map[string]int      `json:"resources"`
	BuildingsQueue []BuildingQueueItem `json:"buildingsQueue"`
}

type BuildingQueueItem struct {
	Building     string    `json:"building"`
	Level        int       `json:"level"`
	CompleteTime time.Time `json:"completeTime"`
}

func (g *game) GetCity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// TODO: use cityID from path or query params instead of hardcoding
	res, err := g.db.GetCity("city_123")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error fetching city data: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(res)
}

// TODO: define this in a OpenAPI spec and generate to have consistent structs with server/client.
type UpgradeBuildingRequest struct {
	CityID   string `json:"cityID"`
	Building string `json:"building"`
	Level    int    `json:"level"`
}

func validateUpgradeBuildingRequest(req *UpgradeBuildingRequest) error {
	if req.CityID == "" || req.Building == "" || req.Level <= 0 {
		return fmt.Errorf("missing or invalid fields in request body")
	}
	if _, ok := buildingUpgradeCosts[req.Building][req.Level]; !ok {
		return fmt.Errorf("invalid building or level")
	}
	return nil
}

func (g *game) UpgradeBuilding(w http.ResponseWriter, r *http.Request) {
	req := &UpgradeBuildingRequest{}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1048576)).Decode(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body: " + err.Error()))
		return
	}

	// basic request validation: the actual resource calculation will only be done in the game loop
	if err := validateUpgradeBuildingRequest(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	upgradeEvent := &UpgradeBuildingEvent{
		CityID:   req.CityID,
		Building: req.Building,
		Level:    req.Level,
	}
	event, err := upgradeEvent.toEvent()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error creating event: " + err.Error()))
		return
	}
	event.Time = g.timeNow() // set the event time on the server side to ensure consistency across events
	event.Tick = g.currentTick()

	err = g.db.WriteEvent(event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error writing event: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
