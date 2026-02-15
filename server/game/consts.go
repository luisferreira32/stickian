package game

import "time"

// resource names
const (
	Population = "population"
	Stone      = "stone"
	Sticks     = "sticks"
	Crystal    = "crystal"
	Gold       = "gold"
)

// building names
const (
	CityHall    = "city_Hall"
	Farm        = "farm"
	Quarry      = "quarry"
	LumberMill  = "lumbermill"
	CrystalMine = "crystal_Mine"
)

var (
	// validation "constants"
	// validResources = map[string]bool{Population: true, Stone: true, Sticks: true, Crystal: true, Gold: true}
	// validBuildings = map[string]bool{CityHall: true, Farm: true, Quarry: true, LumberMill: true, CrystalMine: true}

	// lookup table for resource costs of building upgrade
	// TODO: this should be defined with a function for ideal gameplay and but then
	// precomputed onto a lookup table to speed up calculations
	buildingUpgradeCosts = map[string]map[int]map[string]int{
		CityHall: {
			1: {Stone: 100, Sticks: 100, Crystal: 50, Gold: 20},
			2: {Stone: 200, Sticks: 200, Crystal: 100, Gold: 50},
			3: {Stone: 400, Sticks: 400, Crystal: 200, Gold: 100},
			4: {Stone: 800, Sticks: 800, Crystal: 400, Gold: 200},
			5: {Stone: 1600, Sticks: 1600, Crystal: 800, Gold: 400},
		},
		Farm: {
			1: {Stone: 50, Sticks: 50, Crystal: 20, Gold: 10},
			2: {Stone: 100, Sticks: 100, Crystal: 50, Gold: 20},
			3: {Stone: 200, Sticks: 200, Crystal: 100, Gold: 50},
			4: {Stone: 400, Sticks: 400, Crystal: 200, Gold: 100},
			5: {Stone: 800, Sticks: 800, Crystal: 400, Gold: 200},
		},
		Quarry: {
			1: {Stone: 50, Sticks: 50, Crystal: 20, Gold: 10},
			2: {Stone: 100, Sticks: 100, Crystal: 50, Gold: 20},
			3: {Stone: 200, Sticks: 200, Crystal: 100, Gold: 50},
			4: {Stone: 400, Sticks: 400, Crystal: 200, Gold: 100},
			5: {Stone: 800, Sticks: 800, Crystal: 400, Gold: 200},
		},
		LumberMill: {
			1: {Stone: 50, Sticks: 50, Crystal: 20, Gold: 10},
			2: {Stone: 100, Sticks: 100, Crystal: 50, Gold: 20},
			3: {Stone: 200, Sticks: 200, Crystal: 100, Gold: 50},
			4: {Stone: 400, Sticks: 400, Crystal: 200, Gold: 100},
			5: {Stone: 800, Sticks: 800, Crystal: 400, Gold: 200},
		},
		CrystalMine: {
			1: {Stone: 50, Sticks: 50, Crystal: 20, Gold: 10},
			2: {Stone: 100, Sticks: 100, Crystal: 50, Gold: 20},
			3: {Stone: 200, Sticks: 200, Crystal: 100, Gold: 50},
			4: {Stone: 400, Sticks: 400, Crystal: 200, Gold: 100},
			5: {Stone: 800, Sticks: 800, Crystal: 400, Gold: 200},
		},
	}
	buildingUpgradeTimes = map[string]map[int]time.Duration{
		CityHall: {
			1: 10 * time.Second,
			2: 11 * time.Second,
			3: 12 * time.Second,
			4: 13 * time.Second,
			5: 14 * time.Second,
		},
		Farm: {
			1: 10 * time.Second,
			2: 11 * time.Second,
			3: 12 * time.Second,
			4: 13 * time.Second,
			5: 14 * time.Second,
		},
		Quarry: {
			1: 10 * time.Second,
			2: 11 * time.Second,
			3: 12 * time.Second,
			4: 13 * time.Second,
			5: 14 * time.Second,
		},
		LumberMill: {
			1: 10 * time.Second,
			2: 11 * time.Second,
			3: 12 * time.Second,
			4: 13 * time.Second,
			5: 14 * time.Second,
		},
		CrystalMine: {
			1: 10 * time.Second,
			2: 11 * time.Second,
			3: 12 * time.Second,
			4: 13 * time.Second,
			5: 14 * time.Second,
		},
	}
)
