package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// ShipModel represents the ship's status and components.
type ShipModel struct {
	ShipID            string
	Name              string
	HullHealth        int
	MaxHullHealth     int
	EngineHealth      int
	EngineFuel        int
	HasFTLDrive       bool
	FTLDriveHealth    int
	FTLDriveCharge    int
	ShieldStrength    int
	MaxShieldStrength int
	MaxFuel           int
	Crew              []CrewMember
	Food              int
	Location          data.Location
	Cargo             data.Cargo
	Modules           []data.Module
	Upgrades          data.Upgrades
	Cursor            int

	GameSave *data.FullGameSave
}

// NewShipModel creates a new ShipModel based on saved ship data.
func NewShipModel(savedShip data.Ship) ShipModel {
	return ShipModel{
		ShipID:            savedShip.ShipId,
		Name:              savedShip.ShipName,
		HullHealth:        savedShip.HullIntegrity,
		MaxHullHealth:     savedShip.MaxHullIntegrity,
		EngineHealth:      savedShip.EngineHealth,
		EngineFuel:        savedShip.Fuel,
		HasFTLDrive:       savedShip.HasFTLDrive,
		FTLDriveHealth:    savedShip.FTLDriveHealth,
		FTLDriveCharge:    savedShip.FTLDriveCharge,
		ShieldStrength:    savedShip.ShieldStrength,
		MaxShieldStrength: savedShip.MaxShieldStrength,
		MaxFuel:           savedShip.MaxFuel,
		Crew:              []CrewMember{},
		Food:              savedShip.Food,
		Location:          savedShip.Location,
		Cargo:             savedShip.Cargo,
		Modules:           savedShip.Modules,
		Upgrades:          savedShip.Upgrades,
		Cursor:            0,
	}
}

func (s ShipModel) Init() tea.Cmd {
	return nil
}

// Update handles key inputs.
func (s ShipModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.Cursor > 0 {
				s.Cursor--
			}
		case "down", "j":
			if s.Cursor < 5 { // Number of selectable items.
				s.Cursor++
			}
		case "a":
			// subtract 10 units of EngineFuel.
			s.EngineFuel -= 10
			if s.EngineFuel < 0 {
				s.EngineFuel = 0
			}
		case "d":
			// add 5 units of EngineFuel.
			s.EngineFuel += 5
			if s.EngineFuel > s.MaxFuel {
				s.EngineFuel = s.MaxFuel
			}
		}
	}
	return s, nil
}

func (s ShipModel) View() string {
	// ----- Styles -----
	topPanelWidth := 65
	topPanelHeight := 18
	bottomPanelHeight := 9

	// style for the top panels (List, Details)
	topPanelStyle := lipgloss.NewStyle().
		Width(topPanelWidth).
		Height(topPanelHeight).
		Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	// style for titles within panels
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	// other styles (label, default, arrow) - remain the same
	labelStyle := lipgloss.NewStyle().Bold(true)
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("247"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	// calculate the required width for the bottom panel using GetWidth()
	// this correctly accounts for the top panel's width, padding, and border
	bottomPanelWidth := 132

	// style for the bottom panel (Extra Details)
	bottomPanelStyle := topPanelStyle.Copy().
		Width(bottomPanelWidth).
		Height(bottomPanelHeight).
		Padding(1, 2).
		Border(lipgloss.HiddenBorder())

	// --- Calculate dimensions for the small boxes INSIDE the bottom panel ---
	// get the actual content area width/height by subtracting padding & borders
	bottomPanelContentWidth := bottomPanelWidth - (bottomPanelStyle.GetPaddingLeft() + bottomPanelStyle.GetPaddingRight()) - 10
	bottomPanelContentHeight := bottomPanelHeight - (bottomPanelStyle.GetPaddingTop() + bottomPanelStyle.GetPaddingBottom()) - 2

	// reintroduce spacing between boxes for better separation
	spacingWidth := 2
	numberOfBoxes := 4
	totalSpacingWidth := spacingWidth * (numberOfBoxes - 1)

	// calculate the width for each box
	detailBoxWidth := (bottomPanelContentWidth - totalSpacingWidth) / numberOfBoxes

	// boxes take the full content height of the bottom panel
	detailBoxHeight := bottomPanelContentHeight

	// style for the small detail boxes within the bottom panel
	detailBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(detailBoxWidth).
		Height(detailBoxHeight).
		Align(lipgloss.Left)

	// style for titles within the detail boxes
	boxTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	// style for the main text within the detail boxes
	boxTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	descWidth := topPanelWidth - (topPanelStyle.GetPaddingLeft() + topPanelStyle.GetPaddingRight())
	descriptionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Width(descWidth).
		MarginTop(1)

	// ----- Panel 1: Ship Status List -----
	items := []string{"Hull Health", "Engine Health", "Engine Fuel", "FTL Drive Health", "FTL Drive Charge", "Food"}
	var shipList strings.Builder
	shipList.WriteString(titleStyle.Render("Ship Status") + "\n")
	for i, item := range items {
		cursor := "  "
		style := defaultStyle
		if i == s.Cursor {
			cursor = arrowStyle.Render("> ")
			style = defaultStyle.Copy().Bold(true).Foreground(lipgloss.Color("229"))
		}
		shipList.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(item)))
	}
	leftPanel := topPanelStyle.Render(shipList.String())

	// ----- Panel 2: Selected Detail -----
	var details strings.Builder
	var progressValue float64
	var detailTitle string
	var description string

	switch s.Cursor {
	case 0:
		detailTitle = "Hull Health"
		progressValue = float64(s.HullHealth) / float64(s.MaxHullHealth)
		details.WriteString(fmt.Sprintf("%s %d / %d", labelStyle.Render("Integrity:"), s.HullHealth, s.MaxHullHealth))
		// added Description:
		description = "Overall structural integrity. Protects internal systems and crew from environmental hazards and combat damage. Reaching zero integrity results in ship destruction."

	case 1:
		detailTitle = "Engine Health"
		progressValue = float64(s.EngineHealth) / 100.0
		details.WriteString(fmt.Sprintf("%s %d%%", labelStyle.Render("Status:"), s.EngineHealth))
		// added Description:
		description = "Condition of the main sublight engines. Affects maximum speed, maneuverability, and fuel consumption within star systems. Low health risks critical failure."

	case 2:
		detailTitle = "Engine Fuel"
		progressValue = float64(s.EngineFuel) / float64(s.MaxFuel)
		details.WriteString(fmt.Sprintf("%s %d / %d", labelStyle.Render("Level:"), s.EngineFuel, s.MaxFuel))
		// added Description:
		description = "Propellant reserves for sublight travel. Essential for moving between planets, stations, and jump points. Depletion will leave the ship stranded."

	case 3:
		detailTitle = "FTL Drive Health"
		progressValue = float64(s.FTLDriveHealth) / 100.0
		details.WriteString(fmt.Sprintf("%s %d%%", labelStyle.Render("Status:"), s.FTLDriveHealth))
		// added Description:
		description = "Integrity of the Faster-Than-Light drive system. Required for interstellar jumps. Damage increases charge time, jump inaccuracy, or may prevent jumps entirely."

	case 4:
		detailTitle = "FTL Drive Charge"
		progressValue = float64(s.FTLDriveCharge) / 100.0
		details.WriteString(fmt.Sprintf("%s %d%%", labelStyle.Render("Charge:"), s.FTLDriveCharge))
		// added Description:
		description = "Current energy level accumulated for the next FTL jump. Must reach 100% to initiate warp. Charging speed depends on reactor output and drive health."

	case 5:
		detailTitle = "Food Supply"
		// assuming max food is 200 for progress bar, adjust if needed
		if s.Food > 200 {
			progressValue = 1.0
		} else if s.Food < 0 {
			progressValue = 0.0
		} else {
			progressValue = float64(s.Food) / 200.0
		}
		details.WriteString(fmt.Sprintf("%s %d units", labelStyle.Render("Stock:"), s.Food))
		// added Description:
		description = "Stored nutritional provisions for the crew. Essential for maintaining crew morale and efficiency. Running out leads to starvation and severe penalties."

	}
	if progressValue < 0.0 {
		progressValue = 0.0
	}
	if progressValue > 1.0 {
		progressValue = 1.0
	}
	progressBar := progress.New(progress.WithScaledGradient("#008FE9", "#F00065")).ViewAs(progressValue)
	detailContent := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(detailTitle),
		details.String(),
		"\n",
		progressBar,
		descriptionStyle.Render(description),
	)
	detailPanel := topPanelStyle.Render(detailContent)

	// ----- Bottom Panel: Extra Details (Horizontally Arranged Boxes) -----

	// Location Box
	locationTitle := boxTitleStyle.Render("Location")
	locationText := boxTextStyle.Render(fmt.Sprintf("Star System: %s\nLocation: %s",
		s.Location.StarSystemName,
		s.Location.PlanetName,
	))
	// ensure content fits vertically within detailBoxHeight
	locationBoxContent := lipgloss.NewStyle().Height(detailBoxHeight).Render(lipgloss.JoinVertical(lipgloss.Left, locationTitle, locationText))
	locationBox := detailBoxStyle.Render(locationBoxContent)

	// Cargo Box
	cargoTitle := boxTitleStyle.Render("Cargo Hold")
	cargoText := boxTextStyle.Render(fmt.Sprintf("%d/%d Units\n%d Items",
		s.Cargo.UsedCapacity,
		s.Cargo.Capacity,
		len(s.Cargo.Items),
	))
	cargoBoxContent := lipgloss.NewStyle().Height(detailBoxHeight).Render(lipgloss.JoinVertical(lipgloss.Left, cargoTitle, cargoText))
	cargoBox := detailBoxStyle.Render(cargoBoxContent)

	// Modules Box (Simplified to count)
	modulesTitle := boxTitleStyle.Render("Modules")
	modulesTextContent := fmt.Sprintf("%d Installed", len(s.Modules))
	activeCount := 0
	for _, mod := range s.Modules {
		if mod.Status == "Active" {
			activeCount++
		}
	}
	if len(s.Modules) > 0 {
		modulesTextContent += fmt.Sprintf("\n%d Active", activeCount)
	}
	modulesText := boxTextStyle.Render(modulesTextContent)
	modulesBoxContent := lipgloss.NewStyle().Height(detailBoxHeight).Render(lipgloss.JoinVertical(lipgloss.Left, modulesTitle, modulesText))
	modulesBox := detailBoxStyle.Render(modulesBoxContent)

	// Upgrades Box
	upgradesTitle := boxTitleStyle.Render("Upgrades")
	upgradesText := boxTextStyle.Render(fmt.Sprintf("ENG: %d/%d\nWPN: %d/%d\nCRG: %d/%d",
		s.Upgrades.Engine.CurrentLevel, s.Upgrades.Engine.MaxLevel,
		s.Upgrades.WeaponSystems.CurrentLevel, s.Upgrades.WeaponSystems.MaxLevel,
		s.Upgrades.CargoExpansion.CurrentLevel, s.Upgrades.CargoExpansion.MaxLevel,
	))
	upgradesBoxContent := lipgloss.NewStyle().Height(detailBoxHeight).Render(lipgloss.JoinVertical(lipgloss.Left, upgradesTitle, upgradesText))
	upgradesBox := detailBoxStyle.Render(upgradesBoxContent)

	// arrange boxes horizontally with spacing
	hSpacing := lipgloss.NewStyle().Width(spacingWidth).Render("")

	extraDetailContent := lipgloss.JoinHorizontal(lipgloss.Top,
		locationBox,
		hSpacing,
		cargoBox,
		hSpacing,
		modulesBox,
		hSpacing,
		upgradesBox,
	)

	// render the bottom panel containing the horizontal boxes
	extraPanel := bottomPanelStyle.Render(extraDetailContent)

	// ----- Final Layout -----
	// 1. join the top two panels horizontally
	topRowPanels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, detailPanel)

	// 2. join the top row vertically with the bottom panel
	finalView := lipgloss.JoinVertical(lipgloss.Left,
		topRowPanels,
		extraPanel,
	)

	return finalView
}
