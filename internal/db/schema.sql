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
    parent_task TEXT               -- grouping for related agents
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
