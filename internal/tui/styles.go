package tui

import "github.com/charmbracelet/lipgloss"

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
)
