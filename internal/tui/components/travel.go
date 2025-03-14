package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	model "github.com/dominik-merdzik/project-starbyte/internal/tui/models"
)

// TravelTickMsg is sent when the travel timer ticks
type TravelTickMsg struct{}

// TravelComponent handles the UI and logic for traveling to mission locations
type TravelComponent struct {
	Mission        *model.Mission
	IsTravelling   bool
	StartTime      time.Time
	Duration       time.Duration
	Progress       progress.Model
	TravelComplete bool
	DestLocation   string
}

// NewTravelComponent creates a new travel component
func NewTravelComponent() TravelComponent {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return TravelComponent{
		Progress: p,
	}
}

// StartTravel begins a travel journey to a mission location
func (t *TravelComponent) StartTravel(mission *model.Mission) tea.Cmd {
	t.Mission = mission
	t.IsTravelling = true
	t.TravelComplete = false
	t.StartTime = time.Now()
	t.Duration = 2 * time.Second // Hardcoded 2 seconds travel time
	t.Progress.SetPercent(0)
	t.DestLocation = mission.Location

	return tea.Batch(
		tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return TravelTickMsg{} }),
		//tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return progress.FrameMsg{} }),
		t.Progress.Init(),
	)
}

// Update handles messages for the travel component
func (t *TravelComponent) Update(msg tea.Msg) (TravelComponent, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case progress.FrameMsg:
		if t.IsTravelling {
			progressModel, cmd := t.Progress.Update(msg)
			if p, ok := progressModel.(progress.Model); ok {
				t.Progress = p
			}
			cmds = append(cmds, cmd)
		}

	case TravelTickMsg:
		if t.IsTravelling && t.Mission != nil {
			elapsed := time.Since(t.StartTime)
			percentComplete := float64(elapsed) / float64(t.Duration)

			cmds = append(cmds, t.Progress.SetPercent(percentComplete))

			// Check if travel is complete
			if elapsed >= t.Duration {
				t.IsTravelling = false
				t.TravelComplete = true
				t.Mission.Status = model.MissionStatusInProgress
				return *t, tea.Batch(cmds...)
			}

			// Continue ticking if still traveling
			cmds = append(cmds, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
				return TravelTickMsg{}
			}))
		}
	}

	return *t, tea.Batch(cmds...)
}

// View renders the travel component
func (t *TravelComponent) View() string {
	if !t.IsTravelling || t.Mission == nil {
		return ""
	}

	remainingTime := t.Duration - time.Since(t.StartTime)

	progressBar := t.Progress.View()

	return fmt.Sprintf("Travelling to %s\n\n%s\n\nTime remaining: %v\n",
		t.Mission.Location,
		progressBar,
		remainingTime.Round(time.Millisecond))
}
