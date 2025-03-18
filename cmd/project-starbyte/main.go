package main

import (
	"fmt"
	"log"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
	"github.com/dominik-merdzik/project-starbyte/internal/tui/views"

	configs "github.com/dominik-merdzik/project-starbyte/configs"
	music "github.com/dominik-merdzik/project-starbyte/internal/music"
)

type menuModel struct {
	choices    []string
	cursor     int
	output     string
	configPath string
}

func main() {

	// Define the relative path to your configuration file
	configPath := "config/config.toml"

	// Initialize the config (ensures directory exists, creates default if missing, then loads)
	cfg, err := configs.InitConfig(configPath)
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	// Get the absolute path of the config file
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		log.Printf("Error obtaining absolute config path: %v", err)
		absConfigPath = configPath // fallback to relative path
	}

	// initialize the background music using the loaded config
	music.PlayBackgroundMusicFromEmbed(cfg.Music)

	// Setup menu choices.
	var choices []string
	if data.SaveExists() {
		choices = []string{"Enter Simulation", "Edit Config", "Help", "Exit"}
	} else {
		choices = []string{"Start New Simulation", "Edit Config", "Help", "Exit"}
	}

	// menuModel storing the absolute config path
	model := menuModel{
		choices:    choices,
		configPath: absConfigPath,
	}

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
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "q":
			return m, tea.Quit
		case "enter":
			switch m.choices[m.cursor] {
			case "Start New Simulation":
				return views.NewGameCreationModel(), tea.EnterAltScreen
			case "Enter Simulation":
				return views.NewGameModel(), tea.EnterAltScreen
			case "Edit Config":
				m.output = "You can find and edit your config file at:\n" + m.configPath
			case "Help":
				m.output = "Help Menu:\n - Enter Simulation: Start the game\n - Edit Config: Modify settings\n - Help: Show this menu\n - Exit: Quit the program"
			case "Exit":
				return m, tea.Quit
			}
		}
	}
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

	// ASCII art title
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
			cursor = cursorStyle.Render(">")
		}
		menu += fmt.Sprintf(" %s %s\n", cursor, choiceStyle.Render(choice))
	}

	// render key hints
	hints := hintStyle.Render("[k ↑ j ↓ ~ arrow keys ] Navigate • [Enter] Select • [q] Quit ")

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
