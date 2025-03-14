package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
	"github.com/dominik-merdzik/project-starbyte/internal/tui/components"
	model "github.com/dominik-merdzik/project-starbyte/internal/tui/models"
	"github.com/dominik-merdzik/project-starbyte/internal/utilities"
)

type GameModel struct {
	// components
	ProgressBar components.ProgressBar
	Yuta        components.YutaModel
	Travel      components.TravelComponent
	Dialogue    *components.DialogueComponent

	// additional models
	Ship       model.ShipModel
	Crew       model.CrewModel
	Journal    model.JournalModel
	Map        model.MapModel
	Collection model.CollectionModel // NEW: Collection model

	currentHealth int
	maxHealth     int

	menuItems  []string
	menuCursor int

	selectedItem string
	activeView   ActiveView

	TrackedMission *model.Mission

	isTravelling bool

	locationService *data.LocationService

	Credits int
	Version string

	dirty              bool
	gameSave           *data.FullGameSave
	lastAutoSaveTime   time.Time
	lastManualSaveTime time.Time
	notification       string
	autoSaveInitiated  bool
}

type ActiveView int

const (
	ViewNone ActiveView = iota
	ViewJournal
	ViewCrew
	ViewMap
	ViewShip
	ViewCollection // NEW: Added Collection view
)

type clearNotificationMsg struct{}

type autoSaveMsg time.Time

func (g GameModel) Init() tea.Cmd {
	return nil
}

func (g GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if !g.autoSaveInitiated {
		g.autoSaveInitiated = true
		cmds = append(cmds, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return autoSaveMsg(t)
		}))
	}

	// update active view
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
		newShip, shipCmd := g.Ship.Update(msg)
		if s, ok := newShip.(model.ShipModel); ok {
			g.Ship = s
		}
		cmds = append(cmds, shipCmd)
	case ViewCollection: // NEW: Update Collection view
		newCollection, CollectionCmd := g.Collection.Update(msg)
		if col, ok := newCollection.(model.CollectionModel); ok {
			g.Collection = col
		}
		cmds = append(cmds, CollectionCmd)
	}

	// ---------------------------
	// Listen for messages
	// ---------------------------

	switch msg := msg.(type) {
	case autoSaveMsg:
		// calculate elapsed time since last auto-save
		elapsed := time.Since(g.lastAutoSaveTime)
		addDurationToPlayTime(&g.gameSave.GameMetadata.TotalPlayTime, elapsed)
		g.lastAutoSaveTime = time.Now()

		// sync current game state
		g.syncSaveData()

		// save the game asynchronously with a goroutine
		go func(save *data.FullGameSave) {
			if err := data.SaveGame(save); err != nil {
				fmt.Println("Error auto-saving game:", err)
			}
		}(g.gameSave)

		// schedule the next auto-save tick after 2 seconds *UPDATED TO 5 FOR TESTING*
		return g, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return autoSaveMsg(t) })
	case clearNotificationMsg:
		g.notification = ""
		return g, nil

	// Mission started. Trigger mission travel sequence
	case model.StartMissionMsg:
		g.TrackedMission = &msg.Mission // Set the mission to track

		// If we're not already at the mission location, start travelling there
		if !g.isTravelling {
			g.isTravelling = true

			// // Convert mission location string to Location struct
			// targetPlanet := g.locationService.FindByPlanetName(g.TrackedMission.Location.PlanetName)

			// if targetPlanet == nil {
			// 	fmt.Println("Error: Could not find location:", g.TrackedMission.Location)
			// 	return g, nil
			// }

			// Set the mission on Travel component
			g.Travel.Mission = g.TrackedMission

			return g, g.Travel.StartTravel()
		} else {
			// If we're already there, just set the mission as in progress
			g.TrackedMission.Status = model.MissionStatusInProgress
		}

	// (1/3) Timer for travel component
	case components.TravelTickMsg:
		if g.isTravelling {
			newTravel, cmd := g.Travel.Update(msg)
			g.Travel = newTravel
			cmds = append(cmds, cmd)
		}
	// (2/3) Animation for progress bar in travel component
	case progress.FrameMsg:
		if g.isTravelling {
			newTravel, cmd := g.Travel.Update(msg)
			g.Travel = newTravel
			cmds = append(cmds, cmd)
		}

	// ---------------------------
	// Handle key presses
	// ---------------------------

	case tea.KeyMsg:
		// First, if an active view is set, process escape.
		if g.activeView != ViewNone && msg.String() == "esc" {
			g.activeView = ViewNone
			g.selectedItem = ""
			return g, tea.Batch(cmds...)
		}
		if g.activeView != ViewNone {
			return g, tea.Batch(cmds...)
		}

		// DIALOGUE -- Advance through it with Enter
		if g.TrackedMission != nil && g.TrackedMission.Status == model.MissionStatusInProgress {
			if msg.String() == "enter" {
				if g.Dialogue == nil {
					// initialize dialogue with the first line already shown
					d := components.NewDialogueComponentFromMission(g.TrackedMission.Dialogue)
					g.Dialogue = &d
				} else {
					g.Dialogue.Next()
				}
				// if we've advanced past all dialogue lines, mark mission as completed
				if g.Dialogue.CurrentLine >= len(g.TrackedMission.Dialogue) {
					g.TrackedMission.Status = model.MissionStatusCompleted
					g.Dialogue = nil // Clear dialogue when complete
				}
				return g, nil
			}
		}

		// normal key handling
		switch msg.String() {
		case "q":
			return g, tea.Quit
		case "a":
			g.currentHealth -= 10
			if g.currentHealth < 0 {
				g.currentHealth = 0
			}
			g.dirty = true
		case "h":
			g.currentHealth += 10
			if g.currentHealth > g.maxHealth {
				g.currentHealth = g.maxHealth
			}
			g.dirty = true
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
			switch g.selectedItem {
			case "Journal":
				g.activeView = ViewJournal
			case "Crew":
				g.activeView = ViewCrew
			case "Map":
				g.activeView = ViewMap
			case "Ship":
				g.activeView = ViewShip
			case "Collection": // NEW: Activate Collection view
				g.activeView = ViewCollection
			}
		case "s":

			// Check if 2 seconds have passed since the last manual save
			if time.Since(g.lastManualSaveTime) < 2*time.Second {
				g.notification = "Please wait before saving again."
				return g, nil
			}
			g.lastManualSaveTime = time.Now()

			// update total play time before saving
			elapsed := time.Since(g.lastAutoSaveTime)
			addDurationToPlayTime(&g.gameSave.GameMetadata.TotalPlayTime, elapsed)

			// reset the last auto-save time
			g.lastAutoSaveTime = time.Now()

			// Instead of saving synchronously here, call our helper
			// The syncFunc here updates our in-memory game state (e.g. calling g.syncSaveData())
			return g, utilities.ManualSave(g.gameSave, func() {
				g.syncSaveData()
			})
		case " ":
			// Press SPACE to dismiss mission complete screen
			if g.TrackedMission != nil && g.TrackedMission.Status == model.MissionStatusCompleted {
				g.TrackedMission = nil
				g.Dialogue = nil
			}
		}
	case utilities.SaveRetryMsg:
		// if saving failed, schedule a retry after 2 seconds
		return g, utilities.RetryManualSave(g.gameSave, func() {
			g.syncSaveData()
		})
	case utilities.SaveSuccessMsg:
		// update the notification on a successful save
		g.notification = "Game saved successfully!"
		return g, tea.Batch(tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearNotificationMsg{} }))

	// This message is received when tracking a mission (called from journal.go)
	case model.TrackMissionMsg:
		g.TrackedMission = &msg.Mission
		return g, nil

	// This message is received when travelling (map.go)
	// It will update the ship's location and fuel and trigger a save
	case model.TravelUpdateMsg:
		// Update parent Ship var
		g.Ship.Location = msg.Location
		g.Ship.EngineFuel = msg.Fuel

		// Then save
		return g, utilities.PushSave(g.gameSave, func() {
			g.syncSaveData()       // Sync the save data
			g.Travel.StartTravel() // Show travel screen
		})
	}

	// (3/3) More updates for the travel component
	if g.isTravelling {
		// !!! This case stops the animated progress bar from halting partially
		switch msg.(type) {
		case components.TravelTickMsg, progress.FrameMsg:
		// Already handled these specific message types above
		default:
			newTravel, travelCmd := g.Travel.Update(msg)
			g.Travel = newTravel
			cmds = append(cmds, travelCmd)
		}

		// If travel is complete...
		if g.Travel.TravelComplete {
			g.isTravelling = false
			g.TrackedMission.Status = model.MissionStatusInProgress
			// Then show dialogue
			d := components.NewDialogueComponentFromMission(g.TrackedMission.Dialogue)
			g.Dialogue = &d
		}
	}

	return g, tea.Batch(cmds...)
}

func (g GameModel) View() string {

	// ---------------------------
	// Left Panel: Title & Menu, with Version at the bottom
	// ---------------------------
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Align(lipgloss.Center).
		Width(40).
		Padding(1, 0, 1, 0).
		BorderForeground(lipgloss.Color("63"))
	title := titleStyle.Render("🚀 STARSHIP SIMULATION 🚀")

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
	// left panel top (title and menu)
	leftPanelTop := lipgloss.NewStyle().
		Width(40).
		Height(18).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Left, lipgloss.Top).
		Render(fmt.Sprintf("%s\n\n%s", title, menuView.String()))
	// version text at the bottom of the left panel
	versionText := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Foreground(lipgloss.Color("246")).
		Height(1).
		Width(40).
		PaddingLeft(1).
		Render("Version: " + g.Version)
	leftPanel := lipgloss.JoinVertical(lipgloss.Left, leftPanelTop, versionText)

	// ---------------------------
	// Center Panel: Stats & Progress Bars with Credits at the Bottom
	// ---------------------------

	statLabelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))

	shipHealthText := fmt.Sprintf("%s: %d/%d", statLabelStyle.Render("Ship Health"), g.currentHealth, g.maxHealth)
	healthBar := g.ProgressBar.RenderProgressBar(g.currentHealth, g.maxHealth)

	fuelText := fmt.Sprintf("%s: ", statLabelStyle.Render("Fuel"))
	fuelBar := g.ProgressBar.RenderProgressBar(g.Ship.EngineFuel, g.Ship.MaxFuel)

	foodText := fmt.Sprintf("%s: ", statLabelStyle.Render("Food"))
	foodBar := g.ProgressBar.RenderProgressBar(g.Ship.Food, 100)

	statsContent := fmt.Sprintf("\n%s\n%s\n\n%s\n%s\n\n%s\n%s",
		shipHealthText, healthBar, fuelText, fuelBar, foodText, foodBar)

	creditsContent := fmt.Sprintf("¢redits %d", g.Credits)
	creditsStyled := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("215")).
		Align(lipgloss.Center).
		Render(creditsContent)

	centerStatsPanel := lipgloss.NewStyle().
		Width(50).
		Height(18).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(statsContent)

	centerCreditsPanel := lipgloss.NewStyle().
		Width(50).
		Height(1).
		Align(lipgloss.Center).
		Render(creditsStyled)

	centerPanel := lipgloss.JoinVertical(lipgloss.Center, centerStatsPanel, centerCreditsPanel)

	// ---------------------------
	// Right Panel: Yuta Animation
	// ---------------------------

	rightPanel := lipgloss.NewStyle().
		Width(40).
		Height(19).
		Border(lipgloss.RoundedBorder()).
		Foreground(lipgloss.Color("215")).
		Align(lipgloss.Center).
		Render(g.Yuta.View())

	// ---------------------------
	// Bottom Panel: (Mission Details, etc.)
	// ---------------------------

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
	case "Collection": // NEW: Display Collection view.
		bottomPanelContent = g.Collection.View()
	default:
		if g.TrackedMission != nil {
			// render current task (this might include mission title, objectives, etc.)
			currentTask := components.NewCurrentTaskComponent()
			bottomPanelContent += currentTask.Render(g.TrackedMission)

			// Show travel view if travelling
			if g.isTravelling {
				bottomPanelContent = g.Travel.View()
			}

			// Show dialogue when mission is in progress
			if g.TrackedMission.Status == model.MissionStatusInProgress && g.Dialogue != nil {
				bottomPanelContent = g.Dialogue.View()
				bottomPanelContent += "\n\nPress [Enter] to continue dialogue."
			}

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

	// ---------------------------
	// Hints Row
	// ---------------------------

	selected := g.selectedItem
	if selected == "" {
		selected = "none"
	}

	notificationText := ""
	if g.notification != "" {
		notificationText = " ~ " + g.notification
	}

	selectedText := fmt.Sprintf("Selected [%s] %s", selected, notificationText)
	leftSide := lipgloss.NewStyle().
		Width(58).
		PaddingLeft(2).
		Render(selectedText)
	hints := "[k ↑ | j ↓ | h ← | l → ~ arrow keys] Navigate • [Enter] Select • [q] Quit"
	rightSide := lipgloss.NewStyle().
		Width(76).
		Align(lipgloss.Right).
		PaddingRight(2).
		Render(hints)
	hintsRowContent := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, rightSide)
	hintsRowStyle := lipgloss.NewStyle().
		Width(134).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("15"))
	hintsRow := hintsRowStyle.Render(hintsRowContent)

	// ---------------------------
	// Combine Top Row Panels, Bottom Panel, and Hints Row.
	// ---------------------------

	topRow := lipgloss.JoinHorizontal(lipgloss.Center, leftPanel, centerPanel, rightPanel)
	bottomRows := lipgloss.JoinVertical(lipgloss.Center, bottomPanel, hintsRow)
	mainView := lipgloss.JoinVertical(lipgloss.Center, topRow, bottomRows)

	return mainView
}

func NewGameModel() tea.Model {
	fullSave, err := data.LoadFullGameSave()
	if err != nil || fullSave == nil {
		fmt.Println("Error loading save file or save file not found; using default values")
		// Optionally, set fullSave = data.DefaultFullGameSave() here
	}
	currentHealth := fullSave.Ship.HullIntegrity
	maxHealth := fullSave.Ship.MaxHullIntegrity

	shipModel := model.NewShipModel(fullSave.Ship)
	shipModel.GameSave = fullSave
	crewModel := model.NewCrewModel(fullSave.Crew)
	journalModel := model.NewJournalModel()
	mapModel := model.NewMapModel(fullSave.GameMap, fullSave.Ship)
	mapModel.GameSave = fullSave                                     // Need this to avoid null pointer
	collectionModel := model.NewCollectionModel(fullSave.Collection) // NEW: Initialize Collection model

	return GameModel{
		ProgressBar:      components.NewProgressBar(),
		currentHealth:    currentHealth,
		maxHealth:        maxHealth,
		Yuta:             components.NewYuta(),
		menuItems:        []string{"Ship", "Crew", "Journal", "Map", "Collection", "Exit"},
		menuCursor:       0,
		Ship:             shipModel,
		Crew:             crewModel,
		Journal:          journalModel,
		Collection:       collectionModel, // NEW: Set Collection model
		Map:              mapModel,
		Travel:           components.NewTravelComponent(),
		activeView:       ViewNone,
		isTravelling:     false,
		Credits:          fullSave.Player.Credits,
		Version:          fullSave.GameMetadata.Version,
		dirty:            false,
		gameSave:         fullSave,
		lastAutoSaveTime: time.Now(),
		locationService:  data.NewLocationService(fullSave.GameMap),
	}
}

// startMission updates the ship model based on mission fuel requirements
func StartMission(mission model.Mission, ship model.ShipModel) model.ShipModel {

	// if ship.EngineFuel < mission.FuelNeeded {
	// 	return ship
	// }
	// ship.EngineFuel -= mission.FuelNeeded
	mission.Status = model.MissionStatusInProgress
	return ship
}

// syncSaveData updates the gameSave data with the latest state from the GameModel
// syncSaveData updates gameSave with the latest in-memory state
func (g *GameModel) syncSaveData() {

	// Sync ship values
	g.gameSave.Ship.HullIntegrity = g.currentHealth
	g.gameSave.Ship.Fuel = g.Ship.EngineFuel
	g.gameSave.Ship.EngineHealth = g.Ship.EngineHealth
	g.gameSave.Ship.FTLDriveHealth = g.Ship.FTLDriveHealth
	g.gameSave.Ship.FTLDriveCharge = g.Ship.FTLDriveCharge
	g.gameSave.Ship.ShieldStrength = g.Ship.ShieldStrength
	g.gameSave.Ship.MaxHullIntegrity = g.Ship.MaxHullHealth
	g.gameSave.Ship.MaxFuel = g.Ship.MaxFuel
	g.gameSave.Ship.Food = g.Ship.Food

	g.gameSave.Player.Credits = g.Credits
	g.gameSave.GameMetadata.LastSaveTime = time.Now().Format(time.RFC3339)

	g.gameSave.Ship.Location = g.Ship.Location // Sync location
}

// helper: add elapsed duration to TotalPlayTime, normalizing seconds/minutes/hours
func addDurationToPlayTime(tt *data.TotalPlayTime, d time.Duration) {
	secondsToAdd := int(d.Seconds())
	tt.Seconds += secondsToAdd
	tt.Minutes += tt.Seconds / 60
	tt.Seconds = tt.Seconds % 60
	tt.Hours += tt.Minutes / 60
	tt.Minutes = tt.Minutes % 60
}
