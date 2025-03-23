package data

import (
	"encoding/json"
	"math/rand"
	"os"
)

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

// Loads missions out of mission_templates.json
func LoadMissionTemplates(path string) ([]MissionTemplate, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data struct {
		Missions []MissionTemplate `json:"missions"`
	}

	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data.Missions, nil
}

// Find parent star system for planet
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

// Generate a semi-generated mission
func GenerateMissionFromTemplate(id int, templates []MissionTemplate, planets []PlanetWithSystem) Mission {
	t := templates[rand.Intn(len(templates))] // Random mission template
	p := planets[rand.Intn(len(planets))]     // Pick random planet and parent solar system
	income := 1000 + rand.Intn(1500)          // Random income (Could add mission difficulty multipliers)

	// Build and return the mission
	return Mission{
		Id:          id,
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
