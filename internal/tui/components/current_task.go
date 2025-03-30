package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// CurrentTaskComponent is responsible for rendering the currently tracked mission
type CurrentTaskComponent struct {
	GameSave *data.FullGameSave
}

// NewCurrentTaskComponent creates and returns a new CurrentTaskComponent
func NewCurrentTaskComponent() CurrentTaskComponent {
	return CurrentTaskComponent{}
}

// Render returns a string with the current task data rendered in a styled box
func (c CurrentTaskComponent) Render(task *data.Mission) string {
	if task == nil {
		return "No current task."
	}

	boxStyle := lipgloss.NewStyle().
		Width(55).
		Height(8).
		Align(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)

	content := fmt.Sprintf("%s\n\n%s",
		titleStyle.Render("Tracking Mission: "+task.Title),
		labelStyle.Render("Status:")+" "+task.Status.String(),
	)
	return boxStyle.Render(content)
}
