package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
	"github.com/dominik-merdzik/project-starbyte/internal/tui/components"
)

// newGameModel represents the model for the new game creation form
type newGameModel struct {
	inputs     []textinput.Model
	focusIndex int
	err        error
	showIntro  bool // flag to show the intro exposition
	Dialogue   *components.DialogueComponent
	lines      []string // lines of dialogue to display
}

// NewGameCreationModel initializes the new game creation form
func NewGameCreationModel() tea.Model {
	m := newGameModel{
		inputs:     make([]textinput.Model, 3),
		focusIndex: 0,
		showIntro:  true,
		lines: []string{
			"The year is 2399 and humanity has reached the stars.",
			"Space travel beyond the Solar System is now possible with the use of Faster-Than-Light (FTL) technology.",
			"They are not alone in the universe. Contact has been made with intelligent alien species.",
			"Some are friendly, some hostile, and some are indifferent.",
			"Humanity is now part of a galactic community, but the galaxy is a mysterious place.",
			"Your ship and crew are ready, and the stars await...",
		},
	}

	// Initialize Dialogue component
	d := components.NewDialogueComponentFromMission(m.lines)
	m.Dialogue = &d

	// 1. ship Name
	ti := textinput.New()
	ti.Placeholder = "Enter ship name"
	ti.Focus() // first field gets focus
	ti.CharLimit = 20
	m.inputs[0] = ti

	// 2. game difficulty
	ti2 := textinput.New()
	ti2.Placeholder = "Enter game difficulty (Easy, Normal, Hard)"
	ti2.CharLimit = 10
	m.inputs[1] = ti2

	// 3. starting location
	ti3 := textinput.New()
	ti3.Placeholder = "Enter starting location (default: Earth)"
	ti3.CharLimit = 20
	m.inputs[2] = ti3

	return m
}

func (m newGameModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m newGameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle quit for all key messages
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		// handle Enter key for the dialogue component
		if m.showIntro {
			if msg.String() == "enter" {
				m.Dialogue.Next()
				if m.Dialogue.CurrentLine >= len(m.lines) {
					m.showIntro = false
				}
			}
			if msg.String() == "backspace" {
				m.Dialogue.Previous()
			}
			if msg.String() == " " {
				m.showIntro = false // Skip
			}
			// Must return so that it doesnt input into the stuff below
			return m, nil
		}

		// Converted from switch to if for slight performance gain
		if msg.String() == "tab" || msg.String() == "shift+tab" ||
			msg.String() == "enter" || msg.String() == "up" || msg.String() == "down" {
			// when pressing Enter on the last input, assume the form is complete
			if msg.String() == "enter" && m.focusIndex == len(m.inputs)-1 {
				// gather input values
				shipName := m.inputs[0].Value()
				difficulty := m.inputs[1].Value()
				location := m.inputs[2].Value()
				if strings.TrimSpace(shipName) == "" {
					shipName = "Starship"
				}
				if strings.TrimSpace(difficulty) == "" {
					difficulty = "Normal" // default difficulty
				}
				if strings.TrimSpace(location) == "" {
					location = "Earth" // default starting location
				}

				// create a new full game save populated with all new game data
				if err := data.CreateNewFullGameSave(difficulty, shipName, location); err != nil {
					m.err = err
					return m, nil
				}

				// after creating the save, load the game simulation
				return NewGameModel(), nil
			}

			// handle focus movement (tab/shift+tab/up/down)
			if msg.String() == "tab" || msg.String() == "down" {
				m.focusIndex++
				if m.focusIndex > len(m.inputs)-1 {
					m.focusIndex = 0
				}
			} else if msg.String() == "shift+tab" || msg.String() == "up" {
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs) - 1
				}
			}
		}
	}

	// update all text inputs and set focus accordingly
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m newGameModel) View() string {
	// -----
	// INTRO
	// -----
	if m.showIntro {
		dialogueContent := m.Dialogue.View(80 - 2) // Width: 80 - 2. Assuming user has standard unix 80x24 terminal size

		hints := "[Enter] Next • [Backspace] Previous • [Space] Skip • [Esc] Quit"
		hintsRowStyle := lipgloss.NewStyle().
			Width(80).
			Padding(0, 1).
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("15"))

		// Show hints below dialogue
		return dialogueContent + "\n" + hintsRowStyle.Render(hints)
	}

	// --------------
	// NEW GAME SETUP
	// --------------
	var b strings.Builder

	b.WriteString("=== New Simulation Setup ===\n\n")
	b.WriteString("Please enter the following details:\n\n")

	labels := []string{"Ship Name: ", "Game Difficulty: ", "Starting Location: "}
	for i, input := range m.inputs {
		b.WriteString(labels[i] + input.View() + "\n")
	}
	b.WriteString("\n(Press Enter on the last field to start the simulation)")
	if m.err != nil {
		b.WriteString("\nError: " + m.err.Error())
	}

	style := lipgloss.NewStyle().
		Width(80-2).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	return style.Render(b.String())
}
