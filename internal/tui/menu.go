package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GameModel struct {
	starshipPosition int  // controls the vertical position of the starship
	moveUp           bool // direction of movement (true for up, false for down)
}

func (g GameModel) Init() tea.Cmd {
	// start the hovering effect
	return hoverStarship()
}

func (g GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return g, tea.Quit
		}
	case hoverMsg:
		// update the position and direction for hovering
		if g.moveUp {
			g.starshipPosition--
			if g.starshipPosition <= 0 {
				g.moveUp = false
			}
		} else {
			g.starshipPosition++
			if g.starshipPosition >= 2 { // max hover range
				g.moveUp = true
			}
		}
		return g, hoverStarship()
	}
	return g, nil
}

func (g GameModel) View() string {
	// title styling
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")). // Purple color
		Align(lipgloss.Center).
		Width(80).
		Padding(1, 0, 1, 0).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("63"))

	// static title text
	title := titleStyle.Render("🚀 STARSHIP SIMULATION 🚀")

	// starship styling
	starshipArt := `
       !
       !
       ^
      / \
     /___\
    |=   =|
    |     |
    |     |
    |     |
    |     |
    |     |
    |     |
    |     |
    |     |
    |     |
   /|##!##|\
  / |##!##| \
 /  |##!##|  \
|  / ^ | ^ \  |
| /  ( | )  \ |
|/   ( | )   \|
    ((   ))
   ((  :  ))
   ((  :  ))
    ((   ))
     (( ))
      ( )
       .
       .
       .
`
	starshipStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(80)


	// add blank lines above the starship to simulate hovering
	hoveredStarship := repeat("\n", g.starshipPosition) + starshipStyle.Render(starshipArt)

	// combine the title and the centered starship
	return fmt.Sprintf("%s\n\n%s", title, hoveredStarship)
}

// NewGameModel creates and returns a new GameModel instance
func NewGameModel() tea.Model {
	return GameModel{
		starshipPosition: 1, // Start at the middle position
		moveUp:           true,
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

// hoverMsg is a custom message for triggering the hover effect
type hoverMsg struct{}

// hoverStarship returns a tea.Cmd that triggers a hoverMsg after a delay
func hoverStarship() tea.Cmd {
	return tea.Tick(325*time.Millisecond, func(time.Time) tea.Msg {
		return hoverMsg{}
	})
}

// repeat returns a string repeated n times
func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

