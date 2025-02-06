package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mission defines a single journal mission.
type Mission struct {
	Title        string
	Description  string
	Status       string
	Location     string
	Income       string
	Requirements string
	Received     string
}

// JournalModel represents the state and behavior of our Journal component.
type JournalModel struct {
	Missions         []Mission // full list of missions
	Cursor           int       // index of the currently selected mission in the current display list
	SearchMode       bool      // whether search mode is active (typing a query)
	SearchQuery      string    // current search query text
	FilteredMissions []Mission // missions filtered by the search query
}

// NewJournalModel initializes and returns a new JournalModel with some fun missions.
func NewJournalModel() JournalModel {
	missions := []Mission{
		{
			Title:        "Rescue Operation",
			Description:  "Locate and rescue the stranded astronaut on a rogue asteroid.",
			Status:       "Pending",
			Location:     "Mars",
			Income:       "5000 Credits",
			Requirements: "Advanced Rescue Kit",
			Received:     "Command Center",
		},
		{
			Title:        "Nebula Exploration",
			Description:  "Explore the forbidden nebula to gather scientific data and rare minerals.",
			Status:       "In Progress",
			Location:     "Orion Nebula",
			Income:       "8000 Credits",
			Requirements: "Scientific Vessel",
			Received:     "Galactic Federation",
		},
		{
			Title:        "Artifact Recovery",
			Description:  "Retrieve the ancient artifact from the ruins of a lost civilization.",
			Status:       "Completed",
			Location:     "Xandar",
			Income:       "12000 Credits",
			Requirements: "Archaeological Tools",
			Received:     "Mysterious Benefactor",
		},
	}

	return JournalModel{
		Missions: missions,
		Cursor:   0,
	}
}

// Init is called when the JournalModel is first initialized.
func (j JournalModel) Init() tea.Cmd {
	return nil
}

// Update handles key events for both normal navigation and search input.
func (j JournalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if j.SearchMode {
			switch msg.String() {
			case "backspace":
				if len(j.SearchQuery) > 0 {
					j.SearchQuery = j.SearchQuery[:len(j.SearchQuery)-1]
				}
			case "enter":
				// Finalize search: exit search mode but keep search query and filtered results.
				j.SearchMode = false
				return j, nil
			case "esc":
				// Cancel search: clear query and filtered results but remain in Journal view.
				j.SearchMode = false
				j.SearchQuery = ""
				j.FilteredMissions = nil
				j.Cursor = 0
				return j, nil
			default:
				// Accept only single-character inputs.
				if len(msg.String()) == 1 {
					j.SearchQuery += msg.String()
				}
			}
			// Recompute filtered missions live.
			j.FilteredMissions = nil
			for _, m := range j.Missions {
				if strings.Contains(strings.ToLower(m.Title), strings.ToLower(j.SearchQuery)) ||
					strings.Contains(strings.ToLower(m.Description), strings.ToLower(j.SearchQuery)) {
					j.FilteredMissions = append(j.FilteredMissions, m)
				}
			}
			// Reset cursor if it exceeds the filtered results.
			if j.Cursor >= len(j.FilteredMissions) {
				j.Cursor = 0
			}
			return j, nil
		}

		// When not in search mode, use the (possibly filtered) results for navigation.
		switch msg.String() {
		case "up", "k":
			if j.SearchQuery != "" {
				if j.Cursor > 0 {
					j.Cursor--
				}
			} else {
				if j.Cursor > 0 {
					j.Cursor--
				}
			}
		case "down", "j":
			if j.SearchQuery != "" {
				if j.Cursor < len(j.FilteredMissions)-1 {
					j.Cursor++
				}
			} else {
				if j.Cursor < len(j.Missions)-1 {
					j.Cursor++
				}
			}
		case "s":
			// Activate search mode.
			j.SearchMode = true
			j.SearchQuery = ""
			j.FilteredMissions = nil
			j.Cursor = 0
		}
	}
	return j, nil
}

// View renders the JournalModel as two side-by-side panels with a vertical divider.
func (j JournalModel) View() string {
	// if a search query exists (even if search mode is off), show filtered missions.
	var missionsToDisplay []Mission
	if j.SearchQuery != "" {
		if len(j.FilteredMissions) > 0 {
			missionsToDisplay = j.FilteredMissions
		} else {
			missionsToDisplay = []Mission{}
		}
	} else {
		missionsToDisplay = j.Missions
	}

	// ----------------------------
	// Left Panel: Mission List
	// ----------------------------
	leftStyle := lipgloss.NewStyle().
		Width(60).
		Height(15).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var missionList strings.Builder
	// If in search mode, display the search query at the top.
	if j.SearchMode {
		missionList.WriteString("Search: " + j.SearchQuery + "\n\n")
	}
	for i, mission := range missionsToDisplay {
		if i == j.Cursor {
			// Selected mission: show arrow and hover color.
			missionList.WriteString(fmt.Sprintf("%s %s\n",
				arrowStyle.Render(">"),
				hoverStyle.Render(mission.Title)))
		} else {
			missionList.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(mission.Title)))
		}
	}
	leftPanel := leftStyle.Render(missionList.String())

	// ----------------------------
	// Right Panel: Mission Details
	// ----------------------------
	rightStyle := lipgloss.NewStyle().
		Width(60).
		Height(15).
        Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	// style for the mission title (bold, colored) and for the labels (bold, uncolored).
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)

	var details string
	if len(missionsToDisplay) > 0 {
		selectedMission := missionsToDisplay[j.Cursor]
		details = fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n%s\n%s",
			titleStyle.Render(selectedMission.Title),
			labelStyle.Render("Description:")+" "+selectedMission.Description,
			labelStyle.Render("Status:")+" "+selectedMission.Status,
			labelStyle.Render("Location:")+" "+selectedMission.Location,
			labelStyle.Render("Income:")+" "+selectedMission.Income,
			labelStyle.Render("Requirements:")+" "+selectedMission.Requirements,
			labelStyle.Render("Received:")+" "+selectedMission.Received,
		)
	} else {
		details = "No missions found."
	}
	rightPanel := rightStyle.Render(details)

	// ----------------------------
	// Vertical Divider
	// ----------------------------
	const divider = `
│
│
│
│
│
│
│
│
│
│
│
│
│
│
│
`
	dividerStyle := lipgloss.NewStyle().
		Width(1).
		Height(15).
        Align(lipgloss.Center).
		Foreground(lipgloss.Color("240"))
	div := dividerStyle.Render(divider)

	// ----------------------------
	// Combine Panels
	// ----------------------------
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, div, rightPanel)
}

