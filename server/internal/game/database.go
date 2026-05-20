package game

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luisferreira32/stickian/server/internal/utils"
)

type MapTile struct {
	Q     int
	R     int
	Biome int
}

type GameDatabase interface {
	GetCity(ctx context.Context, id string) (*City, error)
	GetCities(ctx context.Context, q1, r1, q2, r2 int) ([]*City, error)
	CreateCity(ctx context.Context, c *City) error
	GetMap(ctx context.Context, minQ, maxQ, minR, maxR int) ([]*MapTile, error)
	GetNextCitySpot(ctx context.Context) (*MapTile, error)
	GetSettleableTile(ctx context.Context, q, r int) (*MapTile, error)
	CreateMovement(ctx context.Context, m *Movement) error
	GetMovement(ctx context.Context, id string) (*Movement, error)
	GetOutgoingMovements(ctx context.Context, cityID string) ([]*Movement, error)
	GetIncomingMovements(ctx context.Context, cityID string) ([]*Movement, error)
	GetDueMovements(ctx context.Context, now time.Time) ([]*Movement, error)
	CompleteArrival(ctx context.Context, m *Movement) error
	CompleteReturn(ctx context.Context, m *Movement) error
	CancelMovement(ctx context.Context, id string, cutoff float64) (*Movement, error)
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
	cb.workshop, cb.observatory, cb.temple, cb.shrine, cb.cathedral,
	ct.swordsmen, ct.archers, ct.cavalry, ct.ships, ct.spies
	FROM city c
	LEFT JOIN city_resources cr ON cr.city_id = c.id
	LEFT JOIN city_buildings cb ON cb.city_id = c.id
	LEFT JOIN city_troops ct ON ct.city_id = c.id
	WHERE c.id = $1`

func (db *PostgresDatabase) GetCity(ctx context.Context, id string) (*City, error) {
	city := &City{
		Resources: &Resources{},
		Buildings: &Buildings{},
		Troops:    &Troops{},
	}
	err := db.DB.QueryRow(ctx, getCityQuery, id).Scan(
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
		&city.Troops.Swordsmen,
		&city.Troops.Archers,
		&city.Troops.Cavalry,
		&city.Troops.Ships,
		&city.Troops.Spies,
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

const createCityTroopsQuery = `INSERT INTO city_troops (city_id, swordsmen, archers, cavalry, ships, spies)
	VALUES ($1, $2, $3, $4, $5, $6)
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
	_, err = db.DB.Exec(ctx, createCityTroopsQuery,
		c.ID,
		c.Troops.Swordsmen,
		c.Troops.Archers,
		c.Troops.Cavalry,
		c.Troops.Ships,
		c.Troops.Spies,
	)
	if err != nil {
		return fmt.Errorf("city troops: %w", err)
	}

	return nil
}

const getMapQuery = `SELECT q, r, biome FROM world WHERE q BETWEEN $1 AND $2 AND r BETWEEN $3 AND $4`

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

const getNextCitySpotQuery = `
SELECT w.q, w.r, w.biome 
FROM world w
LEFT JOIN city c ON w.q = c.q AND w.r = c.r
WHERE w.settleable=true
AND c.id IS NULL
ORDER BY w.q+w.r
LIMIT 1`

// GetNextCitySpot returns the next available spot for a new city.
//
// The spot is determined by the lowest sum of q and r coordinates, which creates a diagonal pattern
// of city placement starting from the origin (0, 0) and moving outward. Only settleable tiles that
// are not already occupied by a city are considered.
func (db *PostgresDatabase) GetNextCitySpot(ctx context.Context) (*MapTile, error) {
	row := db.DB.QueryRow(ctx, getNextCitySpotQuery)
	var t MapTile
	err := row.Scan(&t.Q, &t.R, &t.Biome)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

const getSettleableTileQuery = `
SELECT w.q, w.r, w.biome
FROM world w
LEFT JOIN city c ON w.q = c.q AND w.r = c.r
WHERE w.q = $1 AND w.r = $2
AND w.settleable = true
AND c.id IS NULL`

// GetSettleableTile returns the tile at (q, r) if it is settleable and has no existing city.
// Returns ErrUserError if the tile does not exist, is not settleable, or is already occupied.
func (db *PostgresDatabase) GetSettleableTile(ctx context.Context, q, r int) (*MapTile, error) {
	var t MapTile
	err := db.DB.QueryRow(ctx, getSettleableTileQuery, q, r).Scan(&t.Q, &t.R, &t.Biome)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: tile is not available for settlement", utils.ErrUserError)
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

const createMovementQuery = `INSERT INTO movement
	(id, city_from, city_to, type, departure_time, arrival_time, is_returning,
	 swordsmen, archers, cavalry, ships, spies,
	 food, sticks, stones, gems)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

func (db *PostgresDatabase) CreateMovement(ctx context.Context, m *Movement) error {
	_, err := db.DB.Exec(ctx, createMovementQuery,
		m.ID,
		m.CityFrom,
		m.CityTo,
		m.Type,
		m.DepartureTime,
		m.ArrivalTime,
		m.IsReturning,
		m.Troops.Swordsmen,
		m.Troops.Archers,
		m.Troops.Cavalry,
		m.Troops.Ships,
		m.Troops.Spies,
		m.Resources.Food,
		m.Resources.Sticks,
		m.Resources.Stones,
		m.Resources.Gems,
	)
	if err != nil {
		return fmt.Errorf("create movement: %w", err)
	}
	return nil
}

const movementColumns = `id, city_from, city_to, type, departure_time, arrival_time, is_returning,
	swordsmen, archers, cavalry, ships, spies,
	food, sticks, stones, gems`

const getMovementQuery = `SELECT ` + movementColumns + ` FROM movement WHERE id = $1`

func scanMovementRow(row pgx.Row) (*Movement, error) {
	m := &Movement{Troops: &Troops{}, Resources: &MaterialResources{}}
	err := row.Scan(
		&m.ID, &m.CityFrom, &m.CityTo, &m.Type, &m.DepartureTime, &m.ArrivalTime, &m.IsReturning,
		&m.Troops.Swordsmen, &m.Troops.Archers, &m.Troops.Cavalry, &m.Troops.Ships, &m.Troops.Spies,
		&m.Resources.Food, &m.Resources.Sticks, &m.Resources.Stones, &m.Resources.Gems,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (db *PostgresDatabase) GetMovement(ctx context.Context, id string) (*Movement, error) {
	m, err := scanMovementRow(db.DB.QueryRow(ctx, getMovementQuery, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

const getOutgoingMovementsQuery = `SELECT ` + movementColumns +
	` FROM movement WHERE city_from = $1 ORDER BY arrival_time`

func (db *PostgresDatabase) GetOutgoingMovements(ctx context.Context, cityID string) ([]*Movement, error) {
	return db.scanMovements(ctx, getOutgoingMovementsQuery, cityID)
}

const getIncomingMovementsQuery = `SELECT ` + movementColumns +
	` FROM movement WHERE city_to = $1 ORDER BY arrival_time`

func (db *PostgresDatabase) GetIncomingMovements(ctx context.Context, cityID string) ([]*Movement, error) {
	return db.scanMovements(ctx, getIncomingMovementsQuery, cityID)
}

const getDueMovementsQuery = `SELECT ` + movementColumns +
	` FROM movement WHERE arrival_time <= $1 ORDER BY arrival_time LIMIT 100`

// GetDueMovements returns movements whose arrival_time has passed.
// The ticker picks these up and dispatches them based on is_returning.
func (db *PostgresDatabase) GetDueMovements(ctx context.Context, now time.Time) ([]*Movement, error) {
	rows, err := db.DB.Query(ctx, getDueMovementsQuery, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*Movement
	for rows.Next() {
		m, err := scanMovementRow(rows)
		if err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}
	return movements, nil
}

const deleteMovementByIDQuery = `DELETE FROM movement WHERE id = $1 AND arrival_time <= $2`

// CompleteArrival finalises an outbound movement that reached its destination.
// Effects at the destination are TODO; for now the row is just deleted.
// The arrival_time check guards against processing a row that was cancelled between
// discovery and execution.
func (db *PostgresDatabase) CompleteArrival(ctx context.Context, m *Movement) error {
	_, err := db.DB.Exec(ctx, deleteMovementByIDQuery, m.ID, time.Now())
	if err != nil {
		return fmt.Errorf("complete arrival: %w", err)
	}
	return nil
}

const refundTroopsQuery = `UPDATE city_troops SET
	swordsmen = swordsmen + $2,
	archers   = archers   + $3,
	cavalry   = cavalry   + $4,
	ships     = ships     + $5,
	spies     = spies     + $6
	WHERE city_id = $1`

const refundResourcesQuery = `UPDATE city_resources SET
	food   = food   + $2,
	sticks = sticks + $3,
	stones = stones + $4,
	gems   = gems   + $5
	WHERE city_id = $1`

// CompleteReturn finalises a returning movement: troops and resources are added
// back to the origin city (city_from), then the movement row is deleted.
// All three operations run in a single transaction so a partial failure cannot
// leak units or resources.
func (db *PostgresDatabase) CompleteReturn(ctx context.Context, m *Movement) error {
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin return tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, refundTroopsQuery, m.CityFrom,
		m.Troops.Swordsmen, m.Troops.Archers, m.Troops.Cavalry, m.Troops.Ships, m.Troops.Spies,
	); err != nil {
		return fmt.Errorf("refund troops: %w", err)
	}
	if _, err := tx.Exec(ctx, refundResourcesQuery, m.CityFrom,
		m.Resources.Food, m.Resources.Sticks, m.Resources.Stones, m.Resources.Gems,
	); err != nil {
		return fmt.Errorf("refund resources: %w", err)
	}
	if _, err := tx.Exec(ctx, deleteMovementByIDQuery, m.ID, time.Now()); err != nil {
		return fmt.Errorf("delete returning movement: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit return tx: %w", err)
	}
	return nil
}

const cancelMovementQuery = `UPDATE movement
	SET is_returning   = TRUE,
	    departure_time = NOW(),
	    arrival_time   = NOW() + (NOW() - departure_time)
	WHERE id = $1
	  AND is_returning = FALSE
	  AND arrival_time > NOW()
	  AND EXTRACT(EPOCH FROM (NOW() - departure_time))
	      < $2 * EXTRACT(EPOCH FROM (arrival_time - departure_time))
	RETURNING ` + movementColumns

// CancelMovement flips an outbound movement into a return trip if it is still
// cancellable (not already returning, not past arrival, progress < cutoff).
// Returns the updated movement on success, or utils.ErrUserError if the movement
// is no longer cancellable.
func (db *PostgresDatabase) CancelMovement(ctx context.Context, id string, cutoff float64) (*Movement, error) {
	m, err := scanMovementRow(db.DB.QueryRow(ctx, cancelMovementQuery, id, cutoff))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: movement is no longer cancellable", utils.ErrUserError)
	}
	if err != nil {
		return nil, fmt.Errorf("cancel movement: %w", err)
	}
	return m, nil
}

func (db *PostgresDatabase) scanMovements(ctx context.Context, query, cityID string) ([]*Movement, error) {
	rows, err := db.DB.Query(ctx, query, cityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*Movement
	for rows.Next() {
		m, err := scanMovementRow(rows)
		if err != nil {
			return nil, err
		}
		movements = append(movements, m)
	}
	return movements, nil
}
