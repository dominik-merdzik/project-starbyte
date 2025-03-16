package data

import (
	"math"
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
	for _, system := range ls.GameMap.StarSystems {
		for _, planet := range system.Planets {
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

// CalculateDistance determines the distance between two locations
// Using euclidean distance to better simulate space travel times
func (ls *LocationService) CalculateDistance(from, to Coordinates, fromStarSys, toStarSys string) int {
	distance := math.Sqrt(
		math.Pow(float64(to.X-from.X), 2) +
			math.Pow(float64(to.Y-from.Y), 2) +
			math.Pow(float64(to.Z-from.Z), 2),
	)

	// Compare star systems
	sameSystem := false
	if fromStarSys == toStarSys {
		sameSystem = true
	}

	// Add travel multiplyer if destination is in different star system
	starSystemMultiplyer := 1.0
	if !sameSystem {
		starSystemMultiplyer = 3.0
	}

	travelTime := math.Round(distance * starSystemMultiplyer)
	return int(travelTime)
}

// GetFuelCost calculates fuel needed to travel between locations
// fuelCost = distance * (2 - (engineHealth/100))
// Makes engine health matter regarding fuel efficiency
func (ls *LocationService) GetFuelCost(from, to Coordinates, fromStarSys, toStarSys string, engineHealth int) int {
	distance := ls.CalculateDistance(from, to, fromStarSys, toStarSys)

	baseFuelCost := distance

	// Engine health modifier, lower engineHealth = higher fuelCost
	engineModifier := 2 - (float64(engineHealth) / 100)
	if engineHealth == 0 {
		engineModifier = 3 // Broken engines very expensive
	}

	fuelCost := float64(baseFuelCost) * engineModifier
	return int(fuelCost)
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

// GetFullPlanet retrieves the complete Planet object for this location
func (loc Location) GetFullPlanet(gameMap GameMap) *Planet {
	for _, system := range gameMap.StarSystems {
		if system.Name == loc.StarSystemName {
			for _, planet := range system.Planets {
				if planet.Name == loc.PlanetName {
					return &planet
				}
			}
		}
	}
	return nil
}
