package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// MapModel represents the map interface, and is displaying star systems and planets
type MapModel struct {
	GameMap        data.GameMap
	Cursor         int
	ViewingPlanets bool
	SelectedSystem data.StarSystem
}

// NewMapModel initializes the star system list
func NewMapModel(gameMap data.GameMap) MapModel {
	return MapModel{
		GameMap:        gameMap,
		Cursor:         0,
		ViewingPlanets: false, // Start in star system view
	}
}

func (m MapModel) Init() tea.Cmd {
	return nil
}

func (m MapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}

		case "down", "j":
			if m.ViewingPlanets {
				if m.Cursor < len(m.SelectedSystem.Planets)-1 {
					m.Cursor++
				}
			} else {
				if m.Cursor < len(m.GameMap.StarSystems)-1 {
					m.Cursor++
				}
			}

		case "enter":
			if !m.ViewingPlanets {
				// Entering planet view
				m.ViewingPlanets = true
				m.SelectedSystem = m.GameMap.StarSystems[m.Cursor]
				m.Cursor = 0 // Reset cursor to the first planet
			}

		case "esc":
			if m.ViewingPlanets {
				// Go back to star system view
				m.ViewingPlanets = false
				m.Cursor = 0 // Reset cursor to the first star system
			}
		}
	}
	return m, nil
}

// View renders the list of star systems or planets
func (m MapModel) View() string {
	if m.ViewingPlanets {
		return m.renderPlanetView()
	}
	return m.renderStarSystemView()
}

// Renders the list of star systems
func (m MapModel) renderStarSystemView() string {
	if len(m.GameMap.StarSystems) == 0 {
		return "No star systems available."
	}

	leftStyle := lipgloss.NewStyle().
		Width(50).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var starList strings.Builder
	for i, system := range m.GameMap.StarSystems {
		titleText := fmt.Sprintf("%s [%s]", system.Name, system.SystemID)
		if i == m.Cursor {
			starList.WriteString(fmt.Sprintf("%s %s\n",
				arrowStyle.Render(">"),
				hoverStyle.Render(titleText)))
		} else {
			starList.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
		}
	}

	leftPanel := leftStyle.Render(starList.String())

	rightStyle := lipgloss.NewStyle().
		Width(50).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)

	var systemDetails string
	if len(m.GameMap.StarSystems) > 0 {
		system := m.GameMap.StarSystems[m.Cursor]
		systemDetails = titleStyle.Render(system.Name) + "\n" +
			labelStyle.Render("System ID: ") + system.SystemID + "\n" +
			labelStyle.Render("Coordinates: ") +
			fmt.Sprintf("(%d, %d, %d)", system.Coordinates.X, system.Coordinates.Y, system.Coordinates.Z) + "\n" +
			labelStyle.Render("Planets: ") + fmt.Sprintf("%d", len(system.Planets))
	}

	rightPanel := rightStyle.Render(systemDetails)

	divider := lipgloss.NewStyle().Width(1).Height(15).Align(lipgloss.Center).
		Foreground(lipgloss.Color("240")).Render(`
│
│
│
│
│
│
│
│
│
│
│
│
│
│
│
`)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, divider, rightPanel)
}

// Renders the list of planets within a selected star system
func (m MapModel) renderPlanetView() string {
	leftStyle := lipgloss.NewStyle().
		Width(40).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var planetList strings.Builder
	for i, planet := range m.SelectedSystem.Planets {
		titleText := fmt.Sprintf("%s [%s]", planet.Name, planet.PlanetID)
		if i == m.Cursor {
			planetList.WriteString(fmt.Sprintf("%s %s\n",
				arrowStyle.Render(">"),
				hoverStyle.Render(titleText)))
		} else {
			planetList.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
		}
	}

	leftPanel := leftStyle.Render(planetList.String())

	rightStyle := lipgloss.NewStyle().
		Width(40).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)

	var planetDetails string
	if len(m.SelectedSystem.Planets) > 0 {
		planet := m.SelectedSystem.Planets[m.Cursor]
		planetDetails = titleStyle.Render(planet.Name) + "\n" +
			labelStyle.Render("Planet ID: ") + planet.PlanetID + "\n" +
			labelStyle.Render("Type: ") + planet.Type + "\n" +
			labelStyle.Render("Coordinates: ") +
			fmt.Sprintf("(%d, %d, %d)", planet.Coordinates.X, planet.Coordinates.Y, planet.Coordinates.Z)
	}

	rightPanel := rightStyle.Render(planetDetails)

	divider := lipgloss.NewStyle().Width(1).Height(15).Align(lipgloss.Center).
		Foreground(lipgloss.Color("240")).Render(`
│
│
│
│
│
│
│
│
│
│
│
│
│
│
│
`)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, divider, rightPanel)
}
