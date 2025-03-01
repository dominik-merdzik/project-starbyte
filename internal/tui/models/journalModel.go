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

type TrackMissionMsg struct {
	Mission Mission
}

// Mission represents a mission in the model.
// Note: This is a different type from data.Mission and data.MainMission
type Mission struct {
	Title             string
	Description       string
	Status            string
	Location          string
	Income            int
	Requirements      string
	Received          string
	Category          string
	TravelTime        int
	FuelNeeded        int
	DestinationPlanet string
}

type JournalModel struct {
	Missions         []Mission // full list of missions
	Cursor           int
	SearchMode       bool
	SearchQuery      string
	FilteredMissions []Mission
	Page             int
	PageSize         int

	// Detail view fields
	DetailView    bool
	DetailCursor  int
	DetailOptions []string
}

// convertDataMission converts a data.Mission (from a received mission) into a model.Mission
func convertDataMission(dm data.Mission) Mission {
	return Mission{
		Title:             dm.Title,
		Description:       dm.Description,
		Status:            dm.Status,
		Location:          dm.Location,
		Income:            dm.Income,
		Requirements:      dm.Requirements,
		Received:          dm.Received,
		Category:          dm.Category,
		TravelTime:        dm.TravelTime,
		FuelNeeded:        dm.FuelNeeded,
		DestinationPlanet: dm.DestinationPlanet,
	}
}

// convertMainMission converts a data.MainMission into a model.Mission
func convertMainMission(mm data.MainMission) Mission {
	return Mission{
		Title:             fmt.Sprintf("Step %d: %s", mm.Step, mm.Title),
		Description:       mm.Description,
		Status:            mm.Status,
		Location:          mm.Location,
		Income:            mm.Income,
		Requirements:      mm.Requirements,
		Received:          mm.Received,
		Category:          mm.Category,
		TravelTime:        mm.TravelTime,
		FuelNeeded:        mm.FuelNeeded,
		DestinationPlanet: mm.DestinationPlanet,
	}
}

// currentList returns the missions to be displayed based on search mode
func (j JournalModel) currentList() []Mission {
	if j.SearchQuery != "" {
		if len(j.FilteredMissions) > 0 {
			return j.FilteredMissions
		}
		return []Mission{}
	}
	return j.Missions
}

func NewJournalModel() JournalModel {
	missionsFile, err := data.LoadMissions()
	if err != nil {
		log.Printf("Error loading missions JSON: %v", err)
	}

	var missions []Mission

	// add main missions.
	for _, mm := range missionsFile.Main {
		missions = append(missions, convertMainMission(mm))
	}

	// add all received missions (for testing, later it will only be missions the player has received)
	for _, loc := range missionsFile.Received {
		for _, npc := range loc.NPCs {
			for _, m := range npc.Missions {
				rec := convertDataMission(m)
				// if location is empty, fill it with the parent's location (for now)
				if rec.Location == "" {
					rec.Location = loc.Location
				}
				missions = append(missions, rec)
			}
		}
	}

	return JournalModel{
		Missions:      missions,
		Cursor:        0,
		Page:          0,
		PageSize:      5,
		DetailView:    false,
		DetailCursor:  0,
		DetailOptions: []string{"Track", "Start Mission", "Abandon", "Back"},
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
				// simulate a back action
				j.DetailView = false
			case "enter":
				// execute the selected option
				selectedOption := j.DetailOptions[j.DetailCursor]
				switch selectedOption {
				case "Back":
					j.DetailView = false
				case "Track":
					// track the mission in the background and exit the detail view
					trackedMission := j.getSelectedMission()
					j.DetailView = false
					return j, func() tea.Msg {
						return TrackMissionMsg{Mission: trackedMission}
					}
				case "Open in Map":
					// TODO: Add map-opening functionality.
					fmt.Println("Opening mission in map...")
				case "Assign to Crew Member":
					// TODO: Add crew assignment functionality.
					fmt.Println("Assigning mission to a crew member...")
				case "Abandon":
					// for demonstration, mark the mission as "Abandoned".
					mission := j.getSelectedMission()
					mission.Status = "Abandoned"
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
				// Finalize search.
				j.SearchMode = false
				j.Page = 0
				j.Cursor = 0
				return j, nil
			case "esc":
				// Cancel search.
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
			// Enter detail view if there is at least one item.
			if pageItemsCount > 0 {
				j.DetailView = true
				j.DetailCursor = 0
			}
		case "/":
			// Activate search mode.
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
func (j JournalModel) getSelectedMission() Mission {
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
	return missionsOnPage[j.Cursor]
}

// updateMission updates the mission in j.Missions by matching the title.
func (j *JournalModel) updateMission(updated Mission) {
	for i, m := range j.Missions {
		if m.Title == updated.Title {
			j.Missions[i] = updated
			break
		}
	}
}

func (j JournalModel) View() string {
	if j.DetailView {
		// detail view: use the current list based on search
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

		// left Panel: mission options
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

		// right panel: Detailed mission information
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
		details := fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s",
			titleStyle.Render(selectedMission.Title),
			labelStyle.Render("Description:")+" "+selectedMission.Description,
			labelStyle.Render("Status:")+" "+selectedMission.Status,
			labelStyle.Render("Location:")+" "+selectedMission.Location,
			labelStyle.Render("Income:")+" "+fmt.Sprintf("%d", selectedMission.Income)+" credits",
			labelStyle.Render("Requirements:")+" "+selectedMission.Requirements,
			labelStyle.Render("Received:")+" "+selectedMission.Received,
			labelStyle.Render("Category:")+" "+selectedMission.Category,
			labelStyle.Render("Travel Time:")+" "+fmt.Sprintf("%d", selectedMission.TravelTime)+" minutes",
			labelStyle.Render("Fuel Needed:")+" "+fmt.Sprintf("%d", selectedMission.FuelNeeded)+" units",
			labelStyle.Render("Destination:")+" "+selectedMission.DestinationPlanet,
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
	var currentList []Mission
	if j.SearchQuery != "" {
		if len(j.FilteredMissions) > 0 {
			currentList = j.FilteredMissions
		} else {
			currentList = []Mission{}
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

	// left panel: mission list
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
	if j.SearchMode {
		missionList.WriteString("Search: " + j.SearchQuery + "\n\n")
	}
	for i, mission := range missionsOnPage {
		titleText := mission.Title
		if strings.ToLower(mission.Status) == "completed" {
			titleText = titleText + " " + "✓"
		}
		if i == j.Cursor {
			missionList.WriteString(fmt.Sprintf("%s %s\n",
				arrowStyle.Render(">"),
				hoverStyle.Render(titleText)))
		} else {
			missionList.WriteString(fmt.Sprintf("  %s\n", defaultStyle.Render(titleText)))
		}
		missionList.WriteString(fmt.Sprintf("   %s\n", subtitleStyle.Render(mission.Category)))
	}
	pageInfo := fmt.Sprintf("Page %d of %d", j.Page+1, totalPages)
	hints := "[/] search  [n] next page  [N] previous page"
	missionList.WriteString("\n" + pageInfo + "\n" + hintStyle.Render(hints))
	leftPanel := leftStyle.Render(missionList.String())

	// right panel: mission details
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
			labelStyle.Render("Status:")+" "+selectedMission.Status,
			labelStyle.Render("Location:")+" "+selectedMission.Location,
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
`
	dividerStyle := lipgloss.NewStyle().
		Width(1).
		Height(18).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("240"))
	div := dividerStyle.Render(dividerStr)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, div, rightPanel)
}
