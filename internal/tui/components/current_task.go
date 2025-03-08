package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	model "github.com/dominik-merdzik/project-starbyte/internal/tui/models"
)

// CurrentTaskComponent is responsible for rendering the currently tracked mission
type CurrentTaskComponent struct{}

// NewCurrentTaskComponent creates and returns a new CurrentTaskComponent
func NewCurrentTaskComponent() CurrentTaskComponent {
	return CurrentTaskComponent{}
}

// Render returns a string with the current task data rendered in a styled box
func (c CurrentTaskComponent) Render(task *model.Mission) string {
	if task == nil {
		return "No current task."
	}

	boxStyle := lipgloss.NewStyle().
		Width(55).
		Height(8).
		Align(lipgloss.Left)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)

	content := fmt.Sprintf("%s\n\n%s",
		titleStyle.Render("Tracking Mission: "+task.Title),
		labelStyle.Render("Status:")+" "+task.Status.String(),
	)
	return boxStyle.Render(content)
}
