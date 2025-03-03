package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data" // Import the package where saved crew data is defined.
)

// a single crew member in our model
type CrewMember struct {
	Name       string
	Role       string // e.g. Captain, Navigator, Engineer, Gunner, Cook, Deckhand
	Degree     int
	Experience int
	Morale     int
	Health     int
	HireCost   int // to be used when recruiting crew members
}

// all of the crew on board the player's ship
type CrewModel struct {
	CrewMembers []CrewMember
	Cursor      int
}

// NewCrewModel creates a new CrewModel based on saved crew data
// it converts a slice of data.CrewMember (from the save file) into the internal CrewMember type
func NewCrewModel(savedCrew []data.CrewMember) CrewModel {
	var crew []CrewMember
	for _, s := range savedCrew {
		crew = append(crew, CrewMember{
			Name:       s.Name,
			Role:       s.Role,
			Degree:     s.Degree,
			Experience: s.Experience,
			Morale:     s.Morale,
			Health:     s.Health,
			HireCost:   100, // idea for later - adjust or compute based on s.Degree
		})
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

func (c CrewModel) View() string {
	// Left Panel: Crew List
	leftStyle := lipgloss.NewStyle().
		Width(60).
		Height(18).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var crewList strings.Builder
	for i, crew := range c.CrewMembers {
		titleText := crew.Name + " ~ " + crew.Role
		if i == c.Cursor {
			crewList.WriteString(fmt.Sprintf("%s %s\n",
				arrowStyle.Render(">"),
				hoverStyle.Render(titleText)))
		} else {
			crewList.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
		}
	}

	leftPanel := leftStyle.Render(crewList.String())

	// Right Panel: Details of the selected crew member
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
			labelStyle.Render("Degree: ") + fmt.Sprintf("%d", crew.Degree) + "\n" +
			labelStyle.Render("Experience: ") + fmt.Sprintf("%d", crew.Experience) + "\n" +
			labelStyle.Render("Morale: ") + fmt.Sprintf("%d", crew.Morale) + "\n" +
			labelStyle.Render("Health: ") + fmt.Sprintf("%d", crew.Health) + "\n"
		//labelStyle.Render("Hire Cost: ") + "100 credits" + "\n"
	}

	rightPanel := rightStyle.Render(crewDetails)

	// Vertical Divider
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

	// Combine Panels
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, div, rightPanel)
}
