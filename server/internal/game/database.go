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
	GetCities(ctx context.Context, q, r int) (*City, error)
	GetMap(ctx context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error)
}

// InMemoryDatabase is a placeholder for an actual database implementation.
type InMemoryDatabase struct{}

func (db *InMemoryDatabase) GetCity(_ context.Context, _ string) (*City, error) {
	// This is just a stub. In a real implementation, you would query your database here.
	return &City{
		Name:     "Stick City",
		Q:        10,
		R:        10,
		Biome:    "plains",
		Points:   100,
		PlayerID: "Stickman",
		Buildings: &Buildings{
			CityHall:    4,
			Farm:        2,
			Quarry:      2,
			Lumbermill:  2,
			CrystalMine: 3,
		},
		Resources: &Resources{
			Population: 45,
			Sticks:     312,
			Stones:     215,
			Gems:       145,
			Food:       500,
			Faith:      18,
		},
	}, nil
}

// GetCities gets the city at the specified Q, R coordinate.
func (db *InMemoryDatabase) GetCities(_ context.Context, _, _ int) (*City, error) {
	// This is just a stub. In a real implementation, you would query your database here.
	return &City{
		Name:     "Stick City",
		Q:        10,
		R:        10,
		Biome:    "plains",
		PlayerID: "Stickman",
	}, nil
}

func (db *InMemoryDatabase) GetMap(_ context.Context, _, _, _, _ int) ([]*MapTile, error) {
	// This is just a stub.
	return []*MapTile{}, nil
}

type PostgresDatabase struct {
	DB *pgx.Conn
}

const getCityQuery = `SELECT
	c.id, c.player_id, c.name, c.q, c.r, c.biome, c.points,
	cr.food, cr.sticks, cr.rocks, cr.gems, cr.population, cr.faith,
	cb.city_hall, cb.embassy, cb.treasury, cb.tavern,
	cb.farm, cb.lumbermill, cb.quarry, cb.crystal_mine,
	cb.warehouse, cb.market, cb.harbor, cb.walls,
	cb.barracks, cb.docks, cb.spy_guild, cb.library,
	cb.workshop, cb.observatory, cb.temple, cb.shrine, cb.cathedral
	FROM city c
	LEFT JOIN city_resources cr ON cr.city_id = c.id
	LEFT JOIN city_buildings cb ON cb.city_id = c.id
	WHERE c.id = $1`

func (db *PostgresDatabase) GetCity(ctx context.Context, id string) (*City, error) {
	city := &City{
		Resources: &Resources{},
		Buildings: &Buildings{},
	}
	err := db.DB.QueryRow(ctx, getCityQuery, id).Scan(
		&city.ID, &city.PlayerID, &city.Name, &city.Q, &city.R, &city.Biome, &city.Points,
		&city.Resources.Food, &city.Resources.Sticks, &city.Resources.Stones,
		&city.Resources.Gems, &city.Resources.Population, &city.Resources.Faith,
		&city.Buildings.CityHall, &city.Buildings.Embassy, &city.Buildings.Treasury, &city.Buildings.Tavern,
		&city.Buildings.Farm, &city.Buildings.Lumbermill, &city.Buildings.Quarry, &city.Buildings.CrystalMine,
		&city.Buildings.Warehouse, &city.Buildings.Market, &city.Buildings.Harbor, &city.Buildings.Walls,
		&city.Buildings.Barracks, &city.Buildings.Docks, &city.Buildings.SpyGuild, &city.Buildings.Library,
		&city.Buildings.Workshop, &city.Buildings.Observatory, &city.Buildings.Temple, &city.Buildings.Shrine, &city.Buildings.Cathedral,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return city, nil
}

const getCitiesByCoordQuery = `SELECT id, player_id, name, q, r, biome, points FROM city WHERE q = $1 AND r = $2`

func (db *PostgresDatabase) GetCities(ctx context.Context, q, r int) (*City, error) {
	city := &City{}
	err := db.DB.QueryRow(ctx, getCitiesByCoordQuery, q, r).Scan(
		&city.ID, &city.PlayerID, &city.Name, &city.Q, &city.R, &city.Biome, &city.Points,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return city, nil
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
