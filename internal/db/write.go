package db

import (
	"fmt"
	"time"
)

// RegisterAgent inserts a new agent with status=running.
func (db *DB) RegisterAgent(id, profile, task string, parentTask *string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	_, err := db.conn.Exec(
		`INSERT INTO agents (id, profile, task, status, started_at, parent_task)
		 VALUES (?, ?, ?, 'running', ?, ?)`,
		id, profile, task, now, parentTask,
	)
	if err != nil {
		return fmt.Errorf("register agent: %w", err)
	}
	return nil
}

// CompleteAgent updates an agent's status and sets finished_at.
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
