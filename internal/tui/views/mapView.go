package views

import (
	tea "github.com/charmbracelet/bubbletea"
)

type mapModel struct {
}

func (m mapModel) Init() tea.Cmd {
	return nil
}

func (m mapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return GameModel{}, nil
		}
	}
	return m, nil
}

func (m mapModel) View() string {
	return "Map View - press q to go back"
}
