package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
    "github.com/dominik-merdzik/project-starbyte/internal/tui"
)

type menuModel struct {
	choices []string // list of menu options
	cursor  int      // current position of the cursor
	output  string   // additional information or output text
}

func main() {
	// initialize the menu model with menu choices
	model := menuModel{
		choices: []string{
			"Enter Simulation",
			"Edit Config",
			"Help",
			"Exit",
		},
	}

	// create a new program using the menu model and start it
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting application: %v\n", err)
	}
}

func (m menuModel) Init() tea.Cmd {
	// no initial commands
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// handle navigation and selection based on key input
		switch msg.String() {
		case "up", "k":
			// move cursor up
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			// move cursor down
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "q":
			// quit the program
			return m, tea.Quit
		case "enter":
			// handle menu item selection
			switch m.choices[m.cursor] {
			case "Enter Simulation":
				// use the views package to launch the simulation view
				return m, views.StartSimulation()
			case "Edit Config":
				m.output = "Configuration editing is currently not implemented."
			case "Help":
				m.output = "Help Menu:\n - Enter Game: Start the game\n - Edit Config: Modify settings\n - Help: Show this menu\n - Exit: Quit the program"
			case "Exit":
				return m, tea.Quit
			}
		}
	}
	// return updated model without additional commands
	return m, nil
}

func (m menuModel) View() string {
	// define styles for various UI elements
	titleStyle := lipgloss.NewStyle().Bold(true).PaddingLeft(2).Foreground(lipgloss.Color("39"))
	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("201"))
	choiceStyle := lipgloss.NewStyle().PaddingBottom(1).Foreground(lipgloss.Color("229"))
	hintStyle := lipgloss.NewStyle().Faint(true).PaddingLeft(1).Foreground(lipgloss.Color("240"))
	outputStyle := lipgloss.NewStyle().PaddingLeft(2).Italic(true).Foreground(lipgloss.Color("45"))
	columnStyle := lipgloss.NewStyle().Padding(0, 2)

	// define the title using ASCII art
	const title = `
██████╗ ██████╗  ██████╗      ██╗███████╗ ██████╗████████╗    ███████╗████████╗ █████╗ ██████╗ ██████╗ ██╗   ██╗████████╗███████╗
██╔══██╗██╔══██╗██╔═══██╗     ██║██╔════╝██╔════╝╚══██╔══╝    ██╔════╝╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝
██████╔╝██████╔╝██║   ██║     ██║█████╗  ██║        ██║       ███████╗   ██║   ███████║██████╔╝██████╔╝ ╚████╔╝    ██║   █████╗
██╔═══╝ ██╔══██╗██║   ██║██   ██║██╔══╝  ██║        ██║       ╚════██║   ██║   ██╔══██║██╔══██╗██╔══██╗  ╚██╔╝     ██║   ██╔══╝
██║     ██║  ██║╚██████╔╝╚█████╔╝███████╗╚██████╗   ██║       ███████║   ██║   ██║  ██║██║  ██║██████╔╝   ██║      ██║   ███████╗
╚═╝     ╚═╝  ╚═╝ ╚═════╝  ╚════╝ ╚══════╝ ╚═════╝   ╚═╝       ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝    ╚═╝      ╚═╝   ╚══════╝
`
	// render the title
	titleView := titleStyle.Render(title) + "\n\n"

	// render menu options
	menu := ""
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			// add cursor to the current selection
			cursor = cursorStyle.Render(">")
		}
		menu += fmt.Sprintf(" %s %s\n", cursor, choiceStyle.Render(choice))
	}

	// render key hints
	hints := hintStyle.Render("[Arrow Keys] Navigate • [Enter] Select • [q] Quit")

	// render output section
	output := "Welcome to Starbyte!\n"
	if m.output != "" {
		output += m.output
	} else {
		output += " "
	}

	// combine menu and output into two columns
	menuColumn := columnStyle.Render(menu + "\n" + hints)
	outputColumn := outputStyle.Render(output)

	// join the two columns side by side
	columns := lipgloss.JoinHorizontal(lipgloss.Top, menuColumn, outputColumn)

	// combine the title and columns
	return titleView + columns
}
