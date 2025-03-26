package model

import (
	"fmt"
	"strings"
	"time"

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
					// Check if already at the selected planet BEFORE switching view
					currentLocation := m.Ship.Location
					destinationPlanet := m.SelectedSystem.Planets[m.PlanetCursor]
					destinationLocation := data.Location{
						StarSystemName: m.SelectedSystem.Name,
						PlanetName:     destinationPlanet.Name,
						Coordinates:    destinationPlanet.Coordinates,
					}
					isAlreadyHere := currentLocation.IsEqual(destinationLocation)

					if !isAlreadyHere { // Only enter confirm view if NOT already here
						m.SelectedPlanet = destinationPlanet
						m.ActiveView = ViewTravelConfirm
						m.ConfirmCursor = 0 // Reset confirm cursor
					} else {
						// Optional: Add a brief notification or sound? For now, just do nothing.
						return m, nil
					}
				}

			case ViewTravelConfirm:
				// Check fuel again for safety, although UI should indicate it
				currentLocation := m.Ship.Location
				destinationPlanet := m.SelectedPlanet // Already selected
				destinationLocation := data.Location{
					StarSystemName: m.SelectedSystem.Name,
					PlanetName:     destinationPlanet.Name,
					Coordinates:    destinationPlanet.Coordinates,
				}
				estimatedFuelOnArrival := m.locationService.GetFuelCost(
					currentLocation.Coordinates, destinationLocation.Coordinates,
					currentLocation.StarSystemName, destinationLocation.StarSystemName,
					m.Ship.EngineHealth, m.GameSave.Ship.Fuel,
				)
				isFuelInsufficient := estimatedFuelOnArrival < 0
				// isAlreadyHere check is implicitly handled because we don't enter this view if already here

				// If user selects "Confirm" (cursor 0) AND fuel is sufficient
				if m.ConfirmCursor == 0 && !isFuelInsufficient {
					// Prepare destination for the message
					destination := destinationLocation // Use the already constructed destination

					// Return a message to signal game.go to START travel
					return m, tea.Batch(
						func() tea.Msg {
							return TravelUpdateMsg{
								Location:   destination,
								Fuel:       0, // This field is unused by game.go now
								ShowTravel: true,
							}
						},
						func() tea.Msg {
							// Exit the map view entirely after confirming travel
							return tea.KeyMsg{Type: tea.KeyEsc} // Send Esc to game.go
						},
					)
				} else if m.ConfirmCursor == 1 { // Cancel selected
					m.ActiveView = ViewPlanets // Go back to planet selection
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
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(" #C0C0C0"))
	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

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
		return panelStyle.Render("Error: Invalid planet index.")
	}
	planet := m.SelectedSystem.Planets[m.PlanetCursor] // potential destination

	// check if already at the selected planet
	currentLocation := m.Ship.Location
	destinationLocation := data.Location{
		StarSystemName: m.SelectedSystem.Name,
		PlanetName:     planet.Name,
		Coordinates:    planet.Coordinates,
	}
	isAlreadyHere := currentLocation.IsEqual(destinationLocation)

	var travelDuration time.Duration
	var estimatedFuelOnArrival int
	var isFuelInsufficient bool

	if !isAlreadyHere { // Only calculate if not already here
		engineLevel := m.GameSave.Ship.Upgrades.Engine.CurrentLevel
		maxEngineLevel := m.GameSave.Ship.Upgrades.Engine.MaxLevel
		travelDuration = m.locationService.CalculateTravelDuration(currentLocation, destinationLocation, engineLevel, maxEngineLevel)

		estimatedFuelOnArrival = m.locationService.GetFuelCost(
			currentLocation.Coordinates, destinationLocation.Coordinates,
			currentLocation.StarSystemName, destinationLocation.StarSystemName,
			m.Ship.EngineHealth, m.GameSave.Ship.Fuel,
		)
		isFuelInsufficient = estimatedFuelOnArrival < 0
	} else {
		// If already here, duration is 0 and fuel is current fuel
		travelDuration = 0
		estimatedFuelOnArrival = m.GameSave.Ship.Fuel
		isFuelInsufficient = false
	}

	var b strings.Builder

	// basic planet info
	b.WriteString(labelStyle.Render("Planet: ") + valueStyle.Render(planet.Name) + "\n")
	b.WriteString(labelStyle.Render("Type: ") + valueStyle.Render(planet.Type) + "\n")
	b.WriteString(labelStyle.Render("Coordinates: ") + valueStyle.Render(
		fmt.Sprintf("(%d, %d, %d)", planet.Coordinates.X, planet.Coordinates.Y, planet.Coordinates.Z)) + "\n\n")

	// display travel duration in seconds
	if isAlreadyHere {
		currentLocStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
		b.WriteString(currentLocStyle.Render("Location: CURRENT") + "\n") // Use special style
	} else {
		b.WriteString(titleStyle.Render(fmt.Sprintf("Est. Travel Time: %.1f sec", travelDuration.Seconds())) + "\n")
	}

	// display estimated fuel on arrival, highlight if insufficient
	fuelStr := fmt.Sprintf("%d units", estimatedFuelOnArrival)
	fuelStyle := titleStyle // Use title style as base
	if isFuelInsufficient {
		fuelStr = fmt.Sprintf("%d units (INSUFFICIENT!)", estimatedFuelOnArrival)
		fuelStyle = fuelStyle.Copy().Foreground(errorStyle.GetForeground())
	} else if isAlreadyHere {
		// if already here, maybe don't label it "Est. Fuel on Arrival"
		fuelStr = fmt.Sprintf("%d units", estimatedFuelOnArrival)
		b.WriteString(fuelStyle.Render("Current Fuel: " + fuelStr + "\n\n"))
	} else {
		b.WriteString(fuelStyle.Render("Est. Fuel on Arrival: " + fuelStr + "\n\n"))
	}

	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", int(panelStyle.GetWidth())-2))
	b.WriteString(divider + "\n")

	// append the requirements section
	b.WriteString(labelStyle.Render("Requirements:") + "\n")
	if len(planet.Requirements) == 0 {
		b.WriteString("  " + valueStyle.Render("None"))
	} else {
		for _, req := range planet.Requirements {
			// check if player meets requirement
			met := data.CheckCrewRequirement(m.GameSave.Crew, req)

			// prepare the requirement string
			reqStr := fmt.Sprintf("%s (Degree %d, Count %d)", req.Role, req.Degree, req.Count)

			pointStyle := bulletStyle
			if !met {
				pointStyle = errorStyle // use error style if not met (e.g., red bullet)

			}
			b.WriteString(fmt.Sprintf("  %s %s\n",
				pointStyle.Render("•"),
				valueStyle.Render(reqStr),
			))
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

		// Define styles used
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	infoStyle := lipgloss.NewStyle().Italic(true)

	// ensure SelectedPlanet is valid
	if m.SelectedPlanet.Name == "" {
		return style.Render("Error: No planet selected for confirmation.")
	}
	planet := m.SelectedPlanet // this is the destination planet

	// check if already at the selected planet
	currentLocation := m.Ship.Location
	destinationLocation := data.Location{
		StarSystemName: m.SelectedSystem.Name,
		PlanetName:     planet.Name,
		Coordinates:    planet.Coordinates,
	}
	isAlreadyHere := currentLocation.IsEqual(destinationLocation)

	// if already here, show info message instead of confirm
	if isAlreadyHere {
		return style.Render(infoStyle.Render(fmt.Sprintf("\nYou are already at %s.\n(Press Esc to return to map)",
			valueStyle.Render(planet.Name))))
	}

	// calculate if not already here
	engineLevel := m.GameSave.Ship.Upgrades.Engine.CurrentLevel
	maxEngineLevel := m.GameSave.Ship.Upgrades.Engine.MaxLevel
	travelDuration := m.locationService.CalculateTravelDuration(currentLocation, destinationLocation, engineLevel, maxEngineLevel)
	estimatedFuelOnArrival := m.locationService.GetFuelCost(
		currentLocation.Coordinates, destinationLocation.Coordinates,
		currentLocation.StarSystemName, destinationLocation.StarSystemName,
		m.Ship.EngineHealth, m.GameSave.Ship.Fuel,
	)
	isFuelInsufficient := estimatedFuelOnArrival < 0

	var content strings.Builder
	content.WriteString(fmt.Sprintf("Confirm travel to %s?\n\n", valueStyle.Render(planet.Name)))
	content.WriteString(fmt.Sprintf("Est. Travel Time: %.1f seconds\n", travelDuration.Seconds()))

	fuelStr := fmt.Sprintf("%d units", estimatedFuelOnArrival)
	fuelStyle := defaultStyle
	if isFuelInsufficient {
		fuelStr = fmt.Sprintf("%d units (INSUFFICIENT!)", estimatedFuelOnArrival)
		fuelStyle = errorStyle
	}
	content.WriteString(fmt.Sprintf("Est. Fuel on Arrival: %s\n\n", fuelStyle.Render(fuelStr)))

	// travel options - disable Confirm if fuel insufficient
	options := []string{"Confirm", "Cancel"}
	for i, option := range options {
		optionStyle := defaultStyle
		prefix := "  "
		isDisabled := (i == 0 && isFuelInsufficient)

		// Allow cursor only on enabled options
		if i == m.ConfirmCursor {
			if !isDisabled {
				optionStyle = hoverStyle
				prefix = arrowStyle.Render("> ")
			} else {
				optionStyle = optionStyle.Copy().Faint(true).Strikethrough(true)
				prefix = "  "
			}
		} else if isDisabled && i == 0 {
			optionStyle = optionStyle.Copy().Faint(true).Strikethrough(true)
		}

		content.WriteString(fmt.Sprintf("%s%s\n", prefix, optionStyle.Render(option)))
	}
	return style.Render(content.String())
}
