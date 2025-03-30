package data

import (
	"math"
	"time"
)

// travel time constants
const (
	// Base scaling: seconds one unit of calculated distance represent
	// adjust this value to make travel generally faster or slower
	secondsPerDistanceUnit = 0.5 // e.g., 1 distance unit = 0.5 seconds base time

	// Engine bonus: how much faster the ship gets per engine level
	engineSpeedBonusFactor = 0.20 // e.g., Each engine level makes travel 20% faster than base speed

	// Min/Max travel time caps (in seconds)
	minTravelSeconds = 3.0  // minimum travel time will be 3 seconds
	maxTravelSeconds = 60.0 // maximum travel time will be 60 seconds
)

// LocationService handles location-related operations
type LocationService struct {
	GameMap GameMap
}

// NewLocationService creates a new location service with the given game map
func NewLocationService(gameMap GameMap) *LocationService {
	return &LocationService{GameMap: gameMap}
}

// FindByPlanetName looks up a location by planet name
func (ls *LocationService) FindByPlanetName(planetName string) *Location {
	for i := range ls.GameMap.StarSystems {
		system := ls.GameMap.StarSystems[i]
		for j := range system.Planets {
			planet := system.Planets[j]
			if planet.Name == planetName {
				return &Location{
					StarSystemName: system.Name,
					PlanetName:     planet.Name,
					Coordinates:    planet.Coordinates,
				}
			}
		}
	}
	return nil
}

// CalculateDistance determines the 'distance value' between two locations
func (ls *LocationService) CalculateDistance(from, to Coordinates, fromStarSys, toStarSys string) int {

	// Euclidean distance calculation
	dx := float64(to.X - from.X)
	dy := float64(to.Y - from.Y)
	dz := float64(to.Z - from.Z)
	distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

	// inter-system travel multiplier (used when star systems differ)
	starSystemMultiplyer := 1.0
	if fromStarSys != toStarSys {
		starSystemMultiplyer = 3.0
	}

	// Use ceil to ensure even short distances result in a non-zero value after multiplier
	finalDistanceValue := math.Ceil(distance * starSystemMultiplyer)

	return int(finalDistanceValue)
}

// Calculates the actual time.Duration for travel based on distance and engine level
func (ls *LocationService) CalculateTravelDuration(fromLoc Location, toLoc Location, engineLevel int, maxEngineLevel int /* Not currently used, but could be */) time.Duration {
	// calculates raw distance value using existing method
	distanceValue := ls.CalculateDistance(fromLoc.Coordinates, toLoc.Coordinates, fromLoc.StarSystemName, toLoc.StarSystemName)

	// handle zero distance case (shouldn't happen if locations differ)
	if distanceValue <= 0 {
		return time.Duration(minTravelSeconds*1000) * time.Millisecond // return min duration if distance is zero
	}

	// calculate base duration in seconds
	baseDurationSeconds := float64(distanceValue) * secondsPerDistanceUnit

	// calculate engine speed multiplier (1.0 = base speed, >1.0 = faster)
	if engineLevel < 0 {
		engineLevel = 0
	}
	speedMultiplier := 1.0 + (float64(engineLevel) * engineSpeedBonusFactor)

	// prevent division by zero or excessively small multipliers if bonus factor is weirdly negative
	if speedMultiplier < 0.1 {
		speedMultiplier = 0.1 // minimum speed multiplier
	}

	// apply engine bonus to get final duration in seconds
	finalDurationSeconds := baseDurationSeconds / speedMultiplier

	// apply Min/Max caps (in seconds)
	if finalDurationSeconds < minTravelSeconds {
		finalDurationSeconds = minTravelSeconds
	}
	if finalDurationSeconds > maxTravelSeconds {
		finalDurationSeconds = maxTravelSeconds
	}

	// convert final seconds to time.Duration (using milliseconds for precision)
	finalDuration := time.Duration(finalDurationSeconds*1000) * time.Millisecond

	//log.Printf("Calculated Travel: DistVal=%d, BaseSec=%.2f, Multi=%.2f, FinalSec=%.2f, Duration=%s", distanceValue, baseDurationSeconds, speedMultiplier, finalDurationSeconds, finalDuration)

	return finalDuration
}

// GetFuelCost calculates fuel remaining after travel.
func (ls *LocationService) GetFuelCost(from, to Coordinates, fromStarSys, toStarSys string, engineHealth int, currentFuel int) int {
	distance := ls.CalculateDistance(from, to, fromStarSys, toStarSys)

	if distance <= 0 {
		return currentFuel
	}

	// ensure engineHealth is within 0-100 range
	if engineHealth < 0 {
		engineHealth = 0
	}
	if engineHealth > 100 {
		engineHealth = 100
	}

	// engine health modifier: lower health means higher cost multiplier
	// multiplier ranges from 1.0 (at 100% health) up to 2.0 (at 0% health)
	engineModifier := 1.0 + (1.0 - (float64(engineHealth) / 100.0))

	// calculate actual fuel cost using Ceil to ensure at least 1 fuel per distance unit base
	fuelCost := math.Ceil(float64(distance) * engineModifier)

	remainingFuel := currentFuel - int(fuelCost)

	// prevent fuel from going below zero
	if remainingFuel < 0 {
		remainingFuel = 0
	}

	return remainingFuel
}

// NewLocationFromPlanet creates a fully populated Location from a Planet and StarSystem
func NewLocationFromPlanet(system StarSystem, planet Planet) Location {
	return Location{
		StarSystemName: system.Name,
		PlanetName:     planet.Name,
		Coordinates:    planet.Coordinates,
	}
}

// IsEqual compares two locations to determine if they refer to the same place
func (loc Location) IsEqual(other Location) bool {
	return loc.StarSystemName == other.StarSystemName &&
		loc.PlanetName == other.PlanetName
}

// GetFullPlanet retrieves the full Planet object from the GameMap
func (loc Location) GetFullPlanet(gameMap GameMap) Planet {
	for i := range gameMap.StarSystems {
		system := gameMap.StarSystems[i]
		if system.Name == loc.StarSystemName {
			for j := range system.Planets {
				planet := system.Planets[j]
				if planet.Name == loc.PlanetName {
					return planet // Found in map
				}
			}
			break // Planet not in this system
		}
	}
	// Not found in map, return placeholder
	// log.Printf("Info: Location '%s' in system '%s' not found in GameMap. Creating placeholder.", loc.PlanetName, loc.StarSystemName)
	return Planet{
		Name:         loc.PlanetName,
		Type:         "Mission Location",
		Coordinates:  loc.Coordinates,
		Resources:    []Resource{},
		Requirements: []CrewRequirement{},
	}
}
