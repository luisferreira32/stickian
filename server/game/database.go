package game

import (
	"encoding/json"
	"time"
)

type event struct {
	Name eventName
	Time time.Time
	Data json.RawMessage
}

type database interface {
	WriteEvent(e *event) error
	GetEvents() ([]*event, error)
	WriteCity(c *City) error
	GetCity(id string) (*City, error)
}

// NewInMemoryDatabase returns a new instance of an in-memory database.
//
// This must only be used for development and testing purposes.
func NewInMemoryDatabase() *inMemoryDatabase {
	i := &inMemoryDatabase{
		cities:  make(map[string]*City),
		events:  make([]*event, 0),
		timeNow: time.Now,
	}

	// setup some hardcoded mock data for now
	c := City{
		ID:   "city_123",
		Name: "Stick City",
		Buildings: map[string]int{
			"city_Hall":    4,
			"farm":         2,
			"quarry":       2,
			"lumbermill":   2,
			"crystal_Mine": 3,
		},
		Resources: map[string]int{
			"population": 45,
			"stone":      215,
			"sticks":     312,
			"crystal":    145,
			"gold":       18,
		},
	}
	_ = i.WriteCity(&c)

	return i
}

// inMemoryDatabase implements the interface for a simple in-memory database.
type inMemoryDatabase struct {
	cities map[string]*City
	events []*event

	// defined as a field to mock in unit tests
	timeNow func() time.Time
}

func (db *inMemoryDatabase) WriteEvent(e *event) error {
	e.Time = db.timeNow() // set the event time on the database level to ensure consistency across events
	db.events = append(db.events, e)
	return nil
}

func (db *inMemoryDatabase) GetEvents() ([]*event, error) {
	return db.events, nil
}

func (db *inMemoryDatabase) WriteCity(c *City) error {
	db.cities[c.ID] = c
	return nil
}

func (db *inMemoryDatabase) GetCity(id string) (*City, error) {
	return db.cities[id], nil
}
