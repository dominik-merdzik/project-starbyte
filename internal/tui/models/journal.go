package model

import (
	"fmt"
	"log"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// TrackMissionMsg is used to signal that a mission is being tracked
type TrackMissionMsg struct {
	Mission data.Mission
}

// Used to signal the start of a mission
type StartMissionMsg struct {
	Mission data.Mission
}

// convertDataMission converts a data.Mission into a model.Mission
func convertDataMission(dm data.Mission) data.Mission {
	return data.Mission{
		Title:        dm.Title,
		Description:  dm.Description,
		Step:         dm.Step,
		Status:       dm.Status,
		Location:     dm.Location,
		Income:       dm.Income,
		Requirements: dm.Requirements,
		Received:     dm.Received,
		Category:     dm.Category,
		Dialogue:     dm.Dialogue,
	}
}

// currentList returns the missions to be displayed based on search mode,
// filtering out any missions with status "complete" or "completed"
func (j JournalModel) currentList() []data.Mission {
	var baseList []data.Mission
	if j.SearchQuery != "" {
		if len(j.FilteredMissions) > 0 {
			baseList = j.FilteredMissions
		} else {
			baseList = []data.Mission{}
		}
	} else {
		baseList = j.Missions
	}

	return baseList
}

// JournalModel represents the mission journal
type JournalModel struct {
	Missions         []data.Mission
	Cursor           int
	SearchMode       bool
	SearchQuery      string
	FilteredMissions []data.Mission
	Page             int
	PageSize         int

	// Detail view fields.
	DetailView    bool
	DetailCursor  int
	DetailOptions []string
	GameSave      *data.FullGameSave
}

// loads the full game save, extracts the missions, and converts them for display
func NewJournalModel() JournalModel {
	fullSave, err := data.LoadFullGameSave()
	if err != nil {
		log.Printf("Error loading full game save: %v", err)
	}
	if fullSave == nil {
		log.Printf("No save file found; initializing with empty mission list")
		// return an empty JournalModel or initialize with default missions
		return JournalModel{
			Missions:      []data.Mission{},
			Cursor:        0,
			Page:          0,
			PageSize:      5,
			DetailView:    false,
			DetailCursor:  0,
			DetailOptions: []string{"Track", "Start Mission", "Abandon", "Back"},
		}
	}

	missionsFile := fullSave.Missions

	var missions []data.Mission

	// Add all missions
	for _, m := range missionsFile {
		missions = append(missions, convertDataMission(m))
	}

	return JournalModel{
		Missions:      missions,
		Cursor:        0,
		Page:          0,
		PageSize:      5,
		DetailView:    false,
		DetailCursor:  0,
		DetailOptions: []string{"Start Mission", "Track", "Abandon", "Back"},
	}
}

func (j JournalModel) Init() tea.Cmd {
	return nil
}

func (j JournalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// -----------------------------
	// Detail View Key Handling
	// -----------------------------
	case tea.KeyMsg:
		if j.DetailView {
			switch msg.String() {
			case "up", "k":
				if j.DetailCursor > 0 {
					j.DetailCursor--
				}
			case "down", "j":
				if j.DetailCursor < len(j.DetailOptions)-1 {
					j.DetailCursor++
				}
			case "b":
				// Simulate a back action.
				j.DetailView = false
			case "enter":
				// Execute the selected option.
				selectedOption := j.DetailOptions[j.DetailCursor]
				switch selectedOption {
				case "Back":
					j.DetailView = false
				case "Track":
					trackedMission := j.getSelectedMission()
					j.DetailView = false
					return j, func() tea.Msg {
						return TrackMissionMsg{Mission: trackedMission}
					}
				case "Start Mission":
					mission := j.getSelectedMission()
					j.DetailView = false
					if mission.Status == data.MissionStatusNotStarted {
						// First return a command to exit the journal view entirely
						return j, tea.Batch(
							func() tea.Msg {
								// Exit the journal view by sending an ESC key message
								return tea.KeyMsg{Type: tea.KeyEsc}
							},
							func() tea.Msg {
								// Then start the mission
								return StartMissionMsg{Mission: mission}
							},
						)
					}
				case "Abandon":
					mission := j.getSelectedMission()
					mission.Status = data.MissionStatusAbandoned
					j.updateMission(mission)
					j.DetailView = false
				}
			case "esc":
				j.DetailView = false
			}
			return j, nil
		}

		// -----------------------------
		// Search Mode Handling
		// -----------------------------
		if j.SearchMode {
			switch msg.String() {
			case "backspace":
				if len(j.SearchQuery) > 0 {
					j.SearchQuery = j.SearchQuery[:len(j.SearchQuery)-1]
				}
			case "enter":
				j.SearchMode = false
				j.Page = 0
				j.Cursor = 0
				return j, nil
			case "/":
				j.SearchMode = false
				j.SearchQuery = ""
				j.FilteredMissions = nil
				j.Page = 0
				j.Cursor = 0
				return j, nil
			default:
				if len(msg.String()) == 1 {
					j.SearchQuery += msg.String()
				}
			}
			// Recompute filtered missions.
			j.FilteredMissions = nil
			for _, m := range j.Missions {
				if strings.Contains(strings.ToLower(m.Title), strings.ToLower(j.SearchQuery)) ||
					strings.Contains(strings.ToLower(m.Description), strings.ToLower(j.SearchQuery)) {
					j.FilteredMissions = append(j.FilteredMissions, m)
				}
			}
			j.Page = 0
			j.Cursor = 0
			return j, nil
		}

		// -----------------------------
		// Normal Mission List Handling
		// -----------------------------
		currentList := j.currentList()
		totalItems := len(currentList)
		startIndex := j.Page * j.PageSize
		if startIndex > totalItems {
			startIndex = totalItems
		}
		endIndex := startIndex + j.PageSize
		if endIndex > totalItems {
			endIndex = totalItems
		}
		pageItemsCount := endIndex - startIndex

		switch msg.String() {
		case "up", "k":
			if j.Cursor > 0 {
				j.Cursor--
			}
		case "down", "j":
			if j.Cursor < pageItemsCount-1 {
				j.Cursor++
			}
		case "n":
			if endIndex < totalItems {
				j.Page++
				j.Cursor = 0
			}
		case "N":
			if j.Page > 0 {
				j.Page--
				j.Cursor = 0
			}
		case "enter":
			// If this mission is already completed, do nothing
			selectedMission := j.getSelectedMission()
			if selectedMission.Status == data.MissionStatusCompleted {
				break
			}
			// Else enter detail view for this mission
			if pageItemsCount > 0 {
				j.DetailView = true
				j.DetailCursor = 0
			}
		case "/":
			j.SearchMode = true
			j.SearchQuery = ""
			j.FilteredMissions = nil
			j.Page = 0
			j.Cursor = 0
		}
	}
	return j, nil
}

// getSelectedMission returns the selected mission from the current page.
func (j JournalModel) getSelectedMission() data.Mission {
	currentList := j.currentList()
	totalItems := len(currentList)

	// Guard against empty list
	if totalItems == 0 {
		return data.Mission{} // Return empty mission instead of causing a panic
	}

	startIndex := j.Page * j.PageSize
	if startIndex > totalItems {
		startIndex = totalItems
	}
	endIndex := startIndex + j.PageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}
	missionsOnPage := currentList[startIndex:endIndex]

	// // Guard against empty page or cursor out of bounds
	// if len(missionsOnPage) == 0 || j.Cursor >= len(missionsOnPage) {
	// 	return data.Mission{} // Return empty mission instead of causing a panic
	// }

	return missionsOnPage[j.Cursor]
}

// updateMission updates the mission in j.Missions by matching the title.
func (j *JournalModel) updateMission(updated data.Mission) {
	for i, m := range j.Missions {
		if m.Title == updated.Title {
			j.Missions[i] = updated
			break
		}
	}
}

func (j JournalModel) View() string {
	if j.DetailView {
		// display options and full mission details
		currentList := j.currentList()
		totalItems := len(currentList)
		startIndex := j.Page * j.PageSize
		if startIndex > totalItems {
			startIndex = totalItems
		}
		endIndex := startIndex + j.PageSize
		if endIndex > totalItems {
			endIndex = totalItems
		}
		missionsOnPage := currentList[startIndex:endIndex]
		if len(missionsOnPage) == 0 {
			return "No missions found."
		}
		selectedMission := missionsOnPage[j.Cursor]

		// left panel mission list
		var optionsList strings.Builder
		for i, option := range j.DetailOptions {
			if i == j.DetailCursor {
				optionsList.WriteString(fmt.Sprintf("> %s\n", option))
			} else {
				optionsList.WriteString(fmt.Sprintf("  %s\n", option))
			}
		}
		leftPanel := lipgloss.NewStyle().
			Width(60).
			Height(18).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Render(optionsList.String())

		// right panel detailed mission information
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
		labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
		details := fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
			titleStyle.Render(selectedMission.Title),
			labelStyle.Render("Description:")+" "+selectedMission.Description,
			labelStyle.Render("Status:")+" "+selectedMission.Status.String(),
			labelStyle.Render("Location:")+" "+selectedMission.Location.PlanetName,
			labelStyle.Render("Income:")+" "+fmt.Sprintf("%d", selectedMission.Income)+" credits",
			labelStyle.Render("Requirements:")+" "+selectedMission.Requirements,
			labelStyle.Render("Received:")+" "+selectedMission.Received,
			labelStyle.Render("Category:")+" "+selectedMission.Category,
		)
		rightPanel := lipgloss.NewStyle().
			Width(60).
			Height(18).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Render(details)

		return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	}

	// normal list view
	var currentList []data.Mission
	if j.SearchQuery != "" {
		if len(j.FilteredMissions) > 0 {
			currentList = j.FilteredMissions
		} else {
			currentList = []data.Mission{}
		}
	} else {
		currentList = j.Missions
	}
	totalItems := len(currentList)
	startIndex := j.Page * j.PageSize
	if startIndex > totalItems {
		startIndex = totalItems
	}
	endIndex := startIndex + j.PageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}
	missionsOnPage := currentList[startIndex:endIndex]
	totalPages := 1
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(j.PageSize)))
	}

	// left panel mission list
	leftStyle := lipgloss.NewStyle().
		Width(60).
		Height(18).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	hoverStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	subtitleStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

	var missionList strings.Builder
	if j.SearchQuery != "" {
		missionList.WriteString("Search: " + j.SearchQuery + "\n\n")
	}
	for i, mission := range missionsOnPage {
		titleText := mission.Title
		if mission.Status == data.MissionStatusCompleted {
			titleText = titleText + " " + "✓"
		}
		if i == j.Cursor {
			missionList.WriteString(fmt.Sprintf("%s %s\n", arrowStyle.Render(">"), hoverStyle.Render(titleText)))
		} else {
			missionList.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
		}
		missionList.WriteString(fmt.Sprintf("   %s\n", subtitleStyle.Render(mission.Category)))
	}
	pageInfo := fmt.Sprintf("Page %d of %d", j.Page+1, totalPages)
	hints := "[/] search  [n] next page  [N] previous page"
	missionList.WriteString("\n" + pageInfo + "\n" + hintStyle.Render(hints))
	leftPanel := leftStyle.Render(missionList.String())

	// right panel mission details
	rightStyle := lipgloss.NewStyle().
		Width(60).
		Height(18).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)
	var details string
	if len(missionsOnPage) > 0 {
		selectedMission := missionsOnPage[j.Cursor]
		details = fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n%s\n%s",
			titleStyle.Render(selectedMission.Title),
			labelStyle.Render("Description:")+" "+selectedMission.Description,
			labelStyle.Render("Status:")+" "+selectedMission.Status.String(),
			labelStyle.Render("Location:")+" "+selectedMission.Location.PlanetName,
			labelStyle.Render("Income:")+" "+fmt.Sprintf("%d", selectedMission.Income),
			labelStyle.Render("Requirements:")+" "+selectedMission.Requirements,
			labelStyle.Render("Received:")+" "+selectedMission.Received,
		)
	} else {
		details = "No missions found."
	}
	rightPanel := rightStyle.Render(details)

	dividerStr := `
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
│
│
`
	dividerStyle := lipgloss.NewStyle().
		Width(1).
		Height(18).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("240"))
	div := dividerStyle.Render(dividerStr)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, div, rightPanel)
}
