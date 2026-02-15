package game

import (
	"encoding/json"
	"fmt"
	"time"
)

type eventType int

type event struct {
	Key  string
	Type eventType
	Time time.Time
	Data json.RawMessage
}

// The list of all possible game events.
//
// Each event should have sufficient self-contained information such that partial tick
// processing is possible and a retry of the same event won't duplicate the state changes.
// The order of the iota is important as it defines the priority of events in case of
// conflicting timestamps.
const (
	UpgradeBuildingComplete eventType = iota + 1
	UpgradeBuilding
)

type UpgradeBuildingCompleteEvent struct {
	CityID   string `json:"cityID"`
	Building string `json:"building"`
	Level    int    `json:"level"`
}

func (e *UpgradeBuildingCompleteEvent) toEvent() (*event, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return &event{
		Key:  fmt.Sprintf("%s-%s-%d", e.CityID, e.Building, e.Level),
		Type: UpgradeBuildingComplete,
		Data: data,
	}, nil
}

type UpgradeBuildingEvent struct {
	CityID   string `json:"cityID"`
	Building string `json:"building"`
	Level    int    `json:"level"`
}

func (e *UpgradeBuildingEvent) toEvent() (*event, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return &event{
		// generate idempotent key to remove duplicates on requests
		Key:  fmt.Sprintf("%s-%s-%d", e.CityID, e.Building, e.Level),
		Type: UpgradeBuilding,
		Data: data,
	}, nil
}
