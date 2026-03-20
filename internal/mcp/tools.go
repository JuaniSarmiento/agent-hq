package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/juani/agent-hq/internal/db"
	"github.com/juani/agent-hq/internal/profiles"
)

// ToolDef describes a tool for the MCP tools/list response.
type ToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

// ToolHandler handles MCP tool calls.
type ToolHandler struct {
	db *db.DB
}

// NewToolHandler creates a new tool handler.
func NewToolHandler(database *db.DB) *ToolHandler {
	return &ToolHandler{db: database}
}

// ListTools returns all available tool definitions.
func (h *ToolHandler) ListTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "agent_list_profiles",
			Description: "List all available agent profiles with their role descriptions",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
			},
		},
		{
			Name:        "agent_get_profile",
			Description: "Get the full markdown content of an agent profile",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Profile name (e.g. python-backend, qa-testing)",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "agent_register",
			Description: "Register a new agent with status=running",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "Unique agent ID (e.g. agt-xxx)",
					},
					"profile": map[string]any{
						"type":        "string",
						"description": "Profile name to use",
					},
					"task": map[string]any{
						"type":        "string",
						"description": "Task description",
					},
					"parent_task": map[string]any{
						"type":        "string",
						"description": "Optional parent task for grouping",
					},
				},
				"required": []string{"id", "profile", "task"},
			},
		},
		{
			Name:        "agent_complete",
			Description: "Mark an agent as completed or failed",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "Agent ID to complete",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "Final status: completed or failed",
						"enum":        []string{"completed", "failed"},
					},
					"result_summary": map[string]any{
						"type":        "string",
						"description": "Optional summary of the result",
					},
				},
				"required": []string{"id", "status"},
			},
		},
		{
			Name:        "agent_log_activity",
			Description: "Log an activity for an agent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent_id": map[string]any{
						"type":        "string",
						"description": "Agent ID",
					},
					"action": map[string]any{
						"type":        "string",
						"description": "Action type: file_read, file_write, search, command, etc",
					},
					"detail": map[string]any{
						"type":        "string",
						"description": "Human-readable description",
					},
					"file_path": map[string]any{
						"type":        "string",
						"description": "File path if applicable",
					},
				},
				"required": []string{"agent_id", "action"},
			},
		},
		{
			Name:        "agent_status",
			Description: "Get status of all agents, optionally filtered by profile",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"profile": map[string]any{
						"type":        "string",
						"description": "Optional profile name to filter by",
					},
				},
				"required": []string{},
			},
		},
	}
}

// toolCallParams is the wrapper for tools/call params.
type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// CallTool executes a tool call and returns the MCP result.
func (h *ToolHandler) CallTool(params json.RawMessage) (any, error) {
	var p toolCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid tool call params: %w", err)
	}

	var text string
	var err error

	switch p.Name {
	case "agent_list_profiles":
		text, err = h.listProfiles()
	case "agent_get_profile":
		text, err = h.getProfile(p.Arguments)
	case "agent_register":
		text, err = h.registerAgent(p.Arguments)
	case "agent_complete":
		text, err = h.completeAgent(p.Arguments)
	case "agent_log_activity":
		text, err = h.logActivity(p.Arguments)
	case "agent_status":
		text, err = h.agentStatus(p.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", p.Name)
	}

	if err != nil {
		return nil, err
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": text,
			},
		},
	}, nil
}

func (h *ToolHandler) listProfiles() (string, error) {
	list, err := profiles.List()
	if err != nil {
		return "", fmt.Errorf("list profiles: %w", err)
	}

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal profiles: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) getProfile(args json.RawMessage) (string, error) {
	var params struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	content, err := profiles.Get(params.Name)
	if err != nil {
		return "", err
	}
	return content, nil
}

func (h *ToolHandler) registerAgent(args json.RawMessage) (string, error) {
	var params struct {
		ID         string  `json:"id"`
		Profile    string  `json:"profile"`
		Task       string  `json:"task"`
		ParentTask *string `json:"parent_task,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.ID == "" || params.Profile == "" || params.Task == "" {
		return "", fmt.Errorf("id, profile, and task are required")
	}

	if err := h.db.RegisterAgent(params.ID, params.Profile, params.Task, params.ParentTask); err != nil {
		return "", err
	}

	return fmt.Sprintf("Agent %s registered with profile=%s, status=running", params.ID, params.Profile), nil
}

func (h *ToolHandler) completeAgent(args json.RawMessage) (string, error) {
	var params struct {
		ID            string  `json:"id"`
		Status        string  `json:"status"`
		ResultSummary *string `json:"result_summary,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.ID == "" || params.Status == "" {
		return "", fmt.Errorf("id and status are required")
	}
	if params.Status != "completed" && params.Status != "failed" {
		return "", fmt.Errorf("status must be completed or failed")
	}

	if err := h.db.CompleteAgent(params.ID, params.Status, params.ResultSummary); err != nil {
		return "", err
	}

	return fmt.Sprintf("Agent %s marked as %s", params.ID, params.Status), nil
}

func (h *ToolHandler) logActivity(args json.RawMessage) (string, error) {
	var params struct {
		AgentID  string  `json:"agent_id"`
		Action   string  `json:"action"`
		Detail   *string `json:"detail,omitempty"`
		FilePath *string `json:"file_path,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.AgentID == "" || params.Action == "" {
		return "", fmt.Errorf("agent_id and action are required")
	}

	if err := h.db.LogActivity(params.AgentID, params.Action, params.Detail, params.FilePath); err != nil {
		return "", err
	}

	// Also log to files_changed if it's a file write action.
	if params.FilePath != nil && isFileWriteAction(params.Action) {
		if err := h.db.LogFileChange(params.AgentID, *params.FilePath, params.Action); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Activity logged for agent %s: %s", params.AgentID, params.Action), nil
}

func (h *ToolHandler) agentStatus(args json.RawMessage) (string, error) {
	var params struct {
		Profile string `json:"profile,omitempty"`
	}
	if args != nil && len(args) > 0 {
		_ = json.Unmarshal(args, &params)
	}

	var agents []agentStatusEntry
	var err error

	if params.Profile != "" {
		rawAgents, qErr := h.db.GetAgentsByProfile(params.Profile)
		if qErr != nil {
			return "", qErr
		}
		for _, a := range rawAgents {
			entry := agentStatusEntry{
				ID:      a.ID,
				Profile: a.Profile,
				Task:    a.Task,
				Status:  string(a.Status),
			}
			if a.StartedAt != nil {
				s := a.StartedAt.Format("2006-01-02 15:04:05")
				entry.StartedAt = &s
			}
			if a.FinishedAt != nil {
				s := a.FinishedAt.Format("2006-01-02 15:04:05")
				entry.FinishedAt = &s
			}
			entry.ResultSummary = a.ResultSummary
			entry.ParentTask = a.ParentTask

			// Include recent activity for running agents.
			if a.Status == "running" {
				activities, aErr := h.db.GetAgentActivity(a.ID)
				if aErr == nil && len(activities) > 0 {
					limit := 5
					if len(activities) < limit {
						limit = len(activities)
					}
					for _, act := range activities[:limit] {
						entry.RecentActivity = append(entry.RecentActivity, activityEntry{
							Action:    act.Action,
							Detail:    act.Detail,
							FilePath:  act.FilePath,
							Timestamp: act.Timestamp.Format("2006-01-02 15:04:05"),
						})
					}
				}
			}

			agents = append(agents, entry)
		}
	} else {
		rawAgents, qErr := h.db.GetAgents()
		if qErr != nil {
			return "", qErr
		}
		for _, a := range rawAgents {
			entry := agentStatusEntry{
				ID:      a.ID,
				Profile: a.Profile,
				Task:    a.Task,
				Status:  string(a.Status),
			}
			if a.StartedAt != nil {
				s := a.StartedAt.Format("2006-01-02 15:04:05")
				entry.StartedAt = &s
			}
			if a.FinishedAt != nil {
				s := a.FinishedAt.Format("2006-01-02 15:04:05")
				entry.FinishedAt = &s
			}
			entry.ResultSummary = a.ResultSummary
			entry.ParentTask = a.ParentTask

			if a.Status == "running" {
				activities, aErr := h.db.GetAgentActivity(a.ID)
				if aErr == nil && len(activities) > 0 {
					limit := 5
					if len(activities) < limit {
						limit = len(activities)
					}
					for _, act := range activities[:limit] {
						entry.RecentActivity = append(entry.RecentActivity, activityEntry{
							Action:    act.Action,
							Detail:    act.Detail,
							FilePath:  act.FilePath,
							Timestamp: act.Timestamp.Format("2006-01-02 15:04:05"),
						})
					}
				}
			}

			agents = append(agents, entry)
		}
	}

	// Group by profile.
	grouped := make(map[string][]agentStatusEntry)
	for _, a := range agents {
		grouped[a.Profile] = append(grouped[a.Profile], a)
	}

	data, err := json.MarshalIndent(grouped, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal status: %w", err)
	}
	return string(data), nil
}

type agentStatusEntry struct {
	ID             string          `json:"id"`
	Profile        string          `json:"profile"`
	Task           string          `json:"task"`
	Status         string          `json:"status"`
	StartedAt      *string         `json:"started_at,omitempty"`
	FinishedAt     *string         `json:"finished_at,omitempty"`
	ResultSummary  *string         `json:"result_summary,omitempty"`
	ParentTask     *string         `json:"parent_task,omitempty"`
	RecentActivity []activityEntry `json:"recent_activity,omitempty"`
}

type activityEntry struct {
	Action    string  `json:"action"`
	Detail    *string `json:"detail,omitempty"`
	FilePath  *string `json:"file_path,omitempty"`
	Timestamp string  `json:"timestamp"`
}

func isFileWriteAction(action string) bool {
	lower := strings.ToLower(action)
	return lower == "file_write" || lower == "file_create" || lower == "file_delete"
}
