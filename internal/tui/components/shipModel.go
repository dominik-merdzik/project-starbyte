package components

import tea "github.com/charmbracelet/bubbletea"

// ShipModel - is a Bubble Tea model representing our "Ship" view/component
type ShipModel struct {
	// for internal state
}

// NewShipModel - creates and returns a new ShipModel
func NewShipModel() ShipModel {
	return ShipModel{}
}

// Init - is called when the ShipModel is first initialized (optional)
func (s ShipModel) Init() tea.Cmd {
	return nil
}

// Update - handles incoming messages and updates the ShipModel state accordingly
func (s ShipModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// for now, this model doesnt need to update anything, so we just return itself
	return s, nil
}

// View - returns the string to display for the ShipModel component
func (s ShipModel) View() string {
	return "Sent from ship component!"
}
