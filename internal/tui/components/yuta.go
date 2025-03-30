package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// Yuta is a robot assistant who will tell the player helpful information on what to do next
type YutaComponent struct {
	Ship           data.Ship
	PlayerName     string
	ShipName       string
	Version        string
	SuggestRefuel  bool
	SuggestRepair  bool
	SuggestCredits bool
}

// creates a new instance of the Yuta model
func NewYutaComponent(ship data.Ship, playerName string, credits int, version string) YutaComponent {
	return YutaComponent{
		Ship:       ship,
		PlayerName: playerName,
		ShipName:   ship.ShipName,
		Version:    version,

		// If fuel 20% or less, suggest refuel
		SuggestRefuel: ship.Fuel <= 20,

		// If hull integrity 50% or less, suggest repair
		SuggestRepair: ship.HullIntegrity <= 50,

		// If money is low, suggest money making
		SuggestCredits: credits <= 100,
	}
}

func (m YutaComponent) View() string {
	var assistantText string

	// Prioritized suggestions
	if m.SuggestRefuel {
		assistantText = fmt.Sprintf("%s, I recommend refueling the %s.\n\nFortunately, fuel prices are below market value at the nearest Station.", m.PlayerName, m.ShipName)
	} else if m.SuggestRepair {
		assistantText = fmt.Sprintf("%s, I recommend repairing the %s's hull.\n\nTechnicians are available at the Station.", m.PlayerName, m.ShipName)
	} else if m.SuggestCredits {
		assistantText = fmt.Sprintf("%s, I recommend earning some credits.\n\nYou can do this by completing new missions.", m.PlayerName)
	} else {
		assistantText = "Everything is in order."
	}
	assistant := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Render("^_^") // Make a cute lil robot with rounded borders

	// TODO crew morale system
	//moraleText := "The crew are in high spirits."

	//weatherText := "Weather report: " + weatherList[1]

	return fmt.Sprintf("%s\n\n%s",
		assistant, assistantText)
}

// TODO weather report for immersion (no gameplay effect)
// var weatherList = []string{"Solar Flares",
// 	"Solar Winds",
// 	"Coronal Mass Ejections",
// 	"Geomagnetic Storms",
// 	"Cosmic Rays",
// 	"Radiation Storms",
// 	"Plasma Ejections",
// 	"Microgravity Dust Storms",
// }
