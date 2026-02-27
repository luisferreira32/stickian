package dummy

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
)

type DummyDatabase interface {
	AddEvent(event Event, timestamp int64) error
	GetEvents(timestamp int64) ([]Event, error)
	GetFoo() (int, error)
	GetBar() (int, error)
	SetFooBar(foo, bar int) error
	Select1(ctx context.Context) error
}

// PosgresDatabase is a concrete implementation for a connection to a Postgres database.
// It does not do anything in the dummy module since there is no actual logic behind it,
// but serves as an example of how the database interface could be implemented for the game.
type PosgresDatabase struct {
	DB *pgx.Conn
}

func (db *PosgresDatabase) AddEvent(event Event, timestamp int64) error {
	return errors.New("not implemented")
}

func (db *PosgresDatabase) GetEvents(timestamp int64) ([]Event, error) {
	return nil, errors.New("not implemented")
}

func (db *PosgresDatabase) GetFoo() (int, error) {
	return 0, errors.New("not implemented")
}

func (db *PosgresDatabase) GetBar() (int, error) {
	return 0, errors.New("not implemented")
}

func (db *PosgresDatabase) SetFooBar(foo, bar int) error {
	return errors.New("not implemented")
}

func (db *PosgresDatabase) Select1(ctx context.Context) error {
	rows, err := db.DB.Query(ctx, "SELECT 1;")
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	elements := make([]int, 0)
	for rows.Next() {
		var element int
		if err := rows.Scan(&element); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		elements = append(elements, element)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error: %w", err)
	}

	// this should only have [1], so let's just check that the database is running smoothly
	if len(elements) != 1 || elements[0] != 1 {
		return fmt.Errorf("unexpected result: %v", elements)
	}

	return nil
}

// InMemoryDatabase is a placeholder for an actual database implementation.
type InMemoryDatabase struct {
	l          sync.Mutex
	Foo        int
	Bar        int
	EventQueue map[int64][]Event
}

type Event struct {
	Key  string
	Type string
}

const (
	EventTrainFoo    = "train_foo"
	EventProducedFoo = "produced_foo"
	EventBuildBar    = "build_bar"
)

func (db *InMemoryDatabase) AddEvent(event Event, timestamp int64) error {
	db.l.Lock()
	defer db.l.Unlock()

	// NOTE: in SQL implementation this logic with the key would be a trivial WHERE clause
	for _, v := range db.EventQueue[timestamp] {
		if v.Key == event.Key {
			// if we already have an event with the same key, we consider this a duplicate and skip it
			// this is important to ensure that in case of retries or duplicate requests we don't end
			// up with duplicated events
			return nil
		}
	}

	db.EventQueue[timestamp] = append(db.EventQueue[timestamp], event)
	return nil
}

func (db *InMemoryDatabase) GetEvents(timestamp int64) ([]Event, error) {
	db.l.Lock()
	defer db.l.Unlock()

	return db.EventQueue[timestamp], nil
}

func (db *InMemoryDatabase) GetFoo() (int, error) {
	return db.Foo, nil
}

func (db *InMemoryDatabase) GetBar() (int, error) {
	return db.Bar, nil
}

func (db *InMemoryDatabase) SetFooBar(foo, bar int) error {
	db.l.Lock()
	defer db.l.Unlock()
	db.Foo = foo
	db.Bar = bar
	return nil
}

func (db *InMemoryDatabase) Select1(_ context.Context) error {
	return nil
}
