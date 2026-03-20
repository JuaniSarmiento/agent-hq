package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Status colors.
	colorRunning   = lipgloss.Color("#00FF00")
	colorQueued    = lipgloss.Color("#FFFF00")
	colorCompleted = lipgloss.Color("#0088FF")
	colorFailed    = lipgloss.Color("#FF0000")
	colorIdle      = lipgloss.Color("#666666")
	colorAccent    = lipgloss.Color("#FF00FF")
	colorSubtle    = lipgloss.Color("#444444")
	colorWhite     = lipgloss.Color("#FFFFFF")
	colorDimWhite  = lipgloss.Color("#AAAAAA")

	// Header style.
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(lipgloss.Color("#333366")).
			Padding(0, 1).
			Align(lipgloss.Center)

	// Office panel border.
	officeBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorSubtle).
				Padding(0, 1)

	officeSelectedBorderStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(colorAccent).
					Padding(0, 1)

	// Office title.
	officeTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite)

	// Agent line styles.
	agentRunningStyle   = lipgloss.NewStyle().Foreground(colorRunning)
	agentQueuedStyle    = lipgloss.NewStyle().Foreground(colorQueued)
	agentCompletedStyle = lipgloss.NewStyle().Foreground(colorCompleted)
	agentFailedStyle    = lipgloss.NewStyle().Foreground(colorFailed)
	agentIdleStyle      = lipgloss.NewStyle().Foreground(colorIdle)

	// Selected agent highlight.
	selectedAgentStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	// Task log styles.
	taskLogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite).
				Padding(0, 1)

	taskLogBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorSubtle).
				Padding(0, 1)

	taskLogSelectedBorderStyle = lipgloss.NewStyle().
					Border(lipgloss.NormalBorder()).
					BorderForeground(colorAccent).
					Padding(0, 1)

	taskLogEntryStyle = lipgloss.NewStyle().
				Foreground(colorDimWhite)

	// Footer / help style.
	helpStyle = lipgloss.NewStyle().
			Foreground(colorDimWhite).
			Padding(0, 1)

	// Detail view styles.
	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWhite).
				Padding(0, 1)

	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorWhite)

	// Token display style (dim, low visual weight).
	tokenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	// Model badge styles by model name.
	modelBadgeOpus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")).
			Background(lipgloss.Color("#BB77FF")).
			Padding(0, 1).
			Bold(true)

	modelBadgeSonnet = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333")).
				Background(lipgloss.Color("#5599FF")).
				Padding(0, 1).
				Bold(true)

	modelBadgeHaiku = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")).
			Background(lipgloss.Color("#44CC77")).
			Padding(0, 1).
			Bold(true)

	modelBadgeDefault = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333")).
				Background(lipgloss.Color("#AAAAAA")).
				Padding(0, 1).
				Bold(true)
)

// modelBadge returns a styled badge for the given model name.
func modelBadge(modelName string) string {
	if modelName == "" {
		return ""
	}
	lower := strings.ToLower(modelName)
	switch {
	case strings.Contains(lower, "opus"):
		return modelBadgeOpus.Render(modelName)
	case strings.Contains(lower, "sonnet"):
		return modelBadgeSonnet.Render(modelName)
	case strings.Contains(lower, "haiku"):
		return modelBadgeHaiku.Render(modelName)
	default:
		return modelBadgeDefault.Render(modelName)
	}
}

// formatTokens formats a token count: >= 1000 as "1.2k", otherwise raw number.
func formatTokens(n int) string {
	if n >= 1000 {
		v := float64(n) / 1000.0
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", v), "0"), ".") + "k"
	}
	return fmt.Sprintf("%d", n)
}
