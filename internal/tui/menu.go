package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GameModel struct {}

func (g GameModel) Init() tea.Cmd {
	// No initialization needed
	return nil
}

func (g GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return g, tea.Quit
		}
	}
	return g, nil
}

func (g GameModel) View() string {
	// Title styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")). // Purple color
		Align(lipgloss.Center).
		Width(40).
		Padding(1, 0, 1, 0).
		BorderForeground(lipgloss.Color("63"))

	title := titleStyle.Render("🚀 STARSHIP SIMULATION 🚀")
	stats := titleStyle.Render("Statistics")
	starshipArt := `
	__
	| \
	=[_|H)--._____
	=[+--,-------'
	[|_/""
`
	leftPanel := lipgloss.NewStyle().
		Width(40).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center).
		Render(title)

	rightPanel := lipgloss.NewStyle().
		Width(40).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("34")).
		Render(starshipArt)

	centerPanel := lipgloss.NewStyle().
		Width(50).
		Height(25).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render(stats)

	bottomPanel := lipgloss.NewStyle().
		Width(134).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		Align(lipgloss.Center).
		Render("This is the bottom panel.")

	mainView := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Center, leftPanel, centerPanel, rightPanel),
		bottomPanel,
	)

	return mainView
}

// NewGameModel creates and returns a new GameModel instance
func NewGameModel() tea.Model {
	return GameModel{}
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
