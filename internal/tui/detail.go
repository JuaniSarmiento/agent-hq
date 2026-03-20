package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/juani/agent-hq/internal/model"
)

// DetailView shows detailed information about a single agent.
type DetailView struct {
	Agent      model.Agent
	Activities []model.Activity
	Files      []model.FileChange
	Artifacts  []model.Artifact
	FileLocks  []model.FileLock
	Viewport   viewport.Model
	Ready      bool
}

// NewDetailView creates a detail view for the given agent.
func NewDetailView(agent model.Agent, activities []model.Activity, files []model.FileChange, artifacts []model.Artifact, locks []model.FileLock, width, height int) DetailView {
	vp := viewport.New(width-4, height-6)
	vp.SetContent(renderDetailContent(agent, activities, files, artifacts, locks, width-6))

	return DetailView{
		Agent:      agent,
		Activities: activities,
		Files:      files,
		Artifacts:  artifacts,
		FileLocks:  locks,
		Viewport:   vp,
		Ready:      true,
	}
}

func renderDetailContent(agent model.Agent, activities []model.Activity, files []model.FileChange, artifacts []model.Artifact, locks []model.FileLock, width int) string {
	var b strings.Builder

	// Agent info header.
	b.WriteString(detailLabelStyle.Render("ID:       "))
	b.WriteString(detailValueStyle.Render(agent.ID))
	b.WriteString("\n")

	b.WriteString(detailLabelStyle.Render("Profile:  "))
	b.WriteString(detailValueStyle.Render(agent.Profile))
	b.WriteString("\n")

	b.WriteString(detailLabelStyle.Render("Task:     "))
	b.WriteString(detailValueStyle.Render(agent.Task))
	b.WriteString("\n")

	b.WriteString(detailLabelStyle.Render("Status:   "))
	icon := statusIcon(agent.Status)
	b.WriteString(fmt.Sprintf("%s %s", icon, agentStyle(agent.Status)(string(agent.Status))))
	b.WriteString("\n")

	// Model field.
	if agent.Model != "" {
		b.WriteString(detailLabelStyle.Render("Model:    "))
		b.WriteString(modelBadge(agent.Model))
		b.WriteString("\n")
	}

	// Timeout field.
	if agent.TimeoutMs > 0 {
		b.WriteString(detailLabelStyle.Render("Timeout:  "))
		b.WriteString(detailValueStyle.Render(fmt.Sprintf("%dms", agent.TimeoutMs)))
		b.WriteString("\n")
	}

	// Retries field.
	if agent.MaxRetries > 0 {
		b.WriteString(detailLabelStyle.Render("Retries:  "))
		b.WriteString(detailValueStyle.Render(fmt.Sprintf("%d/%d", agent.RetryCount, agent.MaxRetries)))
		b.WriteString("\n")
	}

	if agent.StartedAt != nil {
		b.WriteString(detailLabelStyle.Render("Started:  "))
		b.WriteString(detailValueStyle.Render(agent.StartedAt.Format("15:04:05")))
		b.WriteString("\n")
	}
	if agent.FinishedAt != nil {
		b.WriteString(detailLabelStyle.Render("Finished: "))
		b.WriteString(detailValueStyle.Render(agent.FinishedAt.Format("15:04:05")))
		b.WriteString("\n")
	}
	if agent.ResultSummary != nil {
		b.WriteString(detailLabelStyle.Render("Result:   "))
		b.WriteString(detailValueStyle.Render(*agent.ResultSummary))
		b.WriteString("\n")
	}

	// Tokens section.
	if agent.TokenInput > 0 || agent.TokenOutput > 0 {
		b.WriteString("\n")
		b.WriteString(detailTitleStyle.Render("\u26A1 TOKENS"))
		b.WriteString("\n")

		b.WriteString(fmt.Sprintf("  %-12s %s\n",
			detailLabelStyle.Render("Input:"),
			detailValueStyle.Render(formatTokens(agent.TokenInput)),
		))
		b.WriteString(fmt.Sprintf("  %-12s %s\n",
			detailLabelStyle.Render("Output:"),
			detailValueStyle.Render(formatTokens(agent.TokenOutput)),
		))
		b.WriteString(fmt.Sprintf("  %-12s %s\n",
			detailLabelStyle.Render("Total:"),
			detailValueStyle.Render(formatTokens(agent.TokenInput+agent.TokenOutput)),
		))
	}

	// File locks section.
	if len(locks) > 0 {
		b.WriteString("\n")
		b.WriteString(detailTitleStyle.Render(fmt.Sprintf("\U0001F512 File Locks (%d)", len(locks))))
		b.WriteString("\n")

		for _, l := range locks {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				agentQueuedStyle.Render("LOCKED"),
				detailValueStyle.Render(l.FilePath),
			))
		}
	}

	// Artifacts section.
	if len(artifacts) > 0 {
		b.WriteString("\n")
		b.WriteString(detailTitleStyle.Render(fmt.Sprintf("\U0001F4E6 ARTIFACTS (%d)", len(artifacts))))
		b.WriteString("\n")

		for _, a := range artifacts {
			val := a.Value
			maxVal := width - len(a.Key) - 10
			if maxVal > 0 && len(val) > maxVal {
				val = val[:maxVal-3] + "..."
			}
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				detailLabelStyle.Render(a.Key+":"),
				detailValueStyle.Render(val),
			))
		}
	}

	// Files section.
	b.WriteString("\n")
	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("\U0001F4C1 Files Changed (%d)", len(files))))
	b.WriteString("\n")

	if len(files) == 0 {
		b.WriteString(agentIdleStyle.Render("  (none)"))
		b.WriteString("\n")
	}
	for _, f := range files {
		action := f.Action
		diff := fmt.Sprintf("+%d -%d", f.LinesAdded, f.LinesRemoved)
		b.WriteString(fmt.Sprintf("  %-10s %s %s\n",
			agentCompletedStyle.Render(action),
			detailValueStyle.Render(f.FilePath),
			agentIdleStyle.Render(diff),
		))
	}

	// Activity section.
	b.WriteString("\n")
	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("\U0001F4CB Activity Log (%d)", len(activities))))
	b.WriteString("\n")

	if len(activities) == 0 {
		b.WriteString(agentIdleStyle.Render("  (none)"))
		b.WriteString("\n")
	}
	for _, act := range activities {
		ts := act.Timestamp.Format("15:04:05")
		icon := actionIcon(act.Action)
		detail := act.Action
		if act.Detail != nil {
			detail = *act.Detail
		}
		b.WriteString(fmt.Sprintf("  %s %s %s\n", agentIdleStyle.Render(ts), icon, taskLogEntryStyle.Render(detail)))
	}

	return b.String()
}

// Render draws the detail view.
func (d DetailView) Render(width int) string {
	title := detailTitleStyle.Render(fmt.Sprintf("\U0001F50D Agent: %s", d.Agent.ID))

	border := officeSelectedBorderStyle.Width(width)

	return border.Render(
		title + "\n" +
			strings.Repeat("\u2500", width-4) + "\n" +
			d.Viewport.View() + "\n" +
			helpStyle.Render("[esc] back  [j/k] scroll"),
	)
}
