package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type DialogueComponent struct {
	Lines       []string
	CurrentLine int
}

// NewDialogueComponentFromMission creates a new DialogueComponent from a mission's dialogue
func NewDialogueComponentFromMission(lines []string) DialogueComponent {
	return DialogueComponent{
		Lines:       lines,
		CurrentLine: 0,
	}
}

// next advances the dialogue to the next line, if available
func (d *DialogueComponent) Next() {
	if d.CurrentLine < len(d.Lines) {
		d.CurrentLine++
	}
}

// View renders the dialogue component
// Width is an optional parameter
func (d DialogueComponent) View(width ...int) string {
	if len(d.Lines) == 0 || d.CurrentLine >= len(d.Lines) {
		return "End of dialogue."
	}

	// Set default width to 120 if not specified
	dialogueWidth := 120
	if len(width) > 0 && width[0] > 0 {
		dialogueWidth = width[0]
	}

	dialogueStyle := lipgloss.NewStyle().
		Width(dialogueWidth).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	content := d.Lines[d.CurrentLine]

	// Add a progress indicator
	progress := fmt.Sprintf("\n\n[%d/%d]", d.CurrentLine+1, len(d.Lines))

	return dialogueStyle.Render(content + progress)
}
