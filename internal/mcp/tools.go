package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
		// --- TIER 1: Core Spawn + Tokens ---
		{
			Name:        "agent_spawn",
			Description: "Prepare an agent for launch with concurrency checks, profile resolution, and optional DAG dependencies",
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
					"model": map[string]any{
						"type":        "string",
						"description": "Model to use (default: sonnet)",
					},
					"timeout_ms": map[string]any{
						"type":        "integer",
						"description": "Timeout in milliseconds (default: 600000)",
					},
					"max_retries": map[string]any{
						"type":        "integer",
						"description": "Maximum retries (default: 3)",
					},
					"depends_on": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Agent IDs this agent depends on",
					},
				},
				"required": []string{"id", "profile", "task"},
			},
		},
		{
			Name:        "agent_spawn_batch",
			Description: "Spawn multiple agents at once with concurrency validation",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agents": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"id": map[string]any{
									"type":        "string",
									"description": "Unique agent ID",
								},
								"profile": map[string]any{
									"type":        "string",
									"description": "Profile name to use",
								},
								"task": map[string]any{
									"type":        "string",
									"description": "Task description",
								},
								"model": map[string]any{
									"type":        "string",
									"description": "Model to use (default: sonnet)",
								},
								"timeout_ms": map[string]any{
									"type":        "integer",
									"description": "Timeout in milliseconds (default: 600000)",
								},
								"max_retries": map[string]any{
									"type":        "integer",
									"description": "Maximum retries (default: 3)",
								},
								"depends_on": map[string]any{
									"type":        "array",
									"items":       map[string]any{"type": "string"},
									"description": "Agent IDs this agent depends on",
								},
							},
							"required": []string{"id", "profile", "task"},
						},
						"description": "Array of agents to spawn",
					},
					"parent_task": map[string]any{
						"type":        "string",
						"description": "Parent task for grouping all agents",
					},
				},
				"required": []string{"agents", "parent_task"},
			},
		},
		{
			Name:        "agent_update_tokens",
			Description: "Record token usage for an agent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent_id": map[string]any{
						"type":        "string",
						"description": "Agent ID",
					},
					"token_input": map[string]any{
						"type":        "integer",
						"description": "Input tokens consumed",
					},
					"token_output": map[string]any{
						"type":        "integer",
						"description": "Output tokens consumed",
					},
				},
				"required": []string{"agent_id", "token_input", "token_output"},
			},
		},
		{
			Name:        "agent_cost",
			Description: "Get cost summary for a task group or single agent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"parent_task": map[string]any{
						"type":        "string",
						"description": "Parent task to get cost report for",
					},
					"agent_id": map[string]any{
						"type":        "string",
						"description": "Single agent ID to get cost for",
					},
				},
				"required": []string{},
			},
		},
		// --- TIER 2: DAG + Context + Safety ---
		{
			Name:        "dag_define",
			Description: "Define dependency graph edges between agents",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"parent_task": map[string]any{
						"type":        "string",
						"description": "Parent task for the DAG",
					},
					"edges": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"from": map[string]any{
									"type":        "string",
									"description": "Source agent ID",
								},
								"to": map[string]any{
									"type":        "string",
									"description": "Target agent ID",
								},
							},
							"required": []string{"from", "to"},
						},
						"description": "Array of dependency edges",
					},
				},
				"required": []string{"parent_task", "edges"},
			},
		},
		{
			Name:        "dag_next",
			Description: "Get next agents ready to run based on dependency graph",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"parent_task": map[string]any{
						"type":        "string",
						"description": "Parent task for the DAG",
					},
				},
				"required": []string{"parent_task"},
			},
		},
		{
			Name:        "artifact_put",
			Description: "Store an output artifact for an agent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent_id": map[string]any{
						"type":        "string",
						"description": "Agent ID that produced the artifact",
					},
					"key": map[string]any{
						"type":        "string",
						"description": "Artifact key name",
					},
					"value": map[string]any{
						"type":        "string",
						"description": "Artifact value content",
					},
				},
				"required": []string{"agent_id", "key", "value"},
			},
		},
		{
			Name:        "artifact_get",
			Description: "Get artifacts from an agent's dependencies",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent_id": map[string]any{
						"type":        "string",
						"description": "Agent ID to get dependency artifacts for",
					},
					"parent_task": map[string]any{
						"type":        "string",
						"description": "Parent task for DAG context",
					},
				},
				"required": []string{"agent_id", "parent_task"},
			},
		},
		{
			Name:        "file_lock_check",
			Description: "Pre-spawn conflict detection for file locks",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"files": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "File paths to check for locks",
					},
					"agent_id": map[string]any{
						"type":        "string",
						"description": "Optional agent ID to exclude from conflict results",
					},
				},
				"required": []string{"files"},
			},
		},
		{
			Name:        "file_lock_manage",
			Description: "Acquire or release file locks for an agent",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"description": "Lock action to perform",
						"enum":        []string{"acquire", "release", "release_all"},
					},
					"agent_id": map[string]any{
						"type":        "string",
						"description": "Agent ID performing the action",
					},
					"files": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "File paths (required for acquire/release)",
					},
				},
				"required": []string{"action", "agent_id"},
			},
		},
		{
			Name:        "gate_define",
			Description: "Define a quality gate for a pipeline phase",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"parent_task": map[string]any{
						"type":        "string",
						"description": "Parent task for the gate",
					},
					"phase": map[string]any{
						"type":        "string",
						"description": "Pipeline phase name",
					},
					"gate_name": map[string]any{
						"type":        "string",
						"description": "Quality gate name",
					},
					"command": map[string]any{
						"type":        "string",
						"description": "Command to execute for the gate check",
					},
					"required": map[string]any{
						"type":        "boolean",
						"description": "Whether this gate is required to pass (default: true)",
					},
				},
				"required": []string{"parent_task", "phase", "gate_name", "command"},
			},
		},
		{
			Name:        "gate_report",
			Description: "Report a quality gate result and get overall gate status",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"gate_id": map[string]any{
						"type":        "integer",
						"description": "Quality gate ID",
					},
					"status": map[string]any{
						"type":        "string",
						"description": "Gate result status",
						"enum":        []string{"passed", "failed"},
					},
					"output": map[string]any{
						"type":        "string",
						"description": "Optional gate execution output",
					},
				},
				"required": []string{"gate_id", "status"},
			},
		},
		// --- TIER 3: Pipelines ---
		{
			Name:        "pipeline_create",
			Description: "Create a reusable pipeline template",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{
						"type":        "string",
						"description": "Unique pipeline ID",
					},
					"name": map[string]any{
						"type":        "string",
						"description": "Human-readable pipeline name",
					},
					"definition": map[string]any{
						"type":        "object",
						"description": "Pipeline definition with phases array",
					},
				},
				"required": []string{"id", "name", "definition"},
			},
		},
		{
			Name:        "pipeline_run",
			Description: "Start a pipeline execution",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pipeline_id": map[string]any{
						"type":        "string",
						"description": "Pipeline template ID to run",
					},
					"run_id": map[string]any{
						"type":        "string",
						"description": "Unique run ID for this execution",
					},
				},
				"required": []string{"pipeline_id", "run_id"},
			},
		},
		{
			Name:        "pipeline_status",
			Description: "Get pipeline run status with step details",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"run_id": map[string]any{
						"type":        "string",
						"description": "Pipeline run ID",
					},
				},
				"required": []string{"run_id"},
			},
		},
		{
			Name:        "pipeline_history",
			Description: "Get past pipeline runs",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pipeline_id": map[string]any{
						"type":        "string",
						"description": "Pipeline template ID",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of runs to return (default: 10)",
					},
				},
				"required": []string{"pipeline_id"},
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
	// TIER 1
	case "agent_spawn":
		text, err = h.agentSpawn(p.Arguments)
	case "agent_spawn_batch":
		text, err = h.agentSpawnBatch(p.Arguments)
	case "agent_update_tokens":
		text, err = h.agentUpdateTokens(p.Arguments)
	case "agent_cost":
		text, err = h.agentCost(p.Arguments)
	// TIER 2
	case "dag_define":
		text, err = h.dagDefine(p.Arguments)
	case "dag_next":
		text, err = h.dagNext(p.Arguments)
	case "artifact_put":
		text, err = h.artifactPut(p.Arguments)
	case "artifact_get":
		text, err = h.artifactGet(p.Arguments)
	case "file_lock_check":
		text, err = h.fileLockCheck(p.Arguments)
	case "file_lock_manage":
		text, err = h.fileLockManage(p.Arguments)
	case "gate_define":
		text, err = h.gateDefine(p.Arguments)
	case "gate_report":
		text, err = h.gateReport(p.Arguments)
	// TIER 3
	case "pipeline_create":
		text, err = h.pipelineCreate(p.Arguments)
	case "pipeline_run":
		text, err = h.pipelineRun(p.Arguments)
	case "pipeline_status":
		text, err = h.pipelineStatus(p.Arguments)
	case "pipeline_history":
		text, err = h.pipelineHistory(p.Arguments)
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

// --- Existing tool handlers ---

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
		Model      string  `json:"model,omitempty"`
		TimeoutMs  int     `json:"timeout_ms,omitempty"`
		MaxRetries int     `json:"max_retries,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.ID == "" || params.Profile == "" || params.Task == "" {
		return "", fmt.Errorf("id, profile, and task are required")
	}
	if params.Model == "" {
		params.Model = "sonnet"
	}
	if params.TimeoutMs == 0 {
		params.TimeoutMs = 600000
	}
	if params.MaxRetries == 0 {
		params.MaxRetries = 3
	}

	if err := h.db.RegisterAgent(params.ID, params.Profile, params.Task, params.ParentTask, params.Model, params.TimeoutMs, params.MaxRetries); err != nil {
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

// --- TIER 1: Core Spawn + Tokens ---

func (h *ToolHandler) agentSpawn(args json.RawMessage) (string, error) {
	var params struct {
		ID         string   `json:"id"`
		Profile    string   `json:"profile"`
		Task       string   `json:"task"`
		ParentTask *string  `json:"parent_task,omitempty"`
		Model      string   `json:"model,omitempty"`
		TimeoutMs  int      `json:"timeout_ms,omitempty"`
		MaxRetries int      `json:"max_retries,omitempty"`
		DependsOn  []string `json:"depends_on,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.ID == "" || params.Profile == "" || params.Task == "" {
		return "", fmt.Errorf("id, profile, and task are required")
	}
	if params.Model == "" {
		params.Model = "sonnet"
	}
	if params.TimeoutMs == 0 {
		params.TimeoutMs = 600000
	}
	if params.MaxRetries == 0 {
		params.MaxRetries = 3
	}

	// Check concurrency limit.
	maxAgents := getMaxAgents()
	if params.ParentTask != nil {
		count, err := h.db.GetRunningAgentCount(*params.ParentTask)
		if err != nil {
			return "", fmt.Errorf("check running agents: %w", err)
		}
		if count >= maxAgents {
			return "", fmt.Errorf("max concurrent agents reached (%d), wait for one to complete", maxAgents)
		}
	}

	// Resolve profile content.
	profileContent, err := profiles.Get(params.Profile)
	if err != nil {
		return "", fmt.Errorf("profile %q not found: %w", params.Profile, err)
	}

	// Register the agent.
	if err := h.db.RegisterAgent(params.ID, params.Profile, params.Task, params.ParentTask, params.Model, params.TimeoutMs, params.MaxRetries); err != nil {
		return "", err
	}

	// Add DAG edges if depends_on provided.
	if len(params.DependsOn) > 0 && params.ParentTask != nil {
		for _, depID := range params.DependsOn {
			if err := h.db.AddDAGEdge(*params.ParentTask, depID, params.ID); err != nil {
				return "", fmt.Errorf("add dependency %s -> %s: %w", depID, params.ID, err)
			}
		}
	}

	result := map[string]any{
		"agent_id":        params.ID,
		"profile_content": profileContent,
		"model":           params.Model,
		"timeout_ms":      params.TimeoutMs,
		"status":          "spawned",
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal spawn result: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) agentSpawnBatch(args json.RawMessage) (string, error) {
	var params struct {
		Agents []struct {
			ID         string   `json:"id"`
			Profile    string   `json:"profile"`
			Task       string   `json:"task"`
			Model      string   `json:"model,omitempty"`
			TimeoutMs  int      `json:"timeout_ms,omitempty"`
			MaxRetries int      `json:"max_retries,omitempty"`
			DependsOn  []string `json:"depends_on,omitempty"`
		} `json:"agents"`
		ParentTask string `json:"parent_task"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if len(params.Agents) == 0 {
		return "", fmt.Errorf("agents array is required and must not be empty")
	}
	if params.ParentTask == "" {
		return "", fmt.Errorf("parent_task is required")
	}

	// Check concurrency limit for entire batch.
	maxAgents := getMaxAgents()
	count, err := h.db.GetRunningAgentCount(params.ParentTask)
	if err != nil {
		return "", fmt.Errorf("check running agents: %w", err)
	}
	if count+len(params.Agents) > maxAgents {
		return "", fmt.Errorf("batch of %d would exceed max concurrent agents (%d running, %d max)", len(params.Agents), count, maxAgents)
	}

	var results []map[string]any
	parentTask := params.ParentTask

	for _, a := range params.Agents {
		if a.ID == "" || a.Profile == "" || a.Task == "" {
			return "", fmt.Errorf("each agent requires id, profile, and task")
		}
		if a.Model == "" {
			a.Model = "sonnet"
		}
		if a.TimeoutMs == 0 {
			a.TimeoutMs = 600000
		}
		if a.MaxRetries == 0 {
			a.MaxRetries = 3
		}

		// Resolve profile content.
		profileContent, err := profiles.Get(a.Profile)
		if err != nil {
			return "", fmt.Errorf("profile %q not found for agent %s: %w", a.Profile, a.ID, err)
		}

		// Register the agent.
		if err := h.db.RegisterAgent(a.ID, a.Profile, a.Task, &parentTask, a.Model, a.TimeoutMs, a.MaxRetries); err != nil {
			return "", fmt.Errorf("register agent %s: %w", a.ID, err)
		}

		// Add DAG edges if depends_on provided.
		for _, depID := range a.DependsOn {
			if err := h.db.AddDAGEdge(parentTask, depID, a.ID); err != nil {
				return "", fmt.Errorf("add dependency %s -> %s: %w", depID, a.ID, err)
			}
		}

		results = append(results, map[string]any{
			"agent_id":        a.ID,
			"profile_content": profileContent,
			"model":           a.Model,
			"timeout_ms":      a.TimeoutMs,
			"status":          "spawned",
		})
	}

	data, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("marshal batch result: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) agentUpdateTokens(args json.RawMessage) (string, error) {
	var params struct {
		AgentID     string `json:"agent_id"`
		TokenInput  int    `json:"token_input"`
		TokenOutput int    `json:"token_output"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.AgentID == "" {
		return "", fmt.Errorf("agent_id is required")
	}

	if err := h.db.UpdateAgentTokens(params.AgentID, params.TokenInput, params.TokenOutput); err != nil {
		return "", err
	}

	// If agent has a parent_task, check if there's a pipeline run to update too.
	agent, err := h.db.GetAgentByID(params.AgentID)
	if err != nil {
		return "", fmt.Errorf("get agent: %w", err)
	}
	if agent != nil && agent.ParentTask != nil {
		// Try to update pipeline run tokens if one exists for this parent task.
		run, err := h.db.GetPipelineRunByParentTask(*agent.ParentTask)
		if err == nil && run != nil {
			_ = h.db.UpdatePipelineRunTokens(run.ID, params.TokenInput, params.TokenOutput)
		}
	}

	// Get updated totals.
	updatedAgent, err := h.db.GetAgentByID(params.AgentID)
	if err != nil {
		return "", fmt.Errorf("get updated agent: %w", err)
	}

	result := map[string]any{
		"agent_id":     params.AgentID,
		"token_input":  updatedAgent.TokenInput,
		"token_output": updatedAgent.TokenOutput,
		"status":       "updated",
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal token result: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) agentCost(args json.RawMessage) (string, error) {
	var params struct {
		ParentTask *string `json:"parent_task,omitempty"`
		AgentID    *string `json:"agent_id,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}

	if params.ParentTask != nil && *params.ParentTask != "" {
		report, err := h.db.GetCostReport(*params.ParentTask)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal cost report: %w", err)
		}
		return string(data), nil
	}

	if params.AgentID != nil && *params.AgentID != "" {
		agent, err := h.db.GetAgentByID(*params.AgentID)
		if err != nil {
			return "", err
		}
		if agent == nil {
			return "", fmt.Errorf("agent %q not found", *params.AgentID)
		}
		result := map[string]any{
			"agent_id":     agent.ID,
			"profile":      agent.Profile,
			"token_input":  agent.TokenInput,
			"token_output": agent.TokenOutput,
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal agent cost: %w", err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("either parent_task or agent_id is required")
}

// --- TIER 2: DAG + Context + Safety ---

func (h *ToolHandler) dagDefine(args json.RawMessage) (string, error) {
	var params struct {
		ParentTask string `json:"parent_task"`
		Edges      []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"edges"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.ParentTask == "" {
		return "", fmt.Errorf("parent_task is required")
	}
	if len(params.Edges) == 0 {
		return "", fmt.Errorf("edges array is required and must not be empty")
	}

	for _, edge := range params.Edges {
		if edge.From == "" || edge.To == "" {
			return "", fmt.Errorf("each edge requires from and to fields")
		}
		if err := h.db.AddDAGEdge(params.ParentTask, edge.From, edge.To); err != nil {
			return "", fmt.Errorf("add edge %s -> %s: %w", edge.From, edge.To, err)
		}
	}

	result := map[string]any{
		"parent_task": params.ParentTask,
		"edges_added": len(params.Edges),
		"status":      "defined",
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal dag result: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) dagNext(args json.RawMessage) (string, error) {
	var params struct {
		ParentTask string `json:"parent_task"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.ParentTask == "" {
		return "", fmt.Errorf("parent_task is required")
	}

	readyIDs, err := h.db.GetReadyAgents(params.ParentTask)
	if err != nil {
		return "", err
	}

	runningCount, err := h.db.GetRunningAgentCount(params.ParentTask)
	if err != nil {
		return "", fmt.Errorf("get running count: %w", err)
	}

	maxAgents := getMaxAgents()

	result := map[string]any{
		"ready_agents":  readyIDs,
		"running_count": runningCount,
		"max_agents":    maxAgents,
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal dag next: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) artifactPut(args json.RawMessage) (string, error) {
	var params struct {
		AgentID string `json:"agent_id"`
		Key     string `json:"key"`
		Value   string `json:"value"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.AgentID == "" || params.Key == "" || params.Value == "" {
		return "", fmt.Errorf("agent_id, key, and value are required")
	}

	if err := h.db.SaveArtifact(params.AgentID, params.Key, params.Value); err != nil {
		return "", err
	}

	return fmt.Sprintf("Artifact %q saved for agent %s", params.Key, params.AgentID), nil
}

func (h *ToolHandler) artifactGet(args json.RawMessage) (string, error) {
	var params struct {
		AgentID    string `json:"agent_id"`
		ParentTask string `json:"parent_task"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.AgentID == "" || params.ParentTask == "" {
		return "", fmt.Errorf("agent_id and parent_task are required")
	}

	artifacts, err := h.db.GetArtifactsForAgent(params.AgentID, params.ParentTask)
	if err != nil {
		return "", err
	}

	var result []map[string]any
	for _, a := range artifacts {
		result = append(result, map[string]any{
			"key":             a.Key,
			"value":           a.Value,
			"source_agent_id": a.AgentID,
		})
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal artifacts: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) fileLockCheck(args json.RawMessage) (string, error) {
	var params struct {
		Files   []string `json:"files"`
		AgentID *string  `json:"agent_id,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if len(params.Files) == 0 {
		return "", fmt.Errorf("files array is required and must not be empty")
	}

	locks, err := h.db.CheckFileLocks(params.Files)
	if err != nil {
		return "", err
	}

	// Filter out own locks if agent_id provided.
	var conflicts []map[string]any
	for _, l := range locks {
		if params.AgentID != nil && l.AgentID == *params.AgentID {
			continue
		}
		conflicts = append(conflicts, map[string]any{
			"file_path": l.FilePath,
			"locked_by": l.AgentID,
		})
	}

	result := map[string]any{
		"conflicts":     conflicts,
		"has_conflicts": len(conflicts) > 0,
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal lock check: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) fileLockManage(args json.RawMessage) (string, error) {
	var params struct {
		Action  string   `json:"action"`
		AgentID string   `json:"agent_id"`
		Files   []string `json:"files,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.Action == "" || params.AgentID == "" {
		return "", fmt.Errorf("action and agent_id are required")
	}

	switch params.Action {
	case "acquire":
		if len(params.Files) == 0 {
			return "", fmt.Errorf("files are required for acquire action")
		}
		for _, f := range params.Files {
			if err := h.db.AcquireFileLock(f, params.AgentID); err != nil {
				return "", fmt.Errorf("acquire lock on %s: %w", f, err)
			}
		}
		return fmt.Sprintf("Acquired locks on %d files for agent %s", len(params.Files), params.AgentID), nil

	case "release":
		if len(params.Files) == 0 {
			return "", fmt.Errorf("files are required for release action")
		}
		for _, f := range params.Files {
			if err := h.db.ReleaseFileLock(f, params.AgentID); err != nil {
				return "", fmt.Errorf("release lock on %s: %w", f, err)
			}
		}
		return fmt.Sprintf("Released locks on %d files for agent %s", len(params.Files), params.AgentID), nil

	case "release_all":
		if err := h.db.ReleaseAllLocks(params.AgentID); err != nil {
			return "", err
		}
		return fmt.Sprintf("Released all locks for agent %s", params.AgentID), nil

	default:
		return "", fmt.Errorf("action must be acquire, release, or release_all")
	}
}

func (h *ToolHandler) gateDefine(args json.RawMessage) (string, error) {
	var params struct {
		ParentTask string `json:"parent_task"`
		Phase      string `json:"phase"`
		GateName   string `json:"gate_name"`
		Command    string `json:"command"`
		Required   *bool  `json:"required,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.ParentTask == "" || params.Phase == "" || params.GateName == "" || params.Command == "" {
		return "", fmt.Errorf("parent_task, phase, gate_name, and command are required")
	}

	required := true
	if params.Required != nil {
		required = *params.Required
	}

	if err := h.db.CreateQualityGate(params.ParentTask, params.Phase, params.GateName, params.Command, required); err != nil {
		return "", err
	}

	return fmt.Sprintf("Quality gate %q defined for phase %q in task %s", params.GateName, params.Phase, params.ParentTask), nil
}

func (h *ToolHandler) gateReport(args json.RawMessage) (string, error) {
	var params struct {
		GateID int    `json:"gate_id"`
		Status string `json:"status"`
		Output string `json:"output,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.GateID == 0 {
		return "", fmt.Errorf("gate_id is required")
	}
	if params.Status != "passed" && params.Status != "failed" {
		return "", fmt.Errorf("status must be passed or failed")
	}

	// Update the gate.
	if err := h.db.UpdateQualityGate(params.GateID, params.Status, params.Output); err != nil {
		return "", err
	}

	// Get the gate to find its parent_task.
	gate, err := h.db.GetQualityGateByID(params.GateID)
	if err != nil {
		return "", fmt.Errorf("get gate: %w", err)
	}
	if gate == nil {
		return "", fmt.Errorf("gate %d not found after update", params.GateID)
	}

	// Get all gates for the same parent_task.
	allGates, err := h.db.GetQualityGates(gate.ParentTask)
	if err != nil {
		return "", fmt.Errorf("get all gates: %w", err)
	}

	allPassed := true
	var pendingGates []string
	var failedGates []string
	for _, g := range allGates {
		if g.Status == "pending" {
			allPassed = false
			pendingGates = append(pendingGates, g.GateName)
		} else if g.Status == "failed" && g.Required {
			allPassed = false
			failedGates = append(failedGates, g.GateName)
		}
	}

	result := map[string]any{
		"gate_id":          params.GateID,
		"gate_name":        gate.GateName,
		"status":           params.Status,
		"all_gates_passed": allPassed,
		"pending_gates":    pendingGates,
		"failed_gates":     failedGates,
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal gate report: %w", err)
	}
	return string(data), nil
}

// --- TIER 3: Pipelines ---

func (h *ToolHandler) pipelineCreate(args json.RawMessage) (string, error) {
	var params struct {
		ID         string          `json:"id"`
		Name       string          `json:"name"`
		Definition json.RawMessage `json:"definition"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.ID == "" || params.Name == "" || len(params.Definition) == 0 {
		return "", fmt.Errorf("id, name, and definition are required")
	}

	// Validate definition is valid JSON.
	var defCheck any
	if err := json.Unmarshal(params.Definition, &defCheck); err != nil {
		return "", fmt.Errorf("definition must be valid JSON: %w", err)
	}

	if err := h.db.CreatePipeline(params.ID, params.Name, string(params.Definition)); err != nil {
		return "", err
	}

	return fmt.Sprintf("Pipeline %q (%s) created", params.Name, params.ID), nil
}

func (h *ToolHandler) pipelineRun(args json.RawMessage) (string, error) {
	var params struct {
		PipelineID string `json:"pipeline_id"`
		RunID      string `json:"run_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.PipelineID == "" || params.RunID == "" {
		return "", fmt.Errorf("pipeline_id and run_id are required")
	}

	// Load pipeline definition.
	pipeline, err := h.db.GetPipeline(params.PipelineID)
	if err != nil {
		return "", fmt.Errorf("get pipeline: %w", err)
	}
	if pipeline == nil {
		return "", fmt.Errorf("pipeline %q not found", params.PipelineID)
	}

	// Create pipeline run.
	if err := h.db.CreatePipelineRun(params.RunID, params.PipelineID); err != nil {
		return "", err
	}

	// Parse definition to get phases.
	var definition struct {
		Phases []struct {
			Name string `json:"name"`
		} `json:"phases"`
	}
	if err := json.Unmarshal([]byte(pipeline.Definition), &definition); err != nil {
		return "", fmt.Errorf("parse pipeline definition: %w", err)
	}

	// Create steps for each phase.
	var phases []string
	for _, phase := range definition.Phases {
		if _, err := h.db.CreatePipelineStep(params.RunID, phase.Name); err != nil {
			return "", fmt.Errorf("create step for phase %q: %w", phase.Name, err)
		}
		phases = append(phases, phase.Name)
	}

	// Set current phase to the first one.
	var firstPhase *string
	if len(phases) > 0 {
		firstPhase = &phases[0]
		if err := h.db.UpdatePipelineRun(params.RunID, "running", firstPhase); err != nil {
			return "", fmt.Errorf("update run phase: %w", err)
		}
	}

	result := map[string]any{
		"run_id":      params.RunID,
		"pipeline_id": params.PipelineID,
		"phases":      phases,
		"first_phase": firstPhase,
		"status":      "running",
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal run result: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) pipelineStatus(args json.RawMessage) (string, error) {
	var params struct {
		RunID string `json:"run_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.RunID == "" {
		return "", fmt.Errorf("run_id is required")
	}

	run, err := h.db.GetPipelineRun(params.RunID)
	if err != nil {
		return "", fmt.Errorf("get pipeline run: %w", err)
	}
	if run == nil {
		return "", fmt.Errorf("pipeline run %q not found", params.RunID)
	}

	steps, err := h.db.GetPipelineSteps(params.RunID)
	if err != nil {
		return "", fmt.Errorf("get pipeline steps: %w", err)
	}

	var stepResults []map[string]any
	for _, s := range steps {
		step := map[string]any{
			"id":           s.ID,
			"phase":        s.Phase,
			"status":       s.Status,
			"token_input":  s.TokenInput,
			"token_output": s.TokenOutput,
		}
		if s.AgentID != nil {
			step["agent_id"] = *s.AgentID
		}
		if s.StartedAt != nil {
			step["started_at"] = s.StartedAt.Format("2006-01-02 15:04:05")
		}
		if s.FinishedAt != nil {
			step["finished_at"] = s.FinishedAt.Format("2006-01-02 15:04:05")
		}
		stepResults = append(stepResults, step)
	}

	result := map[string]any{
		"run_id":             run.ID,
		"pipeline_id":        run.PipelineID,
		"status":             run.Status,
		"started_at":         run.StartedAt.Format("2006-01-02 15:04:05"),
		"total_token_input":  run.TotalTokenInput,
		"total_token_output": run.TotalTokenOutput,
		"current_phase":      run.CurrentPhase,
		"steps":              stepResults,
	}
	if run.FinishedAt != nil {
		result["finished_at"] = run.FinishedAt.Format("2006-01-02 15:04:05")
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal pipeline status: %w", err)
	}
	return string(data), nil
}

func (h *ToolHandler) pipelineHistory(args json.RawMessage) (string, error) {
	var params struct {
		PipelineID string `json:"pipeline_id"`
		Limit      int    `json:"limit,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if params.PipelineID == "" {
		return "", fmt.Errorf("pipeline_id is required")
	}
	if params.Limit == 0 {
		params.Limit = 10
	}

	runs, err := h.db.GetPipelineHistory(params.PipelineID, params.Limit)
	if err != nil {
		return "", err
	}

	var results []map[string]any
	for _, r := range runs {
		entry := map[string]any{
			"run_id":             r.ID,
			"status":             r.Status,
			"started_at":         r.StartedAt.Format("2006-01-02 15:04:05"),
			"total_token_input":  r.TotalTokenInput,
			"total_token_output": r.TotalTokenOutput,
		}
		if r.FinishedAt != nil {
			entry["finished_at"] = r.FinishedAt.Format("2006-01-02 15:04:05")
			duration := r.FinishedAt.Sub(r.StartedAt).String()
			entry["duration"] = duration
		}
		if r.CurrentPhase != nil {
			entry["current_phase"] = *r.CurrentPhase
		}
		results = append(results, entry)
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal pipeline history: %w", err)
	}
	return string(data), nil
}

// --- Types and helpers ---

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

// getMaxAgents reads AGENTHQ_MAX_AGENTS env var, defaulting to 3.
func getMaxAgents() int {
	envVal := os.Getenv("AGENTHQ_MAX_AGENTS")
	if envVal == "" {
		return 3
	}
	val, err := strconv.Atoi(envVal)
	if err != nil || val <= 0 {
		return 3
	}
	return val
}
