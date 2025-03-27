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

var ftlRequiredSystems = map[string]bool{
	"Alpha Centauri": true,
	"Sirius":         true,
	"Vega":           true,
}

// MapModel represents the map interface
type MapModel struct {
	GameMap         data.GameMap
	Ship            data.Ship
	SystemCursor    int
	PlanetCursor    int
	ConfirmCursor   int
	ActiveView      ActiveView
	ActivePanel     PanelFocus
	SelectedSystem  data.StarSystem
	SelectedPlanet  data.Planet
	GameSave        *data.FullGameSave
	locationService *data.LocationService
}

// NewMapModel initializes the star system list
func NewMapModel(gameMap data.GameMap, ship data.Ship, gameSave *data.FullGameSave) MapModel {
	return MapModel{
		GameMap:         gameMap,
		Ship:            ship,
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

// TravelUpdateMsg signals game.go to update the ship's location and fuel.
// Used when travelling to a planet.
type TravelUpdateMsg struct {
	Location   data.Location
	Fuel       int
	ShowTravel bool
}

func (m MapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// helper function for selecting system and moving focus/view
		attemptSelectSystemAndFocusCenter := func() (MapModel, tea.Cmd) {
			// check bounds and get the system under the cursor
			if m.SystemCursor >= 0 && m.SystemCursor < len(m.GameMap.StarSystems) {
				systemToSelect := m.GameMap.StarSystems[m.SystemCursor]
				requiresFTL := ftlRequiredSystems[systemToSelect.Name]
				isLocked := !m.Ship.HasFTLDrive && requiresFTL

				if isLocked {
					// if the system is locked, do nothing
					return m, nil
				}

				// if NOT locked, proceed to select system, switch view/panel
				m.SelectedSystem = systemToSelect
				m.ActiveView = ViewPlanets
				m.ActivePanel = PanelCenter
				m.PlanetCursor = 0
				// pre-select the first planet
				if len(m.SelectedSystem.Planets) > 0 {
					m.SelectedPlanet = m.SelectedSystem.Planets[0]
				} else {
					m.SelectedPlanet = data.Planet{} // handle empty system
				}
			}
			return m, nil // return updated model
		}

		switch key {
		// horizontal navigation
		case "l", "right":
			// if focus is on Left panel, try to select/move Center
			if m.ActivePanel == PanelLeft {
				return attemptSelectSystemAndFocusCenter() // use shared logic
			}
			// if already Center or Right, 'l' does nothing more horizontally
			return m, nil // explicit return needed

		case "h", "left", "b":
			// only allow moving left from Center panel WHEN in ViewPlanets
			if m.ActiveView == ViewPlanets && m.ActivePanel == PanelCenter {
				m.ActivePanel = PanelLeft // move focus left
				// reset selections when focus returns to the Star System list
				m.SelectedSystem = data.StarSystem{}
				m.SelectedPlanet = data.Planet{}
				m.PlanetCursor = 0
			}
			// if already Left, 'h' does nothing
			return m, nil // explicit return needed

		// vertical navigation
		case "up", "k":
			// PRIORITIZE checking if we are in the confirmation view
			if m.ActiveView == ViewTravelConfirm {
				if m.ConfirmCursor > 0 {
					m.ConfirmCursor--
				}
				// if NOT in confirmation view, THEN handle panel-based navigation
			} else if m.ActivePanel == PanelLeft {
				// navigate Star Systems list
				if m.SystemCursor > 0 {
					m.SystemCursor--
				}
			} else if m.ActivePanel == PanelCenter {
				// navigate Planet list (only reachable if not ViewTravelConfirm)
				if m.PlanetCursor > 0 {
					m.PlanetCursor--
					// update SelectedPlanet only if valid
					if m.SelectedSystem.Name != "" && m.PlanetCursor >= 0 && m.PlanetCursor < len(m.SelectedSystem.Planets) {
						m.SelectedPlanet = m.SelectedSystem.Planets[m.PlanetCursor]
					}
				}
			}

		case "down", "j":
			// PRIORITIZE checking if we are in the confirmation view
			if m.ActiveView == ViewTravelConfirm {
				if m.ConfirmCursor < 1 { // only 2 options: 0 (Confirm), 1 (Cancel)
					m.ConfirmCursor++
				}
				// if NOT in confirmation view, THEN handle panel-based navigation
			} else if m.ActivePanel == PanelLeft {
				// navigate Star Systems list
				if m.SystemCursor < len(m.GameMap.StarSystems)-1 {
					m.SystemCursor++
				}
			} else if m.ActivePanel == PanelCenter {
				// navigate Planet list (only reachable if not ViewTravelConfirm)
				if m.SelectedSystem.Name != "" && m.PlanetCursor < len(m.SelectedSystem.Planets)-1 {
					m.PlanetCursor++
					m.SelectedPlanet = m.SelectedSystem.Planets[m.PlanetCursor]
				}
			}

		// action key
		case "enter":
			// PRIORITIZE checking if we are ALREADY in the confirmation view FIRST
			if m.ActiveView == ViewTravelConfirm {
				// confirm/cancel logic
				currentLocation := m.Ship.Location
				destinationPlanet := m.SelectedPlanet
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

				if m.ConfirmCursor == 0 && !isFuelInsufficient { // confirm selected and fuel OK
					destination := destinationLocation
					return m, tea.Batch(
						func() tea.Msg {
							return TravelUpdateMsg{
								Location:   destination,
								Fuel:       0,
								ShowTravel: true,
							}
						},
						func() tea.Msg {
							return tea.KeyMsg{Type: tea.KeyEsc}
						},
					)
				} else if m.ConfirmCursor == 1 { // cancel selected
					m.ActiveView = ViewPlanets
					return m, nil
				} else { // confirm selected but insufficient fuel, or invalid cursor
					return m, nil
				}

			} else if m.ActivePanel == PanelLeft {
				// select system logic
				return attemptSelectSystemAndFocusCenter()

			} else if m.ActivePanel == PanelCenter {
				// enter confirm logic
				currentLocation := m.Ship.Location
				var destinationPlanet data.Planet
				if m.SelectedSystem.Name != "" && m.PlanetCursor >= 0 && m.PlanetCursor < len(m.SelectedSystem.Planets) {
					destinationPlanet = m.SelectedSystem.Planets[m.PlanetCursor]
				} else {
					return m, nil // invalid planet selection
				}

				destinationLocation := data.Location{
					StarSystemName: m.SelectedSystem.Name,
					PlanetName:     destinationPlanet.Name,
					Coordinates:    destinationPlanet.Coordinates,
				}
				isAlreadyHere := currentLocation.IsEqual(destinationLocation)
				isInterstellarTravel := m.SelectedSystem.Name != currentLocation.StarSystemName
				needsFTL := isInterstellarTravel

				if !m.Ship.HasFTLDrive && needsFTL {
					return m, nil // FTL required but not available
				}

				if !isAlreadyHere { // only enter confirm if not already there
					m.SelectedPlanet = destinationPlanet
					m.ActiveView = ViewTravelConfirm
					m.ConfirmCursor = 0
					return m, nil // let next View render confirm screen
				} else { // already at the destination
					return m, nil
				}
			}

			// fallback for 'enter' if no conditions met
			return m, nil

		case "esc":
			// always reset selections when escaping views within the map
			m.SelectedSystem = data.StarSystem{}
			m.SelectedPlanet = data.Planet{}
			m.PlanetCursor = 0 // reset planet cursor too

			switch m.ActiveView {
			case ViewPlanets:
				// go back to system view/panel
				m.ActiveView = ViewStarSystems
				m.ActivePanel = PanelLeft
			case ViewTravelConfirm:
				// go back to planet view (center panel remains active)
				m.ActiveView = ViewPlanets
				// case ViewStarSystems:
				// Esc in the top view (StarSystems) might exit the map model entirely,
				// which should be handled by the parent model receiving the tea.KeyMsg.
			}
		} // end outer switch key

	} // end case tea.KeyMsg

	return m, nil // return default if no specific key handling occurred
}

// View renders the map
func (m MapModel) View() string {
	// render the composite three-panel view for non-modal states
	if m.ActiveView == ViewStarSystems || m.ActiveView == ViewPlanets {
		leftPanel := m.renderStarSystemList()
		centerPanel := m.renderPlanetList()
		rightPanel := m.renderPlanetDetails()
		return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, centerPanel, rightPanel)
	}

	// render only the modal for travel confirmation
	return m.renderTravelConfirm()
}

// renderStarSystemList renders the star system list (Left Panel)
func (m MapModel) renderStarSystemList() string {
	panelStyle := lipgloss.NewStyle().
		Width(30).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder())

	// highlight border when the panel is active or potentially active
	if m.ActiveView == ViewStarSystems || (m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft) {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("215")).Bold(true)
	} else {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("240")) // dim border otherwise
	}

	// define text styles
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))               // default item color
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))                 // highlighted/selected item color
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))                  // cursor arrow color
	greyedOutStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217")).Faint(true) // inaccessible item color

	var sb strings.Builder
	for i, system := range m.GameMap.StarSystems {
		titleText := system.Name
		style := defaultStyle // start with default style

		// check accessibility
		requiresFTL := ftlRequiredSystems[system.Name]
		isLocked := !m.Ship.HasFTLDrive && requiresFTL

		// determine if the cursor is here or if the system is selected (even if panel focus moved)
		cursorIsHere := (m.ActiveView == ViewStarSystems || (m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft)) && i == m.SystemCursor
		systemIsSelected := m.ActiveView == ViewPlanets && m.SelectedSystem.Name == system.Name

		// apply styling based on state
		if isLocked {
			style = greyedOutStyle // locked systems are always greyed out
		} else if cursorIsHere || systemIsSelected {
			style = hoverStyle // highlight if cursor is here OR if it's the selected system
		}

		// determine prefix (arrow or spaces)
		prefix := "  " // default: two spaces
		// show arrow only when this panel is actively being navigated
		if cursorIsHere {
			prefix = arrowStyle.Render("> ")
		}

		sb.WriteString(fmt.Sprintf("%s%s\n", prefix, style.Render(titleText)))
	}

	return panelStyle.Render(sb.String())
}

// renderPlanetList renders the planet list (Center Panel)
func (m MapModel) renderPlanetList() string {
	panelStyle := lipgloss.NewStyle().
		Width(30).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder())

	// highlight border only when center panel is the active focus point
	if m.ActiveView == ViewPlanets && m.ActivePanel == PanelCenter {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("215")).Bold(true)
	} else {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("240")) // dim border otherwise
	}

	// ftl warning check
	// check if we should show the FTL warning instead of planets
	shouldCheckFTLWarning := m.ActiveView == ViewStarSystems || (m.ActiveView == ViewPlanets && m.ActivePanel == PanelLeft)
	if shouldCheckFTLWarning {
		if m.SystemCursor >= 0 && m.SystemCursor < len(m.GameMap.StarSystems) {
			hoveredSystem := m.GameMap.StarSystems[m.SystemCursor]
			requiresFTL := ftlRequiredSystems[hoveredSystem.Name]
			isLocked := !m.Ship.HasFTLDrive && requiresFTL

			if isLocked {
				// render warning message instead of planets
				errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				warningMsg := "Upgrade your ship to travel to new planets - Must have FTL Drive"
				// center the warning message within the panel
				return panelStyle.Align(lipgloss.Center, lipgloss.Center).Render(errorStyle.Render(warningMsg))
			}
		}
	}
	// end FTL warning check

	// normal planet listing (only runs if FTL warning wasn't shown)
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217")) // default item color
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))   // highlighted item color
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))    // cursor arrow color

	var sb strings.Builder

	// handle cases where no system is selected or the system has no planets
	if m.SelectedSystem.Name == "" || len(m.SelectedSystem.Planets) == 0 {
		// render an empty panel or a placeholder message
		// return panelStyle.Render("Select a system.") // example placeholder
		return panelStyle.Render("") // keep it empty for now
	}

	// list planets for the selected system
	for i, planet := range m.SelectedSystem.Planets {
		titleText := planet.Name
		style := defaultStyle
		prefix := "  " // default: two spaces

		// highlight and show arrow only if center panel is active and cursor is on this planet
		if m.ActiveView == ViewPlanets && m.ActivePanel == PanelCenter && i == m.PlanetCursor {
			style = hoverStyle
			prefix = arrowStyle.Render("> ")
		}

		sb.WriteString(fmt.Sprintf("%s%s\n", prefix, style.Render(titleText)))
	}

	return panelStyle.Render(sb.String())
}

// renderPlanetDetails renders planet details (Right Panel)
func (m MapModel) renderPlanetDetails() string {
	panelStyle := lipgloss.NewStyle().
		Width(40).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder())

	// highlight border (optional, right panel is usually informational)
	// if m.ActiveView == ViewPlanets && m.ActivePanel == PanelRight {
	// 	panelStyle = panelStyle.BorderForeground(lipgloss.Color("215")).Bold(true)
	// } else {
	panelStyle = panelStyle.BorderForeground(lipgloss.Color("240")) // dim border
	// }

	// define text styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))      // titles like "Est. Travel Time"
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))      // labels like "Planet:", "Type:"
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(" #C0C0C0"))           // values next to labels (use #C0C0C0 for silver/grey)
	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))                // green bullet for met requirements
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))                // red for errors/unmet requirements/insufficient fuel
	currentLocStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46")) // green for "CURRENT" location

	// handle cases where no system/planet is properly selected for display
	if m.SelectedSystem.Name == "" || len(m.SelectedSystem.Planets) == 0 {
		// render placeholder art or message
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
        `) // placeholder art
	}

	// safely get the planet under the cursor
	if m.PlanetCursor < 0 || m.PlanetCursor >= len(m.SelectedSystem.Planets) {
		// this case should ideally not happen if cursor logic is correct, but handle defensively
		return panelStyle.Render(errorStyle.Render("Error: Invalid planet index."))
	}
	planet := m.SelectedSystem.Planets[m.PlanetCursor]

	// calculate travel details
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

	if !isAlreadyHere { // only perform calculations if it's a potential trip
		engineLevel := m.GameSave.Ship.Upgrades.Engine.CurrentLevel
		maxEngineLevel := m.GameSave.Ship.Upgrades.Engine.MaxLevel
		travelDuration = m.locationService.CalculateTravelDuration(currentLocation, destinationLocation, engineLevel, maxEngineLevel)
		estimatedFuelOnArrival = m.locationService.GetFuelCost(
			currentLocation.Coordinates, destinationLocation.Coordinates,
			currentLocation.StarSystemName, destinationLocation.StarSystemName,
			m.Ship.EngineHealth, m.GameSave.Ship.Fuel,
		)
		isFuelInsufficient = estimatedFuelOnArrival < 0
	} else { // if already here
		travelDuration = 0
		estimatedFuelOnArrival = m.GameSave.Ship.Fuel // fuel is just current fuel
		isFuelInsufficient = false
	}
	// end calculation

	var b strings.Builder

	// basic planet info
	b.WriteString(labelStyle.Render("Planet: ") + valueStyle.Render(planet.Name) + "\n")
	b.WriteString(labelStyle.Render("Type: ") + valueStyle.Render(planet.Type) + "\n")
	b.WriteString(labelStyle.Render("Coordinates: ") + valueStyle.Render(
		fmt.Sprintf("(%d, %d, %d)", planet.Coordinates.X, planet.Coordinates.Y, planet.Coordinates.Z)) + "\n\n")

	// travel time / location status
	if isAlreadyHere {
		b.WriteString(currentLocStyle.Render("Location: CURRENT") + "\n")
	} else {
		b.WriteString(titleStyle.Render(fmt.Sprintf("Est. Travel Time: %.1f sec", travelDuration.Seconds())) + "\n")
	}

	// fuel info - reuse logic from renderTravelConfirm's content building for consistency
	// (note: content builder not used here, just the msg/style part)
	var fuelMsg string
	fuelStyle := titleStyle // use title style as base, modify if needed

	if isFuelInsufficient {
		fuelMsg = fmt.Sprintf("%d units (INSUFFICIENT!)", estimatedFuelOnArrival)
		fuelStyle = errorStyle // use error style for insufficient fuel
	} else {
		fuelMsg = fmt.Sprintf("%d units", estimatedFuelOnArrival)
		// fuelStyle remains titleStyle
	}

	// add fuel information line
	if isAlreadyHere {
		b.WriteString(labelStyle.Render("Current Fuel: ") + fuelStyle.Render(fuelMsg) + "\n\n") // show current fuel
	} else {
		b.WriteString(labelStyle.Render("Est. Fuel on Arrival: ") + fuelStyle.Render(fuelMsg) + "\n\n") // show estimate
	}

	// divider
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("─", panelStyle.GetWidth()-int(panelStyle.GetHorizontalPadding())))
	b.WriteString(divider + "\n")

	// requirements section
	b.WriteString(labelStyle.Render("Requirements:") + "\n")
	if len(planet.Requirements) == 0 {
		b.WriteString("  " + valueStyle.Render("None") + "\n") // indent "None"
	} else {
		for _, req := range planet.Requirements {
			met := data.CheckCrewRequirement(m.GameSave.Crew, req)
			reqStr := fmt.Sprintf("%s (Degree %d, Count %d)", req.Role, req.Degree, req.Count)

			pointStyle := bulletStyle // assume met (green bullet)
			if !met {
				pointStyle = errorStyle // use error style (red bullet) if not met
			}

			b.WriteString(fmt.Sprintf("  %s %s\n", // indent requirements
				pointStyle.Render("•"),
				valueStyle.Render(reqStr),
			))
		}
	}

	return panelStyle.Render(b.String())
}

// renderTravelConfirm renders the travel confirmation pop-up modal
func (m MapModel) renderTravelConfirm() string {
	// modal Style
	style := lipgloss.NewStyle().
		Width(60).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")). // use a noticeable border color
		Align(lipgloss.Center)                  // center content within the modal

	// define text styles used within the modal
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217")) // default text
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))   // selected option text
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))    // cursor arrow
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))   // insufficient fuel text
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))   // planet name, values
	infoStyle := lipgloss.NewStyle().Italic(true)                         // info messages like "Already here"

	// ensure a planet is actually selected for confirmation
	if m.SelectedPlanet.Name == "" {
		return style.Render(errorStyle.Render("Error: No planet selected for confirmation."))
	}
	planet := m.SelectedPlanet // the destination

	// calculate travel details (again, for the modal view)
	currentLocation := m.Ship.Location
	destinationLocation := data.Location{
		StarSystemName: m.SelectedSystem.Name,
		PlanetName:     planet.Name,
		Coordinates:    planet.Coordinates,
	}
	isAlreadyHere := currentLocation.IsEqual(destinationLocation)

	// if already here, show an info message instead of the confirmation options
	if isAlreadyHere {
		// this case should ideally be prevented by the Update logic, but render defensively
		return style.Render(infoStyle.Render(fmt.Sprintf("\nYou are already at %s.\n(Press Esc to return to map)",
			valueStyle.Render(planet.Name))))
	}

	// perform calculations only if not already here
	engineLevel := m.GameSave.Ship.Upgrades.Engine.CurrentLevel
	maxEngineLevel := m.GameSave.Ship.Upgrades.Engine.MaxLevel
	travelDuration := m.locationService.CalculateTravelDuration(currentLocation, destinationLocation, engineLevel, maxEngineLevel)
	estimatedFuelOnArrival := m.locationService.GetFuelCost(
		currentLocation.Coordinates, destinationLocation.Coordinates,
		currentLocation.StarSystemName, destinationLocation.StarSystemName,
		m.Ship.EngineHealth, m.GameSave.Ship.Fuel,
	)
	isFuelInsufficient := estimatedFuelOnArrival < 0
	// end calculation

	var content strings.Builder

	// confirmation message
	content.WriteString(fmt.Sprintf("Confirm travel to %s?\n\n", valueStyle.Render(planet.Name)))
	content.WriteString(fmt.Sprintf("Est. Travel Time: %.1f seconds\n", travelDuration.Seconds()))

	// fuel information
	fuelStr := fmt.Sprintf("%d units", estimatedFuelOnArrival)
	fuelStyle := defaultStyle // use default style as base
	if isFuelInsufficient {
		fuelStr = fmt.Sprintf("%d units (INSUFFICIENT!)", estimatedFuelOnArrival)
		fuelStyle = errorStyle // use error style for insufficient fuel
	}
	content.WriteString(fmt.Sprintf("Est. Fuel on Arrival: %s\n\n", fuelStyle.Render(fuelStr)))

	// render Confirm/Cancel Options
	options := []string{"Confirm", "Cancel"}
	for i, option := range options {
		optionStyle := defaultStyle
		prefix := "  "                               // default prefix
		isDisabled := (i == 0 && isFuelInsufficient) // confirm is disabled if fuel is insufficient

		// apply styling based on cursor position and disabled state
		if i == m.ConfirmCursor { // if the cursor is on this option
			if !isDisabled {
				// highlight if not disabled
				optionStyle = hoverStyle
				prefix = arrowStyle.Render("> ")
			} else {
				// style as disabled (faint, strikethrough) if cursor is here but option disabled
				optionStyle = optionStyle.Copy().Faint(true).Strikethrough(true)
				prefix = "  " // no arrow for disabled options
			}
		} else if isDisabled && i == 0 { // style Confirm as disabled even if cursor isn't there
			optionStyle = optionStyle.Copy().Faint(true).Strikethrough(true)
		}

		content.WriteString(fmt.Sprintf("%s%s\n", prefix, optionStyle.Render(option)))
	}

	return style.Render(content.String())
}
