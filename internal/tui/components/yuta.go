package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

const (
	fps       = 30  // Frames per second - doesn't seem to change much in our case
	height    = 15  // max height for bouncing
	frequency = 1.5 // higher frequency for faster bouncing
	damping   = 0.1 // less damping for more sustained motion
)

var (
	yutaStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("210"))
)

type frameMsg time.Time

// trigger a new frame at the desired FPS
func animate() tea.Cmd {
	return tea.Tick(time.Second/fps, func(t time.Time) tea.Msg {
		return frameMsg(t)
	})
}

type YutaModel struct {
	yPos   float64          // current Y position
	yVel   float64          // current velocity
	spring harmonica.Spring // harmonica spring for bounce physics
}

// creates a new instance of the Yuta model
func NewYuta() YutaModel {
	return YutaModel{
		spring: harmonica.NewSpring(harmonica.FPS(fps), frequency, damping),
	}
}

// initializes Yuta and starts the animation
func (m YutaModel) Init() tea.Cmd {
	return animate()
}

func (m YutaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case frameMsg:
		const targetY = float64(height)

		// update Y position and velocity using the spring
		m.yPos, m.yVel = m.spring.Update(m.yPos, m.yVel, targetY)

		// instead of stopping, reverse direction when reaching the target
		if m.yPos >= targetY && m.yVel > 0 {
			m.yVel = -m.yVel
		} else if m.yPos <= 0 && m.yVel < 0 {
			m.yVel = -m.yVel
		}

		// continue animation indefinitely ** might chanage this to stop after a certain time
		return m, animate()

	default:
		return m, nil
	}
}

func (m YutaModel) View() string {
	var out strings.Builder

	fmt.Fprintf(&out, "[debug] yPos=%.2f, yVel=%.2f\n\n", m.yPos, m.yVel)

	fmt.Fprint(&out, "\n")

	// calculating vertical position
	y := int(math.Round(m.yPos))
	if y < 0 {
		y = 0
	}

	for i := 0; i < y; i++ {
		fmt.Fprintln(&out) // empty lines to simulate vertical position
	}

	cube := []string{
		"╔═══════╗",
		"║║ ʘ‿ʘ ║║",
		"║       ║",
		"║       ║",
		"╚═══════╝",
	}

	// rendering the cube art
	for _, line := range cube {
		fmt.Fprintln(&out, yutaStyle.Render(line))
	}

	return out.String()
}
