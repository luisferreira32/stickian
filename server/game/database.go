package game

import (
	"sync"
)

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
		cities: make(map[string]*City),
		events: make(map[string]*event),
	}

	// setup some hardcoded mock data for now
	c := City{
		ID:   "city_123",
		Name: "Stick City",
		Buildings: map[string]int{
			CityHall:    4,
			Farm:        2,
			Quarry:      2,
			LumberMill:  2,
			CrystalMine: 3,
		},
		Resources: map[string]int{
			Population: 645,
			Stone:      2115,
			Sticks:     3342,
			Crystal:    1455,
			Gold:       1812,
		},
	}
	_ = i.WriteCity(&c)

	return i
}

// inMemoryDatabase implements the interface for a simple in-memory database.
type inMemoryDatabase struct {
	cities map[string]*City
	events map[string]*event
	evLock sync.Mutex
}

func (db *inMemoryDatabase) WriteEvent(e *event) error {
	db.evLock.Lock()
	defer db.evLock.Unlock()
	db.events[e.Key] = e
	return nil
}

func (db *inMemoryDatabase) GetEvents() ([]*event, error) {
	db.evLock.Lock()
	defer db.evLock.Unlock()

	events := make([]*event, 0, len(db.events))
	for _, e := range db.events {
		events = append(events, e)
	}
	return events, nil
}

func (db *inMemoryDatabase) WriteCity(c *City) error {
	db.cities[c.ID] = c
	return nil
}

func (db *inMemoryDatabase) GetCity(id string) (*City, error) {
	return db.cities[id], nil
}
