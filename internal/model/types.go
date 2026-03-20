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
	Model         string
	TimeoutMs     int
	RetryCount    int
	MaxRetries    int
	TokenInput    int
	TokenOutput   int
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

// DAGEdge represents a dependency edge between two agents within a parent task.
type DAGEdge struct {
	ID          int
	ParentTask  string
	FromAgentID string
	ToAgentID   string
}

// Artifact represents a key-value context artifact produced by an agent.
type Artifact struct {
	ID        int
	AgentID   string
	Key       string
	Value     string
	Timestamp time.Time
}

// FileLock represents a file lock held by an agent for conflict detection.
type FileLock struct {
	FilePath string
	AgentID  string
	LockedAt time.Time
}

// QualityGate represents a quality gate check within a pipeline phase.
type QualityGate struct {
	ID         int
	ParentTask string
	Phase      string
	GateName   string
	Command    string
	Required   bool
	Status     string // pending, passed, failed
	Output     *string
	ExecutedAt *time.Time
}

// Pipeline represents a reusable pipeline template.
type Pipeline struct {
	ID         string
	Name       string
	Definition string // JSON
	CreatedAt  time.Time
}

// PipelineRun represents a single execution of a pipeline.
type PipelineRun struct {
	ID               string
	PipelineID       string
	Status           string // running, completed, failed, paused
	StartedAt        time.Time
	FinishedAt       *time.Time
	TotalTokenInput  int
	TotalTokenOutput int
	CurrentPhase     *string
}

// PipelineStep represents an individual step within a pipeline run.
type PipelineStep struct {
	ID          int
	RunID       string
	AgentID     *string
	Phase       string
	Status      string // pending, running, completed, failed
	StartedAt   *time.Time
	FinishedAt  *time.Time
	TokenInput  int
	TokenOutput int
}

// CostReport aggregates token usage across agents for a parent task.
type CostReport struct {
	ParentTask       string
	TotalTokenInput  int
	TotalTokenOutput int
	AgentCount       int
	Agents           []AgentCost
}

// AgentCost represents token usage for a single agent.
type AgentCost struct {
	AgentID     string
	Profile     string
	TokenInput  int
	TokenOutput int
	Duration    *time.Duration
}
