package game

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound = errors.New("not found")
)

type MapTile struct {
	Q     int
	R     int
	Biome int
}

type GameDatabase interface {
	GetCity(ctx context.Context, id string) (*City, error)
	GetCities(ctx context.Context, a, b Location) ([]*City, error)
	GetMap(ctx context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error)
}

// InMemoryDatabase is a placeholder for an actual database implementation.
type InMemoryDatabase struct{}

func (db *InMemoryDatabase) GetCity(_ context.Context, _ string) (*City, error) {
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
func (db *InMemoryDatabase) GetCities(_ context.Context, _, _ Location) ([]*City, error) {
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

func (db *InMemoryDatabase) GetMap(_ context.Context, _, _, _, _ int) ([]*MapTile, error) {
	// This is just a stub.
	return []*MapTile{}, nil
}

type PostgresDatabase struct {
	DB *pgx.Conn
}

func (db *PostgresDatabase) GetCity(_ context.Context, _ string) (*City, error) {
	// TODO: implement this
	return nil, errors.New("not implemented")
}

func (db *PostgresDatabase) GetCities(_ context.Context, _, _ Location) ([]*City, error) {
	// TODO: implement this
	return nil, errors.New("not implemented")
}

const getMapQuery = "SELECT q, r, biome FROM world WHERE q BETWEEN $1 AND $2 AND r BETWEEN $3 AND $4"

func (db *PostgresDatabase) GetMap(ctx context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error) {
	rows, err := db.DB.Query(ctx, getMapQuery, minQ, maxQ, minR, maxR)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiles []*MapTile
	for rows.Next() {
		var t MapTile
		err := rows.Scan(&t.Q, &t.R, &t.Biome)
		if err != nil {
			return nil, err
		}
		tiles = append(tiles, &t)
	}

	return tiles, nil
}
