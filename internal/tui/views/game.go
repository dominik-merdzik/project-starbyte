package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
	"github.com/dominik-merdzik/project-starbyte/internal/tui/components"
	model "github.com/dominik-merdzik/project-starbyte/internal/tui/models"
)

// -----------------------------------------------------------------------------
// Flags to determine which view is currently active
type ActiveView int

// This is like a constant enum in C# or Java
// These are the possible views that can be active
const (
	ViewNone ActiveView = iota
	ViewJournal
	ViewCrew
	ViewMap
	ViewShip
)

// -----------------------------------------------------------------------------

type GameModel struct {
	// components
	ProgressBar components.ProgressBar
	Yuta        components.YutaModel
	spinner     spinner.Model

	// additional models
	Ship    model.ShipModel
	Crew    model.CrewModel
	Journal model.JournalModel
	Map     model.MapModel

	currentHealth int
	maxHealth     int

	menuItems  []string // list of menu options
	menuCursor int      // current position of the cursor

	// selectedItem tracks which menu item is selected
	selectedItem string

	activeView ActiveView // The currently active view

	// tracked mission (if any)
	TrackedMission *model.Mission

	isTravelling    bool
	travelStartTime time.Time
	travelDuration  time.Duration
	travelProgress  progress.Model
}

type travelTickMsg struct{}

func (g GameModel) Init() tea.Cmd {
	// Init spinner
	return g.spinner.Tick
	// initialize Yuta's animation (seem to be broken ATM)
	// return tea.Batch(
	// 	g.Yuta.Init(),
	// 	g.spinner.Tick, // Initializing the spinner from here doesn't work for some reason.
	// )
}

func (g GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// always update Yuta and Ship, regardless of focus.
	newYuta, yutaCmd := g.Yuta.Update(msg)
	if y, ok := newYuta.(components.YutaModel); ok {
		g.Yuta = y
	}
	cmds = append(cmds, yutaCmd)

	newShip, shipCmd := g.Ship.Update(msg)
	if s, ok := newShip.(model.ShipModel); ok {
		g.Ship = s
	}
	cmds = append(cmds, shipCmd)

	// Update active view if needed
	switch g.activeView {
	case ViewJournal:
		newJournal, journalCmd := g.Journal.Update(msg)
		if j, ok := newJournal.(model.JournalModel); ok {
			g.Journal = j
		}
		cmds = append(cmds, journalCmd)
	case ViewCrew:
		newCrew, crewCmd := g.Crew.Update(msg)
		if c, ok := newCrew.(model.CrewModel); ok {
			g.Crew = c
		}
		cmds = append(cmds, crewCmd)
	case ViewMap:
		newMap, mapCmd := g.Map.Update(msg)
		if m, ok := newMap.(model.MapModel); ok {
			g.Map = m
		}
		cmds = append(cmds, mapCmd)
	case ViewShip:
		newShip, shipCmd := g.Map.Update(msg)
		if m, ok := newShip.(model.MapModel); ok {
			g.Map = m
		}
		cmds = append(cmds, shipCmd)
	}

	// process key messages.
	switch msg := msg.(type) {
	// Spinner ticks
	case spinner.TickMsg:
		var cmd tea.Cmd
		g.spinner, cmd = g.spinner.Update(msg)
		return g, cmd
	case model.TrackMissionMsg:
		// store the tracked mission and update selectedItem.
		g.TrackedMission = &msg.Mission
	case tea.KeyMsg:
		// Handle escape key for any active view
		if g.activeView != ViewNone && msg.String() == "esc" {
			g.activeView = ViewNone
			g.selectedItem = ""
			return g, tea.Batch(cmds...)
		}

		// If we have an active view, don't process main menu keys
		if g.activeView != ViewNone {
			return g, tea.Batch(cmds...)
		}

		// main menu key handling when not in Journal mode.
		switch msg.String() {
		case "q":
			return g, tea.Quit

		case "a":
			// simulate damage
			g.currentHealth -= 10
			if g.currentHealth < 0 {
				g.currentHealth = 0
			}

		case "h":
			// simulate healing
			g.currentHealth += 10
			if g.currentHealth > g.maxHealth {
				g.currentHealth = g.maxHealth
			}

		// main menu navigation
		case "up", "k":
			if g.menuCursor > 0 {
				g.menuCursor--
			}
		case "down", "j":
			if g.menuCursor < len(g.menuItems)-1 {
				g.menuCursor++
			}
		case "enter":
			g.selectedItem = g.menuItems[g.menuCursor]

			// Switch to the selected view using the iota
			switch g.selectedItem {
			case "Journal":
				g.activeView = ViewJournal
			case "Crew":
				g.activeView = ViewCrew
			case "Map":
				g.activeView = ViewMap
			case "Ship":
				g.activeView = ViewShip
			}
			// Press SPACE to cycle through mission status
		case " ":
			// If TrackedMission is not null and is not already travelling on a mission
			if g.TrackedMission != nil && !g.isTravelling {
				if g.TrackedMission.Status == "Not Started" {
					// Start travel when mission begins
					g.isTravelling = true
					g.travelStartTime = time.Now()
					g.travelDuration = time.Duration(g.TrackedMission.TravelTime) * time.Second

					// Reset progress bar
					g.travelProgress.SetPercent(0)

					// Return commands for both travel tick and progress animation
					return g, tea.Batch(
						tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return travelTickMsg{} }),
						g.travelProgress.Init(),
					)
				} else if g.TrackedMission.Status == "In Progress" {
					g.TrackedMission.Status = "Completed"
				}
			}
		}
	case progress.FrameMsg:
		// Handle progress bar animation frames
		if g.isTravelling {
			progressModel, cmd := g.travelProgress.Update(msg)
			if p, ok := progressModel.(progress.Model); ok {
				g.travelProgress = p
			}
			cmds = append(cmds, cmd)
		}
	case travelTickMsg:
		if g.isTravelling && g.TrackedMission != nil {
			elapsed := time.Since(g.travelStartTime)

			// Update progress percentage based on elapsed time
			percentComplete := float64(elapsed) / float64(g.travelDuration)
			if percentComplete > 1.0 {
				percentComplete = 1.0
			}

			cmd := g.travelProgress.SetPercent(percentComplete)

			if elapsed >= g.travelDuration {
				// Travel complete
				g.isTravelling = false
				g.TrackedMission.Status = "In Progress"
				return g, nil
			}

			// Continue checking travel status
			return g, tea.Batch(
				tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return travelTickMsg{} }),
				cmd,
			)
		}
	}
	cmds = append(cmds, g.spinner.Tick) // This is needed to animate the spinner

	return g, tea.Batch(cmds...)
}

func (g GameModel) View() string {

	//-------------------------------------------------------------------
	// title style
	//-------------------------------------------------------------------
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Align(lipgloss.Center).
		Width(40).
		Padding(1, 0, 1, 0).
		BorderForeground(lipgloss.Color("63"))
	title := titleStyle.Render("🚀 STARSHIP SIMULATION 🚀")

	//-------------------------------------------------------------------
	// stats & progress Bar
	//-------------------------------------------------------------------
	stats := fmt.Sprintf("Ship Health: %d/%d", g.currentHealth, g.maxHealth)
	healthBar := g.ProgressBar.RenderProgressBar(g.currentHealth, g.maxHealth)

	//-------------------------------------------------------------------
	// menu items
	//-------------------------------------------------------------------
	menuItemStyle := lipgloss.NewStyle().
		Bold(true).
		PaddingLeft(1).
		Foreground(lipgloss.Color("217"))
	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		PaddingLeft(2).
		Bold(true)
	var menuView strings.Builder
	for i, item := range g.menuItems {
		cursor := "_"
		menuItemStyle = menuItemStyle.Foreground(lipgloss.Color("217"))
		if i == g.menuCursor {
			menuItemStyle = menuItemStyle.Foreground(lipgloss.Color("215"))
			cursor = ">"
		}
		styledItem := menuItemStyle.Render(strings.ToUpper(item))
		styledCursor := cursorStyle.Render(cursor)
		menuView.WriteString(fmt.Sprintf("%s %s\n", styledCursor, styledItem))
	}

	//-------------------------------------------------------------------
	// panels: left, center, right
	//-------------------------------------------------------------------
	leftPanelStyle := lipgloss.NewStyle().
		Width(40).
		Height(20).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Left, lipgloss.Top)
	leftPanel := leftPanelStyle.Render(fmt.Sprintf("%s\n\n%s", title, menuView.String()))

	centerContent := fmt.Sprintf("%s\n\n%s", stats, healthBar)
	centerPanel := lipgloss.NewStyle().
		Width(50).
		Height(20).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(centerContent)

	rightPanel := lipgloss.NewStyle().
		Width(40).
		Height(20).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("34")).
		Align(lipgloss.Center).
		Render(g.Yuta.View())

	//-------------------------------------------------------------------
	// bottom panel
	//-------------------------------------------------------------------
	var bottomPanelContent string
	switch g.selectedItem {
	case "Ship":
		bottomPanelContent = g.Ship.View()
	case "Crew":
		bottomPanelContent = g.Crew.View()
	case "Journal":
		bottomPanelContent = g.Journal.View()
	case "Map":
		bottomPanelContent = g.Map.View()
	default:
		if g.TrackedMission != nil {
			// Starting a mission
			if g.isTravelling {
				// Call StartMission (this func is not working rn)
				StartMission(*g.TrackedMission, g.Ship)

				// NOTE: Don't run timers from here. Do it from Update(). Else it will hold up the whole app.
				remainingTime := g.travelDuration - time.Since(g.travelStartTime)
				if remainingTime < 0 {
					remainingTime = 0
				}

				// Create travel progress display with animated bar
				progressBar := g.travelProgress.View()

				// Construct the content with spinner, progress bar and remaining time
				bottomPanelContent = fmt.Sprintf("%s Travelling to %s\n\n%s\n\nTime remaining: %v\n",
					g.spinner.View(),
					g.TrackedMission.Location,
					progressBar,
					remainingTime.Round(time.Millisecond))
			}
			// if g.TrackedMission.Status == "In Progress" {
			// 	bottomPanelContent = fmt.Sprintf(g.spinner.View())
			// }
			// Append current task after
			currentTask := components.NewCurrentTaskComponent()
			bottomPanelContent += currentTask.Render(g.TrackedMission)

		} else {
			bottomPanelContent = "This is the bottom panel."
		}
	}
	bottomPanel := lipgloss.NewStyle().
		Width(134).
		Height(21).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(bottomPanelContent)

	//-------------------------------------------------------------------
	// hints row: selected item (left) and hints (right)
	//-------------------------------------------------------------------
	selected := g.selectedItem
	if selected == "" {
		selected = "none"
	}
	selectedText := fmt.Sprintf("Selected [%s]", selected)
	leftSide := lipgloss.NewStyle().
		Width(65).
		PaddingLeft(2).
		Render(selectedText)
	hints := "[k ↑ j ↓ arrow keys] Navigate • [Enter] Select • [q] Quit"
	rightSide := lipgloss.NewStyle().
		Width(69).
		Align(lipgloss.Right).
		PaddingRight(2).
		Render(hints)
	hintsRowContent := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, rightSide)
	hintsRowStyle := lipgloss.NewStyle().
		Width(134).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("15"))
	hintsRow := hintsRowStyle.Render(hintsRowContent)

	//-------------------------------------------------------------------
	// combine the top row, bottom panel, and hints row
	//-------------------------------------------------------------------
	topRow := lipgloss.JoinHorizontal(lipgloss.Center, leftPanel, centerPanel, rightPanel)
	bottomRows := lipgloss.JoinVertical(lipgloss.Center, bottomPanel, hintsRow)
	mainView := lipgloss.JoinVertical(lipgloss.Center, topRow, bottomRows)

	return mainView
}

// NewGameModel creates and returns a new GameModel instance
func NewGameModel() tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot                                         // Spinner style CAN CHANGE
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("217")) // Spinner color

	// Initialize progress bar with a custom gradient
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	// Load the full game save from your save file
	fullSave, err := data.LoadFullGameSave()
	if err != nil || fullSave == nil {
		// handle error or absence of save file
		fmt.Println("Error loading save file or save file not found; using default values")
		// we can create a default FullGameSave here or call a helper like data.DefaultFullGameSave()
		//fullSave = data.DefaultFullGameSave() <-- define this function in data package
	}

	// map the saved ship data to the game model fields
	currentHealth := fullSave.Ship.HullIntegrity
	maxHealth := fullSave.Ship.MaxHullIntegrity

	// create the models from the loaded save data
	shipModel := model.NewShipModel(fullSave.Ship)
	crewModel := model.NewCrewModel(fullSave.Crew)
	journalModel := model.NewJournalModel()
	//mapModel := model.NewMapModel()

	// initialize and return the main GameModel with values from the save file
	return GameModel{
		ProgressBar:   components.NewProgressBar(),
		currentHealth: currentHealth,
		maxHealth:     maxHealth,
		Yuta:          components.NewYuta(),
		menuItems:     []string{"Ship", "Crew", "Journal", "Map", "Exit"},
		menuCursor:    0,
		Ship:          shipModel,
		Crew:          crewModel,
		Journal:       journalModel,
		//Map:            mapModel,
		activeView:     ViewNone,
		spinner:        s,
		isTravelling:   false,
		travelProgress: p,
	}
}

// Function to start a mission
// Param: mission - update mission status
// Param: ShipModel - modify fuel stat
// TODO: actually make it update our shipModel instance
func StartMission(mission model.Mission, ship model.ShipModel) model.ShipModel {
	// Check if ship has enough fuel
	if ship.EngineFuel < mission.FuelNeeded {
		// Print message that ship does not have enough fuel
		return ship
	}
	ship.EngineFuel -= mission.FuelNeeded // Deduct fuel from ship

	mission.Status = "In Progress" // Update mission status to "In Progress"

	// Animate progress bar to simulate travel time

	// Arrived

	// Do mission objective / do event
	// call event function

	// Return to base

	// mission.Status = "Completed" // Update mission status to "Completed"

	return ship
}
