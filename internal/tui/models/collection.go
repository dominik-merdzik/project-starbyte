package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

type CollectionModel struct {
	Collection data.Collection

	GameSave *data.FullGameSave
}

func NewCollectionModel(c data.Collection) CollectionModel {
	return CollectionModel{
		Collection: c,
	}
}

func (m CollectionModel) Init() tea.Cmd {
	return nil
}

func (m CollectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View renders the Collection view
func (m CollectionModel) View() string {
	var b strings.Builder

	// Title for the view
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	b.WriteString(titleStyle.Render("PLAYER Collection") + "\n\n")

	// Display overall collection capacity
	b.WriteString(fmt.Sprintf("Capacity: %d / %d\n\n", m.Collection.UsedCapacity, m.Collection.MaxCapacity))
	b.WriteString("Research Notes:\n")

	// List each research note tier
	for _, tier := range m.Collection.ResearchNotes {
		b.WriteString(fmt.Sprintf(" - %s: %d notes (XP: %d)\n", tier.Name, tier.Quantity, tier.XP))
	}

	return b.String()
}
