package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/tui/components"
	"github.com/dominik-merdzik/project-starbyte/internal/tui/models"
)

type GameModel struct {

	// components
	ProgressBar components.ProgressBar
	Yuta        components.YutaModel
	// how-to: 1) add Ship field to GameModel struct
	Ship model.ShipModel

	currentHealth int
	maxHealth     int

	menuItems    []string // list of menu options
	menuCursor   int      // current position of the cursor
	selectedItem string
}

func (g GameModel) Init() tea.Cmd {
	// initialize Yuta's animation (seem to be broken ATM)
	return tea.Batch(
		g.Yuta.Init(),
	)
}

func (g GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// update Yuta (collect its commands for future use)
	newYuta, yutaCmd := g.Yuta.Update(msg)
	if yutaModel, ok := newYuta.(components.YutaModel); ok {
		g.Yuta = yutaModel
	}
	cmds = append(cmds, yutaCmd)

	// how-to: 3) update Ship field
	newShip, shipCmd := g.Ship.Update(msg)
	if s, ok := newShip.(model.ShipModel); ok {
		g.Ship = s
	}
	cmds = append(cmds, shipCmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// quit the entire application
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

		// menu navigation inside the game
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
	// define style for the menu items
	menuItemStyle := lipgloss.NewStyle().
		Bold(true).
		PaddingLeft(1).
		Foreground(lipgloss.Color("217")) // chanage colour to theme later !!

	// define style for cursor
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
		Height(25).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Left, lipgloss.Top) // <-- Horizontal=Left, Vertical=Top

	leftPanel := leftPanelStyle.Render(
		fmt.Sprintf("%s\n\n%s", title, menuView.String()),
	)

	centerContent := fmt.Sprintf("%s\n\n%s", stats, healthBar)
	centerPanel := lipgloss.NewStyle().
		Width(50).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(centerContent)

	rightPanel := lipgloss.NewStyle().
		Width(40).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("34")).
		Align(lipgloss.Center).
		Render(g.Yuta.View())

	//-------------------------------------------------------------------
	// bottom panel (just a placeholder for now)
	//-------------------------------------------------------------------
	// how-to: 4) update bottom panel (where we want different components)
	var bottomPanelContent string
	switch g.selectedItem {
	case "Ship":
		// show the ShipModel view
		bottomPanelContent = g.Ship.View()
	default:
		// default fallback
		bottomPanelContent = "This is the bottom panel."
	}

	bottomPanel := lipgloss.NewStyle().
		Width(134).
		Height(12).
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

	// left side (selected item)
	leftSide := lipgloss.NewStyle().
		Width(65). // half-ish of 134
		PaddingLeft(2).
		Render(selectedText)

	// right side (hints)
	hints := "[k ↑ j ↓ arrow keys] Navigate • [Enter] Select • [q] Quit"
	rightSide := lipgloss.NewStyle().
		Width(69). // the other half
		Align(lipgloss.Right).
		PaddingRight(2).
		Render(hints)

	// join both sides horizontally
	hintsRowContent := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, rightSide)

	// style for the entire hints row
	hintsRowStyle := lipgloss.NewStyle().
		Width(134).
		Background(lipgloss.Color("236")). // dark gray
		Foreground(lipgloss.Color("15"))   // white

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
	return GameModel{
		ProgressBar:   components.NewProgressBar(),
		currentHealth: 62,                   // example initial health
		maxHealth:     100,                  // example max health
		Yuta:          components.NewYuta(), // initialize Yuta
		menuItems:     []string{"Ship", "Crew", "Journal", "Map", "Exit"},
		menuCursor:    0, // start cursor at the first menu item

		// how-to: 2) initialize Ship field
		Ship: model.NewShipModel(),
	}
}
