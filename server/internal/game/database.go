package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/luisferreira32/stickian/server/internal/utils"
)

type MapTile struct {
	Q     int
	R     int
	Biome int
}

type GameDatabase interface {
	GetCity(ctx context.Context, id string, userID string) (*City, error)
	GetCities(ctx context.Context, q1, r1, q2, r2 int) ([]*City, error)
	CreateCity(ctx context.Context, c *City) error
	GetMap(ctx context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error)
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
	WHERE c.id = $1 AND c.player_id = $2`

func (db *PostgresDatabase) GetCity(ctx context.Context, id string, userID string) (*City, error) {
	city := &City{
		Resources: &Resources{},
		Buildings: &Buildings{},
	}
	err := db.DB.QueryRow(ctx, getCityQuery, id, userID).Scan(
		&city.ID,
		&city.PlayerID,
		&city.Name,
		&city.Q,
		&city.R,
		&city.Biome,
		&city.Points,
		&city.Resources.Food,
		&city.Resources.Sticks,
		&city.Resources.Stones,
		&city.Resources.Gems,
		&city.Resources.Population,
		&city.Resources.Faith,
		&city.Buildings.CityHall,
		&city.Buildings.Embassy,
		&city.Buildings.Treasury,
		&city.Buildings.Tavern,
		&city.Buildings.Farm,
		&city.Buildings.Lumbermill,
		&city.Buildings.Quarry,
		&city.Buildings.CrystalMine,
		&city.Buildings.Warehouse,
		&city.Buildings.Market,
		&city.Buildings.Harbor,
		&city.Buildings.Walls,
		&city.Buildings.Barracks,
		&city.Buildings.Docks,
		&city.Buildings.SpyGuild,
		&city.Buildings.Library,
		&city.Buildings.Workshop,
		&city.Buildings.Observatory,
		&city.Buildings.Temple,
		&city.Buildings.Shrine,
		&city.Buildings.Cathedral,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, utils.ErrNotFound
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

const createCityQuery = `INSERT INTO city (id, player_id, name, q, r, biome, points)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT DO NOTHING`

const createCityResourcesQuery = `INSERT INTO city_resources (city_id, food, sticks, stones, gems, population, faith)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT DO NOTHING`

const createCityBuildingsQuery = `INSERT INTO city_buildings (city_id, city_hall, embassy, treasury, tavern, farm, lumbermill, quarry, crystal_mine, warehouse, market, harbor, walls, barracks, docks, spy_guild, library, workshop, observatory, temple, shrine, cathedral)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	ON CONFLICT DO NOTHING`

func (db *PostgresDatabase) CreateCity(ctx context.Context, c *City) error {
	_, err := db.DB.Exec(ctx, createCityQuery,
		c.ID,
		c.PlayerID,
		c.Name,
		c.Q,
		c.R,
		c.Biome,
		c.Points,
	)
	if err != nil {
		return fmt.Errorf("city creation: %w", err)
	}
	_, err = db.DB.Exec(ctx, createCityResourcesQuery,
		c.ID,
		c.Resources.Food,
		c.Resources.Sticks,
		c.Resources.Stones,
		c.Resources.Gems,
		c.Resources.Population,
		c.Resources.Faith,
	)
	if err != nil {
		return fmt.Errorf("city resources: %w", err)
	}
	_, err = db.DB.Exec(ctx, createCityBuildingsQuery,
		c.ID,
		c.Buildings.CityHall,
		c.Buildings.Embassy,
		c.Buildings.Treasury,
		c.Buildings.Tavern,
		c.Buildings.Farm,
		c.Buildings.Lumbermill,
		c.Buildings.Quarry,
		c.Buildings.CrystalMine,
		c.Buildings.Warehouse,
		c.Buildings.Market,
		c.Buildings.Harbor,
		c.Buildings.Walls,
		c.Buildings.Barracks,
		c.Buildings.Docks,
		c.Buildings.SpyGuild,
		c.Buildings.Library,
		c.Buildings.Workshop,
		c.Buildings.Observatory,
		c.Buildings.Temple,
		c.Buildings.Shrine,
		c.Buildings.Cathedral,
	)
	if err != nil {
		return fmt.Errorf("city buildings: %w", err)
	}

	return nil
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
