package db

import (
	"database/sql"
	"fmt"
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
		SELECT id, profile, task, status, started_at, finished_at, result_summary, parent_task
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
		SELECT id, profile, task, status, started_at, finished_at, result_summary, parent_task
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

		if err := rows.Scan(&a.ID, &a.Profile, &a.Task, &a.Status, &startedAt, &finishedAt, &resultSummary, &parentTask); err != nil {
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
