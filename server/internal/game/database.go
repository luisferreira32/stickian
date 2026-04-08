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
	GetCities(ctx context.Context, q1, r1, q2, r2 int) ([]*City, error)
	GetMap(ctx context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error)
}

// InMemoryDatabase is a fake in-memory store for local development and testing.
type InMemoryDatabase struct {
	cities map[string]*City
}

// NewInMemoryDatabase returns an InMemoryDatabase pre-populated with fake cities.
func NewInMemoryDatabase() *InMemoryDatabase {
	cities := []*City{
		{
			ID:       "city-001",
			PlayerID: "player-001",
			Name:     "Big Stickland",
			Q:        10,
			R:        10,
			Biome:    "plains",
			Points:   350,
			Buildings: &Buildings{
				CityHall: 4, Farm: 2, Quarry: 2, Lumbermill: 2, CrystalMine: 3,
				Market: 1, Warehouse: 1,
			},
			Resources: &Resources{
				Population: 45, Sticks: 312, Stones: 215, Gems: 145, Food: 500, Faith: 18,
			},
		},
		{
			ID:       "city-002",
			PlayerID: "player-001",
			Name:     "Wowsticks",
			Q:        5,
			R:        3,
			Biome:    "mountain",
			Points:   210,
			Buildings: &Buildings{
				CityHall: 3, Quarry: 4, CrystalMine: 5, Walls: 2, Barracks: 1,
			},
			Resources: &Resources{
				Population: 28, Sticks: 80, Stones: 640, Gems: 310, Food: 200, Faith: 5,
			},
		},
		{
			ID:       "city-003",
			PlayerID: "player-002",
			Name:     "Port Cove",
			Q:        15, R: 8,
			Biome:  "coast",
			Points: 480,
			Buildings: &Buildings{
				CityHall: 6, Farm: 5, Quarry: 5, Lumbermill: 5, CrystalMine: 5, Harbor: 1, Docks: 2, Market: 2, Tavern: 1, Embassy: 1,
			},
			Resources: &Resources{
				Population: 72, Sticks: 150, Stones: 90, Gems: 60, Food: 890, Faith: 42,
			},
		},
		{
			ID:       "city-004",
			PlayerID: "player-002",
			Name:     "Verdant Vale",
			Q:        20, R: 14,
			Biome:  "plains",
			Points: 125,
			Buildings: &Buildings{
				CityHall: 2, Farm: 3, Lumbermill: 3, Shrine: 1,
			},
			Resources: &Resources{
				Population: 18, Sticks: 420, Stones: 55, Gems: 10, Food: 750, Faith: 88,
			},
		},
		{
			ID:       "city-005",
			PlayerID: "player-003",
			Name:     "Ironhold",
			Q:        3, R: 18,
			Biome:  "mountain",
			Points: 560,
			Buildings: &Buildings{
				CityHall: 6, Quarry: 5, Walls: 4, Barracks: 3, Workshop: 2, Observatory: 1,
			},
			Resources: &Resources{
				Population: 60, Sticks: 200, Stones: 980, Gems: 500, Food: 300, Faith: 12,
			},
		},
	}

	m := make(map[string]*City, len(cities))
	for _, c := range cities {
		m[c.ID] = c
	}
	return &InMemoryDatabase{cities: m}
}

func (db *InMemoryDatabase) GetCity(_ context.Context, id string) (*City, error) {
	if c, ok := db.cities[id]; ok {
		return c, nil
	}
	return nil, ErrNotFound
}

// GetCities returns all cities whose coordinates lie within the bounding box
// defined by vertices (q1, r1) and (q2, r2) (inclusive). Only city table
// fields are returned — Buildings and Resources are omitted.
func (db *InMemoryDatabase) GetCities(_ context.Context, q1, r1, q2, r2 int) ([]*City, error) {
	minQ, maxQ := q1, q2
	if minQ > maxQ {
		minQ, maxQ = maxQ, minQ
	}
	minR, maxR := r1, r2
	if minR > maxR {
		minR, maxR = maxR, minR
	}
	var cities []*City
	for _, c := range db.cities {
		if c.Q >= minQ && c.Q <= maxQ && c.R >= minR && c.R <= maxR {
			cities = append(cities, &City{
				ID:       c.ID,
				PlayerID: c.PlayerID,
				Name:     c.Name,
				Q:        c.Q,
				R:        c.R,
				Biome:    c.Biome,
				Points:   c.Points,
			})
		}
	}
	return cities, nil
}

func (db *InMemoryDatabase) GetMap(_ context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error) {
	var tiles []*MapTile
	for _, c := range db.cities {
		if c.Q >= minQ && c.Q <= maxQ && c.R >= minR && c.R <= maxR {
			tiles = append(tiles, &MapTile{Q: c.Q, R: c.R})
		}
	}
	return tiles, nil
}

type PostgresDatabase struct {
	DB *pgx.Conn
}

const getCityQuery = `SELECT
	c.id, c.player_id, c.name, c.q, c.r, c.biome, c.points,
	cr.food, cr.sticks, cr.stones, cr.gems, cr.population, cr.faith,
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

const getCitiesByBoundsQuery = `SELECT id, player_id, name, q, r, biome, points FROM city
	WHERE q BETWEEN $1 AND $2 AND r BETWEEN $3 AND $4`

// GetCities returns all cities within the bounding box defined by vertices
// (q1, r1) and (q2, r2). The range is normalised so order does not matter.
// Only city table fields are returned — Buildings and Resources are omitted.
func (db *PostgresDatabase) GetCities(ctx context.Context, q1, r1, q2, r2 int) ([]*City, error) {
	minQ, maxQ := q1, q2
	if minQ > maxQ {
		minQ, maxQ = maxQ, minQ
	}
	minR, maxR := r1, r2
	if minR > maxR {
		minR, maxR = maxR, minR
	}
	rows, err := db.DB.Query(ctx, getCitiesByBoundsQuery, minQ, maxQ, minR, maxR)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []*City
	for rows.Next() {
		city := &City{}
		if err := rows.Scan(
			&city.ID, &city.PlayerID, &city.Name, &city.Q, &city.R, &city.Biome, &city.Points,
		); err != nil {
			return nil, err
		}
		cities = append(cities, city)
	}
	return cities, nil
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
