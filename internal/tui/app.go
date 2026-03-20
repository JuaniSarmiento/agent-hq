package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/juani/agent-hq/internal/db"
	"github.com/juani/agent-hq/internal/model"
)

// View represents which screen is active.
type View int

const (
	ViewMain View = iota
	ViewDetail
)

// FocusPanel represents which panel has focus.
type FocusPanel int

const (
	FocusOffices FocusPanel = iota
	FocusTaskLog
)

// tickMsg is sent on each poll interval.
type tickMsg time.Time

// Model is the main Bubbletea model for Agent HQ.
type Model struct {
	db       *db.DB
	keys     KeyMap
	width    int
	height   int
	view     View
	focus    FocusPanel
	showHelp bool

	// Main view state.
	offices       []Office
	officeIdx     int
	agentCursor   int
	taskLog       TaskLog
	taskLogCursor int
	startTime     time.Time

	// Detail view state.
	detail DetailView

	// Error state.
	err error
}

// New creates a new Model with the given database connection.
func New(database *db.DB) Model {
	return Model{
		db:        database,
		keys:      DefaultKeyMap(),
		startTime: time.Now(),
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.pollData(),
		m.tick(),
	)
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// pollDataMsg carries refreshed data from the database.
type pollDataMsg struct {
	agents     []model.Agent
	activities []model.Activity
	err        error
}

func (m Model) pollData() tea.Cmd {
	return func() tea.Msg {
		agents, err := m.db.GetAgents()
		if err != nil {
			return pollDataMsg{err: err}
		}
		activities, err := m.db.GetRecentActivity(50)
		if err != nil {
			return pollDataMsg{err: err}
		}
		return pollDataMsg{agents: agents, activities: activities}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		return m, tea.Batch(m.pollData(), m.tick())

	case pollDataMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.refreshOffices(msg.agents)
		m.taskLog.Activities = msg.activities
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Forward to detail viewport if active.
	if m.view == ViewDetail && m.detail.Ready {
		var cmd tea.Cmd
		m.detail.Viewport, cmd = m.detail.Viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys.
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.ToggleHelp):
		m.showHelp = !m.showHelp
		return m, nil
	}

	// View-specific keys.
	switch m.view {
	case ViewDetail:
		if key.Matches(msg, m.keys.Back) {
			m.view = ViewMain
			return m, nil
		}
		// Forward j/k to viewport for scrolling.
		var cmd tea.Cmd
		m.detail.Viewport, cmd = m.detail.Viewport.Update(msg)
		return m, cmd

	case ViewMain:
		return m.handleMainKeys(msg)
	}

	return m, nil
}

func (m Model) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Tab):
		if m.focus == FocusOffices {
			m.focus = FocusTaskLog
		} else {
			m.focus = FocusOffices
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.focus == FocusOffices {
			if len(m.offices) > 0 && m.officeIdx < len(m.offices) {
				office := m.offices[m.officeIdx]
				if m.agentCursor < len(office.Agents)-1 {
					m.agentCursor++
				}
			}
		} else {
			if m.taskLogCursor < len(m.taskLog.Activities)-1 {
				m.taskLogCursor++
			}
		}
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.focus == FocusOffices {
			if m.agentCursor > 0 {
				m.agentCursor--
			}
		} else {
			if m.taskLogCursor > 0 {
				m.taskLogCursor--
			}
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if m.focus == FocusOffices && len(m.offices) > 0 {
			office := m.offices[m.officeIdx]
			if m.agentCursor < len(office.Agents) {
				agent := office.Agents[m.agentCursor]
				return m.openDetail(agent)
			}
		}
		return m, nil

	// Left/right to switch offices via h/l.
	case msg.String() == "h" || msg.String() == "left":
		if m.focus == FocusOffices && m.officeIdx > 0 {
			m.officeIdx--
			m.agentCursor = 0
		}
		return m, nil

	case msg.String() == "l" || msg.String() == "right":
		if m.focus == FocusOffices && m.officeIdx < len(m.offices)-1 {
			m.officeIdx++
			m.agentCursor = 0
		}
		return m, nil
	}

	return m, nil
}

func (m Model) openDetail(agent model.Agent) (tea.Model, tea.Cmd) {
	activities, _ := m.db.GetAgentActivity(agent.ID)
	files, _ := m.db.GetAgentFiles(agent.ID)

	m.detail = NewDetailView(agent, activities, files, m.width, m.height)
	m.view = ViewDetail

	return m, nil
}

func (m *Model) refreshOffices(agents []model.Agent) {
	profileMap := make(map[string][]model.Agent)
	var profileOrder []string

	for _, a := range agents {
		if _, exists := profileMap[a.Profile]; !exists {
			profileOrder = append(profileOrder, a.Profile)
		}
		profileMap[a.Profile] = append(profileMap[a.Profile], a)
	}

	offices := make([]Office, 0, len(profileOrder))
	for _, p := range profileOrder {
		offices = append(offices, Office{
			Profile: p,
			Agents:  profileMap[p],
		})
	}

	m.offices = offices

	// Clamp cursors.
	if m.officeIdx >= len(m.offices) {
		m.officeIdx = max(0, len(m.offices)-1)
	}
	if len(m.offices) > 0 && m.agentCursor >= len(m.offices[m.officeIdx].Agents) {
		m.agentCursor = max(0, len(m.offices[m.officeIdx].Agents)-1)
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	switch m.view {
	case ViewDetail:
		return m.detail.Render(m.width)
	default:
		return m.renderMain()
	}
}

func (m Model) renderMain() string {
	var sections []string

	// Header.
	elapsed := time.Since(m.startTime)
	header := headerStyle.Width(m.width).Render(
		fmt.Sprintf("\U0001F3E2 AGENT HQ                        \u23F1 %s", formatDuration(elapsed)),
	)
	sections = append(sections, header)

	// Offices grid.
	officeRows := m.renderOfficeGrid()
	sections = append(sections, officeRows)

	// Task log (takes remaining space, min 8 lines).
	taskLogHeight := m.height - lipgloss.Height(strings.Join(sections, "\n")) - 3
	if taskLogHeight < 4 {
		taskLogHeight = 4
	}

	// Limit activities shown based on available height.
	displayActivities := m.taskLog.Activities
	maxEntries := taskLogHeight - 2
	if maxEntries > 0 && len(displayActivities) > maxEntries {
		displayActivities = displayActivities[:maxEntries]
	}

	tl := TaskLog{Activities: displayActivities, Selected: m.taskLogCursor}
	sections = append(sections, tl.Render(m.width-2, m.focus == FocusTaskLog, m.taskLogCursor))

	// Footer.
	footer := m.renderFooter()
	sections = append(sections, footer)

	return strings.Join(sections, "\n")
}

func (m Model) renderOfficeGrid() string {
	if len(m.offices) == 0 {
		return agentIdleStyle.Render("  No agents found. Waiting for data...")
	}

	// Calculate column width based on terminal width and number of offices.
	numCols := m.width / 30
	if numCols < 1 {
		numCols = 1
	}
	if numCols > len(m.offices) {
		numCols = len(m.offices)
	}

	colWidth := (m.width - 2) / numCols
	if colWidth < 20 {
		colWidth = 20
	}

	var rows []string
	for i := 0; i < len(m.offices); i += numCols {
		var cols []string
		for j := 0; j < numCols && i+j < len(m.offices); j++ {
			idx := i + j
			office := m.offices[idx]
			focused := m.focus == FocusOffices && idx == m.officeIdx
			cursor := 0
			if focused {
				cursor = m.agentCursor
			}
			cols = append(cols, office.Render(colWidth-4, focused, cursor))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderFooter() string {
	if m.showHelp {
		return helpStyle.Width(m.width).Render(
			"[tab] cycle panel  [h/l] switch office  [j/k] navigate  [enter] detail  [esc] back  [?] close help  [q] quit",
		)
	}
	return helpStyle.Width(m.width).Render(
		"[tab] panel  [enter] detail  [?] help  [q] quit",
	)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
