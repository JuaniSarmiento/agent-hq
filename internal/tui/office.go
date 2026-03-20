package tui

import (
	"fmt"
	"strings"

	"github.com/juani/agent-hq/internal/model"
)

// Office represents a group of agents sharing the same profile.
type Office struct {
	Profile  string
	Agents   []model.Agent
	Selected int
}

// statusIcon returns an emoji for the agent status.
func statusIcon(status model.AgentStatus) string {
	switch status {
	case model.StatusRunning:
		return "\U0001F7E2" // green circle
	case model.StatusQueued:
		return "\U0001F7E1" // yellow circle
	case model.StatusCompleted:
		return "\u2705" // check mark
	case model.StatusFailed:
		return "\u274C" // red X
	default:
		return "\u26AB" // black circle
	}
}

// statusLabel returns a display label for the status.
func statusLabel(status model.AgentStatus) string {
	switch status {
	case model.StatusRunning:
		return "BUSY"
	case model.StatusQueued:
		return "WAIT"
	case model.StatusCompleted:
		return "DONE"
	case model.StatusFailed:
		return "FAIL"
	default:
		return "IDLE"
	}
}

// agentStyle returns the appropriate lipgloss style for a status.
func agentStyle(status model.AgentStatus) func(strs ...string) string {
	switch status {
	case model.StatusRunning:
		return agentRunningStyle.Render
	case model.StatusQueued:
		return agentQueuedStyle.Render
	case model.StatusCompleted:
		return agentCompletedStyle.Render
	case model.StatusFailed:
		return agentFailedStyle.Render
	default:
		return agentIdleStyle.Render
	}
}

// Render draws the office panel with its agents.
func (o Office) Render(width int, focused bool, cursor int) string {
	var b strings.Builder

	title := officeTitleStyle.Render(fmt.Sprintf(" %s", o.Profile))
	b.WriteString(title)
	b.WriteString("\n")

	if len(o.Agents) == 0 {
		b.WriteString(agentIdleStyle.Render("  (no agents)"))
		b.WriteString("\n")
	}

	for i, agent := range o.Agents {
		icon := statusIcon(agent.Status)
		label := statusLabel(agent.Status)
		render := agentStyle(agent.Status)

		shortID := agent.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		// Build agent line: icon, shortID, status, and model badge.
		line := fmt.Sprintf(" %s %s %s", icon, shortID, render(label))
		if agent.Model != "" {
			line += " " + modelBadge(agent.Model)
		}

		if focused && i == cursor {
			line = selectedAgentStyle.Render("> ") + line
		} else {
			line = "  " + line
		}

		b.WriteString(line)
		b.WriteString("\n")

		// Show task on next line for running agents.
		if agent.Status == model.StatusRunning {
			task := agent.Task
			maxLen := width - 6
			if maxLen > 0 && len(task) > maxLen {
				task = task[:maxLen-3] + "..."
			}
			b.WriteString(fmt.Sprintf("      %s", agentRunningStyle.Render(fmt.Sprintf("%q", task))))
			b.WriteString("\n")
		}

		// Show token usage for running or completed agents.
		if (agent.Status == model.StatusRunning || agent.Status == model.StatusCompleted) &&
			(agent.TokenInput > 0 || agent.TokenOutput > 0) {
			tokenInfo := tokenStyle.Render(
				fmt.Sprintf("      \u27E8%s/%s\u27E9", formatTokens(agent.TokenInput), formatTokens(agent.TokenOutput)),
			)
			// Show retry count if > 0.
			if agent.RetryCount > 0 {
				tokenInfo += tokenStyle.Render(fmt.Sprintf(" retry %d/%d", agent.RetryCount, agent.MaxRetries))
			}
			b.WriteString(tokenInfo)
			b.WriteString("\n")
		}
	}

	content := b.String()

	if focused {
		return officeSelectedBorderStyle.Width(width).Render(content)
	}
	return officeBorderStyle.Width(width).Render(content)
}
