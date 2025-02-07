package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	//ViewMap
	//ViewShip
)

// -----------------------------------------------------------------------------

type GameModel struct {
	// components
	ProgressBar components.ProgressBar
	Yuta        components.YutaModel

	// additional models
	Ship    model.ShipModel
	Crew    model.CrewModel
	Journal model.JournalModel

	currentHealth int
	maxHealth     int

	menuItems  []string // list of menu options
	menuCursor int      // current position of the cursor

	// selectedItem tracks which menu item is selected
	selectedItem string

	activeView ActiveView // The currently active view
}

func (g GameModel) Init() tea.Cmd {
	// initialize Yuta's animation (seem to be broken ATM)
	return tea.Batch(
		g.Yuta.Init(),
	)
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
	}

	// process key messages.
	switch msg := msg.(type) {
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
			}
		}
	}

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
	default:
		bottomPanelContent = "This is the bottom panel."
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

// NewGameModel creates and returns a new GameModel instance.
func NewGameModel() tea.Model {
	return GameModel{
		ProgressBar:   components.NewProgressBar(),
		currentHealth: 62,                   // example initial health
		maxHealth:     100,                  // example max health
		Yuta:          components.NewYuta(), // initialize Yuta
		menuItems:     []string{"Ship", "Crew", "Journal", "Map", "Exit"},
		menuCursor:    0,
		Ship:          model.NewShipModel(),
		Crew:          model.NewCrewModel(),
		Journal:       model.NewJournalModel(),
		activeView:    ViewNone, // No active view initially
	}
}
