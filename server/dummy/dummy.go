package dummy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Echo is an example handler that simply echoes back the request body.
func Echo(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	_, _ = w.Write(append(b, []byte("\n")...))
}

// Panic is an example handler that will panic when called.
func Panic(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	panic("this is a panic: " + string(b))
}

// Hello is an example handler that only answers to GET requests with "hello world".
func Hello(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello world\n"))
}

// TODO: add a more robust ID generator, this is just for testing purposes and should not be used in production
func genid() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// DummyService is a dummy service to test ticker functionality for our tick based game loop.
type DummyService struct {
	DummyDatabase *InMemoryDatabase

	tickWrite    int64
	tickRead     int64
	tickLock     sync.RWMutex
	tickDuration int64
}

// TrainFoo is a dummy handler to test event ticker functionality.
//
// To train a Foo you spend 1 Bar and it takes 3 ticks to complete.
func (s *DummyService) TrainFoo(w http.ResponseWriter, r *http.Request) {
	// first do basic non-binding validation
	bar, err := s.DummyDatabase.GetBar()
	if err != nil {
		log.Printf("failed to get bar: %v", err)
		http.Error(w, "failed to get Bar", http.StatusInternalServerError)
		return
	}
	if bar < 1 {
		http.Error(w, "not enough Bar to train Foo", http.StatusBadRequest)
		return
	}

	s.tickLock.RLock()
	defer s.tickLock.RUnlock()
	s.DummyDatabase.AddEvent(Event{Type: EventTrainFoo, Key: genid()}, s.tickWrite)
}

// BuildBar is a dummy handler to test event ticker functionality.
//
// To build 2 Bar you spend 1 Bar. And you always produce a baseline of 1 Bar per tick.
func (s *DummyService) BuildBar(w http.ResponseWriter, r *http.Request) {
	// first do basic non-binding validation
	bar, err := s.DummyDatabase.GetBar()
	if err != nil {
		log.Printf("failed to get bar: %v", err)
		http.Error(w, "failed to get Bar", http.StatusInternalServerError)
		return
	}
	if bar < 2 {
		http.Error(w, "not enough Bar to build Bar", http.StatusBadRequest)
		return
	}

	s.tickLock.RLock()
	defer s.tickLock.RUnlock()
	s.DummyDatabase.AddEvent(Event{Type: EventBuildBar, Key: genid()}, s.tickWrite)
}

// GetFooBar is a dummy handler to test event ticker functionality.
func (s *DummyService) GetFooBar(w http.ResponseWriter, r *http.Request) {
	foo, err1 := s.DummyDatabase.GetFoo()
	bar, err2 := s.DummyDatabase.GetBar()
	if err1 != nil || err2 != nil {
		log.Printf("failed to get state: foo err: %v, bar err: %v", err1, err2)
		http.Error(w, "failed to get state", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(
		`{"foo": ` + string(rune(foo)) + `, "bar": ` + string(rune(bar)) + `}` + "\n",
	))
}

// Run starts the dummy service ticker loop.
func (s *DummyService) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.tickDuration))
	defer ticker.Stop()

	forceTick := make(chan struct{}, 10)
	defer close(forceTick)

	for {
		select {
		case <-forceTick:
			err := s.tick()
			if err != nil {
				log.Printf("tick err: %v", err)
			}
		case <-ticker.C:
			err := s.tick()
			if err != nil {
				log.Printf("tick err: %v", err)
			}
			if s.tickRead != s.tickWrite {
				select {
				case forceTick <- struct{}{}:
				default:
					// if we reach this situation means that processing is falling behind a lot!
					// either due to very long tick processing time or a persistent error
					log.Printf("skipping tick, forceTick channel is full")
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *DummyService) tick() error {
	// prevent new events from being added while we process current tick
	s.tickLock.Lock()
	s.tickWrite++
	s.tickLock.Unlock()

	// get all events for the current tick under processing
	events, err := s.DummyDatabase.GetEvents(s.tickRead)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// process events on the tick and keep in memory the necessary state to apply the effects of the events
	// at the end in a single transaction to the correct tables - since we're not doing propper event sourcing
	// this is required to ensure we don't end up with an inconsistent state in case of errors during the
	// processing of the tick

	foo, err1 := s.DummyDatabase.GetFoo()
	bar, err2 := s.DummyDatabase.GetBar()
	if err1 != nil || err2 != nil {
		return fmt.Errorf("failed to get state: foo err: %v, bar err: %v", err1, err2)
	}

	for _, event := range events {
		switch event.Type {
		case EventTrainFoo:
			if bar < 1 {
				// not enough Bar to train Foo, skip this event, it is up to the front-end to
				// fetch latests accurate state to revert the optimistic updates
				continue
			}
			// NOTE: future event keys must be determinisitc such that processing the same events
			// a second time would not add additional events, but just re-write the same ones
			err = s.DummyDatabase.AddEvent(Event{Type: EventProducedFoo, Key: event.Key}, s.tickRead+3)
			if err != nil {
				return fmt.Errorf("failed to add produced foo event: %w", err)
			}
			bar--
		case EventProducedFoo:
			foo++
		case EventBuildBar:
			if bar < 1 {
				// not enough Bar to build Bar, skip this event, it is up to the front-end to
				// fetch latests accurate state to revert the optimistic updates
				continue
			}
			bar++ // mock: +2 - 1 = +1
		}
	}

	// baseline production
	bar += 1

	// NOTE: update to the database must be transactional, otherwise we might fail partially and
	// end up with inconsistent state for the re-processing of the tick
	err = s.DummyDatabase.SetFooBar(foo, bar)
	if err != nil {
		return fmt.Errorf("failed to set foo and bar: %w", err)
	}

	s.tickRead++
	return nil
}
