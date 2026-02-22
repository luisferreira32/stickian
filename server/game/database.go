package game

type GameDatabase interface {
	GetCity() (*City, error)
}

// InMemoryDatabase is a placeholder for an actual database implementation.
type InMemoryDatabase struct{}

func (db *InMemoryDatabase) GetCity() (*City, error) {
	// This is just a stub. In a real implementation, you would query your database here.
	return &City{
		Name: "Stick City",
		Buildings: Buildings{
			CityHall:    4,
			Farm:        2,
			Quarry:      2,
			Lumbermill:  2,
			CrystalMine: 3,
		},
		Resources: Resources{
			Population: 45,
			Stone:      215,
			Sticks:     312,
			Crystal:    145,
			Gold:       18,
		},
	}, nil
}
