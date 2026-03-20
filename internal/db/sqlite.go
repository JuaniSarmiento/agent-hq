package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/juani/agent-hq/internal/model"

	_ "modernc.org/sqlite"
)

// DB wraps a sql.DB connection to the Agent HQ SQLite database.
type DB struct {
	conn *sql.DB
}

// Open opens or creates the SQLite database at the given path.
func Open(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetAgents returns all agents ordered by status (running first) then started_at.
func (db *DB) GetAgents() ([]model.Agent, error) {
	query := `
		SELECT id, profile, task, status, started_at, finished_at, result_summary, parent_task,
		       model, timeout_ms, retry_count, max_retries, token_input, token_output
		FROM agents
		ORDER BY
			CASE status
				WHEN 'running' THEN 0
				WHEN 'queued' THEN 1
				WHEN 'completed' THEN 2
				WHEN 'failed' THEN 3
			END,
			started_at DESC`

	return db.scanAgents(query)
}

// GetAgentsByProfile returns agents filtered by profile name.
func (db *DB) GetAgentsByProfile(profile string) ([]model.Agent, error) {
	query := `
		SELECT id, profile, task, status, started_at, finished_at, result_summary, parent_task,
		       model, timeout_ms, retry_count, max_retries, token_input, token_output
		FROM agents
		WHERE profile = ?
		ORDER BY
			CASE status
				WHEN 'running' THEN 0
				WHEN 'queued' THEN 1
				WHEN 'completed' THEN 2
				WHEN 'failed' THEN 3
			END,
			started_at DESC`

	return db.scanAgents(query, profile)
}

// GetRecentActivity returns the most recent activity log entries.
func (db *DB) GetRecentActivity(limit int) ([]model.Activity, error) {
	query := `
		SELECT id, agent_id, timestamp, action, detail, file_path
		FROM activity_log
		ORDER BY timestamp DESC
		LIMIT ?`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent activity: %w", err)
	}
	defer rows.Close()

	return scanActivities(rows)
}

// GetAgentActivity returns all activity for a specific agent.
func (db *DB) GetAgentActivity(agentID string) ([]model.Activity, error) {
	query := `
		SELECT id, agent_id, timestamp, action, detail, file_path
		FROM activity_log
		WHERE agent_id = ?
		ORDER BY timestamp DESC`

	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("query agent activity: %w", err)
	}
	defer rows.Close()

	return scanActivities(rows)
}

// GetAgentFiles returns all file changes for a specific agent.
func (db *DB) GetAgentFiles(agentID string) ([]model.FileChange, error) {
	query := `
		SELECT id, agent_id, file_path, action, lines_added, lines_removed, timestamp
		FROM files_changed
		WHERE agent_id = ?
		ORDER BY timestamp DESC`

	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("query agent files: %w", err)
	}
	defer rows.Close()

	var files []model.FileChange
	for rows.Next() {
		var f model.FileChange
		var ts string
		if err := rows.Scan(&f.ID, &f.AgentID, &f.FilePath, &f.Action, &f.LinesAdded, &f.LinesRemoved, &ts); err != nil {
			return nil, fmt.Errorf("scan file change: %w", err)
		}
		f.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		files = append(files, f)
	}

	return files, rows.Err()
}

// GetProfiles returns distinct profile names from all agents.
func (db *DB) GetProfiles() ([]string, error) {
	rows, err := db.conn.Query("SELECT DISTINCT profile FROM agents ORDER BY profile")
	if err != nil {
		return nil, fmt.Errorf("query profiles: %w", err)
	}
	defer rows.Close()

	var profiles []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		profiles = append(profiles, p)
	}

	return profiles, rows.Err()
}

func (db *DB) scanAgents(query string, args ...any) ([]model.Agent, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()

	var agents []model.Agent
	for rows.Next() {
		var a model.Agent
		var startedAt, finishedAt sql.NullString
		var resultSummary, parentTask sql.NullString

		if err := rows.Scan(
			&a.ID, &a.Profile, &a.Task, &a.Status, &startedAt, &finishedAt, &resultSummary, &parentTask,
			&a.Model, &a.TimeoutMs, &a.RetryCount, &a.MaxRetries, &a.TokenInput, &a.TokenOutput,
		); err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}

		if startedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", startedAt.String)
			a.StartedAt = &t
		}
		if finishedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", finishedAt.String)
			a.FinishedAt = &t
		}
		if resultSummary.Valid {
			a.ResultSummary = &resultSummary.String
		}
		if parentTask.Valid {
			a.ParentTask = &parentTask.String
		}

		agents = append(agents, a)
	}

	return agents, rows.Err()
}

// GetDAGEdges returns all dependency edges for a parent task.
func (db *DB) GetDAGEdges(parentTask string) ([]model.DAGEdge, error) {
	query := `SELECT id, parent_task, from_agent_id, to_agent_id FROM dag_edges WHERE parent_task = ?`
	rows, err := db.conn.Query(query, parentTask)
	if err != nil {
		return nil, fmt.Errorf("query dag edges: %w", err)
	}
	defer rows.Close()

	var edges []model.DAGEdge
	for rows.Next() {
		var e model.DAGEdge
		if err := rows.Scan(&e.ID, &e.ParentTask, &e.FromAgentID, &e.ToAgentID); err != nil {
			return nil, fmt.Errorf("scan dag edge: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// GetReadyAgents returns agent IDs whose ALL dependencies are completed.
func (db *DB) GetReadyAgents(parentTask string) ([]string, error) {
	query := `
		SELECT DISTINCT de.to_agent_id
		FROM dag_edges de
		WHERE de.parent_task = ?
		AND NOT EXISTS (
			SELECT 1 FROM dag_edges de2
			JOIN agents a ON a.id = de2.from_agent_id
			WHERE de2.parent_task = de.parent_task
			AND de2.to_agent_id = de.to_agent_id
			AND a.status != 'completed'
		)
		AND EXISTS (
			SELECT 1 FROM agents a2 WHERE a2.id = de.to_agent_id AND a2.status = 'queued'
		)`
	rows, err := db.conn.Query(query, parentTask)
	if err != nil {
		return nil, fmt.Errorf("query ready agents: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan ready agent: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetArtifactsByAgent returns artifacts produced by an agent.
func (db *DB) GetArtifactsByAgent(agentID string) ([]model.Artifact, error) {
	query := `SELECT id, agent_id, key, value, timestamp FROM artifacts WHERE agent_id = ? ORDER BY timestamp DESC`
	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("query artifacts by agent: %w", err)
	}
	defer rows.Close()

	var artifacts []model.Artifact
	for rows.Next() {
		var a model.Artifact
		var ts string
		if err := rows.Scan(&a.ID, &a.AgentID, &a.Key, &a.Value, &ts); err != nil {
			return nil, fmt.Errorf("scan artifact: %w", err)
		}
		a.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		artifacts = append(artifacts, a)
	}
	return artifacts, rows.Err()
}

// GetArtifactsForAgent returns artifacts from all dependency agents via dag_edges.
func (db *DB) GetArtifactsForAgent(agentID string, parentTask string) ([]model.Artifact, error) {
	query := `
		SELECT ar.id, ar.agent_id, ar.key, ar.value, ar.timestamp
		FROM artifacts ar
		JOIN dag_edges de ON de.from_agent_id = ar.agent_id
		WHERE de.to_agent_id = ? AND de.parent_task = ?
		ORDER BY ar.timestamp DESC`
	rows, err := db.conn.Query(query, agentID, parentTask)
	if err != nil {
		return nil, fmt.Errorf("query artifacts for agent: %w", err)
	}
	defer rows.Close()

	var artifacts []model.Artifact
	for rows.Next() {
		var a model.Artifact
		var ts string
		if err := rows.Scan(&a.ID, &a.AgentID, &a.Key, &a.Value, &ts); err != nil {
			return nil, fmt.Errorf("scan artifact: %w", err)
		}
		a.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		artifacts = append(artifacts, a)
	}
	return artifacts, rows.Err()
}

// CheckFileLocks checks which of the given paths are currently locked.
func (db *DB) CheckFileLocks(paths []string) ([]model.FileLock, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	query := "SELECT file_path, agent_id, locked_at FROM file_locks WHERE file_path IN (?" + strings.Repeat(",?", len(paths)-1) + ")"
	args := make([]any, len(paths))
	for i, p := range paths {
		args[i] = p
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query file locks: %w", err)
	}
	defer rows.Close()

	var locks []model.FileLock
	for rows.Next() {
		var l model.FileLock
		var ts string
		if err := rows.Scan(&l.FilePath, &l.AgentID, &ts); err != nil {
			return nil, fmt.Errorf("scan file lock: %w", err)
		}
		l.LockedAt, _ = time.Parse("2006-01-02 15:04:05", ts)
		locks = append(locks, l)
	}
	return locks, rows.Err()
}

// GetQualityGates returns all quality gates for a parent task.
func (db *DB) GetQualityGates(parentTask string) ([]model.QualityGate, error) {
	query := `SELECT id, parent_task, phase, gate_name, command, required, status, output, executed_at
		FROM quality_gates WHERE parent_task = ? ORDER BY id`
	rows, err := db.conn.Query(query, parentTask)
	if err != nil {
		return nil, fmt.Errorf("query quality gates: %w", err)
	}
	defer rows.Close()

	var gates []model.QualityGate
	for rows.Next() {
		var g model.QualityGate
		var reqInt int
		var output, executedAt sql.NullString
		if err := rows.Scan(&g.ID, &g.ParentTask, &g.Phase, &g.GateName, &g.Command, &reqInt, &g.Status, &output, &executedAt); err != nil {
			return nil, fmt.Errorf("scan quality gate: %w", err)
		}
		g.Required = reqInt != 0
		if output.Valid {
			g.Output = &output.String
		}
		if executedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", executedAt.String)
			g.ExecutedAt = &t
		}
		gates = append(gates, g)
	}
	return gates, rows.Err()
}

// GetPipeline returns a pipeline by ID.
func (db *DB) GetPipeline(id string) (*model.Pipeline, error) {
	query := `SELECT id, name, definition, created_at FROM pipelines WHERE id = ?`
	var p model.Pipeline
	var ts string
	err := db.conn.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Definition, &ts)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query pipeline: %w", err)
	}
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", ts)
	return &p, nil
}

// ListPipelines returns all pipeline templates.
func (db *DB) ListPipelines() ([]model.Pipeline, error) {
	query := `SELECT id, name, definition, created_at FROM pipelines ORDER BY created_at DESC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query pipelines: %w", err)
	}
	defer rows.Close()

	var pipelines []model.Pipeline
	for rows.Next() {
		var p model.Pipeline
		var ts string
		if err := rows.Scan(&p.ID, &p.Name, &p.Definition, &ts); err != nil {
			return nil, fmt.Errorf("scan pipeline: %w", err)
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", ts)
		pipelines = append(pipelines, p)
	}
	return pipelines, rows.Err()
}

// GetPipelineRun returns a pipeline run by ID.
func (db *DB) GetPipelineRun(id string) (*model.PipelineRun, error) {
	query := `SELECT id, pipeline_id, status, started_at, finished_at, total_token_input, total_token_output, current_phase
		FROM pipeline_runs WHERE id = ?`
	var r model.PipelineRun
	var startedAt string
	var finishedAt, currentPhase sql.NullString
	err := db.conn.QueryRow(query, id).Scan(
		&r.ID, &r.PipelineID, &r.Status, &startedAt, &finishedAt,
		&r.TotalTokenInput, &r.TotalTokenOutput, &currentPhase,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query pipeline run: %w", err)
	}
	r.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)
	if finishedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", finishedAt.String)
		r.FinishedAt = &t
	}
	if currentPhase.Valid {
		r.CurrentPhase = &currentPhase.String
	}
	return &r, nil
}

// GetPipelineSteps returns all steps for a pipeline run.
func (db *DB) GetPipelineSteps(runID string) ([]model.PipelineStep, error) {
	query := `SELECT id, run_id, agent_id, phase, status, started_at, finished_at, token_input, token_output
		FROM pipeline_steps WHERE run_id = ? ORDER BY id`
	rows, err := db.conn.Query(query, runID)
	if err != nil {
		return nil, fmt.Errorf("query pipeline steps: %w", err)
	}
	defer rows.Close()

	var steps []model.PipelineStep
	for rows.Next() {
		var s model.PipelineStep
		var agentID, startedAt, finishedAt sql.NullString
		if err := rows.Scan(&s.ID, &s.RunID, &agentID, &s.Phase, &s.Status, &startedAt, &finishedAt, &s.TokenInput, &s.TokenOutput); err != nil {
			return nil, fmt.Errorf("scan pipeline step: %w", err)
		}
		if agentID.Valid {
			s.AgentID = &agentID.String
		}
		if startedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", startedAt.String)
			s.StartedAt = &t
		}
		if finishedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", finishedAt.String)
			s.FinishedAt = &t
		}
		steps = append(steps, s)
	}
	return steps, rows.Err()
}

// GetPipelineHistory returns recent runs for a pipeline, ordered by start time.
func (db *DB) GetPipelineHistory(pipelineID string, limit int) ([]model.PipelineRun, error) {
	query := `SELECT id, pipeline_id, status, started_at, finished_at, total_token_input, total_token_output, current_phase
		FROM pipeline_runs WHERE pipeline_id = ? ORDER BY started_at DESC LIMIT ?`
	rows, err := db.conn.Query(query, pipelineID, limit)
	if err != nil {
		return nil, fmt.Errorf("query pipeline history: %w", err)
	}
	defer rows.Close()

	var runs []model.PipelineRun
	for rows.Next() {
		var r model.PipelineRun
		var startedAt string
		var finishedAt, currentPhase sql.NullString
		if err := rows.Scan(
			&r.ID, &r.PipelineID, &r.Status, &startedAt, &finishedAt,
			&r.TotalTokenInput, &r.TotalTokenOutput, &currentPhase,
		); err != nil {
			return nil, fmt.Errorf("scan pipeline run: %w", err)
		}
		r.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)
		if finishedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", finishedAt.String)
			r.FinishedAt = &t
		}
		if currentPhase.Valid {
			r.CurrentPhase = &currentPhase.String
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// GetCostReport aggregates token usage by agent for a parent task.
func (db *DB) GetCostReport(parentTask string) (*model.CostReport, error) {
	query := `
		SELECT id, profile, token_input, token_output, started_at, finished_at
		FROM agents WHERE parent_task = ?`
	rows, err := db.conn.Query(query, parentTask)
	if err != nil {
		return nil, fmt.Errorf("query cost report: %w", err)
	}
	defer rows.Close()

	report := &model.CostReport{ParentTask: parentTask}
	for rows.Next() {
		var ac model.AgentCost
		var startedAt, finishedAt sql.NullString
		if err := rows.Scan(&ac.AgentID, &ac.Profile, &ac.TokenInput, &ac.TokenOutput, &startedAt, &finishedAt); err != nil {
			return nil, fmt.Errorf("scan agent cost: %w", err)
		}
		if startedAt.Valid && finishedAt.Valid {
			st, _ := time.Parse("2006-01-02 15:04:05", startedAt.String)
			ft, _ := time.Parse("2006-01-02 15:04:05", finishedAt.String)
			d := ft.Sub(st)
			ac.Duration = &d
		}
		report.TotalTokenInput += ac.TokenInput
		report.TotalTokenOutput += ac.TokenOutput
		report.AgentCount++
		report.Agents = append(report.Agents, ac)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return report, nil
}

// GetRunningAgentCount returns the count of running agents for a parent task.
func (db *DB) GetRunningAgentCount(parentTask string) (int, error) {
	var count int
	err := db.conn.QueryRow(
		`SELECT COUNT(*) FROM agents WHERE parent_task = ? AND status = 'running'`,
		parentTask,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query running agent count: %w", err)
	}
	return count, nil
}

// HasCyclicDependency detects if adding an edge from fromID to toID would create a cycle.
func (db *DB) HasCyclicDependency(parentTask string, fromID string, toID string) (bool, error) {
	// BFS from toID following edges: if we can reach fromID, adding fromID->toID creates a cycle.
	visited := map[string]bool{toID: true}
	queue := []string{toID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		rows, err := db.conn.Query(
			`SELECT to_agent_id FROM dag_edges WHERE parent_task = ? AND from_agent_id = ?`,
			parentTask, current,
		)
		if err != nil {
			return false, fmt.Errorf("query cycle detection: %w", err)
		}

		for rows.Next() {
			var next string
			if err := rows.Scan(&next); err != nil {
				rows.Close()
				return false, fmt.Errorf("scan cycle detection: %w", err)
			}
			if next == fromID {
				rows.Close()
				return true, nil
			}
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return false, err
		}
	}
	return false, nil
}

// GetAgentByID returns a single agent by ID, or nil if not found.
func (db *DB) GetAgentByID(id string) (*model.Agent, error) {
	query := `
		SELECT id, profile, task, status, started_at, finished_at, result_summary, parent_task,
		       model, timeout_ms, retry_count, max_retries, token_input, token_output
		FROM agents WHERE id = ?`
	agents, err := db.scanAgents(query, id)
	if err != nil {
		return nil, err
	}
	if len(agents) == 0 {
		return nil, nil
	}
	return &agents[0], nil
}

// GetQualityGateByID returns a single quality gate by ID, or nil if not found.
func (db *DB) GetQualityGateByID(id int) (*model.QualityGate, error) {
	query := `SELECT id, parent_task, phase, gate_name, command, required, status, output, executed_at
		FROM quality_gates WHERE id = ?`
	var g model.QualityGate
	var reqInt int
	var output, executedAt sql.NullString
	err := db.conn.QueryRow(query, id).Scan(&g.ID, &g.ParentTask, &g.Phase, &g.GateName, &g.Command, &reqInt, &g.Status, &output, &executedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query quality gate: %w", err)
	}
	g.Required = reqInt != 0
	if output.Valid {
		g.Output = &output.String
	}
	if executedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", executedAt.String)
		g.ExecutedAt = &t
	}
	return &g, nil
}

// GetPipelineRunByPipelineAndAgent finds a running pipeline run that might be associated with a parent task.
func (db *DB) GetPipelineRunByParentTask(parentTask string) (*model.PipelineRun, error) {
	// Convention: pipeline run ID matches the parent task.
	return db.GetPipelineRun(parentTask)
}

func scanActivities(rows *sql.Rows) ([]model.Activity, error) {
	var activities []model.Activity
	for rows.Next() {
		var a model.Activity
		var ts string
		var detail, filePath sql.NullString

		if err := rows.Scan(&a.ID, &a.AgentID, &ts, &a.Action, &detail, &filePath); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}

		a.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		if detail.Valid {
			a.Detail = &detail.String
		}
		if filePath.Valid {
			a.FilePath = &filePath.String
		}

		activities = append(activities, a)
	}

	return activities, rows.Err()
}
