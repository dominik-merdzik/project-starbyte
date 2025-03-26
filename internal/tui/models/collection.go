package model

import (
	"fmt"
	"log"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

type CollectionModel struct {
	GameSave *data.FullGameSave
}

// NewCollectionModel now accepts the main GameSave pointer
func NewCollectionModel(gameSave *data.FullGameSave) CollectionModel {
	if gameSave == nil {
		log.Println("Warning: NewCollectionModel received nil gameSave. Initializing with placeholder.")
		// return a model pointing to a default/empty save to prevent nil panics later
		// ensure data.DefaultFullGameSave() exists and provides a valid structure
		return CollectionModel{GameSave: data.DefaultFullGameSave()}
	}
	return CollectionModel{
		GameSave: gameSave, // Store the pointer
	}
}

func (m CollectionModel) Init() tea.Cmd {
	return nil
}

// update remains simple for now, potentially add sorting/filtering later
func (m CollectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// view renders the Collection view, reading directly from GameSave
func (m CollectionModel) View() string {
	if m.GameSave == nil {
		return "Error: Collection data unavailable (GameSave is nil)."
	}

	collection := m.GameSave.Collection

	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	b.WriteString(titleStyle.Render("PLAYER Collection") + "\n\n")

	var calculatedUsedCapacity int
	// add capacity used by research notes
	for _, note := range collection.ResearchNotes {
		calculatedUsedCapacity += note.Quantity
	}
	// add capacity used by items
	for _, item := range collection.Items {
		calculatedUsedCapacity += item.Quantity
	}

	currentUsedCapacity := calculatedUsedCapacity
	maxCapacity := collection.MaxCapacity

	capacityStyle := lipgloss.NewStyle()
	if maxCapacity > 0 {
		usageRatio := float64(currentUsedCapacity) / float64(maxCapacity)
		if usageRatio > 0.9 {
			capacityStyle = capacityStyle.Foreground(lipgloss.Color("208")) // Orange
		}
		if usageRatio >= 1.0 {
			capacityStyle = capacityStyle.Foreground(lipgloss.Color("196")) // Red
		}
	}
	b.WriteString(capacityStyle.Render(fmt.Sprintf("Capacity: %d / %d", currentUsedCapacity, maxCapacity)) + "\n\n")

	// display research notes
	noteTitleStyle := lipgloss.NewStyle().Bold(true)
	b.WriteString(noteTitleStyle.Render("Research Notes:") + "\n")
	notesFound := false
	sort.SliceStable(collection.ResearchNotes, func(i, j int) bool {
		return collection.ResearchNotes[i].Tier < collection.ResearchNotes[j].Tier
	})
	for _, tier := range collection.ResearchNotes {
		// only display tiers if the player actually has notes of that tier
		if tier.Quantity > 0 {
			b.WriteString(fmt.Sprintf(" - %s: %d (XP: %d)\n", tier.Name, tier.Quantity, tier.XP))
			notesFound = true
		}
	}
	if !notesFound {
		b.WriteString("   None\n")
	}

	itemTitleStyle := lipgloss.NewStyle().Bold(true)
	b.WriteString("\n" + itemTitleStyle.Render("Items:") + "\n")
	if len(collection.Items) == 0 {
		b.WriteString("   None\n")
	} else {
		for _, item := range collection.Items {
			if item.Quantity > 0 { // Only show items they have
				b.WriteString(fmt.Sprintf(" - %s: %d\n", item.Name, item.Quantity))
			}
		}
	}

	return b.String()
}
