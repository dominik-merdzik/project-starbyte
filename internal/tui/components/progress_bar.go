package components

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

type ProgressBar struct {
	Model progress.Model
}

// NewProgressBar initializes and returns a new ProgressBar
func NewProgressBar() ProgressBar {

	// rogress bar styles
	p := progress.New(
		progress.WithScaledGradient("#F00065", "#008FE9"),
		progress.WithWidth(46),
		//tesing this out:: progress.WithoutPercentage(),
		//progress.WithFillCharacters('░', '▓'), // Tested the runes fill
	)
	return ProgressBar{Model: p}
}

// RenderProgressBar renders the progress bar given the current and max values
func (p ProgressBar) RenderProgressBar(current, max int) string {

	// calculate percentage progress
	percent := float64(current) / float64(max)
	if percent > 1.0 {
		percent = 1.0
	} else if percent < 0.0 {
		percent = 0.0
	}

	// render
	barStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).PaddingTop(1)

	return barStyle.Render(p.Model.ViewAs(percent))
}
