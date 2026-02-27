package dummy

import (
	"sync"
)

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
