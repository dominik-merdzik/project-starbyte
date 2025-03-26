package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// PanelFocus represents which panel is currently focused
type PanelFocus int

const (
	PanelLeft PanelFocus = iota
	PanelCenter
	PanelRight
)

// MapModel represents the map interface
type MapModel struct {
	GameMap        data.GameMap
	Ship           data.Ship
	SystemCursor   int
	PlanetCursor   int
	ConfirmCursor  int
	ActiveView     ActiveView
	ActivePanel    PanelFocus
	SelectedSystem data.StarSystem
	SelectedPlanet data.Planet

	GameSave        *data.FullGameSave
	locationService *data.LocationService
}

// NewMapModel initializes the star system list
func NewMapModel(gameMap data.GameMap, ship data.Ship, gameSave *data.FullGameSave) MapModel {
	return MapModel{
		GameMap:         gameMap,
		Ship:            ship, // A copy of the global ship
		GameSave:        gameSave,
		SystemCursor:    0,
		PlanetCursor:    0,
		ActiveView:      ViewStarSystems,
		ActivePanel:     PanelLeft,
		locationService: data.NewLocationService(gameMap),
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

// This signals game.go to update the ship's location and fuel
// Used when travelling to a planet
type TravelUpdateMsg struct {
	Location   data.Location
	Fuel       int
	ShowTravel bool
}

func (m MapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		// horizontal navigation only in planet view
		if m.ActiveView == ViewPlanets {
			switch key {
			case "l", "right":
				if m.ActivePanel < PanelRight {
					m.ActivePanel++
				}
				return m, nil
			case "h", "left":
				if m.ActivePanel > PanelLeft {
					m.ActivePanel--
				}
				return m, nil
			}
		}

		switch key {
		case "up", "k":
			// vertical navigation
			if m.ActiveView == ViewStarSystems || (m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft) {
				if m.SystemCursor > 0 {
					m.SystemCursor--
				}
				if m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft {
					m.SelectedSystem = m.GameMap.StarSystems[m.SystemCursor]
					m.PlanetCursor = 0
					m.SelectedPlanet = m.SelectedSystem.Planets[m.PlanetCursor]
				}
			} else if m.ActiveView == ViewPlanets && (m.ActivePanel == PanelCenter || m.ActivePanel == PanelRight) {
				if m.PlanetCursor > 0 {
					m.PlanetCursor--
					m.SelectedPlanet = m.SelectedSystem.Planets[m.PlanetCursor]
				}
			} else if m.ActiveView == ViewTravelConfirm {
				if m.ConfirmCursor > 0 {
					m.ConfirmCursor--
				}
			}
		case "down", "j":
			if m.ActiveView == ViewStarSystems || (m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft) {
				if m.SystemCursor < len(m.GameMap.StarSystems)-1 {
					m.SystemCursor++
				}
				if m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft {
					m.SelectedSystem = m.GameMap.StarSystems[m.SystemCursor]
					m.PlanetCursor = 0
					m.SelectedPlanet = m.SelectedSystem.Planets[m.PlanetCursor]
				}
			} else if m.ActiveView == ViewPlanets && (m.ActivePanel == PanelCenter || m.ActivePanel == PanelRight) {
				if m.PlanetCursor < len(m.SelectedSystem.Planets)-1 {
					m.PlanetCursor++
					m.SelectedPlanet = m.SelectedSystem.Planets[m.PlanetCursor]
				}
			} else if m.ActiveView == ViewTravelConfirm {
				if m.ConfirmCursor < 1 {
					m.ConfirmCursor++
				}
			}
		case "enter":
			switch m.ActiveView {
			case ViewStarSystems:
				// select star system and switch to planet view
				m.SelectedSystem = m.GameMap.StarSystems[m.SystemCursor]
				m.ActiveView = ViewPlanets
				m.ActivePanel = PanelCenter
				m.PlanetCursor = 0
			case ViewPlanets:
				// only allow selecting a planet if the center panel is active
				if m.ActivePanel == PanelCenter {
					m.SelectedPlanet = m.SelectedSystem.Planets[m.PlanetCursor]
					m.ActiveView = ViewTravelConfirm
					m.ConfirmCursor = 0
				}
			case ViewTravelConfirm:
				// If user selects "Confirm" to travel
				if m.ConfirmCursor == 0 {
					// Create location
					destination := data.Location{
						StarSystemName: m.SelectedSystem.Name,
						PlanetName:     m.SelectedPlanet.Name,
						Coordinates:    m.SelectedPlanet.Coordinates,
					}

					// Calculate fuel after traveling
					newFuel := m.locationService.GetFuelCost(m.Ship.Location.Coordinates, m.SelectedPlanet.Coordinates, m.Ship.Location.StarSystemName, m.SelectedSystem.Name, m.Ship.EngineHealth, m.GameSave.Ship.Fuel)

					// Update ship location and fuel
					m.Ship.Location = destination
					m.Ship.Fuel = newFuel

					// Return a message to signal an update to the parent state in game.go
					return m, tea.Batch(
						func() tea.Msg {
							return TravelUpdateMsg{
								Location:   destination,
								Fuel:       m.Ship.Fuel,
								ShowTravel: true, // Add this flag to indicate travel animation should show
							}
						},
						func() tea.Msg {
							// Exit the modal view by sending an ESC key message (janky method)
							return tea.KeyMsg{Type: tea.KeyEsc}
						},
					)
				} else {
					// Else go back
					m.ActiveView = ViewPlanets
				}
			}
		case "esc":
			switch m.ActiveView {
			case ViewPlanets:
				m.ActiveView = ViewStarSystems
				m.ActivePanel = PanelLeft
			case ViewTravelConfirm:
				m.ActiveView = ViewPlanets
			}
		}
	}
	return m, nil
}

func (m MapModel) View() string {
	// in non-modal modes, render the composite three-panel view
	if m.ActiveView == ViewStarSystems || m.ActiveView == ViewPlanets {
		leftPanel := m.renderStarSystemList()
		centerPanel := m.renderPlanetList()
		rightPanel := m.renderPlanetDetails()
		return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, centerPanel, rightPanel)
	}
	// in travel confirmation mode, show only the modal
	return m.renderTravelConfirm()
}

// Renders the star system list.
func (m MapModel) renderStarSystemList() string {
	panelStyle := lipgloss.NewStyle().
		Width(30).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder())
	// use a highlighted border when the panel is active
	if m.ActiveView == ViewStarSystems || (m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft) {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("215")).Bold(true)
	} else {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("63"))
	}

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var sb strings.Builder
	for i, system := range m.GameMap.StarSystems {
		titleText := system.Name
		// Wwen in star system view or if left panel is active in planet view, use SystemCursor
		if m.ActiveView == ViewStarSystems || (m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft) {
			if i == m.SystemCursor {
				sb.WriteString(fmt.Sprintf("%s %s\n", arrowStyle.Render(">"), hoverStyle.Render(titleText)))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
			}
		} else if m.ActiveView == ViewPlanets {
			if m.SelectedSystem.Name == system.Name {
				sb.WriteString(fmt.Sprintf("%s %s\n", arrowStyle.Render(">"), hoverStyle.Render(titleText)))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
			}
		}
	}
	return panelStyle.Render(sb.String())
}

// renders the planet list
func (m MapModel) renderPlanetList() string {
	panelStyle := lipgloss.NewStyle().
		Width(30).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder())

	if m.ActiveView == ViewPlanets && m.ActivePanel == PanelCenter {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("215")).Bold(true)
	} else {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("63"))
	}

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var sb strings.Builder
	for i, planet := range m.SelectedSystem.Planets {
		titleText := planet.Name

		// only show the arrow if the center panel is active
		if m.ActiveView == ViewPlanets && m.ActivePanel == PanelCenter && i == m.PlanetCursor {
			sb.WriteString(fmt.Sprintf("%s %s\n", arrowStyle.Render(">"), hoverStyle.Render(titleText)))
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
		}
	}
	return panelStyle.Render(sb.String())
}

// renders planet details
func (m MapModel) renderPlanetDetails() string {
	panelStyle := lipgloss.NewStyle().
		Width(40). // Adjusted width slightly for more space
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder())
	if m.ActiveView == ViewPlanets && m.ActivePanel == PanelRight {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("215")).Bold(true)
	} else {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("63"))
	}

	// Define text styles.
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red for errors/warnings

	// If no system is selected or there are no planets, return a default message.
	if m.SelectedSystem.Name == "" || len(m.SelectedSystem.Planets) == 0 {
		return panelStyle.Render(` 
                .::.
                  .:'  .:
        ,MMM8&&&.:'   .:'
       MMMMM88&&&&  .:'
      MMMMM88&&&&&&:'
      MMMMM88&&&&&&
    .:MMMMM88&&&&&&
  .:'  MMMMM88&&&&
.:'   .:'MMM8&&&'
:'  .:'
'::'  
		`)
	}

	// Safely get the planet using PlanetCursor.
	if m.PlanetCursor < 0 || m.PlanetCursor >= len(m.SelectedSystem.Planets) {
		return panelStyle.Render("Error: Invalid planet index.") // safety check
	}
	planet := m.SelectedSystem.Planets[m.PlanetCursor] // potential destination

	// calculate actual travel duration
	currentLocation := m.Ship.Location
	destinationLocation := data.Location{
		StarSystemName: m.SelectedSystem.Name,
		PlanetName:     planet.Name,
		Coordinates:    planet.Coordinates,
	}
	// get engine details directly from the GameSave pointer for most up-to-date info
	engineLevel := m.GameSave.Ship.Upgrades.Engine.CurrentLevel
	maxEngineLevel := m.GameSave.Ship.Upgrades.Engine.MaxLevel // Ensure this exists in data.UpgradeLevel
	travelDuration := m.locationService.CalculateTravelDuration(currentLocation, destinationLocation, engineLevel, maxEngineLevel)

	// calculate estimated fuel on arrival ---
	estimatedFuelOnArrival := m.locationService.GetFuelCost(
		currentLocation.Coordinates,
		destinationLocation.Coordinates,
		currentLocation.StarSystemName,
		destinationLocation.StarSystemName,
		m.Ship.EngineHealth,  // Use current EngineHealth from MapModel's Ship copy
		m.GameSave.Ship.Fuel, // Use current fuel from the GameSave pointer
	)
	// check if travel is impossible due to fuel
	isFuelInsufficient := estimatedFuelOnArrival < 0

	var b strings.Builder

	// basic planet info
	b.WriteString(labelStyle.Render("Planet: ") + valueStyle.Render(planet.Name) + "\n")
	b.WriteString(labelStyle.Render("Type: ") + valueStyle.Render(planet.Type) + "\n")
	b.WriteString(labelStyle.Render("Coordinates: ") + valueStyle.Render(
		fmt.Sprintf("(%d, %d, %d)", planet.Coordinates.X, planet.Coordinates.Y, planet.Coordinates.Z)) + "\n\n") // Added newline

	// display travel duration in seconds
	b.WriteString(titleStyle.Render(fmt.Sprintf("Est. Travel Time: %.1f sec", travelDuration.Seconds())) + "\n") // <-- Use calculated duration

	// display ESTIMATED fuel on arrival, highlight if insufficient
	fuelStr := fmt.Sprintf("%d units", estimatedFuelOnArrival)
	fuelStyle := titleStyle
	if isFuelInsufficient {
		fuelStr = fmt.Sprintf("%d units (INSUFFICIENT!)", estimatedFuelOnArrival)
		fuelStyle = titleStyle.Copy().Foreground(errorStyle.GetForeground())
	}
	b.WriteString(fuelStyle.Render("Est. Fuel on Arrival: " + fuelStr + "\n\n"))

	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", int(panelStyle.GetWidth())-2)) // Adjust width
	b.WriteString(divider + "\n")

	// append the requirements section
	b.WriteString(labelStyle.Render("Requirements:") + "\n")
	if len(planet.Requirements) == 0 {
		b.WriteString("  " + valueStyle.Render("None"))
	} else {
		for _, req := range planet.Requirements {
			// Check if player meets requirement (example logic - needs access to crew data)
			// met := checkCrewRequirement(m.GameSave.Crew, req) // You'd need a helper function
			// reqStr := fmt.Sprintf("%s (Degree %d, Count %d)", req.Role, req.Degree, req.Count)
			// reqStyle := bulletStyle
			// if !met { reqStyle = errorStyle }
			// b.WriteString(fmt.Sprintf("  • %s\n", reqStyle.Render(reqStr)))

			b.WriteString(fmt.Sprintf("  %s %s (Degree %d, Count %d)\n",
				bulletStyle.Render("•"), valueStyle.Render(req.Role), req.Degree, req.Count))
		}
	}

	return panelStyle.Render(b.String())
}

// renders the travel confirmation pop-up modal
func (m MapModel) renderTravelConfirm() string {
	style := lipgloss.NewStyle().
		Width(60).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center)

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	var content strings.Builder

	// ensure SelectedPlanet is valid
	if m.SelectedPlanet.Name == "" {
		return style.Render("Error: No planet selected for confirmation.")
	}
	planet := m.SelectedPlanet // this is the destination planet

	// travel duration
	currentLocation := m.Ship.Location
	destinationLocation := data.Location{
		StarSystemName: m.SelectedSystem.Name,
		PlanetName:     planet.Name,
		Coordinates:    planet.Coordinates,
	}
	engineLevel := m.GameSave.Ship.Upgrades.Engine.CurrentLevel
	maxEngineLevel := m.GameSave.Ship.Upgrades.Engine.MaxLevel
	travelDuration := m.locationService.CalculateTravelDuration(currentLocation, destinationLocation, engineLevel, maxEngineLevel)

	// estimated fuel on arrival
	estimatedFuelOnArrival := m.locationService.GetFuelCost(
		currentLocation.Coordinates,
		destinationLocation.Coordinates,
		currentLocation.StarSystemName,
		destinationLocation.StarSystemName,
		m.Ship.EngineHealth,
		m.GameSave.Ship.Fuel,
	)
	isFuelInsufficient := estimatedFuelOnArrival < 0

	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	content.WriteString(fmt.Sprintf("Confirm travel to %s?\n\n", valueStyle.Render(planet.Name)))

	// Display ACTUAL travel duration in seconds
	content.WriteString(fmt.Sprintf("Est. Travel Time: %.1f seconds\n", travelDuration.Seconds()))

	// estimated fuel on arrival, highlight if insufficient
	fuelStr := fmt.Sprintf("%d units", estimatedFuelOnArrival)
	fuelStyle := defaultStyle // default style for value
	if isFuelInsufficient {
		fuelStr = fmt.Sprintf("%d units (INSUFFICIENT!)", estimatedFuelOnArrival)
		fuelStyle = errorStyle // error style for value if insufficient
	}
	content.WriteString(fmt.Sprintf("Est. Fuel on Arrival: %s\n\n", fuelStyle.Render(fuelStr)))

	// Travel options highlight insufficient fuel and disable confirm
	options := []string{"Confirm", "Cancel"}
	for i, option := range options {
		optionStyle := defaultStyle
		prefix := "  "
		if i == m.ConfirmCursor {
			optionStyle = hoverStyle
			prefix = arrowStyle.Render("> ")
		}

		if i == 0 && isFuelInsufficient {
			optionStyle = optionStyle.Copy().Faint(true).Strikethrough(true)
		}

		content.WriteString(fmt.Sprintf("%s%s\n", prefix, optionStyle.Render(option)))
	}
	return style.Render(content.String())
}
