package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// CollectionModel represents the model for displaying the player's Collection.
type CollectionModel struct {
	Collection data.Collection
	// Optionally, you could add a cursor here if you wish to navigate through research note tiers.
}

// NewCollectionModel creates a new CollectionModel with the given collection data.
func NewCollectionModel(c data.Collection) CollectionModel {
	return CollectionModel{
		Collection: c,
	}
}

// Init implements the tea.Model interface.
func (m CollectionModel) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface.
// Add key handling here if you need interactive features.
func (m CollectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// For now, no interaction is required.
	return m, nil
}

// View renders the Collection view.
func (m CollectionModel) View() string {
	var b strings.Builder

	// Title for the view.
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	b.WriteString(titleStyle.Render("PLAYER Collection") + "\n\n")

	// Display overall collection capacity.
	b.WriteString(fmt.Sprintf("Capacity: %d / %d\n\n", m.Collection.UsedCapacity, m.Collection.MaxCapacity))
	b.WriteString("Research Notes:\n")

	// List each research note tier.
	for _, tier := range m.Collection.ResearchNotes {
		b.WriteString(fmt.Sprintf(" - %s: %d notes (XP: %d)\n", tier.Name, tier.Quantity, tier.XP))
	}

	return b.String()
}
