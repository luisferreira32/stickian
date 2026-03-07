package game

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

type GameDatabase interface {
	GetCity(id string) (*City, error)
	GetCities(a, b Location) ([]*City, error)
}

// InMemoryDatabase is a placeholder for an actual database implementation.
type InMemoryDatabase struct{}

func (db *InMemoryDatabase) GetCity(_ string) (*City, error) {
	// This is just a stub. In a real implementation, you would query your database here.
	return &City{
		Name: "Stick City",
		Buildings: &Buildings{
			CityHall:    4,
			Farm:        2,
			Quarry:      2,
			Lumbermill:  2,
			CrystalMine: 3,
		},
		Resources: &Resources{
			Population: 45,
			Stone:      215,
			Sticks:     312,
			Crystal:    145,
			Gold:       18,
		},
		Location: &Location{
			X: 10,
			Y: 10,
		},
		PlayerID: "Stickman",
	}, nil
}

// GetCities gets all cities within the specified area defined by two locations (a and b).
func (db *InMemoryDatabase) GetCities(_, _ Location) ([]*City, error) {
	// This is just a stub. In a real implementation, you would query your database here.
	city1 := &City{
		Name: "Stick City",
		Location: &Location{
			X: 10,
			Y: 10,
		},
		PlayerID: "Stickman",
	}
	city2 := &City{
		Name: "Stickville",
		Location: &Location{
			X: 15,
			Y: 11,
		},
		PlayerID: "Stickwoman",
	}
	return []*City{city1, city2}, nil
}
