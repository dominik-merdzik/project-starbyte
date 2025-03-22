package model

import (
	"fmt"
	"math/rand"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

type SpaceStationModel struct {
	Ship       data.Ship
	Tabs       []string
	TabContent []string
	ActiveTab  int

	// Fields for refuel flow
	refuelMode    bool
	refuelConfirm bool
	desiredFuel   int
	fuelPrice     int

	// Fields for upgrade
	upgradeCursor  int // Tracks which upgrade is selected
	upgradeConfirm bool

	// Fields for crew member
	GeneratedRecruits []data.CrewMember // Array of procedurally generated options
	RecruitCursor     int               // Tracks selected crew member
	showingCrewDetail bool              // True when crew member popup open
	confirmHire       bool

	// General fields
	Credits      int
	ErrorMessage string // Stores feedback

}

func NewSpaceStationModel(ship data.Ship, credits int) SpaceStationModel {
	model := SpaceStationModel{
		Ship:       ship,
		Credits:    credits,
		Tabs:       []string{"Hire Crew", "Missions", "Upgrade Ship", "Refuel"},
		TabContent: []string{"Hire new crew members.", "Browse available missions.", "Upgrade your ship.", "Refuel before leaving. [Enter]"},
		ActiveTab:  0,
		fuelPrice:  5,
	}

	if model.Tabs[model.ActiveTab] == "Hire Crew" {
		model.GeneratedRecruits = generateRandomRecruits(rand.Intn(5) + 1) // Generates random amount of recruits between 1-5
		model.RecruitCursor = 0
	}

	return model
}

// Global variable for base prices for ship upgrades
var baseUpgradeCosts = []int{100, 200, 300} // Engine, Weapons, Cargo

func (m SpaceStationModel) Init() tea.Cmd {
	return nil
}

// This signals game.go to update the ships fuel
// Used when travelling to a planet
type RefuelUpdateMsg struct {
	Fuel    int
	Credits int
}

// This signals game.go to update the ships upgrades
// Used when upgrading in space station
type UpgradeUpdateMsg struct {
	UpgradeCursor int
	NewLevel      int
	Credits       int
}

// This signals game.go to update the crew members
// Used when hiring new crew member in space station
type HireCrewMsg struct {
	Crew    data.CrewMember
	Credits int
}

func (m SpaceStationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "right", "l":
			if !m.refuelMode {
				m.ActiveTab = min(m.ActiveTab+1, len(m.Tabs)-1)
			}
			return m, nil
		case "left", "h":
			if !m.refuelMode {
				m.ActiveTab = max(m.ActiveTab-1, 0)
			}
			return m, nil
		case "enter":
			if m.Tabs[m.ActiveTab] == "Refuel" {
				if !m.refuelMode {
					// Enter fuel selection mode
					m.refuelMode = true
					m.desiredFuel = 10
				} else if !m.refuelConfirm {
					// Enter confirmation mode
					m.refuelConfirm = true
				} else {
					// Perform refuel
					totalCost := m.desiredFuel * m.fuelPrice

					if m.SpendCredits(totalCost) {
						m.Ship.Fuel = min(m.Ship.Fuel+m.desiredFuel, m.Ship.MaxFuel)

						// Reset UI state
						m.refuelMode = false
						m.refuelConfirm = false
						m.desiredFuel = 0

						return m, tea.Batch(
							func() tea.Msg {
								return RefuelUpdateMsg{
									Fuel:    m.Ship.Fuel,
									Credits: m.Credits,
								}
							},
							func() tea.Msg {
								return tea.KeyMsg{Type: tea.KeyEsc}
							},
						)
					} else {
						// Not enough credits
						m.refuelConfirm = false
						m.desiredFuel = 0
						m.refuelMode = false
						return m, nil
					}
				}
				return m, nil
			}
			if m.Tabs[m.ActiveTab] == "Upgrade Ship" {
				if !m.upgradeConfirm {
					m.upgradeConfirm = true // Confirm mode
				} else {
					// Apply the upgrade
					success := m.ApplyUpgrade(m.upgradeCursor)
					m.upgradeConfirm = false // Exit confirm mode
					if success {
						return m, tea.Batch(
							func() tea.Msg {
								return UpgradeUpdateMsg{
									UpgradeCursor: m.upgradeCursor,
									NewLevel:      m.GetUpgradeLevel(m.upgradeCursor),
									Credits:       m.Credits,
								}
							},
							func() tea.Msg {
								return tea.KeyMsg{Type: tea.KeyEsc}
							},
						)
					}
				}
			}
			if m.Tabs[m.ActiveTab] == "Hire Crew" {
				if !m.showingCrewDetail {
					m.showingCrewDetail = true
				} else if !m.confirmHire {
					m.confirmHire = true
				} else {
					// Confirm the hire
					recruit := m.GeneratedRecruits[m.RecruitCursor]
					cost := getHireCost(recruit.Degree, string(recruit.Role))

					// Not enough credits
					if m.Credits < cost {
						m.confirmHire = false
						m.ErrorMessage = "Not enough credits!"
						return m, nil
					}

					m.Credits -= cost

					// Remove from recruit list
					m.GeneratedRecruits = append(m.GeneratedRecruits[:m.RecruitCursor], m.GeneratedRecruits[m.RecruitCursor+1:]...)
					if m.RecruitCursor > 0 {
						m.RecruitCursor--
					}
					m.confirmHire = false
					m.showingCrewDetail = false

					return m, func() tea.Msg {
						return HireCrewMsg{
							Crew:    recruit,
							Credits: m.Credits,
						}
					}
				}
			}
		case "esc":
			if m.refuelConfirm {
				m.refuelConfirm = false
			} else if m.refuelMode {
				m.refuelMode = false
				m.desiredFuel = 0
			}
			if m.confirmHire {
				m.confirmHire = false
			} else if m.showingCrewDetail {
				m.showingCrewDetail = false
			}
			return m, nil

		case "up", "k":
			// Increase fuel
			if m.refuelMode && !m.refuelConfirm {
				m.desiredFuel = min(m.desiredFuel+1, m.Ship.MaxFuel-m.Ship.Fuel)
			}
			// Higher upgrade in list
			if m.Tabs[m.ActiveTab] == "Upgrade Ship" {
				m.upgradeCursor = max(m.upgradeCursor-1, 0)
				m.ErrorMessage = ""
			}
			// Higher crew member in list
			if m.Tabs[m.ActiveTab] == "Hire Crew" && len(m.GeneratedRecruits) > 0 {
				m.RecruitCursor = max(m.RecruitCursor-1, 0)
			}
			return m, nil

		case "down", "j":
			// Decrease fuel
			if m.refuelMode && !m.refuelConfirm {
				m.desiredFuel = max(m.desiredFuel-1, 1)
			}
			// Lower upgrade in list
			if m.Tabs[m.ActiveTab] == "Upgrade Ship" {
				m.upgradeCursor = min(m.upgradeCursor+1, 2)
				m.ErrorMessage = ""
			}
			// Lower crew member in list
			if m.Tabs[m.ActiveTab] == "Hire Crew" && len(m.GeneratedRecruits) > 0 {
				m.RecruitCursor = min(m.RecruitCursor+1, len(m.GeneratedRecruits)-1)
			}
			return m, nil
		}
	}
	return m, nil
}

// Border styling function
func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

// Styling variables
var (
	highlightColor    = lipgloss.AdaptiveColor{Light: "217", Dark: "215"}
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	warningStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	labelStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
)

func (m SpaceStationModel) View() string {
	doc := strings.Builder{}
	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.ActiveTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	var content string

	// Refuel section
	if m.Tabs[m.ActiveTab] == "Refuel" && m.refuelMode {
		if m.refuelConfirm {
			totalCost := m.desiredFuel * m.fuelPrice
			content = fmt.Sprintf(
				"Confirm refueling %d units?\nCost: %d¢  |  You have: %d¢\n[Enter] Confirm  [Esc] Cancel",
				m.desiredFuel,
				totalCost,
				m.Credits,
			)

			// Add warning if player cannot afford
			if totalCost > m.Credits {
				content += "\n\n" + warningStyle.Render(fmt.Sprintf("\n\nNot enough credits! (%d¢ needed)", totalCost))
			}
		} else {
			content = fmt.Sprintf(
				"How much fuel do you want to buy?\n[ %d ] units (Cost: %d¢)\nYou have: %d¢\n[↑/↓] Adjust  [Enter] Confirm  [Esc] Cancel",
				m.desiredFuel,
				m.desiredFuel*m.fuelPrice,
				m.Credits,
			)
		}
	} else {
		content = m.TabContent[m.ActiveTab]
	}
	// Upgrade section
	if m.Tabs[m.ActiveTab] == "Upgrade Ship" {
		var upgradeList []string
		upgradeNames := []string{"Engine", "Weapon Systems", "Cargo Expansion"}

		for i, name := range upgradeNames {
			level := m.GetUpgradeLevel(i)
			var line string

			// Show maxed out message
			if level >= 5 {
				line = fmt.Sprintf("%s (Lv %d) - MAXED OUT", name, level)
			} else {
				cost := baseUpgradeCosts[i] * (level + 1)
				line = fmt.Sprintf("%s (Lv %d) - Cost: %d¢", name, level, cost)
			}

			// Highlight selected
			if i == m.upgradeCursor {
				line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("215")).Render("> " + line)
			}

			upgradeList = append(upgradeList, line)
		}

		content = strings.Join(upgradeList, "\n")

		// Show confirmation message
		if m.upgradeConfirm {
			level := m.GetUpgradeLevel(m.upgradeCursor)
			if level >= 5 {
				content += "\n\n" + lipgloss.NewStyle().
					Foreground(lipgloss.Color("8")).
					Italic(true).
					Render("This upgrade is already at maximum level.")
			} else {
				content += fmt.Sprintf(
					"\n\nConfirm upgrading %s to Lv %d for %d¢?\n[Enter] Confirm  [Esc] Cancel",
					upgradeNames[m.upgradeCursor],
					level+1,
					baseUpgradeCosts[m.upgradeCursor]*(level+1),
				)
			}
		}

		// Warning message
		content += "\n\n" + warningStyle.Render(m.ErrorMessage)

	}
	// Hire Crew section
	if m.Tabs[m.ActiveTab] == "Hire Crew" {
		if m.Tabs[m.ActiveTab] == "Hire Crew" {
			var recruitLines []string

			// Display list of recruits with name, role and degree
			for i, r := range m.GeneratedRecruits {
				line := fmt.Sprintf("%-10s  %s ~ Degree %d", r.Name, r.Role, r.Degree)

				if i == m.RecruitCursor {
					line = lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.Color("33")).
						Render("> " + line)
				} else {
					line = "  " + line
				}

				recruitLines = append(recruitLines, line)
			}

			content = lipgloss.NewStyle().
				Padding(1, 2).
				Render(strings.Join(recruitLines, "\n"))
		}

		// List of detailed crew member information
		if m.showingCrewDetail && len(m.GeneratedRecruits) > 0 {
			r := m.GeneratedRecruits[m.RecruitCursor]

			lines := []string{
				fmt.Sprintf("%s %s", labelStyle.Render("Name:"), r.Name),
				fmt.Sprintf("%s %s", labelStyle.Render("Role:"), r.Role),
				fmt.Sprintf("%s %d", labelStyle.Render("Degree:"), r.Degree),
				fmt.Sprintf("%s %d", labelStyle.Render("Experience:"), r.Experience),
				fmt.Sprintf("%s %d", labelStyle.Render("Master Work Level:"), r.MasterWorkLevel),
				fmt.Sprintf("%s %d", labelStyle.Render("Morale:"), r.Morale),
				fmt.Sprintf("%s %d", labelStyle.Render("Health:"), r.Health),
				"",
				fmt.Sprintf("%s %s", labelStyle.Render("Buffs:"), r.Buffs),
				fmt.Sprintf("%s %s", labelStyle.Render("Debuffs:"), r.Debuffs),
				fmt.Sprintf("%s %d", labelStyle.Render("Hire Cost:"), getHireCost(r.Degree, string(r.Role))),
				"",
				func() string {
					if m.confirmHire {
						return "[Enter] Confirm Hire    [Esc] Cancel"
					}
					return "[Enter] Hire    [Esc] Cancel"
				}(),
			}
			if m.ErrorMessage != "" {
				lines = append(lines, "", warningStyle.Render(m.ErrorMessage))
			}

			content = strings.Join(lines, "\n")
		}
	}

	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(content))
	return docStyle.Render(doc.String())
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Subtracts from the players credits
func (m *SpaceStationModel) SpendCredits(amount int) bool {
	if m.Credits >= amount {
		m.Credits -= amount
		return true
	}
	return false // Return false if player does not have enough
}

//***************************************
//        Upgrade functions
//***************************************

// Applies upgrades to the ship
func (m *SpaceStationModel) ApplyUpgrade(index int) bool {
	currentLevel := m.GetUpgradeLevel(index)

	if currentLevel >= 5 {
		m.ErrorMessage = "Already maxed out!"
		return false
	}

	// Calc cost
	cost := baseUpgradeCosts[index] * (currentLevel + 1)

	// Check if player has enough credits
	if !m.SpendCredits(cost) {
		m.ErrorMessage = "Not enough credits!"
		return false
	}

	// Apply upgrade
	m.SetUpgradeLevel(index, currentLevel+1)
	return true
}

// Returns the upgrade level for each module
func (m *SpaceStationModel) GetUpgradeLevel(index int) int {
	switch index {
	case 0:
		return m.Ship.Upgrades.Engine.CurrentLevel
	case 1:
		return m.Ship.Upgrades.WeaponSystems.CurrentLevel
	case 2:
		return m.Ship.Upgrades.CargoExpansion.CurrentLevel
	default:
		return 0
	}
}

// Sets the upgrade level for each module
func (m *SpaceStationModel) SetUpgradeLevel(index int, newLevel int) {
	switch index {
	case 0:
		m.Ship.Upgrades.Engine.CurrentLevel = newLevel
	case 1:
		m.Ship.Upgrades.WeaponSystems.CurrentLevel = newLevel
	case 2:
		m.Ship.Upgrades.CargoExpansion.CurrentLevel = newLevel
	}
}

//***************************************
//        Crew functions
//***************************************

// Generates random recruits
func generateRandomRecruits(n int) []data.CrewMember {
	// Random list of names (can make bigger)
	names := []string{
		"Alice", "Bob", "Junko", "Nash", "Kira", "Maeve", "Cass", "Yuri", "Andrew", "Dominik", "Khanh", "Theoren",
		"Vega", "Talon", "Orion", "Lyra", "Zarek", "Nova", "Soren", "Kael", "Vyn", "Thalos", "Riven",
		"Sylari", "Xan", "Astra", "Zyra", "Nyx", "Thrae", "Ilyon", "Serix", "Kalix", "Liora", "Drayen", "Qira",
		"Byte", "Echo", "Glim", "Frax", "Zip", "Lumen", "Jett", "Neon", "Plex", "Rune",
	}

	roles := []string{"Pilot", "Engineer", "Scientist"}

	// Generate n number of recruits
	recruits := make([]data.CrewMember, 0, n)
	for i := 0; i < n; i++ {
		name := names[rand.Intn(len(names))]
		role := roles[rand.Intn(len(roles))]

		degree := 1 // 60% chance of degree 1
		roll := rand.Intn(100)
		if roll > 90 { // 10% chance of degree 3
			degree = 3
		} else if roll > 60 { // 30% chance of degree 2
			degree = 2
		}

		id := fmt.Sprintf("CREW_%06d", rand.Intn(999999))

		recruit := data.CrewMember{
			CrewId:          id,
			Name:            name,
			Role:            data.CrewRole(role),
			Degree:          degree,
			Experience:      0,
			Morale:          100,
			Health:          100,
			MasterWorkLevel: 0,
			Buffs:           []string{},
			Debuffs:         []string{},
			AssignedTaskId:  nil,
		}

		recruits = append(recruits, recruit)
	}

	return recruits
}

// Calculates hire cost of crew member and returns cost
func getHireCost(degree int, role string) int {
	var roleMult int
	switch {
	case role == "Pilot":
		roleMult = 3
	case role == "Engineer":
		roleMult = 2
	case role == "Scientist":
		roleMult = 4
	default:
		roleMult = 1
	}

	return (100 * roleMult) * degree
}
