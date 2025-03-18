package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominik-merdzik/project-starbyte/internal/data"
)

// TravelTickMsg is sent when the travel timer ticks
type TravelTickMsg struct{}

// TravelComponent handles the UI and logic for traveling to mission locations
type TravelComponent struct {
	Mission        *data.Mission
	IsTravelling   bool
	StartTime      time.Time
	Duration       time.Duration
	Progress       progress.Model
	TravelComplete bool
	DestLocation   data.Location
}

// NewTravelComponent creates a new travel component
func NewTravelComponent() TravelComponent {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(80),
	)

	return TravelComponent{
		Progress: p,
	}
}

// Called when you want to show the travel UI component
func (t *TravelComponent) StartTravel(destination data.Location) tea.Cmd {
	t.IsTravelling = true
	t.TravelComplete = false
	t.StartTime = time.Now()
	t.Duration = 2 * time.Second // Hardcoded 2 seconds travel time
	t.Progress.SetPercent(0)
	t.DestLocation = destination // Store the destination location

	return tea.Batch(
		tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return TravelTickMsg{} }),
		//tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return progress.FrameMsg{} }), // IDK if this line is necessary
		t.Progress.Init(),
	)
}

// Update handles messages for the travel component
func (t *TravelComponent) Update(msg tea.Msg) (TravelComponent, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case progress.FrameMsg:
		// Update the progress bar animation
		progressModel, cmd := t.Progress.Update(msg)
		t.Progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case TravelTickMsg:
		// Calculate elapsed time and update progress
		elapsed := time.Since(t.StartTime)

		// Update progress percentage
		percentComplete := float64(elapsed) / float64(t.Duration)
		if percentComplete > 1.0 {
			percentComplete = 1.0
			t.TravelComplete = true
		}

		// Update progress bar
		cmds = append(cmds, t.Progress.SetPercent(percentComplete))

		// Continue ticking if not complete
		if !t.TravelComplete {
			cmds = append(cmds, tea.Tick(100*time.Millisecond,
				func(time.Time) tea.Msg { return TravelTickMsg{} }))
		} else {
			// Reset travel state when complete
			t.IsTravelling = false
		}
	}

	return *t, tea.Batch(cmds...)
}

// View renders the travel component
func (t *TravelComponent) View() string {
	if !t.IsTravelling {
		return ""
	}

	remainingTime := t.Duration - time.Since(t.StartTime)

	progressBar := t.Progress.View()

	planet := t.DestLocation.PlanetName
	system := t.DestLocation.StarSystemName

	// Build travel view with border and styling
	travelView := fmt.Sprintf("\nTRAVELLING TO %s, %s\n\n%s\n\nTime remaining: %.2f seconds\n",
		planet,
		system,
		progressBar,
		remainingTime.Seconds())

	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(travelView)
}
