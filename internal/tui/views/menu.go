package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/tui/components"
)

type GameModel struct {
	ProgressBar   components.ProgressBar
	currentHealth int
	maxHealth     int
	Yuta          components.YutaModel
}

func (g GameModel) Init() tea.Cmd {
	// initialize Yuta's animation
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
			return g, tea.Quit
		case "a": // simulate damage and healing for testing
			g.currentHealth -= 10
			if g.currentHealth < 0 {
				g.currentHealth = 0
			}
		case "h":
			g.currentHealth += 10
			if g.currentHealth > g.maxHealth {
				g.currentHealth = g.maxHealth
			}
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

	// rendering the progress bar
	healthBar := g.ProgressBar.RenderProgressBar(g.currentHealth, g.maxHealth)

	// left panel
	leftPanel := lipgloss.NewStyle().
		Width(40).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center).
		Render(title)

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
		Render(g.Yuta.View()) // Render Yuta here

	// bottom panel
	bottomPanel := lipgloss.NewStyle().
		Width(134).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render("This is the bottom panel.")

	// combine panels into the main view
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
	}
}

// StartSimulation initializes and starts the simulation TUI
func StartSimulation() tea.Cmd {
	return func() tea.Msg {
		p := tea.NewProgram(NewGameModel(), tea.WithAltScreen())
		if err := p.Start(); err != nil {
			fmt.Printf("Error starting simulation TUI: %v\n", err)
		}
		return nil
	}
}
