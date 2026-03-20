-- Agent HQ Database Schema
-- Tracks agent activity for the Agent HQ TUI dashboard

-- Agents table: tracks each agent instance
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,           -- unique agent ID
    profile TEXT NOT NULL,         -- profile name (python-backend, qa-testing, etc)
    task TEXT NOT NULL,            -- what was assigned
    status TEXT NOT NULL DEFAULT 'queued',  -- queued, running, completed, failed
    started_at DATETIME,
    finished_at DATETIME,
    result_summary TEXT,           -- brief result when done
    parent_task TEXT,              -- grouping for related agents
    model TEXT DEFAULT 'sonnet',          -- which Claude model to use
    timeout_ms INTEGER DEFAULT 600000,    -- 10 min default
    retry_count INTEGER DEFAULT 0,        -- current retry attempt
    max_retries INTEGER DEFAULT 3,        -- max retries before giving up
    token_input INTEGER DEFAULT 0,        -- input tokens consumed
    token_output INTEGER DEFAULT 0        -- output tokens consumed
);

-- Activity log: what each agent is doing
CREATE TABLE IF NOT EXISTS activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL REFERENCES agents(id),
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    action TEXT NOT NULL,          -- tool_call, file_read, file_write, search, etc
    detail TEXT,                   -- human readable description
    file_path TEXT                 -- if applicable
);

-- Files changed: track all file modifications
CREATE TABLE IF NOT EXISTS files_changed (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL REFERENCES agents(id),
    file_path TEXT NOT NULL,
    action TEXT NOT NULL,          -- created, modified, deleted
    lines_added INTEGER DEFAULT 0,
    lines_removed INTEGER DEFAULT 0,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_agents_parent ON agents(parent_task);
CREATE INDEX IF NOT EXISTS idx_activity_agent ON activity_log(agent_id);
CREATE INDEX IF NOT EXISTS idx_activity_timestamp ON activity_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_files_agent ON files_changed(agent_id);
CREATE INDEX IF NOT EXISTS idx_files_path ON files_changed(file_path);

-- DAG dependency edges
CREATE TABLE IF NOT EXISTS dag_edges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_task TEXT NOT NULL,
    from_agent_id TEXT NOT NULL,
    to_agent_id TEXT NOT NULL,
    UNIQUE(parent_task, from_agent_id, to_agent_id)
);
CREATE INDEX IF NOT EXISTS idx_dag_parent ON dag_edges(parent_task);

-- Artifact bus: context passing between agents
CREATE TABLE IF NOT EXISTS artifacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT NOT NULL REFERENCES agents(id),
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_artifacts_agent ON artifacts(agent_id);

-- File locks for conflict detection
CREATE TABLE IF NOT EXISTS file_locks (
    file_path TEXT NOT NULL,
    agent_id TEXT NOT NULL REFERENCES agents(id),
    locked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (file_path)
);

-- Quality gates
CREATE TABLE IF NOT EXISTS quality_gates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_task TEXT NOT NULL,
    phase TEXT NOT NULL,
    gate_name TEXT NOT NULL,
    command TEXT NOT NULL,
    required INTEGER DEFAULT 1,
    status TEXT DEFAULT 'pending',
    output TEXT,
    executed_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_gates_parent ON quality_gates(parent_task);

-- Pipeline templates
CREATE TABLE IF NOT EXISTS pipelines (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    definition TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Pipeline execution runs
CREATE TABLE IF NOT EXISTS pipeline_runs (
    id TEXT PRIMARY KEY,
    pipeline_id TEXT NOT NULL REFERENCES pipelines(id),
    status TEXT DEFAULT 'running',
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    total_token_input INTEGER DEFAULT 0,
    total_token_output INTEGER DEFAULT 0,
    current_phase TEXT
);

-- Individual steps within a pipeline run
CREATE TABLE IF NOT EXISTS pipeline_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL REFERENCES pipeline_runs(id),
    agent_id TEXT REFERENCES agents(id),
    phase TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    started_at DATETIME,
    finished_at DATETIME,
    token_input INTEGER DEFAULT 0,
    token_output INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_steps_run ON pipeline_steps(run_id);
