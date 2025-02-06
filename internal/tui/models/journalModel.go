package model

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mission struct {
	Title        string
	Description  string
	Status       string
	Location     string
	Income       string
	Requirements string
	Received     string
	Category     string
}

// JournalModel represents the state and behavior of our Journal component.
type JournalModel struct {
	Missions         []Mission // full list of missions
	Cursor           int       // index of the currently selected mission in the current page list
	SearchMode       bool      // whether search mode is active (typing a query)
	SearchQuery      string    // current search query text
	FilteredMissions []Mission // missions filtered by the search query
	Page             int       // current page index (0-based)
	PageSize         int       // number of items per page
}

// NewJournalModel initializes and returns a new JournalModel - we will store this in a file later**
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
			Category:     "Task",
		},
		{
			Title:        "Nebula Exploration",
			Description:  "Explore the forbidden nebula to gather scientific data and rare minerals.",
			Status:       "In Progress",
			Location:     "Orion Nebula",
			Income:       "8000 Credits",
			Requirements: "Scientific Vessel",
			Received:     "Galactic Federation",
			Category:     "Exploration",
		},
		{
			Title:        "Artifact Recovery",
			Description:  "Retrieve the ancient artifact from the ruins of a lost civilization.",
			Status:       "Completed",
			Location:     "Xandar",
			Income:       "12000 Credits",
			Requirements: "Archaeological Tools",
			Received:     "Mysterious Benefactor",
			Category:     "Retrieve",
		},
		{
			Title:        "Deep Space Survey",
			Description:  "Conduct a detailed survey of deep space regions for unknown phenomena.",
			Status:       "Scheduled",
			Location:     "Alpha Centauri",
			Income:       "6000 Credits",
			Requirements: "Survey Drone",
			Received:     "Research Lab",
			Category:     "Mission",
		},
		{
			Title:        "Cosmic Anomaly Investigation",
			Description:  "Investigate unusual cosmic anomalies detected near a distant galaxy.",
			Status:       "Pending",
			Location:     "Andromeda",
			Income:       "7000 Credits",
			Requirements: "Advanced Sensors",
			Received:     "Space Agency",
			Category:     "Investigation",
		},
		{
			Title:        "Solar Flare Response",
			Description:  "Monitor and respond to unpredictable solar flare activities.",
			Status:       "In Progress",
			Location:     "Sun",
			Income:       "4000 Credits",
			Requirements: "Shielded Satellite",
			Received:     "Energy Commission",
			Category:     "Task",
		},
		{
			Title:        "Black Hole Monitoring",
			Description:  "Observe the behavior of a nearby black hole and its effects on space-time.",
			Status:       "Scheduled",
			Location:     "Sagittarius A*",
			Income:       "9000 Credits",
			Requirements: "High-Resolution Telescope",
			Received:     "Astro Research Institute",
			Category:     "Monitoring",
		},
		{
			Title:        "Alien Artifact Analysis",
			Description:  "Examine recovered alien artifacts for technological insights.",
			Status:       "Completed",
			Location:     "Lunar Base",
			Income:       "11000 Credits",
			Requirements: "X-Ray Scanner",
			Received:     "Archaeology Dept.",
			Category:     "Analysis",
		},
		{
			Title:        "Interstellar Diplomatic Mission",
			Description:  "Establish diplomatic relations with a newly discovered alien civilization.",
			Status:       "Pending",
			Location:     "Proxima Centauri",
			Income:       "15000 Credits",
			Requirements: "Diplomatic Credentials",
			Received:     "United Earth Council",
			Category:     "Diplomacy",
		},
		{
			Title:        "Quantum Rift Study",
			Description:  "Study the mysterious quantum rift found in deep space.",
			Status:       "In Progress",
			Location:     "Draco Constellation",
			Income:       "13000 Credits",
			Requirements: "Quantum Analyzer",
			Received:     "Quantum Lab",
			Category:     "Study",
		},
		{
			Title:        "Wormhole Stabilization",
			Description:  "Investigate and attempt to stabilize a naturally occurring wormhole.",
			Status:       "Scheduled",
			Location:     "Orion Arm",
			Income:       "14000 Credits",
			Requirements: "Stabilization Module",
			Received:     "Space Federation",
			Category:     "Stabilization",
		},
		{
			Title:        "Meteor Shower Defense",
			Description:  "Deploy defenses against a predicted intense meteor shower.",
			Status:       "Completed",
			Location:     "Earth Orbit",
			Income:       "8000 Credits",
			Requirements: "Defensive Array",
			Received:     "Military Command",
			Category:     "Defense",
		},
		{
			Title:        "Stellar Cartography",
			Description:  "Map uncharted star systems and create new navigation charts.",
			Status:       "Completed",
			Location:     "Milky Way",
			Income:       "5000 Credits",
			Requirements: "Advanced Mapping Software",
			Received:     "Navigation Bureau",
			Category:     "Mapping",
		},
		{
			Title:        "Galactic Trade Negotiation",
			Description:  "Negotiate new trade routes and terms with intergalactic partners.",
			Status:       "Pending",
			Location:     "Andromeda",
			Income:       "16000 Credits",
			Requirements: "Trade Agreement",
			Received:     "Economic Council",
			Category:     "Negotiation",
		},
		{
			Title:        "Exoplanet Colonization",
			Description:  "Prepare a detailed plan for colonizing a promising exoplanet.",
			Status:       "Scheduled",
			Location:     "Kepler-452b",
			Income:       "20000 Credits",
			Requirements: "Colonization Fleet",
			Received:     "Colonial Office",
			Category:     "Colonization",
		},
	}

	return JournalModel{
		Missions: missions,
		Cursor:   0,
		Page:     0,
		PageSize: 5,
	}
}

func (j JournalModel) Init() tea.Cmd {
	return nil
}

func (j JournalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// when in search mode, process search input
		if j.SearchMode {
			switch msg.String() {
			case "backspace":
				if len(j.SearchQuery) > 0 {
					j.SearchQuery = j.SearchQuery[:len(j.SearchQuery)-1]
				}
			case "enter":
				// finalize search: exit search mode but keep results
				j.SearchMode = false
				j.Page = 0
				j.Cursor = 0
				return j, nil
			case "esc":
				// cancel search: clear query and results
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
			// recompute filtered missions live
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

		// when not in search mode, determine the current list
		var currentList []Mission
		if j.SearchQuery != "" && len(j.FilteredMissions) > 0 {
			currentList = j.FilteredMissions
		} else {
			currentList = j.Missions
		}
		totalItems := len(currentList)
		startIndex := j.Page * j.PageSize
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
		case "/":
			// activate search mode - all of these key's are tailored to vim ATM
			j.SearchMode = true
			j.SearchQuery = ""
			j.FilteredMissions = nil
			j.Page = 0
			j.Cursor = 0
		}
	}
	return j, nil
}

func (j JournalModel) View() string {
	// determine which mission list to display
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

	// ----------------------------
	// Left Panel: Mission List with Titles, Subtitles (Category), and Page Info
	// ----------------------------
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
	// loop through the missions on the current page
	for i, mission := range missionsOnPage {
		// if the mission is completed, append a check mark
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
		// subtitle (category) on the next line
		missionList.WriteString(fmt.Sprintf("   %s\n", subtitleStyle.Render(mission.Category)))
	}
	// page indicator and hints
	pageInfo := fmt.Sprintf("Page %d of %d", j.Page+1, totalPages)
	hints := "[/] search  [n] next page  [N] previous page"
	missionList.WriteString("\n" + pageInfo + "\n" + hintStyle.Render(hints))
	leftPanel := leftStyle.Render(missionList.String())

	// ----------------------------
	// Right Panel: Mission Details
	// ----------------------------
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
			labelStyle.Render("Income:")+" "+selectedMission.Income,
			labelStyle.Render("Requirements:")+" "+selectedMission.Requirements,
			labelStyle.Render("Received:")+" "+selectedMission.Received,
		)
	} else {
		details = "No missions found."
	}
	rightPanel := rightStyle.Render(details)

	// ----------------------------
	// Vertical Divider.
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
		Height(18).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("240"))
	div := dividerStyle.Render(divider)

	// ----------------------------
	// Combine Panels.
	// ----------------------------
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, div, rightPanel)
}

