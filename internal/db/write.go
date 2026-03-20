package db

import (
	"fmt"
	"time"
)

// RegisterAgent inserts a new agent with status=running.
func (db *DB) RegisterAgent(id, profile, task string, parentTask *string, model string, timeoutMs int, maxRetries int) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err := db.conn.Exec(
		`INSERT INTO agents (id, profile, task, status, started_at, parent_task, model, timeout_ms, max_retries)
		 VALUES (?, ?, ?, 'running', ?, ?, ?, ?, ?)`,
		id, profile, task, now, parentTask, model, timeoutMs, maxRetries,
	)
	if err != nil {
		return fmt.Errorf("register agent: %w", err)
	}
	return nil
}

// CompleteAgent updates an agent's status, sets finished_at, and releases all file locks.
func (db *DB) CompleteAgent(id, status string, resultSummary *string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	res, err := db.conn.Exec(
		`UPDATE agents SET status = ?, finished_at = ?, result_summary = ? WHERE id = ?`,
		status, now, resultSummary, id,
	)
	if err != nil {
		return fmt.Errorf("complete agent: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("agent %q not found", id)
	}
	if err := db.ReleaseAllLocks(id); err != nil {
		return fmt.Errorf("release locks on complete: %w", err)
	}
	return nil
}

// LogActivity inserts an activity log entry.
func (db *DB) LogActivity(agentID, action string, detail, filePath *string) error {
	_, err := db.conn.Exec(
		`INSERT INTO activity_log (agent_id, action, detail, file_path)
		 VALUES (?, ?, ?, ?)`,
		agentID, action, detail, filePath,
	)
	if err != nil {
		return fmt.Errorf("log activity: %w", err)
	}
	return nil
}

// LogFileChange inserts a file change record.
func (db *DB) LogFileChange(agentID, filePath, action string) error {
	_, err := db.conn.Exec(
		`INSERT INTO files_changed (agent_id, file_path, action)
		 VALUES (?, ?, ?)`,
		agentID, filePath, action,
	)
	if err != nil {
		return fmt.Errorf("log file change: %w", err)
	}
	return nil
}

// AddDAGEdge inserts a dependency edge with cycle detection.
func (db *DB) AddDAGEdge(parentTask, fromAgentID, toAgentID string) error {
	hasCycle, err := db.HasCyclicDependency(parentTask, fromAgentID, toAgentID)
	if err != nil {
		return fmt.Errorf("check cycle: %w", err)
	}
	if hasCycle {
		return fmt.Errorf("adding edge %s -> %s would create a cycle", fromAgentID, toAgentID)
	}
	_, err = db.conn.Exec(
		`INSERT INTO dag_edges (parent_task, from_agent_id, to_agent_id) VALUES (?, ?, ?)`,
		parentTask, fromAgentID, toAgentID,
	)
	if err != nil {
		return fmt.Errorf("add dag edge: %w", err)
	}
	return nil
}

// SaveArtifact inserts a key-value artifact for an agent.
func (db *DB) SaveArtifact(agentID, key, value string) error {
	_, err := db.conn.Exec(
		`INSERT INTO artifacts (agent_id, key, value) VALUES (?, ?, ?)`,
		agentID, key, value,
	)
	if err != nil {
		return fmt.Errorf("save artifact: %w", err)
	}
	return nil
}

// AcquireFileLock attempts to lock a file path for an agent. Fails if already locked by a different agent.
func (db *DB) AcquireFileLock(filePath, agentID string) error {
	_, err := db.conn.Exec(
		`INSERT INTO file_locks (file_path, agent_id) VALUES (?, ?)`,
		filePath, agentID,
	)
	if err != nil {
		// Check if it's already locked by a different agent.
		var existingAgent string
		qErr := db.conn.QueryRow(
			`SELECT agent_id FROM file_locks WHERE file_path = ?`, filePath,
		).Scan(&existingAgent)
		if qErr == nil && existingAgent == agentID {
			return nil // Already locked by same agent, idempotent.
		}
		if qErr == nil {
			return fmt.Errorf("file %q already locked by agent %q", filePath, existingAgent)
		}
		return fmt.Errorf("acquire file lock: %w", err)
	}
	return nil
}

// ReleaseFileLock releases a file lock held by an agent.
func (db *DB) ReleaseFileLock(filePath, agentID string) error {
	res, err := db.conn.Exec(
		`DELETE FROM file_locks WHERE file_path = ? AND agent_id = ?`,
		filePath, agentID,
	)
	if err != nil {
		return fmt.Errorf("release file lock: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no lock on %q held by agent %q", filePath, agentID)
	}
	return nil
}

// ReleaseAllLocks releases all file locks held by an agent.
func (db *DB) ReleaseAllLocks(agentID string) error {
	_, err := db.conn.Exec(`DELETE FROM file_locks WHERE agent_id = ?`, agentID)
	if err != nil {
		return fmt.Errorf("release all locks: %w", err)
	}
	return nil
}

// CreateQualityGate inserts a new quality gate.
func (db *DB) CreateQualityGate(parentTask, phase, gateName, command string, required bool) error {
	reqInt := 0
	if required {
		reqInt = 1
	}
	_, err := db.conn.Exec(
		`INSERT INTO quality_gates (parent_task, phase, gate_name, command, required) VALUES (?, ?, ?, ?, ?)`,
		parentTask, phase, gateName, command, reqInt,
	)
	if err != nil {
		return fmt.Errorf("create quality gate: %w", err)
	}
	return nil
}

// UpdateQualityGate updates a quality gate's status and output.
func (db *DB) UpdateQualityGate(id int, status string, output string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	res, err := db.conn.Exec(
		`UPDATE quality_gates SET status = ?, output = ?, executed_at = ? WHERE id = ?`,
		status, output, now, id,
	)
	if err != nil {
		return fmt.Errorf("update quality gate: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quality gate %d not found", id)
	}
	return nil
}

// CreatePipeline inserts a new pipeline template.
func (db *DB) CreatePipeline(id, name, definition string) error {
	_, err := db.conn.Exec(
		`INSERT INTO pipelines (id, name, definition) VALUES (?, ?, ?)`,
		id, name, definition,
	)
	if err != nil {
		return fmt.Errorf("create pipeline: %w", err)
	}
	return nil
}

// CreatePipelineRun inserts a new pipeline run.
func (db *DB) CreatePipelineRun(id, pipelineID string) error {
	_, err := db.conn.Exec(
		`INSERT INTO pipeline_runs (id, pipeline_id) VALUES (?, ?)`,
		id, pipelineID,
	)
	if err != nil {
		return fmt.Errorf("create pipeline run: %w", err)
	}
	return nil
}

// UpdatePipelineRun updates a pipeline run's status and current phase.
func (db *DB) UpdatePipelineRun(id, status string, currentPhase *string) error {
	var err error
	if status == "completed" || status == "failed" {
		now := time.Now().UTC().Format("2006-01-02 15:04:05")
		_, err = db.conn.Exec(
			`UPDATE pipeline_runs SET status = ?, current_phase = ?, finished_at = ? WHERE id = ?`,
			status, currentPhase, now, id,
		)
	} else {
		_, err = db.conn.Exec(
			`UPDATE pipeline_runs SET status = ?, current_phase = ? WHERE id = ?`,
			status, currentPhase, id,
		)
	}
	if err != nil {
		return fmt.Errorf("update pipeline run: %w", err)
	}
	return nil
}

// UpdatePipelineRunTokens adds token counts to a pipeline run's totals.
func (db *DB) UpdatePipelineRunTokens(id string, tokenInput, tokenOutput int) error {
	_, err := db.conn.Exec(
		`UPDATE pipeline_runs SET total_token_input = total_token_input + ?, total_token_output = total_token_output + ? WHERE id = ?`,
		tokenInput, tokenOutput, id,
	)
	if err != nil {
		return fmt.Errorf("update pipeline run tokens: %w", err)
	}
	return nil
}

// CreatePipelineStep inserts a new pipeline step and returns its ID.
func (db *DB) CreatePipelineStep(runID, phase string) (int, error) {
	res, err := db.conn.Exec(
		`INSERT INTO pipeline_steps (run_id, phase) VALUES (?, ?)`,
		runID, phase,
	)
	if err != nil {
		return 0, fmt.Errorf("create pipeline step: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get pipeline step id: %w", err)
	}
	return int(id), nil
}

// UpdatePipelineStep updates a pipeline step's agent, status, and token counts.
func (db *DB) UpdatePipelineStep(id int, agentID *string, status string, tokenInput, tokenOutput int) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	var err error
	if status == "running" {
		_, err = db.conn.Exec(
			`UPDATE pipeline_steps SET agent_id = ?, status = ?, started_at = ?, token_input = ?, token_output = ? WHERE id = ?`,
			agentID, status, now, tokenInput, tokenOutput, id,
		)
	} else if status == "completed" || status == "failed" {
		_, err = db.conn.Exec(
			`UPDATE pipeline_steps SET agent_id = ?, status = ?, finished_at = ?, token_input = ?, token_output = ? WHERE id = ?`,
			agentID, status, now, tokenInput, tokenOutput, id,
		)
	} else {
		_, err = db.conn.Exec(
			`UPDATE pipeline_steps SET agent_id = ?, status = ?, token_input = ?, token_output = ? WHERE id = ?`,
			agentID, status, tokenInput, tokenOutput, id,
		)
	}
	if err != nil {
		return fmt.Errorf("update pipeline step: %w", err)
	}
	return nil
}

// UpdateAgentTokens adds token counts to an agent's totals.
func (db *DB) UpdateAgentTokens(agentID string, tokenInput, tokenOutput int) error {
	_, err := db.conn.Exec(
		`UPDATE agents SET token_input = token_input + ?, token_output = token_output + ? WHERE id = ?`,
		tokenInput, tokenOutput, agentID,
	)
	if err != nil {
		return fmt.Errorf("update agent tokens: %w", err)
	}
	return nil
}

// IncrementRetryCount increments the retry count for an agent.
func (db *DB) IncrementRetryCount(agentID string) error {
	_, err := db.conn.Exec(
		`UPDATE agents SET retry_count = retry_count + 1 WHERE id = ?`,
		agentID,
	)
	if err != nil {
		return fmt.Errorf("increment retry count: %w", err)
	}
	return nil
}
