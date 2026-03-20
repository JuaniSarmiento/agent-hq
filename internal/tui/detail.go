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
	Viewport   viewport.Model
	Ready      bool
}

// NewDetailView creates a detail view for the given agent.
func NewDetailView(agent model.Agent, activities []model.Activity, files []model.FileChange, width, height int) DetailView {
	vp := viewport.New(width-4, height-6)
	vp.SetContent(renderDetailContent(agent, activities, files, width-6))

	return DetailView{
		Agent:      agent,
		Activities: activities,
		Files:      files,
		Viewport:   vp,
		Ready:      true,
	}
}

func renderDetailContent(agent model.Agent, activities []model.Activity, files []model.FileChange, width int) string {
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
