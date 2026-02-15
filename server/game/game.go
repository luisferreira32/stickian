package game

import (
	"context"
	"encoding/json"
	"log"
	"slices"
	"sync"
	"time"
)

// NewGame returns a new instance of the game struct with the provided database.
func NewGame(db database) *game {
	return &game{
		db:      db,
		timeNow: time.Now,
	}
}

// game is the singleton struct that will hold all the ephemeral game state, logic
// and needed clients (e.g., database).
type game struct {
	db database

	// we use a tick system for event processing. each tick is expected to take one second
	// but we will process all events in the tick before moving forward to the next one, therefore
	// it is necessary to keep a tick counter to be able to fetch only the relevant events for each
	// tick and allow other events to be added to the database while processing the current tick.
	tick     int64
	tickLock sync.RWMutex

	// defined as a field to mock in unit tests
	timeNow func() time.Time
}

func (g *game) Run(ctx context.Context) {
	clock := time.NewTicker(1 * time.Second)
	defer clock.Stop()

	for {
		select {
		case <-clock.C:
			g.tickLock.Lock()
			tickToProcess := g.tick
			g.tick++
			g.tickLock.Unlock()

			if err := g.processTick(tickToProcess); err != nil {
				log.Printf("error processing tick: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (g *game) currentTick() int64 {
	g.tickLock.RLock()
	defer g.tickLock.RUnlock()
	return g.tick
}

func (g *game) processTick(tick int64) error {
	// 1. get events from the database

	events, err := g.db.GetEvents(tick)
	if err != nil {
		return err
	}

	// 2. get current state

	city, err := g.db.GetCity("city_123")
	if err != nil {
		return err
	}

	// 3. calculate new state based on events

	slices.SortFunc(events, func(a, b *event) int {
		timeDiff := a.Time.Sub(b.Time)
		if timeDiff != 0 {
			return int(timeDiff)
		}
		return int(a.Type) - int(b.Type) // if events have the same timestamp, sort by event type priority
	})

	generatedEvents := []*event{}

	for _, e := range events {
		// TODO: there needs to be a way to communicate with the front-end about failed events
		// front-end should be doing optimistic updates based on Accepted status and won't undo display
		// until an explicit refresh or a server-side sent event indicates the failure.
		switch e.Type {
		case UpgradeBuilding:
			upgradeEvent := UpgradeBuildingEvent{}
			err := json.Unmarshal(e.Data, &upgradeEvent)
			if err != nil {
				log.Printf("error unmarshaling event data: %v", err)
				continue // skip this event but continue processing the rest
			}
			currentLevel := city.Buildings[upgradeEvent.Building]
			if currentLevel >= upgradeEvent.Level {
				continue // seems the upgrade was already processed and we're seeing an old event
			}

			// check if the city suffered any "damages" since request, and skip upgrade if pre-conditions are not met
			if currentLevel+1 != upgradeEvent.Level {
				log.Printf("building level mismatch for event: %v, current city state: %v", upgradeEvent, city)
				continue // this means something happened between the time the event was created and processed
			}
			neededResources, ok := buildingUpgradeCosts[upgradeEvent.Building][upgradeEvent.Level]
			if !ok {
				log.Printf("invalid building or level in event: %v", upgradeEvent)
				continue // should not happen
			}
			if !hasSufficientResources(city.Resources, neededResources) {
				log.Printf("not enough resources for upgrade event: %v", upgradeEvent)
				continue // this means something happened between the time the event was created and processed
			}
			upgradeComplete := UpgradeBuildingCompleteEvent{
				CityID:   upgradeEvent.CityID,
				Building: upgradeEvent.Building,
				Level:    upgradeEvent.Level,
			}
			buildingTime := buildingUpgradeTimes[upgradeEvent.Building][upgradeEvent.Level]
			upgradeCompleteTime := e.Time.Add(buildingTime)
			completeEvent, err := upgradeComplete.toEvent()
			if err != nil {
				log.Printf("error creating complete event: %v", err)
				continue
			}
			completeEvent.Time = upgradeCompleteTime                  // this will be in the future
			completeEvent.Tick = tick + int64(buildingTime.Seconds()) // assign tick based on building time to ensure it gets processed in the correct order
			generatedEvents = append(generatedEvents, completeEvent)
			city.Resources = deductResources(city.Resources, neededResources)
			city.BuildingsQueue = append(city.BuildingsQueue, BuildingQueueItem{
				Building:     upgradeEvent.Building,
				Level:        upgradeEvent.Level,
				CompleteTime: upgradeCompleteTime,
			})

		case UpgradeBuildingComplete:
			upgradeCompleteEvent := UpgradeBuildingCompleteEvent{}
			err := json.Unmarshal(e.Data, &upgradeCompleteEvent)
			if err != nil {
				log.Printf("error unmarshaling event data: %v", err)
				continue // skip this event but continue processing the rest
			}
			currentLevel := city.Buildings[upgradeCompleteEvent.Building]
			if currentLevel >= upgradeCompleteEvent.Level {
				continue // seems the upgrade was already processed and we're seeing an old event
			}
			if currentLevel+1 != upgradeCompleteEvent.Level {
				log.Printf("building level mismatch for complete event: %v, current city state: %v", upgradeCompleteEvent, city)
				continue // this means something happened between the time the event was created and processed
			}
			city.Buildings[upgradeCompleteEvent.Building] = upgradeCompleteEvent.Level
			// remove from queue
			newQueue := []BuildingQueueItem{}
			for _, item := range city.BuildingsQueue {
				if item.Building == upgradeCompleteEvent.Building && item.Level == upgradeCompleteEvent.Level {
					continue
				}
				newQueue = append(newQueue, item)
			}
			city.BuildingsQueue = newQueue

		default:
			log.Printf("unknown event type: %v", e.Type)
		}
	}

	// 4. write generated events and the new state
	// TODO: might make sense to have this in a transaction in case of partial failure?
	// or really make sure events are idempotent and front-end can handle partial errors / state updates
	for _, e := range generatedEvents {
		if err := g.db.WriteEvent(e); err != nil {
			log.Printf("error writing generated event: %v", err)
			continue
		}
	}
	if err := g.db.WriteCity(city); err != nil {
		return err
	}

	return nil
}

func hasSufficientResources(currentResources, neededResources map[string]int) bool {
	for resourceType, neededAmmount := range neededResources {
		if currentResources[resourceType] < neededAmmount {
			return false
		}
	}
	return true
}

func deductResources(currentResources, neededResources map[string]int) map[string]int {
	newResources := make(map[string]int)
	for resourceType, ammount := range currentResources {
		newResources[resourceType] = ammount - neededResources[resourceType]
	}
	return newResources
}
