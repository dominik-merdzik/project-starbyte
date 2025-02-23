package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MapModel struct {
	width  int
	height int
}

func (m MapModel) Init() tea.Cmd {
	return nil
}

func (m MapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize messages
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m MapModel) View() string {
	ship := `
^
/ \
/___\
|=   =|
|  O  |
|     |
|  O  |
/|##!##|\
/ |##!##| \
/  |##!##|  \
|  / ^ | ^ \  |
| /  ( | )  \ |
|/   ( | )   \|
((   ))
`
	travelLine := "----------------->"
	planet := `
,MMM8&&&.
_...MMMMM88&&&&..._
.::'''MMMMM88&&&&&&'''::.
::     MMMMM88&&&&&&     ::
'::....MMMMM88&&&&&&....::'
''''MMMMM88&&&&''''
'MMM8&&&'
`
	// Define labels at top
	shipLabel := "Ship"                         // INCLUDE SHIP NAME
	travelTimeLabel := "Travel time remaining:" // INCLUDE TIME LEFT TO TRAVEL TO CURRENT PLANET
	planetLabel := "Destination:"               // INCLUDE DESTINATION/PLANET NAME

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Width(40).
		Align(lipgloss.Center)

	// Width for all 3 sections
	sectionWidth := 40

	shipStyle := lipgloss.NewStyle().
		Width(sectionWidth).
		Height(12).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	travelLineStyle := lipgloss.NewStyle().
		Width(30).
		Height(12).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	planetStyle := lipgloss.NewStyle().
		Width(sectionWidth).
		Height(12).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	// Render sections with only the ASCII art centered
	shipContent := shipStyle.Render(ship)
	travelLineContent := travelLineStyle.Render(travelLine)
	planetContent := planetStyle.Render(planet)

	mapContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		shipContent,
		travelLineContent,
		planetContent,
	)

	// Render labels above
	labelsRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		labelStyle.Width(sectionWidth).Render(shipLabel),
		labelStyle.Width(30).Render(travelTimeLabel),
		labelStyle.Width(sectionWidth).Render(planetLabel),
	)

	// Wrap inside panel
	panel := lipgloss.NewStyle().
		Width(120).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(mapContent)

	return lipgloss.JoinVertical(lipgloss.Left, labelsRow, panel)
}
