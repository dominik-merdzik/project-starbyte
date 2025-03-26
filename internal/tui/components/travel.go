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

// StartTravel begins a new travel to the given destination with the given duration
func (t *TravelComponent) StartTravel(destination data.Location, travelDuration time.Duration) tea.Cmd { // <-- Added travelDuration arg
	t.IsTravelling = true
	t.TravelComplete = false
	t.StartTime = time.Now()
	t.Duration = travelDuration

	// prevent division by zero issues later if duration is somehow zero or negative
	if t.Duration <= 0 {
		t.Duration = 100 * time.Millisecond
	}

	t.Progress.SetPercent(0)
	t.DestLocation = destination

	// combine ticks: Start the first tick and initialize the progress bar animation
	return tea.Batch(
		func() tea.Msg { return TravelTickMsg{} },
		t.Progress.Init(),
	)
}

// Update handles messages for the travel component
func (t *TravelComponent) Update(msg tea.Msg) (TravelComponent, tea.Cmd) {
	var cmds []tea.Cmd

	// if travel is already marked complete, don't process further ticks/frames for it
	if t.TravelComplete {
		// allow progress bar frame updates briefly after completion for visuals
		if _, ok := msg.(progress.FrameMsg); ok {
			progressModel, cmd := t.Progress.Update(msg)
			if newProgress, ok := progressModel.(progress.Model); ok {
				t.Progress = newProgress
			}
			return *t, cmd
		}
		return *t, nil
	}

	// handle messages only if travel is ongoing
	switch msg := msg.(type) {
	case progress.FrameMsg:
		// update the progress bar animation
		progressModel, cmd := t.Progress.Update(msg)
		if newProgress, ok := progressModel.(progress.Model); ok {
			t.Progress = newProgress
		}
		cmds = append(cmds, cmd)

	case TravelTickMsg:
		elapsed := time.Since(t.StartTime)
		percentComplete := float64(elapsed) / float64(t.Duration) // duration checked > 0 in StartTravel

		if percentComplete >= 1.0 {
			percentComplete = 1.0
			t.TravelComplete = true // mark as complete
		}

		cmds = append(cmds, t.Progress.SetPercent(percentComplete))

		// continue timer ticking if not yet complete
		if !t.TravelComplete {
			cmds = append(cmds, tea.Tick(100*time.Millisecond, // tick interval
				func(time.Time) tea.Msg { return TravelTickMsg{} }))
		}
	}

	// batch up any commands collected during this update cycle
	return *t, tea.Batch(cmds...)
}

func (t *TravelComponent) View() string {
	// only show if actively travelling or very recently completed
	showWhileTravelling := t.IsTravelling && !t.TravelComplete
	showBrieflyAfterArrival := t.TravelComplete && time.Since(t.StartTime) <= t.Duration+2*time.Second // Show for 2s after scheduled end

	if !showWhileTravelling && !showBrieflyAfterArrival {
		return "" // Not visible
	}

	// ... rest of the View function remains the same ...

	elapsed := time.Since(t.StartTime)
	remainingTime := t.Duration - elapsed
	if remainingTime < 0 {
		remainingTime = 0
	}

	progressBar := t.Progress.View()
	planet := t.DestLocation.PlanetName
	system := t.DestLocation.StarSystemName
	status := "TRAVELLING TO"

	// display slightly differently upon arrival if desired
	if t.TravelComplete {
		status = "ARRIVED AT"
		remainingTime = 0 // force remaining time to 0 on completion view
	}

	travelView := fmt.Sprintf("\n%s %s, %s\n\n%s\n\nTime remaining: %.1f seconds\n",
		status,
		planet,
		system,
		progressBar,
		remainingTime.Seconds(),
	)

	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render(travelView)
}
