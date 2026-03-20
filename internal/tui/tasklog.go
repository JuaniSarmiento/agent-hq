package tui

import (
	"fmt"
	"strings"

	"github.com/juani/agent-hq/internal/model"
)

// TaskLog renders the bottom panel showing recent activity.
type TaskLog struct {
	Activities []model.Activity
	Selected   int
}

// actionIcon returns an icon for the activity action type.
func actionIcon(action string) string {
	switch action {
	case "file_write":
		return "\U0001F4DD" // memo
	case "file_read":
		return "\U0001F4C4" // page
	case "tool_call":
		return "\U0001F527" // wrench
	case "search":
		return "\U0001F50D" // magnifying glass
	default:
		return "\u2022" // bullet
	}
}

// Render draws the task log panel.
func (tl TaskLog) Render(width int, focused bool, cursor int) string {
	var b strings.Builder

	title := taskLogTitleStyle.Render("\U0001F4CB TASK LOG")
	b.WriteString(title)
	b.WriteString("\n")

	if len(tl.Activities) == 0 {
		b.WriteString(taskLogEntryStyle.Render("  (no activity yet)"))
		b.WriteString("\n")
	}

	for i, act := range tl.Activities {
		ts := act.Timestamp.Format("15:04")
		icon := actionIcon(act.Action)

		detail := act.Action
		if act.Detail != nil && *act.Detail != "" {
			detail = *act.Detail
		}

		maxDetail := width - 30
		if maxDetail > 0 && len(detail) > maxDetail {
			detail = detail[:maxDetail-3] + "..."
		}

		agentID := act.AgentID
		if len(agentID) > 8 {
			agentID = agentID[:8]
		}

		line := fmt.Sprintf("  %s  %s  %s %s", ts, agentID, icon, detail)

		if act.FilePath != nil && *act.FilePath != "" {
			line += fmt.Sprintf(" %s", agentIdleStyle.Render(*act.FilePath))
		}

		if focused && i == cursor {
			line = selectedAgentStyle.Render("> ") + line[2:]
		}

		b.WriteString(taskLogEntryStyle.Render(line))
		b.WriteString("\n")
	}

	content := b.String()

	if focused {
		return taskLogSelectedBorderStyle.Width(width).Render(content)
	}
	return taskLogBorderStyle.Width(width).Render(content)
}
