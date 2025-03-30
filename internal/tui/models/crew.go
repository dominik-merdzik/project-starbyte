package model

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// CrewMember represents a single crew member in our internal model
type CrewMember struct {
	CrewId          string
	Name            string
	Role            string
	Degree          int
	Experience      int
	Morale          int
	Health          int
	MasterWorkLevel int      // acts as a prestige level after level 10
	HireCost        int      // used when recruiting crew members
	Buffs           []string // new: list of buffs
	Debuffs         []string // new: list of debuffs

	GameSave *data.FullGameSave
}

// CrewUpdateMsg signals that a crew upgrade has occurred and the model should update
type CrewUpdateMsg struct{}

// CrewModel contains all crew on board the player's ship and handles modal states
type CrewModel struct {
	CrewMembers         []CrewMember
	Cursor              int
	PopupActive         bool               // whether a modal is open
	PopupState          string             // "main", "research", or "receipt"
	PopupOptions        []string           // options in the main modal
	PopupCursor         int                // cursor for main modal selection
	ResearchPopupCursor int                // cursor for research notes selection
	ResearchUseCount    int                // how many research notes to use
	ReceiptMessage      string             // message to show after applying research
	GameSave            *data.FullGameSave // Updating crew list after hiring new member
}

func (c CrewModel) Init() tea.Cmd {
	return nil
}

// NewCrewModel creates a new CrewModel based on saved crew data
func NewCrewModel(savedCrew []data.CrewMember, save *data.FullGameSave) CrewModel {
	var crew []CrewMember
	for _, s := range savedCrew {
		crew = append(crew, CrewMember{
			CrewId:          s.CrewId,
			Name:            s.Name,
			Role:            string(s.Role),
			Degree:          s.Degree,
			Experience:      s.Experience,
			Morale:          s.Morale,
			Health:          s.Health,
			MasterWorkLevel: s.MasterWorkLevel,
			HireCost:        100, // placeholder value
			Buffs:           s.Buffs,
			Debuffs:         s.Debuffs,
			GameSave:        save,
		})
	}
	return CrewModel{
		CrewMembers:         crew,
		Cursor:              0,
		PopupActive:         false,
		PopupState:          "",
		PopupOptions:        nil,
		PopupCursor:         0,
		ResearchPopupCursor: 0,
		ResearchUseCount:    0,
		ReceiptMessage:      "",
		GameSave:            save,
	}
}

func (c CrewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// if a modal is active, handle its input exclusively
		if c.PopupActive {
			// in the receipt state, any key dismisses the modal
			if c.PopupState == "receipt" {
				c.PopupActive = false
				c.PopupState = ""
				c.PopupOptions = nil
				c.PopupCursor = 0
				c.ResearchPopupCursor = 0
				c.ResearchUseCount = 0
				return c, nil
			}

			switch c.PopupState {
			case "main":
				switch msg.String() {
				case "up", "k":
					if c.PopupCursor > 0 {
						c.PopupCursor--
					}
				case "down", "j":
					if c.PopupCursor < len(c.GameSave.Crew)-1 {
						c.PopupCursor++
					}
				case "enter":
					selectedOption := c.PopupOptions[c.PopupCursor]
					if selectedOption == "Do Research" {
						// transition to research state and initialize the use count
						c.PopupState = "research"
						c.ResearchPopupCursor = 0
						c.ResearchUseCount = 1
					} else if selectedOption == "Back" {
						// close the modal
						c.PopupActive = false
						c.PopupState = ""
						c.PopupOptions = nil
						c.PopupCursor = 0
					}
				case "b":
					// close the modal
					c.PopupActive = false
					c.PopupState = ""
					c.PopupOptions = nil
					c.PopupCursor = 0
				}
			case "research":
				// build list of available research notes (quantity > 0)
				availableNotes := []data.ResearchNoteTier{}
				if c.GameSave != nil {
					for _, note := range c.GameSave.Collection.ResearchNotes {
						if note.Quantity > 0 {
							availableNotes = append(availableNotes, note)
						}
					}
				}
				if len(availableNotes) == 0 {
					// no research notes available; close the modal
					c.PopupActive = false
					c.PopupState = ""
					c.PopupOptions = nil
					c.PopupCursor = 0
					return c, nil
				}
				switch msg.String() {
				case "up", "k":
					if c.ResearchPopupCursor > 0 {
						c.ResearchPopupCursor--
						// reset use count when switching note
						c.ResearchUseCount = 1
					}
				case "down", "j":
					if c.ResearchPopupCursor < len(availableNotes)-1 {
						c.ResearchPopupCursor++
						c.ResearchUseCount = 1
					}
				case "left", "h":
					if c.ResearchUseCount > 1 {
						c.ResearchUseCount--
					}
				case "right", "l":
					// increase use count but do not exceed available quantity
					if c.ResearchUseCount < availableNotes[c.ResearchPopupCursor].Quantity {
						c.ResearchUseCount++
					}
				case "enter":
					// apply the selected number of research notes
					initialDegree := c.GameSave.Crew[c.Cursor].Degree
					useCount := c.ResearchUseCount
					available := availableNotes[c.ResearchPopupCursor].Quantity
					if useCount > available {
						useCount = available
					}
					c.GameSave.Crew[c.Cursor].Degree += useCount
					// update the game save
					if c.GameSave != nil {
						// update research note quantity
						for i, note := range c.GameSave.Collection.ResearchNotes {
							if note.Tier == availableNotes[c.ResearchPopupCursor].Tier {
								c.GameSave.Collection.ResearchNotes[i].Quantity -= useCount
								break
							}
						}
						// update the crew member in the game save
						var dataCrew *data.CrewMember
						for i, savedCrew := range c.GameSave.Crew {
							if savedCrew.CrewId == c.GameSave.Crew[c.Cursor].CrewId {
								dataCrew = &c.GameSave.Crew[i]
								// mirror the new degree
								c.GameSave.Crew[i].Degree = c.GameSave.Crew[c.Cursor].Degree
								break
							}
						}
						// award modifiers if thresholds were crossed
						modifierReceipt := ""
						if dataCrew != nil {
							modifierReceipt = data.AwardModifier(dataCrew, initialDegree, c.GameSave.Crew[c.Cursor].Degree)
							// sync the internal model's Buffs/Debuffs with the saved crew
							c.GameSave.Crew[c.Cursor].Buffs = dataCrew.Buffs
							c.GameSave.Crew[c.Cursor].Debuffs = dataCrew.Debuffs
						}
						c.ReceiptMessage = fmt.Sprintf("Degree %d → %d\n%s", initialDegree, c.GameSave.Crew[c.Cursor].Degree, modifierReceipt)
					} else {
						c.ReceiptMessage = fmt.Sprintf("Degree %d → %d", initialDegree, c.GameSave.Crew[c.Cursor].Degree)
					}
					c.PopupState = "receipt"
				case "b":
					c.PopupState = "main"
				}
			}
			return c, nil
		} else {
			// no modal active – normal crew navigation
			switch msg.String() {
			case "up", "k":
				if c.Cursor > 0 {
					c.Cursor--
				}
			case "down", "j":
				if c.Cursor < len(c.GameSave.Crew)-1 {
					c.Cursor++
				}
			case "enter":
				// open the modal for the currently selected crew member
				c.PopupActive = true
				c.PopupState = "main"
				c.PopupOptions = []string{"Do Research", "Back"}
				c.PopupCursor = 0
			}
		}
	}
	return c, nil
}

// RenderCrewModal renders a modal overlay similar to map.go's travel confirmation
func (c CrewModel) renderCrewModal() string {
	var style lipgloss.Style
	if c.PopupState == "receipt" {
		style = lipgloss.NewStyle().
			Width(60).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Align(lipgloss.Center)
		return style.Render(fmt.Sprintf("%s\n\nPress any key to continue", c.ReceiptMessage))
	}
	// for main and research states
	style = lipgloss.NewStyle().
		Width(60).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Center)

	var modalContent strings.Builder
	if c.PopupState == "main" {
		modalContent.WriteString(lipgloss.NewStyle().Bold(true).Render("Crew Action") + "\n\n")
		for i, option := range c.PopupOptions {
			if i == c.PopupCursor {
				modalContent.WriteString(fmt.Sprintf("> %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("215")).Render(option)))
			} else {
				modalContent.WriteString(fmt.Sprintf("  %s\n", option))
			}
		}
	} else if c.PopupState == "research" {
		modalContent.WriteString(lipgloss.NewStyle().Bold(true).Render("Select Research Note") + "\n")
		modalContent.WriteString("(Adjust quantity with ←/→ keys)\n\n")
		availableNotes := []data.ResearchNoteTier{}
		if c.GameSave != nil {
			for _, note := range c.GameSave.Collection.ResearchNotes {
				if note.Quantity > 0 {
					availableNotes = append(availableNotes, note)
				}
			}
		}
		if len(availableNotes) == 0 {
			modalContent.WriteString("No research notes available.")
		} else {
			for i, note := range availableNotes {
				noteStr := fmt.Sprintf("%s (Qty: %d)", note.Name, note.Quantity)
				if i == c.ResearchPopupCursor {
					// append current use count for the selected note
					noteStr += fmt.Sprintf("  | Use: %d", c.ResearchUseCount)
					modalContent.WriteString(fmt.Sprintf("> %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("215")).Render(noteStr)))
				} else {
					modalContent.WriteString(fmt.Sprintf("  %s\n", noteStr))
				}
			}
		}
	}
	return style.Render(modalContent.String())
}

func (c CrewModel) View() string {
	if c.PopupActive {
		return c.renderCrewModal()
	}

	// left panel: render each crew member in their own container
	var crewContainers []string
	for i, crew := range c.GameSave.Crew {
		var containerStyle lipgloss.Style
		if i == c.Cursor {
			containerStyle = lipgloss.NewStyle().
				Width(57).
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("205"))
		} else {
			containerStyle = lipgloss.NewStyle().
				Width(57).
				Padding(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("15"))
		}

		containerText := lipgloss.NewStyle().Bold(true).Render(crew.Name) + "\n" +
			fmt.Sprintf("%s ~ Degree %d", crew.Role, crew.Degree)
		crewContainers = append(crewContainers, containerStyle.Render(containerText))
	}
	leftPanel := lipgloss.JoinVertical(lipgloss.Top, crewContainers...)

	// right panel: details of the selected crew member with increased height
	rightStyle := lipgloss.NewStyle().
		Width(57).
		Height(20).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	labelStyle := lipgloss.NewStyle().Bold(true)

	var crewDetails strings.Builder
	if len(c.GameSave.Crew) > 0 {
		crew := c.GameSave.Crew[c.Cursor]
		crewDetails.WriteString(titleStyle.Render(crew.Name) + "\n")
		crewDetails.WriteString(labelStyle.Render("Role: ") + string(crew.Role) + "\n")
		crewDetails.WriteString(labelStyle.Render("Degree: ") + fmt.Sprintf("%d", crew.Degree) + "\n")
		crewDetails.WriteString(labelStyle.Render("Experience: ") + fmt.Sprintf("%d", crew.Experience) + "\n")
		crewDetails.WriteString(labelStyle.Render("Master Work Level: ") + fmt.Sprintf("%d", crew.MasterWorkLevel) + "\n")
		crewDetails.WriteString(labelStyle.Render("Morale: ") + fmt.Sprintf("%d", crew.Morale) + "\n")
		crewDetails.WriteString(labelStyle.Render("Health: ") + fmt.Sprintf("%d", crew.Health) + "\n\n")

		// aggregate Buffs
		crewDetails.WriteString(labelStyle.Render("Buffs:") + "\n")
		buffCounts := make(map[string]int)
		for _, buff := range crew.Buffs {
			buffCounts[buff]++
		}
		if len(buffCounts) > 0 {
			type countEntry struct {
				key   string
				count int
			}
			var buffEntries []countEntry
			for key, count := range buffCounts {
				buffEntries = append(buffEntries, countEntry{key, count})
			}
			// sort descending by count, and alphabetically as tiebreaker
			sort.Slice(buffEntries, func(i, j int) bool {
				if buffEntries[i].count == buffEntries[j].count {
					return buffEntries[i].key < buffEntries[j].key
				}
				return buffEntries[i].count > buffEntries[j].count
			})
			for _, entry := range buffEntries {
				crewDetails.WriteString(fmt.Sprintf("  ■ %s x %d\n", entry.key, entry.count))
			}
		} else {
			crewDetails.WriteString("  None\n")
		}

		// aggregate Debuffs
		crewDetails.WriteString("\n" + labelStyle.Render("Debuffs:") + "\n")
		debuffCounts := make(map[string]int)
		for _, debuff := range crew.Debuffs {
			debuffCounts[debuff]++
		}
		if len(debuffCounts) > 0 {
			type countEntry struct {
				key   string
				count int
			}
			var debuffEntries []countEntry
			for key, count := range debuffCounts {
				debuffEntries = append(debuffEntries, countEntry{key, count})
			}
			// sort descending by count, and alphabetically as tiebreaker
			sort.Slice(debuffEntries, func(i, j int) bool {
				if debuffEntries[i].count == debuffEntries[j].count {
					return debuffEntries[i].key < debuffEntries[j].key
				}
				return debuffEntries[i].count > debuffEntries[j].count
			})
			for _, entry := range debuffEntries {
				crewDetails.WriteString(fmt.Sprintf("  ■ %s x %d\n", entry.key, entry.count))
			}
		} else {
			crewDetails.WriteString("  None\n")
		}

	}

	rightPanel := rightStyle.Render(crewDetails.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}
