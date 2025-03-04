package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type DialogueComponent struct{}

func NewDialogueComponent() DialogueComponent {
	return DialogueComponent{}
}

// Renders dialogue like so:
// > Dialogue line 1
// > Dialogue line 2
// > Dialogue line n
func (c DialogueComponent) Render(dialogue string) string {
	style := lipgloss.NewStyle().Align(lipgloss.Left)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))

	lines := strings.Split(dialogue, "\n")
	var content string

	for _, line := range lines {
		if line != "" {
			content += fmt.Sprintf("%s\n", titleStyle.Render("> ")+line)
		}
	}
	return style.Render(content)
}
