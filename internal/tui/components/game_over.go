package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// This struct is only necessary if we want this component to control game over logic
type GameOverComponent struct {
	// TODO: Show other stats we care about
	Ship data.Ship
	//DeathMessage string // We could pass in a death message here to specify the cause of death
}

func NewGameOverComponent(ship data.Ship) GameOverComponent {
	return GameOverComponent{
		Ship: ship,
	}
}

func (g GameOverComponent) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Render("GAME OVER")
	detailText := "The crew of the " + g.Ship.ShipName + " have perished in the depths of space.\n\nThank you for playing Starbyte.\n\n Press [Q] to quit."

	// TODO: List crew member names

	return fmt.Sprintf("%s\n\n%s\n\n", title, detailText)
}
