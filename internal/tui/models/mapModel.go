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
	Ship           data.Ship
	Cursor         int
	ConfirmCursor  int
	ActiveView     ActiveView
	SelectedSystem data.StarSystem
	SelectedPlanet data.Planet
}

// NewMapModel initializes the star system list
func NewMapModel(gameMap data.GameMap, ship data.Ship) MapModel {
	return MapModel{
		GameMap:    gameMap,
		Ship:       ship,
		Cursor:     0,
		ActiveView: ViewStarSystems, // Start in star system view
	}
}

func (m MapModel) Init() tea.Cmd {
	return nil
}

type ActiveView int

const (
	ViewStarSystems ActiveView = iota
	ViewPlanets
	ViewTravelConfirm
)

func (m MapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			switch m.ActiveView {
			case ViewStarSystems:
				if m.Cursor > 0 {
					m.Cursor--
				}
			case ViewPlanets:
				if m.Cursor > 0 {
					m.Cursor--
				}
			case ViewTravelConfirm:
				if m.ConfirmCursor > 0 {
					m.ConfirmCursor-- // Only two options: confirm or cancel
				}
			}

		case "down", "j":
			switch m.ActiveView {
			case ViewStarSystems:
				if m.Cursor < len(m.GameMap.StarSystems)-1 {
					m.Cursor++
				}
			case ViewPlanets:
				if m.Cursor < len(m.SelectedSystem.Planets)-1 {
					m.Cursor++
				}
			case ViewTravelConfirm:
				if m.ConfirmCursor < 1 {
					m.ConfirmCursor++ // Only two options: confirm or cancel
				}
			}

		case "enter":
			switch m.ActiveView {
			case ViewStarSystems:
				// Entering planet view
				m.ActiveView = ViewPlanets
				m.SelectedSystem = m.GameMap.StarSystems[m.Cursor]
				m.Cursor = 0 // Reset cursor to the first planet
			case ViewPlanets:
				// Entering planet travel confirmation view
				m.ActiveView = ViewTravelConfirm
				m.SelectedPlanet = m.SelectedSystem.Planets[m.Cursor]
				m.Cursor = 0 // Reset cursor to the first star system

			case ViewTravelConfirm:
				if m.ConfirmCursor == 0 {
					// travel() function
					// Exit mapModel
				}
				// Else nothing happens, go back to planet view
				m.ActiveView = ViewPlanets
			}

		case "esc":
			if m.ActiveView == ViewPlanets {
				// Go back to star system view
				m.ActiveView = ViewStarSystems
				m.Cursor = 0 // Reset cursor to the first star system
			} else if m.ActiveView == ViewTravelConfirm {
				// Go back to planet view
				m.ActiveView = ViewPlanets
			}
		}
	}

	return m, nil
}

// View renders the list of star systems or planets
func (m MapModel) View() string {
	switch m.ActiveView {
	case ViewStarSystems:
		return m.renderStarSystemView()
	case ViewPlanets:
		return m.renderPlanetView()
	case ViewTravelConfirm:
		return m.renderTravelConfirm()
	default:
		return "Unknown view"
	}
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
		Width(50).
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
		Width(50).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)

	var planetDetails string

	if len(m.SelectedSystem.Planets) > 0 {
		planet := m.SelectedSystem.Planets[m.Cursor]
		distance := data.GetDistance(m.GameMap, m.Ship, planet.Name)
		planetDetails = titleStyle.Render(planet.Name) + "\n" +
			labelStyle.Render("Planet ID: ") + planet.PlanetID + "\n" +
			labelStyle.Render("Type: ") + planet.Type + "\n" +
			labelStyle.Render("Coordinates: ") +
			fmt.Sprintf("(%d, %d, %d)", planet.Coordinates.X, planet.Coordinates.Y, planet.Coordinates.Z) + "\n" +
			titleStyle.Render(fmt.Sprintf("Travel Time: %d hours", distance))
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

func (m MapModel) renderTravelConfirm() string {
	// Render list of two options: confirm or cancel
	style := lipgloss.NewStyle().
		Width(60).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center)

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var content strings.Builder

	planet := m.SelectedSystem.Planets[m.ConfirmCursor]
	distance := data.GetDistance(m.GameMap, m.Ship, planet.Name)

	content.WriteString(fmt.Sprintf("Confirm travel to %s?\n", planet.Name))
	content.WriteString(fmt.Sprintf("Travel time: %d hours\n\n", distance))

	// Travel options
	options := []string{"Confirm", "Cancel"}
	for i, option := range options {
		if i == m.ConfirmCursor {
			content.WriteString(fmt.Sprintf("%s %s\n",
				arrowStyle.Render(">"),
				hoverStyle.Render(option)))
		} else {
			content.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(option)))
		}
	}

	return style.Render(content.String())
}

// Update the location of the ship in the save file
func (m MapModel) travel() {
	// Get the selected planet
	selectedPlanet := m.SelectedPlanet

	// Update the ship's location
	m.Ship.Location.PlanetId = selectedPlanet.PlanetID
	m.Ship.Location.Coordinates = selectedPlanet.Coordinates

	// TODO: Save the game
}
