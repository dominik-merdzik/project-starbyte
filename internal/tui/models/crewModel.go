package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// A single crew member
type CrewMember struct {
	Name     string
	Role     string // Captain, Navigator, Engineer, Gunner, Cook, Deckhand
	Level    int
	HireCost int // To be used when recruiting crew members
}

// All of the crew on board the player's ship
type CrewModel struct {
	CrewMembers []CrewMember
	Cursor      int // Index of the currently selected crew member
}

// NewCrewModel creates a new CrewModel with a default crew
func NewCrewModel() CrewModel {
	crew := []CrewMember{
		{
			Name:     "Alice",
			Role:     "Captain",
			Level:    1,
			HireCost: 100,
		},
		{
			Name:     "Bob",
			Role:     "Navigator",
			Level:    1,
			HireCost: 100,
		},
		{
			Name:     "Charlie",
			Role:     "Engineer",
			Level:    1,
			HireCost: 100,
		},
		{
			Name:     "David",
			Role:     "Gunner",
			Level:    1,
			HireCost: 100,
		},
		{
			Name:     "Eve",
			Role:     "Cook",
			Level:    1,
			HireCost: 100,
		},
		{
			Name:     "Frank",
			Role:     "Deckhand",
			Level:    1,
			HireCost: 100,
		},
	}
	return CrewModel{
		CrewMembers: crew,
		Cursor:      0,
	}
}

func (c CrewModel) Init() tea.Cmd {
	return nil
}

func (c CrewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if c.Cursor > 0 {
				c.Cursor--
			}
		case "down", "j":
			if c.Cursor < len(c.CrewMembers)-1 {
				c.Cursor++
			}
		}
	}
	return c, nil
}

// Much of this is copied from JournalModel.View
func (c CrewModel) View() string {
	// ----------------------------
	// Left Panel: Crew List
	// ----------------------------
	leftStyle := lipgloss.NewStyle().
		Width(60).
		Height(18).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	// Render each crew member per line in the left panel.
	// Prepend the cursor with a '>' character.
	var crewList strings.Builder
	for i, crew := range c.CrewMembers {
		titleText := crew.Name
		if i == c.Cursor {
			crewList.WriteString(fmt.Sprintf("%s %s\n",
				arrowStyle.Render(">"),
				hoverStyle.Render(titleText)))
		} else {
			crewList.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
		}
	}

	leftPanel := leftStyle.Render(crewList.String())

	// ----------------------------
	// Right Panel: Display details of the selected crew member
	// ----------------------------
	rightStyle := lipgloss.NewStyle().
		Width(60).
		Height(18).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)

	var crewDetails string
	if len(c.CrewMembers) > 0 {
		crew := c.CrewMembers[c.Cursor]
		crewDetails = titleStyle.Render(crew.Name) + "\n" +
			labelStyle.Render("Role: ") + crew.Role + "\n" +
			labelStyle.Render("Level: ") + fmt.Sprintf("%d", crew.Level) + "\n" +
			labelStyle.Render("Hire Cost: ") + "100 credits" + "\n"
	}

	rightPanel := rightStyle.Render(crewDetails)

	// ----------------------------
	// Vertical Divider.
	// ----------------------------
	const divider = `
│
│
│
│
│
│
│
│
│
│
│
│
│
│
│
`
	dividerStyle := lipgloss.NewStyle().
		Width(1).
		Height(18).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("240"))
	div := dividerStyle.Render(divider)

	// ----------------------------
	// Combine Panels.
	// ----------------------------
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, div, rightPanel)
}
