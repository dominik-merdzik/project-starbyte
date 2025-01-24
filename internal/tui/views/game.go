package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/tui/components"
)

type GameModel struct {
	ProgressBar components.ProgressBar
	Yuta        components.YutaModel

	currentHealth int
	maxHealth     int

	menuItems    []string // list of menu options
	menuCursor   int      // current position of the cursor
	selectedItem string
}

func (g GameModel) Init() tea.Cmd {
	// initialize Yuta’s animation (seem to be broken ATM)
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
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Align(lipgloss.Center).
		Width(40).
		Padding(1, 0, 1, 0).
		BorderForeground(lipgloss.Color("63"))

	title := titleStyle.Render("🚀 STARSHIP SIMULATION 🚀")

	stats := fmt.Sprintf("Ship Health: %d/%d", g.currentHealth, g.maxHealth)

	// render the progress bar
	healthBar := g.ProgressBar.RenderProgressBar(g.currentHealth, g.maxHealth)

	var menuView strings.Builder
	for i, item := range g.menuItems {
		cursor := "_" // update later
		if i == g.menuCursor {
			cursor = ">"
		}
		menuView.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
	}

	// left panel
	leftPanel := lipgloss.NewStyle().
		Width(40).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center).
		Render(fmt.Sprintf("%s\n\n%s", title, menuView.String()))

	// center panel for stats and the progress bar
	centerContent := fmt.Sprintf("%s\n\n%s", stats, healthBar)
	centerPanel := lipgloss.NewStyle().
		Width(50).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(centerContent)

	// right panel with Yuta
	rightPanel := lipgloss.NewStyle().
		Width(40).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("34")).
		Align(lipgloss.Center).
		Render(g.Yuta.View())

	// bottom panel
	bottomPanelContent := "This is the bottom panel."
	if g.selectedItem != "" {
		bottomPanelContent = fmt.Sprintf("You selected: %s", g.selectedItem)
	}
	bottomPanel := lipgloss.NewStyle().
		Width(134).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(bottomPanelContent)

	// Combine panels into the main view
	mainView := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center, leftPanel, centerPanel, rightPanel),
		bottomPanel,
	)

	return mainView
}

// NewGameModel creates and returns a new GameModel instance
func NewGameModel() tea.Model {
	return GameModel{
		ProgressBar:   components.NewProgressBar(),
		currentHealth: 62,                   // example initial health
		maxHealth:     100,                  // example max health
		Yuta:          components.NewYuta(), // initialize Yuta
		menuItems:     []string{"Option 1", "Option 2", "Option 3"},
		menuCursor:    0, // Start cursor at the first menu item
	}
}
