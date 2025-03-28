package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

type EventModel struct {
	event      *data.Event
	currentIdx int
	selected   bool
	Active     bool
}

// Message to tell game.go to exit event
type EventFinishedMsg struct{}

// Update ship values
type ApplyEffectsMsg struct {
	Effects map[string]int
}

func NewEventModel(event *data.Event) *EventModel {
	return &EventModel{
		event:      event,
		currentIdx: 0,
		selected:   false,
		Active:     true,
	}
}

func (m *EventModel) Update(msg tea.Msg) (*EventModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if !m.selected && m.currentIdx > 0 {
				m.currentIdx--
			}
		case "down", "j":
			if !m.selected && m.currentIdx < len(m.event.Choices)-1 {
				m.currentIdx++
			}
		case "enter":
			if !m.selected {
				m.selected = true
				return m, func() tea.Msg {
					return ApplyEffectsMsg{Effects: m.event.Choices[m.currentIdx].Effects}
				}
			} else {
				return m, func() tea.Msg { return EventFinishedMsg{} }
			}
		case "q":
			return m, func() tea.Msg { return EventFinishedMsg{} }
		}
	}
	return m, nil
}

func (m *EventModel) View() string {
	style := lipgloss.NewStyle().
		Width(80).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	var content strings.Builder
	content.WriteString(fmt.Sprintf("%s\n\n", m.event.Title))

	for _, line := range m.event.Dialogue {
		content.WriteString(line + "\n")
	}

	if m.selected {
		content.WriteString(fmt.Sprintf("\nOutcome: %s", m.event.Choices[m.currentIdx].Outcome))
		content.WriteString("\n\nPress [Enter] to continue.")
	} else {
		content.WriteString("\nWhat will you do?\n")
		for i, choice := range m.event.Choices {
			cursor := "  "
			if i == m.currentIdx {
				cursor = "> "
			}
			content.WriteString(fmt.Sprintf("%s%s\n", cursor, choice.Text))
		}
	}

	return style.Render(content.String())
}
