package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ShipModel represents the ship's status and components

type ShipModel struct {
	Name           string
	HullHealth     int
	EngineHealth   int
	EngineFuel     int
	FTLDriveHealth int
	FTLDriveCharge int
	Crew           []CrewMember
	Food           int
	Cursor         int // Index of the currently selected ship component
}

// NewShipModel creates and returns a new ShipModel
func NewShipModel() ShipModel {
	return ShipModel{
		Name:           "Voyager 3",
		HullHealth:     100,
		EngineHealth:   100,
		EngineFuel:     100,
		FTLDriveHealth: 100,
		FTLDriveCharge: 0,
		Crew:           []CrewMember{},
		Food:           100,
		Cursor:         0,
	}
}

// Init initializes the model
func (s ShipModel) Init() tea.Cmd {
	return nil
}

// Update handles key inputs
func (s ShipModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.Cursor > 0 {
				s.Cursor--
			}
		case "down", "j":
			if s.Cursor < 5 { // Number of selectable items
				s.Cursor++
			}
		}
	}
	return s, nil
}

// View renders the ship model UI
func (s ShipModel) View() string {
	items := []string{
		"Hull Health",
		"Engine Health",
		"Engine Fuel",
		"FTL Drive Health",
		"FTL Drive Charge",
		"Food",
	}

	// Styling
	panelStyle := lipgloss.NewStyle().
		Width(60).
		Height(18).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)
	//cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	details := ""

	var shipList strings.Builder
	shipList.WriteString(titleStyle.Render("Ship Status") + "\n\n")

	for i, item := range items {
		cursor := "  "
		if i == s.Cursor {
			cursor = arrowStyle.Render("> ")
		}
		shipList.WriteString(fmt.Sprintf("%s%s\n", cursor, defaultStyle.Render(item)))
	}

	// Detailed panel based on selection
	switch s.Cursor {
	case 0:
		details = fmt.Sprintf("%s\n\n%s %d",
			titleStyle.Render("Hull Health"),
			labelStyle.Render("Current: "), s.HullHealth)
	case 1:
		details = fmt.Sprintf("%s\n\n%s %d",
			titleStyle.Render("Engine Health"),
			labelStyle.Render("Current: "), s.EngineHealth)
	case 2:
		details = fmt.Sprintf("%s\n\n%s %d",
			titleStyle.Render("Engine Fuel"),
			labelStyle.Render("Current: "), s.EngineFuel)
	case 3:
		details = fmt.Sprintf("%s\n\n%s %d",
			titleStyle.Render("FTL Drive Health"),
			labelStyle.Render("Current: "), s.FTLDriveHealth)
	case 4:
		details = fmt.Sprintf("%s\n\n%s %d",
			titleStyle.Render("FTL Drive Charge"),
			labelStyle.Render("Current: "), s.FTLDriveCharge)
	case 5:
		details = fmt.Sprintf("%s\n\n%s %d",
			titleStyle.Render("Food Supply"),
			labelStyle.Render("Current: "), s.Food)
	}

	detailPanel := panelStyle.Render(details)
	leftPanel := panelStyle.Render(shipList.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, detailPanel)
}
