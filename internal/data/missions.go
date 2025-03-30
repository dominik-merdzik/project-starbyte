package data

import (
	"bytes"
	"encoding/json"
	"math/rand"

	_ "embed"
)

//go:embed mission_templates.json
var embeddedMissionTemplates []byte

type MissionTemplate struct {
	Step         int           `json:"Step,omitempty"`
	Id           int           `json:"Id"`
	Title        string        `json:"Title"`
	Description  string        `json:"Description"`
	Status       MissionStatus `json:"Status"`
	Location     Location      `json:"Location"`
	Income       int           `json:"Income"`
	Requirements string        `json:"Requirements"`
	Received     string        `json:"Received"`
	Category     string        `json:"Category"`
	Dialogue     []string      `json:"dialogue"`
}

type PlanetWithSystem struct {
	StarSystemName string
	Planet         Planet
}

// loads missions from the embedded mission_templates.json data
func LoadMissionTemplates() ([]MissionTemplate, error) {
	var data struct {
		Missions []MissionTemplate `json:"missions"`
	}

	// decode directly from the embedded byte slice using a bytes.Reader
	err := json.NewDecoder(bytes.NewReader(embeddedMissionTemplates)).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data.Missions, nil
}

// find parent star system for planet
func FlattenPlanetsWithSystems(systems []StarSystem) []PlanetWithSystem {
	var result []PlanetWithSystem
	for _, system := range systems {
		for _, planet := range system.Planets {
			result = append(result, PlanetWithSystem{
				StarSystemName: system.Name,
				Planet:         planet,
			})
		}
	}
	return result
}

// generate a semi-generated mission
func GenerateMissionFromTemplate(id int, templates []MissionTemplate, planets []PlanetWithSystem, currentStarSystem string) Mission {
	t := templates[rand.Intn(len(templates))] // Random mission template

	// Filter planets to only include those in the current star system
	var localPlanets []PlanetWithSystem
	for _, planet := range planets {
		if planet.StarSystemName == currentStarSystem {
			localPlanets = append(localPlanets, planet)
		}
	}

	// Fallback: If no planets in current system, just use all
	if len(localPlanets) == 0 {
		localPlanets = planets
	}

	p := localPlanets[rand.Intn(len(localPlanets))] // Pick random planet in same star system
	income := 1000 + rand.Intn(1500)                // Random income (Could add mission difficulty multipliers)

	// Build and return the mission
	return Mission{
		Id:          id,
		Step:        t.Step,
		Title:       t.Title,
		Description: t.Description,
		Status:      0,
		Location: Location{
			StarSystemName: p.StarSystemName,
			PlanetName:     p.Planet.Name,
			Coordinates:    p.Planet.Coordinates,
		},
		Income:       income,
		Requirements: t.Requirements,
		Received:     t.Received,
		Category:     t.Category,
		Dialogue:     t.Dialogue,
	}
}
