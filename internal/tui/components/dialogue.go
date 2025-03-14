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

// Renders dialogue like so:
// > Dialogue line 1
// > Dialogue line 2
// > Dialogue line n
func (d DialogueComponent) View() string {
	// var content strings.Builder
	// prefixStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	// containerStyle := lipgloss.NewStyle().Align(lipgloss.Left)
	// for i := 0; i < d.CurrentLine && i < len(d.Lines); i++ {
	// 	content.WriteString(fmt.Sprintf("%s %s\n", prefixStyle.Render(">"), d.Lines[i]))
	// }
	// return containerStyle.Render(content.String())

	if len(d.Lines) == 0 || d.CurrentLine >= len(d.Lines) {
		return "End of dialogue."
	}

	dialogueStyle := lipgloss.NewStyle().
		Width(120).
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	content := fmt.Sprintf("%s", d.Lines[d.CurrentLine])

	// Add a progress indicator
	progress := fmt.Sprintf("\n\n[%d/%d]", d.CurrentLine+1, len(d.Lines))

	return dialogueStyle.Render(content + progress)
}
