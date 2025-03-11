package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
	"github.com/dominik-merdzik/project-starbyte/internal/utilities"
)

// ShipModel represents the ship's status and components.
type ShipModel struct {
	ShipID            string // From save.
	Name              string
	HullHealth        int
	MaxHullHealth     int
	EngineHealth      int
	EngineFuel        int
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
	Cursor            int // Index of the currently selected ship component.

	GameSave *data.FullGameSave
}

// NewShipModelDefaults creates a ShipModel with default values.
func NewShipModelDefaults() ShipModel {
	return ShipModel{
		ShipID:            "DEFAULT_SHIP_ID",
		Name:              "Voyager 3",
		HullHealth:        100,
		MaxHullHealth:     100,
		EngineHealth:      100,
		EngineFuel:        80,
		FTLDriveHealth:    70,
		FTLDriveCharge:    0,
		ShieldStrength:    50,
		MaxShieldStrength: 50,
		MaxFuel:           200,
		Crew:              []CrewMember{},
		Food:              100,
		Location:          data.Location{StarSystemId: "SYS_DEFAULT", PlanetId: "Earth", Coordinates: data.Coordinates{X: 0, Y: 0, Z: 0}},
		Cargo:             data.Cargo{Capacity: 100, UsedCapacity: 0, Items: []data.CargoItem{}},
		Modules:           []data.Module{},
		Upgrades:          data.Upgrades{Engine: data.UpgradeLevel{CurrentLevel: 1, MaxLevel: 5}, WeaponSystems: data.UpgradeLevel{CurrentLevel: 0, MaxLevel: 5}, CargoExpansion: data.UpgradeLevel{CurrentLevel: 0, MaxLevel: 5}},
		Cursor:            0,
	}
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
		FTLDriveHealth:    savedShip.FTLDriveHealth,
		FTLDriveCharge:    savedShip.FTLDriveCharge,
		ShieldStrength:    savedShip.ShieldStrength,
		MaxShieldStrength: savedShip.MaxShieldStrength,
		MaxFuel:           savedShip.MaxFuel,
		Crew:              []CrewMember{}, // Loaded separately.
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
			// Subtract 10 units of EngineFuel.
			s.EngineFuel -= 10
			if s.EngineFuel < 0 {
				s.EngineFuel = 0
			}
		case "d":
			// Add 5 units of EngineFuel.
			s.EngineFuel += 5
			if s.EngineFuel > s.MaxFuel {
				s.EngineFuel = s.MaxFuel
			}
		case "w":
			// Set engine fuel to 0 immediately *FOR DEMO PURPOSES ONLY*
			s.EngineFuel = 0
			// push the save
			return s, utilities.PushSave(s.GameSave, func() {
				s.GameSave.Ship.Fuel = s.EngineFuel
			})

		}
	}
	return s, nil
}

// View renders the ship model UI.
func (s ShipModel) View() string {
	items := []string{
		"Hull Health",
		"Engine Health",
		"Engine Fuel",
		"FTL Drive Health",
		"FTL Drive Charge",
		"Food",
	}

	// Styling for the main panel.
	panelStyle := lipgloss.NewStyle().
		Width(60).
		Height(22).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle := lipgloss.NewStyle().Bold(true)
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("217"))
	arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	details := ""

	var shipList strings.Builder
	shipList.WriteString(titleStyle.Render("Ship Status") + "\n\n")
	for i, item := range items {
		cursor := "  "
		if i == s.Cursor {
			cursor = arrowStyle.Render("> ")
		}
		shipList.WriteString(fmt.Sprintf("%s%s\n", cursor, defaultStyle.Render(item)))
	}

	// Detailed panel based on selection.
	var progressValue float64
	switch s.Cursor {
	case 0:
		details = fmt.Sprintf("%s\n\n%s %d", titleStyle.Render("Hull Health"), labelStyle.Render("Current:"), s.HullHealth)
		progressValue = float64(s.HullHealth) / float64(s.MaxHullHealth)
	case 1:
		details = fmt.Sprintf("%s\n\n%s %d", titleStyle.Render("Engine Health"), labelStyle.Render("Current:"), s.EngineHealth)
		progressValue = float64(s.EngineHealth) / 100
	case 2:
		details = fmt.Sprintf("%s\n\n%s %d", titleStyle.Render("Engine Fuel"), labelStyle.Render("Current:"), s.EngineFuel)
		progressValue = float64(s.EngineFuel) / float64(s.MaxFuel)
	case 3:
		details = fmt.Sprintf("%s\n\n%s %d", titleStyle.Render("FTL Drive Health"), labelStyle.Render("Current:"), s.FTLDriveHealth)
		progressValue = float64(s.FTLDriveHealth) / 100
	case 4:
		details = fmt.Sprintf("%s\n\n%s %d", titleStyle.Render("FTL Drive Charge"), labelStyle.Render("Current:"), s.FTLDriveCharge)
		progressValue = float64(s.FTLDriveCharge) / 100
	case 5:
		details = fmt.Sprintf("%s\n\n%s %d", titleStyle.Render("Food Supply"), labelStyle.Render("Current:"), s.Food)
		progressValue = float64(s.Food) / 100
	}
	progressBar := progress.New(progress.WithScaledGradient("#008FE9", "#F00065")).ViewAs(progressValue)
	detailPanel := panelStyle.Render(details + "\n\n" + progressBar)

	// extra details for Location, Cargo, Modules, and Upgrades
	var extraDetails strings.Builder
	extraDetails.WriteString(titleStyle.Render("Extra Details") + "\n\n")
	// location
	extraDetails.WriteString(fmt.Sprintf("%s: %s\n%s: %s\n",
		labelStyle.Render("Star System"),
		s.Location.StarSystemId,
		labelStyle.Render("Planet"),
		s.Location.PlanetId))
	// cargo
	extraDetails.WriteString(fmt.Sprintf("%s: %d/%d used, %d items\n",
		labelStyle.Render("Cargo"),
		s.Cargo.UsedCapacity,
		s.Cargo.Capacity,
		len(s.Cargo.Items)))
	// modules
	extraDetails.WriteString(labelStyle.Render("Modules: ") + " ")
	for i, mod := range s.Modules {
		extraDetails.WriteString(fmt.Sprintf("%s (%s)", mod.Name, mod.Status))
		if i < len(s.Modules)-1 {
			extraDetails.WriteString(", ")
		}
	}
	extraDetails.WriteString("\n")
	// upgrades
	extraDetails.WriteString(fmt.Sprintf("%s: Engine %d/%d, Weapon %d/%d, Cargo %d/%d\n",
		labelStyle.Render("Upgrades"),
		s.Upgrades.Engine.CurrentLevel, s.Upgrades.Engine.MaxLevel,
		s.Upgrades.WeaponSystems.CurrentLevel, s.Upgrades.WeaponSystems.MaxLevel,
		s.Upgrades.CargoExpansion.CurrentLevel, s.Upgrades.CargoExpansion.MaxLevel))

	extraPanelStyle := lipgloss.NewStyle().
		Width(60).
		Height(12).
		Align(lipgloss.Center).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))
	extraPanel := extraPanelStyle.Render(extraDetails.String())

	// combine the detail panels vertically
	combinedDetails := lipgloss.JoinVertical(lipgloss.Left, detailPanel, extraPanel)

	// render the left panel (menu) and the combined details side by side
	leftPanel := panelStyle.Render(shipList.String())
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, combinedDetails)
}
