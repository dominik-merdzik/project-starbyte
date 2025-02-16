package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ShipModel - is a Bubble Tea model representing our "Ship" view/component
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

// NewShipModel - creates and returns a new ShipModel
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

// Init - is called when the ShipModel is first initialized (optional)
func (s ShipModel) Init() tea.Cmd {
	return nil
}

// Update - handles incoming messages and updates the ShipModel state accordingly
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

// View - returns the string to display for the ShipModel component
func (s ShipModel) View() string {
	items := []string{
		"Hull Health",
		"Engine Health",
		"Engine Fuel",
		"FTL Drive Health",
		"FTL Drive Charge",
		"Food",
	}

	var view strings.Builder
	for i, item := range items {
		cursor := " "
		if i == s.Cursor {
			cursor = ">"
		}
		view.WriteString(fmt.Sprintf("%s %s\n", cursor, item))
	}

	return view.String()
}
