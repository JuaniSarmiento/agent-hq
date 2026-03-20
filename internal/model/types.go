package model

import "time"

// AgentStatus represents the current state of an agent.
type AgentStatus string

const (
	StatusQueued    AgentStatus = "queued"
	StatusRunning   AgentStatus = "running"
	StatusCompleted AgentStatus = "completed"
	StatusFailed    AgentStatus = "failed"
)

// Agent represents a Claude Code sub-agent instance.
type Agent struct {
	ID            string
	Profile       string
	Task          string
	Status        AgentStatus
	StartedAt     *time.Time
	FinishedAt    *time.Time
	ResultSummary *string
	ParentTask    *string
}

// Activity represents a single action taken by an agent.
type Activity struct {
	ID        int
	AgentID   string
	Timestamp time.Time
	Action    string
	Detail    *string
	FilePath  *string
}

// FileChange represents a file modification made by an agent.
type FileChange struct {
	ID           int
	AgentID      string
	FilePath     string
	Action       string
	LinesAdded   int
	LinesRemoved int
	Timestamp    time.Time
}
