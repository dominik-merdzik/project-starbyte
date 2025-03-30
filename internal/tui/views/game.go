package views

import (
	"fmt"
	"log"
	"math/rand"
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
	Yuta        components.YutaComponent
	Travel      components.TravelComponent
	Dialogue    *components.DialogueComponent
	GameOver    components.GameOverComponent

	// additional models
	Ship         model.ShipModel
	Crew         model.CrewModel
	Journal      model.JournalModel
	Map          model.MapModel
	Collection   model.CollectionModel   // NEW: Collection model
	SpaceStation model.SpaceStationModel // NEW: SpaceStation model

	menuItems  []MenuItem
	menuCursor int

	selectedItem MenuItem
	activeView   ActiveView

	TrackedMission *data.Mission

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

	MissionTemplates []data.MissionTemplate
	// Events
	Event *components.EventModel

	playerLostGame bool
}

type ActiveView int

const (
	ViewNone ActiveView = iota
	ViewJournal
	ViewCrew
	ViewMap
	ViewShip
	ViewCollection   // NEW: Added Collection view
	ViewSpaceStation // NEW: Added SpaceStation view
	ViewEvent        // Random events
)

type MenuItem int

const (
	MenuNone MenuItem = iota
	MenuJournal
	MenuShip
	MenuCrew
	MenuMap
	MenuCollection
	MenuSpaceStation
	MenuExit
)

type clearNotificationMsg struct{}

type autoSaveMsg time.Time

func (g GameModel) Init() tea.Cmd {
	return nil
}

func (g GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Firstly, always allow the player to quit the game no matter the state
	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "ctrl+c" || msg.String() == "q" {
		// Save the game before quitting
		g.syncSaveData()
		saveGameAsync(g.gameSave)

		return g, tea.Quit
	}

	// The player loses the game under these conditions
	if g.Ship.HullHealth <= 0 || g.Ship.EngineFuel <= 0 {
		g.playerLostGame = true
		g.activeView = ViewNone
		g.syncSaveData()

		saveGameAsync(g.gameSave)

		return g, nil // Skip everything else in Update()
	}

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
	case ViewSpaceStation: // NEW: Update SpaceStation view
		newSpaceStation, SpaceStationCmd := g.SpaceStation.Update(msg)
		if ss, ok := newSpaceStation.(model.SpaceStationModel); ok {
			g.SpaceStation = ss
		}
		cmds = append(cmds, SpaceStationCmd)
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

		saveGameAsync(g.gameSave)

		// schedule the next auto-save tick after 2 seconds *UPDATED TO 5 FOR TESTING*
		return g, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return autoSaveMsg(t) })
	case clearNotificationMsg:
		g.notification = ""
		return g, nil

		// Mission started. Trigger mission travel sequence
	case model.StartMissionMsg:
		g.TrackedMission = &msg.Mission // Set the mission to track
		destination := g.TrackedMission.Location
		currentLocation := g.Ship.Location

		// Check if travel is needed
		needsTravel := !currentLocation.IsEqual(destination)

		if needsTravel && !g.isTravelling {
			// --- Calculate Duration ---
			engineLevel := g.gameSave.Ship.Upgrades.Engine.CurrentLevel
			maxEngineLevel := g.gameSave.Ship.Upgrades.Engine.MaxLevel // Make sure this value is correct in your save/defaults
			travelDuration := g.locationService.CalculateTravelDuration(currentLocation, destination, engineLevel, maxEngineLevel)
			// --------------------------

			//log.Printf("Starting mission travel to %s (%s). Duration: %s", destination.PlanetName, destination.StarSystemName, travelDuration) // Debug log

			// Update model state for travel
			g.isTravelling = true
			g.Travel.Mission = g.TrackedMission
			g.Travel.DestLocation = destination

			// Start travel component, passing calculated duration
			cmds = append(cmds, g.Travel.StartTravel(destination, travelDuration))

		} else if !needsTravel {
			// If we're already there, set the mission as in progress and init dialogue
			g.TrackedMission.Status = data.MissionStatusInProgress
			if len(g.TrackedMission.Dialogue) > 0 {
				d := components.NewDialogueComponentFromMission(g.TrackedMission.Dialogue)
				g.Dialogue = &d
			} else {
				g.Dialogue = nil
			}
			// Optionally, clear isTravelling if it was somehow true but no travel needed
			g.isTravelling = false
		}

		// (1/3) Timer for travel component
	case components.TravelTickMsg:
		// Only update if actively travelling AND not yet complete
		if g.isTravelling && !g.Travel.TravelComplete {
			newTravel, cmd := g.Travel.Update(msg)
			g.Travel = newTravel
			cmds = append(cmds, cmd)

			// Check if travel is complete
			if g.Travel.TravelComplete {
				// g.isTravelling = false // Reset travel state
				// g.Ship.Location = g.Travel.DestLocation
				// g.syncSaveData()

				// Trigger a Random Event 30% chance
				if rand.Intn(100) < 30 {
					cmds = append(cmds, TriggerRandomEvent())
				}
			}
		}
	// (2/3) Animation for progress bar in travel component
	case progress.FrameMsg:
		// Update progress bar animation only while the travel view might be shown
		if g.isTravelling || (g.Travel.TravelComplete && time.Since(g.Travel.StartTime) <= g.Travel.Duration+2*time.Second) {
			newTravel, cmd := g.Travel.Update(msg)
			g.Travel = newTravel
			cmds = append(cmds, cmd)
		}
	// Updates ship values from random event
	case components.ApplyEffectsMsg:
		for key, value := range msg.Effects {
			switch key {
			case "fuel": // Fuel between 0-MaxFuel
				if g.Ship.EngineFuel+value <= g.Ship.MaxFuel && g.Ship.EngineFuel+value >= 0 {
					g.Ship.EngineFuel += value
				} else if g.Ship.EngineFuel+value > g.Ship.MaxFuel {
					g.Ship.EngineFuel = g.Ship.MaxFuel
				} else if g.Ship.EngineFuel+value < 0 {
					g.Ship.EngineFuel = 0
				}
				g.gameSave.Ship.Fuel = g.Ship.EngineFuel
			case "credits": // May be able to go into credit debt, so can be negative
				g.Credits += value
				g.gameSave.Player.Credits = g.Credits
			case "morale": // Morale between 0-100
				for i := range g.Crew.CrewMembers {
					if g.Crew.CrewMembers[i].Morale+value <= 100 && g.Crew.CrewMembers[i].Morale+value >= 0 {
						g.Crew.CrewMembers[i].Morale += value
					} else if g.Crew.CrewMembers[i].Morale+value > 100 {
						g.Crew.CrewMembers[i].Morale = 100
					} else if g.Crew.CrewMembers[i].Morale+value < 0 {
						g.Crew.CrewMembers[i].Morale = 0
					}
					g.gameSave.Crew[i].Morale = g.Crew.CrewMembers[i].Morale
				}
			case "food": // Food > 0
				if g.Ship.Food+value >= 0 {
					g.Ship.Food += value
				} else {
					g.Ship.Food = 0
				}
				g.gameSave.Ship.Food = g.Ship.Food
			case "hull": // Hull between 0-MaxHullHealth
				if g.Ship.HullHealth+value <= g.Ship.MaxHullHealth && g.Ship.HullHealth+value >= 0 {
					g.Ship.HullHealth += value
				} else if g.Ship.HullHealth+value > g.Ship.MaxHullHealth {
					g.Ship.HullHealth = g.Ship.MaxHullHealth
				} else if g.Ship.HullHealth+value < 0 {
					g.Ship.HullHealth = 0
				}
				g.gameSave.Ship.HullIntegrity = g.Ship.HullHealth
			}
		}
		g.syncSaveData() // Ensure updates are saved
		return g, nil

	// Random event dialogue started
	case StartEventMsg:
		g.activeView = ViewEvent
		g.Event = components.NewEventModel(msg.Event)
		return g, nil

	// Random event dialogue finished
	case components.EventFinishedMsg:
		g.activeView = ViewNone
		g.Event = nil

		// Ensure travel is completed after an event
		if g.isTravelling {
			//g.isTravelling = false
			//g.Ship.Location = g.Travel.DestLocation
			//g.syncSaveData()
		}

		return g, nil

	// ---------------------------
	// Handle key presses
	// ---------------------------

	case tea.KeyMsg:
		// Send key presses to event.go for random event dialogue
		if g.activeView == ViewEvent && g.Event != nil {
			var cmd tea.Cmd
			g.Event, cmd = g.Event.Update(msg)
			return g, cmd
		}

		// First, if an active view is set, process escape.
		if g.activeView != ViewNone && msg.String() == "esc" {
			g.activeView = ViewNone
			g.selectedItem = MenuNone
			return g, tea.Batch(cmds...)
		}
		if g.activeView != ViewNone {
			return g, tea.Batch(cmds...)
		}

		// DIALOGUE -- Advance through it with Enter
		if g.TrackedMission != nil && g.TrackedMission.Status == data.MissionStatusInProgress {
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
					g.TrackedMission.Status = data.MissionStatusCompleted
					g.Dialogue = nil // Clear dialogue when complete
				}
				return g, nil
			}
		}

		// normal key handling
		switch msg.String() {
		case "up", "k":
			// Skip over space station if not at one
			planet := g.gameSave.Ship.Location.GetFullPlanet(g.gameSave.GameMap)
			hasStation := planet.Type == "Space Station"

			for {

				if g.menuCursor > 0 {
					g.menuCursor--
				}
				if g.menuItems[g.menuCursor] == MenuSpaceStation && !hasStation {
					continue
				}
				break
			}
		case "down", "j":
			// Skip over space station if not at one
			planet := g.gameSave.Ship.Location.GetFullPlanet(g.gameSave.GameMap)
			hasStation := planet.Type == "Space Station"

			for {
				if g.menuCursor < len(g.menuItems)-1 {
					g.menuCursor++
				}
				if g.menuItems[g.menuCursor] == MenuSpaceStation && !hasStation {
					continue
				}
				break
			}
		case "enter":
			g.notification = "" // Clear notification on menu selection
			g.selectedItem = g.menuItems[g.menuCursor]
			switch g.selectedItem {
			case MenuJournal:
				g.activeView = ViewJournal
			case MenuCrew:
				g.activeView = ViewCrew
			case MenuMap:
				g.Map.Ship.Location = g.Ship.Location // Ensure Map has the latest ship location before activating the view
				g.activeView = ViewMap
			case MenuShip:
				g.activeView = ViewShip
			case MenuCollection: // NEW: Activate Collection view
				g.activeView = ViewCollection
			case MenuSpaceStation: // NEW: Activate SpaceStation view
				// Check if current planet is a space station
				planet := g.gameSave.Ship.Location.GetFullPlanet(g.gameSave.GameMap)
				if planet.Type == "Space Station" {
					g.activeView = ViewSpaceStation
				} else {
					//Idk what to put here
				}
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
			// // Press SPACE to dismiss mission complete screen
			// if g.TrackedMission != nil && g.TrackedMission.Status == model.MissionStatusCompleted {
			// 	g.TrackedMission = nil
			// 	g.Dialogue = nil
			// }
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
		// Start travel animation if requested AND not already travelling
		if msg.ShowTravel && !g.isTravelling {
			destination := msg.Location        // Destination from map selection
			currentLocation := g.Ship.Location // Ship's current location

			// --- Calculate Duration ---
			engineLevel := g.gameSave.Ship.Upgrades.Engine.CurrentLevel
			maxEngineLevel := g.gameSave.Ship.Upgrades.Engine.MaxLevel
			travelDuration := g.locationService.CalculateTravelDuration(currentLocation, destination, engineLevel, maxEngineLevel)
			// --------------------------

			//log.Printf("Starting map travel to %s (%s). Duration: %s", destination.PlanetName, destination.StarSystemName, travelDuration) // Debug log

			// Update model state for travel
			g.isTravelling = true
			g.Travel.Mission = nil // Clear mission tracking for map travel
			g.Travel.DestLocation = destination

			// Start travel component, passing calculated duration
			cmds = append(cmds, g.Travel.StartTravel(destination, travelDuration))
		}
	case model.CrewUpdateMsg:
		return g, utilities.PushSave(g.gameSave, func() {
			g.syncSaveData() // Sync save data after crew upgrade
		})
	case model.RefuelUpdateMsg:
		// Fuel
		amount := msg.Amount
		totalCost := msg.Credit
		current := g.Ship.EngineFuel

		// Calculate fuel here to avoid using old states
		newFuel := current + amount
		if newFuel > 200 {
			newFuel = 200
		}

		g.Ship.EngineFuel = newFuel
		g.gameSave.Ship.Fuel = newFuel

		g.Credits -= totalCost
		g.gameSave.Player.Credits = g.Credits

		g.syncSaveData() // Sync save data
		return g, utilities.PushSave(g.gameSave, func() {
			g.syncSaveData() // Sync the save data
		})
	case model.UpgradeUpdateMsg:
		switch msg.UpgradeCursor {
		case 0:
			g.Ship.Upgrades.Engine.CurrentLevel = msg.NewLevel          // Update current ship
			g.gameSave.Ship.Upgrades.Engine.CurrentLevel = msg.NewLevel // Then update game save
		case 1:
			g.Ship.Upgrades.WeaponSystems.CurrentLevel = msg.NewLevel
			g.gameSave.Ship.Upgrades.WeaponSystems.CurrentLevel = msg.NewLevel
		case 2:
			g.Ship.Upgrades.CargoExpansion.CurrentLevel = msg.NewLevel
			g.gameSave.Ship.Upgrades.CargoExpansion.CurrentLevel = msg.NewLevel
		}

		g.Credits = msg.Credits
		g.gameSave.Player.Credits = msg.Credits

		g.syncSaveData()
		return g, utilities.PushSave(g.gameSave, func() {
			g.syncSaveData()
		})
	case model.HireCrewMsg:
		g.gameSave.Crew = append(g.gameSave.Crew, msg.Crew)
		g.Credits = msg.Credits
		g.gameSave.Player.Credits = msg.Credits

		g.syncSaveData()
		return g, utilities.PushSave(g.gameSave, func() {
			g.syncSaveData()
		})
	case model.AcceptMissionMsg:
		g.Journal.Missions = append(g.Journal.Missions, msg.Mission)
		g.gameSave.Missions = append(g.gameSave.Missions, msg.Mission)

		g.syncSaveData()
		return g, utilities.PushSave(g.gameSave, func() {
			g.syncSaveData()
		})

	}

	// (3/3) More updates for the travel component
	if g.isTravelling && g.Travel.TravelComplete {

		arrivalLocation := g.Travel.DestLocation // Store where we arrived
		//log.Printf("Travel complete. Arrived at %s (%s).", arrivalLocation.PlanetName, arrivalLocation.StarSystemName) // Debug log
		startLocation := g.Ship.Location // Location before travel

		// Calculate fuel remaining AFTER the trip
		newFuel := g.locationService.GetFuelCost(
			startLocation.Coordinates, arrivalLocation.Coordinates,
			startLocation.StarSystemName, arrivalLocation.StarSystemName,
			g.Ship.EngineHealth, g.gameSave.Ship.Fuel,
		)

		// --- Update State Post-Arrival ---
		g.Ship.EngineFuel = newFuel
		g.Ship.Location = arrivalLocation
		g.isTravelling = false

		// --- Improvement: Reset TravelComplete flag ---
		g.Travel.TravelComplete = false // Reset the flag now that arrival is handled

		// Sync state to the save structure
		g.syncSaveData()
		// Trigger an asynchronous save for the updated state
		cmds = append(cmds, utilities.PushSave(g.gameSave, g.syncSaveData))

		// Handle Mission Arrival / Map Arrival Notification
		missionTravelCompleted := (g.TrackedMission != nil && g.Travel.Mission != nil && g.Travel.Mission.Title == g.TrackedMission.Title)

		if missionTravelCompleted && g.Ship.Location.IsEqual(g.TrackedMission.Location) {
			// Arrived at the specific tracked mission destination
			// ... (mission status and dialogue logic) ...
			g.TrackedMission.Status = data.MissionStatusInProgress
			if len(g.TrackedMission.Dialogue) > 0 {
				d := components.NewDialogueComponentFromMission(g.TrackedMission.Dialogue)
				g.Dialogue = &d
			} else {
				g.Dialogue = nil
			}
		} else {
			// Arrived via map travel (or mission travel to non-final location)
			// ... (notification logic) ...
			g.notification = fmt.Sprintf("Arrived at %s", arrivalLocation.PlanetName)
			cmds = append(cmds, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
				return clearNotificationMsg{}
			}))
		}

		// Clear the mission associated with the travel component
		g.Travel.Mission = nil
	}

	// When a mission is completed
	if g.TrackedMission != nil && g.TrackedMission.Status == data.MissionStatusCompleted {
		// Update mission status in the journal
		for i, mission := range g.Journal.Missions {
			if mission.Title == g.TrackedMission.Title {
				g.Journal.Missions[i].Status = data.MissionStatusCompleted
			}
		}
		// Reward player with credits
		g.Credits += g.TrackedMission.Income
		// Reward with research note (if any)
		g.addRandomResearchNote()

		// --- NEW: Assign next mission if available ---
		nextStep := g.TrackedMission.Step + 1
		var nextMission *data.Mission = nil
		// Optionally, you can also check for the same mission "line" by comparing categories or a custom ID.
		for _, tmpl := range g.MissionTemplates {
			if tmpl.Category == g.TrackedMission.Category && tmpl.Step == nextStep {
				nextMission = &data.Mission{
					Title:       tmpl.Title,
					Description: tmpl.Description,
					Category:    tmpl.Category,
					Step:        tmpl.Step,
					Location:    tmpl.Location,
					Dialogue:    tmpl.Dialogue,
					Income:      tmpl.Income,
					Status:      data.MissionStatusNotStarted,
				}
				break
			}
		}
		if nextMission != nil {
			// Create a copy of the mission and initialize its status.
			newMission := *nextMission
			newMission.Status = data.MissionStatusNotStarted
			// Append the new mission to the journal and game save
			g.Journal.Missions = append(g.Journal.Missions, newMission)
			g.gameSave.Missions = append(g.gameSave.Missions, newMission)
			// Optionally, notify the player
			g.notification = fmt.Sprintf("New mission available: %s", newMission.Title)
		}

		g.TrackedMission = nil // Clear the tracked mission
	}

	// Game over: ship destroyed
	if g.Ship.HullHealth <= 0 {

	}

	return g, tea.Batch(cmds...)
}

func (g GameModel) View() string {

	//---------------------------
	//Left Panel: Title & Menu, with Version at the bottom
	//---------------------------
	titleStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("63")).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("63"))

	title := titleStyle.Render("WHAT ARE YOUR ORDERS?")
	if g.playerLostGame {
		title = titleStyle.Render("#[*> FATAL ERROR <*]#")
	}

	menuItemStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("217"))
	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true)

	var menuView strings.Builder
	planet := g.gameSave.Ship.Location.GetFullPlanet(g.gameSave.GameMap)
	hasStation := planet.Type == "Space Station"

	for i, item := range g.menuItems {
		cursor := "-"
		style := menuItemStyle.Foreground(lipgloss.Color("217")) // Normal color
		if i == g.menuCursor {
			cursor = ">"
			style = style.Foreground(lipgloss.Color("215")) // Highlight color
		}
		if item == MenuSpaceStation && !hasStation {
			style = style.Foreground(lipgloss.Color("240")) // Gray it out
		}

		// Get display text - use corrupted text if game over
		itemText := item.String()
		if g.playerLostGame {
			// Corrupt the text directly here
			switch item {
			case MenuJournal:
				itemText = "J0*rn%l"
			case MenuShip:
				itemText = "S#!p"
			case MenuCrew:
				itemText = "Cr3>>"
			case MenuMap:
				itemText = "M<p"
			case MenuCollection:
				itemText = "C0ll*ct!0n"
			case MenuSpaceStation:
				itemText = "Sp@c3 St@t!*n"
			case MenuExit:
				itemText = "EX1T"
			}
		}

		styledItem := style.Render(strings.ToUpper(itemText))
		styledCursor := cursorStyle.Render(cursor)

		menuView.WriteString(fmt.Sprintf("%s %s\n", styledCursor, styledItem))
	}

	// left panel top (title and menu)
	leftPanelTop := lipgloss.NewStyle().
		Width(27).
		Height(16).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(title + "\n\n" + menuView.String())

	// ---------------------------
	// Center Panel: Stats & Progress Bars with Credits at the Bottom
	// ---------------------------

	statLabelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))

	centerWidth := 48

	// Create hull health header with centered text
	hullLabelText := "HULL INTEGRITY"
	hullPaddingLeft := (centerWidth - len(hullLabelText) - 4) / 2 // -4 for "┌── " and " ──┐"
	hullPaddingRight := centerWidth - len(hullLabelText) - 4 - hullPaddingLeft
	shipHealthText := statLabelStyle.Render("┌" + strings.Repeat("─", hullPaddingLeft) + " " + hullLabelText + " " + strings.Repeat("─", hullPaddingRight) + "┐")
	healthBar := g.ProgressBar.RenderProgressBar(g.Ship.HullHealth, g.Ship.MaxHullHealth)

	// Create fuel header with centered text
	fuelLabelText := "FUEL"
	fuelPaddingLeft := (centerWidth - len(fuelLabelText) - 4) / 2 // -4 for "┌── " and " ──┐"
	fuelPaddingRight := centerWidth - len(fuelLabelText) - 4 - fuelPaddingLeft
	fuelText := statLabelStyle.Render("┌" + strings.Repeat("─", fuelPaddingLeft) + " " + fuelLabelText + " " + strings.Repeat("─", fuelPaddingRight) + "┐")
	fuelBar := g.ProgressBar.RenderProgressBar(g.Ship.EngineFuel, g.Ship.MaxFuel)

	// Display location
	var locationText string
	// If at space station
	if planet.Type == "Space Station" {
		locationText = fmt.Sprintf("Docked at %s, %s System", g.Ship.Location.PlanetName, g.Ship.Location.StarSystemName)
	} else {
		locationText = fmt.Sprintf("Orbiting %s, %s System", g.Ship.Location.PlanetName, g.Ship.Location.StarSystemName)
	}

	moduleStatusText := "Modules: Engine (OK), Weapons (OK), Cargo (OK)"

	creditsText := fmt.Sprintf("Credits: %d", g.Credits)

	statsContent := fmt.Sprintf("%s\n%s\n\n%s\n%s\n\n\n%s\n\n%s\n\n%s",
		shipHealthText, healthBar, fuelText, fuelBar, locationText, moduleStatusText, creditsText)

	centerStatsPanel := lipgloss.NewStyle().
		Width(60).
		Height(16).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(statsContent)

	// ---------------------------
	// Right Panel: Yuta and game version
	// ---------------------------

	// Game version displayed by Yuta OS
	// Might need to adjust the fancy chars to fit within the width when the version number grows
	versionText := fmt.Sprintf("<== Y//UTA v%s ==>", g.Version)
	yutaContent := g.Yuta.View()

	if g.playerLostGame {
		yutaContent = "" // Turn it blank
	}

	rightYutaPanel := lipgloss.NewStyle().
		Width(27).
		Height(16).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		Foreground(lipgloss.Color("215")).
		Render(versionText + "\n\n" + yutaContent)

	// ---------------------------
	// Bottom Panel: (Mission Details, etc.)
	// ---------------------------

	var bottomPanelContent string
	switch g.selectedItem {
	case MenuShip:
		bottomPanelContent = g.Ship.View()
	case MenuCrew:
		bottomPanelContent = g.Crew.View()
	case MenuJournal:
		bottomPanelContent = g.Journal.View()
	case MenuMap:
		bottomPanelContent = g.Map.View()
	case MenuCollection: // NEW: Display Collection view.
		bottomPanelContent = g.Collection.View()
	case MenuSpaceStation: // NEW: Display SpaceStation view
		bottomPanelContent = g.SpaceStation.View()
	default:
		// Show travel view if travelling, regardless of mission
		if g.isTravelling {
			bottomPanelContent = g.Travel.View()
		} else if g.activeView == ViewEvent && g.Event != nil { // Event in bottom panel
			bottomPanelContent = g.Event.View()
			bottomPanelContent += "\n\nPress [Enter] to continue or [Q] to exit."
		} else if g.TrackedMission != nil {
			// Show current task
			currentTask := components.NewCurrentTaskComponent()
			bottomPanelContent = currentTask.Render(g.TrackedMission)

			// If mission in progress, show dialogue
			switch g.TrackedMission.Status {
			case data.MissionStatusInProgress:
				// Show dialogue ONLY if it's initialized
				if g.Dialogue != nil { // <--- ADD THIS CHECK
					bottomPanelContent = g.Dialogue.View()
					bottomPanelContent += "\n\nPress [Enter] to continue dialogue."
				} else {
					// Optional: Show a message indicating the mission is starting or dialogue is loading
					bottomPanelContent = fmt.Sprintf("Mission '%s' starting...\n(Press Enter to begin dialogue if available)", g.TrackedMission.Title)
					// Or just leave it blank if preferred:
					// bottomPanelContent = "" // Or keep existing content from currentTask rendering
				}
			case data.MissionStatusCompleted:
				// Show mission complete screen
				// Ensure TrackedMission is not nil here too for safety, although the outer check covers it.
				if g.TrackedMission != nil {
					bottomPanelContent = fmt.Sprintf("Mission Complete!\n\nYou were rewarded %d credits.\n\nPress [Space] to continue.", g.TrackedMission.Income)
				}
			}
		} else {
			bottomPanelContent = "" // Clear bottom panel if nothing else is active
		}

		if g.playerLostGame {
			bottomPanelContent = g.GameOver.View()
		}
	}
	bottomPanel := lipgloss.NewStyle().
		Width(120 - 2).
		Height(20).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(bottomPanelContent)

	// ---------------------------
	// Hints Row
	// ---------------------------

	// Contextual hint based on selected item or notification
	var contextualHint string
	if g.notification != "" {
		contextualHint = g.notification
	} else {
		contextualHint = fmt.Sprintf("Selected [%s]", g.selectedItem.String())
	}

	leftSide := lipgloss.NewStyle().
		Width(58).
		PaddingLeft(1).
		Render(contextualHint)
	// TODO: show vim keys if user set to vim mode in config
	//hints := "[k ↑ | j ↓ | h ← | l →] Navigate • [Enter] Select • [q] Quit"
	hints := "[↑ | ↓ | ← | →] Navigate • [Enter] Select • [q] Quit & Save"
	rightSide := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Width(60).
		PaddingRight(1).
		Render(hints)
	hintsRowContent := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, rightSide)
	hintsRowStyle := lipgloss.NewStyle().
		Width(118).
		Margin(0, 1).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("15"))
	hintsRow := hintsRowStyle.Render(hintsRowContent)

	// ---------------------------
	// Combine Top Row Panels, Bottom Panel, and Hints Row.
	// ---------------------------

	topRow := lipgloss.JoinHorizontal(lipgloss.Left, leftPanelTop, centerStatsPanel, rightYutaPanel)
	bottomRows := lipgloss.JoinVertical(lipgloss.Left, bottomPanel, hintsRow)
	mainView := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRows)

	return mainView
}

func NewGameModel() tea.Model {
	fullSave, err := data.LoadFullGameSave()
	if err != nil || fullSave == nil {
		fmt.Println("Error loading save file or save file not found; using default values")
	}

	// Load events from events.json
	err = data.LoadEvents()
	if err != nil {
		fmt.Println("Error failed to load events:", err)
	}

	// Load mission templates file
	missionTemplates, err := data.LoadMissionTemplates()
	if err != nil {
		log.Fatal("Failed to load mission templates:", err)
	}

	shipModel := model.NewShipModel(fullSave.Ship)
	crewModel := model.NewCrewModel(fullSave.Crew, fullSave)
	journalModel := model.NewJournalModel()
	mapModel := model.NewMapModel(fullSave.GameMap, fullSave.Ship, fullSave)
	collectionModel := model.NewCollectionModel(fullSave)
	spaceStationModel := model.NewSpaceStationModel(fullSave.Ship, fullSave.Player.Credits, missionTemplates, fullSave.GameMap.StarSystems)

	return GameModel{
		ProgressBar:      components.NewProgressBar(),
		menuItems:        []MenuItem{MenuJournal, MenuShip, MenuCrew, MenuMap, MenuCollection, MenuSpaceStation, MenuExit},
		menuCursor:       0,
		Ship:             shipModel,
		Crew:             crewModel,
		Journal:          journalModel,
		Collection:       collectionModel,
		SpaceStation:     spaceStationModel,
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
		MissionTemplates: missionTemplates,
		Yuta:             components.NewYutaComponent(fullSave.Ship, fullSave.Player.PlayerName, fullSave.Player.Credits, fullSave.GameMetadata.Version),
		GameOver:         components.NewGameOverComponent(fullSave.Ship),
	}
}

// syncSaveData updates the gameSave data with the latest state from the GameModel
// syncSaveData updates gameSave with the latest in-memory state
func (g *GameModel) syncSaveData() {

	// Sync ship values
	g.gameSave.Ship.HullIntegrity = g.Ship.HullHealth
	g.gameSave.Ship.Fuel = g.Ship.EngineFuel
	g.gameSave.Ship.EngineHealth = g.Ship.EngineHealth
	g.gameSave.Ship.FTLDriveHealth = g.Ship.FTLDriveHealth
	g.gameSave.Ship.FTLDriveCharge = g.Ship.FTLDriveCharge
	g.gameSave.Ship.ShieldStrength = g.Ship.ShieldStrength
	g.gameSave.Ship.MaxHullIntegrity = g.Ship.MaxHullHealth
	g.gameSave.Ship.MaxFuel = g.Ship.MaxFuel
	g.gameSave.Ship.Food = g.Ship.Food
	g.gameSave.Ship.Location = g.Ship.Location

	g.gameSave.Player.Credits = g.Credits
	g.gameSave.GameMetadata.LastSaveTime = time.Now().Format(time.RFC3339)

	g.gameSave.Missions = g.Journal.Missions // Sync missions

	g.gameSave.GameMetadata.GameOver = g.playerLostGame // Sync game over state
}

// saveGameAsync saves the game in a goroutine to avoid blocking the UI
func saveGameAsync(save *data.FullGameSave) {
	go func(save *data.FullGameSave) {
		if err := data.SaveGame(save); err != nil {
			fmt.Println("Error saving game:", err)
		}
	}(save)
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

// Gives a random tier research note to the player
func (g *GameModel) addRandomResearchNote() {
	// Generate a random number between 0-100 to determine tier
	roll := rand.Intn(101)

	// GACHA!
	var tier int
	switch {
	case roll < 5: // 5% chance for highest tier
		tier = 5
	case roll < 15: // 10% chance for tier 4
		tier = 4
	case roll < 35: // 20% chance for tier 3
		tier = 3
	case roll < 65: // 30% chance for tier 2
		tier = 2
	default: // 35% chance for tier 1
		tier = 1
	}

	for i := range g.gameSave.Collection.ResearchNotes { // Iterate using index on GameSave collection
		// Check if this is the correct tier
		if g.gameSave.Collection.ResearchNotes[i].Tier == tier {
			// Increment quantity directly in GameSave
			g.gameSave.Collection.ResearchNotes[i].Quantity++

			g.gameSave.Collection.UsedCapacity++
			if g.gameSave.Collection.UsedCapacity > g.gameSave.Collection.MaxCapacity {
				log.Printf("Warning: UsedCapacity (%d) exceeds MaxCapacity (%d)",
					g.gameSave.Collection.UsedCapacity, g.gameSave.Collection.MaxCapacity)
			}

			// Create notification
			g.notification = fmt.Sprintf("Received %s research note!", g.gameSave.Collection.ResearchNotes[i].Name)

			break // Note found and updated
		}
	}

}

// Triggers random event from events.json
func TriggerRandomEvent() tea.Cmd {
	event := data.GetRandomEvent()
	if event != nil {
		return func() tea.Msg {
			return StartEventMsg{Event: event}
		}
	} else {
		log.Println("No events found in GetRandomEvent()")
	}
	return nil
}

type StartEventMsg struct {
	Event *data.Event
}

// ToString method for MenuItem enum
func (m MenuItem) String() string {
	switch m {
	case MenuJournal:
		return "Journal"
	case MenuShip:
		return "Ship"
	case MenuCrew:
		return "Crew"
	case MenuMap:
		return "Map"
	case MenuCollection:
		return "Collection"
	case MenuSpaceStation:
		return "Space Station"
	case MenuExit:
		return "Exit"
	default:
		return ""
	}
}
