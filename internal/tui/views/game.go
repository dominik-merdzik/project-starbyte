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

// activeView indicates which view is currently active
type ActiveView int

const (
	ViewNone ActiveView = iota
	ViewJournal
	ViewCrew
	ViewMap
	ViewShip
)

// gameModel is the top-level model for the game
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

	menuItems  []string
	menuCursor int

	selectedItem string
	activeView   ActiveView

	TrackedMission *model.Mission

	isTravelling    bool
	travelStartTime time.Time
	travelDuration  time.Duration
	travelProgress  progress.Model

	Credits int
	Version string
}

type travelTickMsg struct{}

func (g GameModel) Init() tea.Cmd {
	return g.spinner.Tick
}

func (g GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// always update Yuta and Ship
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
		newShip, shipCmd := g.Map.Update(msg)
		if m, ok := newShip.(model.MapModel); ok {
			g.Map = m
		}
		cmds = append(cmds, shipCmd)
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		g.spinner, cmd = g.spinner.Update(msg)
		return g, cmd
	case model.TrackMissionMsg:
		g.TrackedMission = &msg.Mission
	case tea.KeyMsg:
		if g.activeView != ViewNone && msg.String() == "esc" {
			g.activeView = ViewNone
			g.selectedItem = ""
			return g, tea.Batch(cmds...)
		}
		if g.activeView != ViewNone {
			return g, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q":
			return g, tea.Quit
		case "a":
			g.currentHealth -= 10
			if g.currentHealth < 0 {
				g.currentHealth = 0
			}
		case "h":
			g.currentHealth += 10
			if g.currentHealth > g.maxHealth {
				g.currentHealth = g.maxHealth
			}
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
			}
		case " ":
			if g.TrackedMission != nil && !g.isTravelling {
				if g.TrackedMission.Status == "Not Started" {
					g.isTravelling = true
					g.travelStartTime = time.Now()
					g.travelDuration = time.Duration(g.TrackedMission.TravelTime) * time.Second
					g.travelProgress.SetPercent(0)
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
			percentComplete := float64(elapsed) / float64(g.travelDuration)
			if percentComplete > 1.0 {
				percentComplete = 1.0
			}
			cmd := g.travelProgress.SetPercent(percentComplete)
			if elapsed >= g.travelDuration {
				g.isTravelling = false
				g.TrackedMission.Status = "In Progress"
				return g, nil
			}
			return g, tea.Batch(
				tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return travelTickMsg{} }),
				cmd,
			)
		}
	}
	cmds = append(cmds, g.spinner.Tick)
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
	default:
		if g.TrackedMission != nil {
			if g.isTravelling {
				StartMission(*g.TrackedMission, g.Ship)
				remainingTime := g.travelDuration - time.Since(g.travelStartTime)
				if remainingTime < 0 {
					remainingTime = 0
				}
				progressBar := g.travelProgress.View()
				bottomPanelContent = fmt.Sprintf("%s Travelling to %s\n\n%s\n\nTime remaining: %v\n",
					g.spinner.View(), g.TrackedMission.Location, progressBar, remainingTime.Round(time.Millisecond))
			}
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

	// ---------------------------
	// Hints Row
	// ---------------------------

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

	// ---------------------------
	// Combine Top Row Panels, Bottom Panel, and Hints Row.
	// ---------------------------

	topRow := lipgloss.JoinHorizontal(lipgloss.Center, leftPanel, centerPanel, rightPanel)
	bottomRows := lipgloss.JoinVertical(lipgloss.Center, bottomPanel, hintsRow)
	mainView := lipgloss.JoinVertical(lipgloss.Center, topRow, bottomRows)

	return mainView
}

// NewGameModel creates and returns a new GameModel instance
func NewGameModel() tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("217"))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	fullSave, err := data.LoadFullGameSave()
	if err != nil || fullSave == nil {
		fmt.Println("Error loading save file or save file not found; using default values")
		// Optionally, set fullSave = data.DefaultFullGameSave() here
	}
	currentHealth := fullSave.Ship.HullIntegrity
	maxHealth := fullSave.Ship.MaxHullIntegrity

	shipModel := model.NewShipModel(fullSave.Ship)
	crewModel := model.NewCrewModel(fullSave.Crew)
	journalModel := model.NewJournalModel()
	// mapModel := model.NewMapModel()

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
		// Map:         mapModel,
		activeView:     ViewNone,
		spinner:        s,
		isTravelling:   false,
		travelProgress: p,
		Credits:        fullSave.Player.Credits,
		Version:        fullSave.GameMetadata.Version,
	}
}

// StartMission updates the ship model based on mission fuel requirements.
func StartMission(mission model.Mission, ship model.ShipModel) model.ShipModel {
	if ship.EngineFuel < mission.FuelNeeded {
		return ship
	}
	ship.EngineFuel -= mission.FuelNeeded
	mission.Status = "In Progress"
	return ship
}
