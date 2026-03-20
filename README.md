# Agent HQ

A specialized multi-agent coordination system with a real-time TUI dashboard for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

Claude Code is powerful, but it uses ONE generalist agent for everything. Agent HQ changes that: **14 specialized agents** -- each with defined roles, skills, and workflows -- coordinated through an MCP server with dependency graphs, pipeline orchestration, file locking, quality gates, and token tracking. All monitored in real time from a terminal dashboard.

Think of it as a virtual engineering floor: each agent sits in their "office" with their expertise, and you watch them work in real time.

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Agent HQ                                              ?: help  q: quit│
├───────────────────────────────┬─────────────────────────────────────────┤
│  OFFICES                      │  TASK LOG                              │
│                               │                                        │
│  ┌──────────┐ ┌──────────┐   │  14:02:31  python   Read auth/mod..    │
│  │ Python   │ │ Frontend │   │  14:02:33  python   Write models/u..   │
│  │ Backend  │ │          │   │  14:02:35  qa       Read tests/tes..   │
│  │ ● BUSY   │ │ ○ IDLE   │   │  14:02:36  python   Edit routes/au..   │
│  │ sonnet   │ │          │   │  14:02:38  qa       Write tests/te..   │
│  │ 12k tok  │ │          │   │  14:02:40  python   ✓ Completed       │
│  └──────────┘ └──────────┘   │  14:02:41  qa       ✓ Completed       │
│  ┌──────────┐ ┌──────────┐   │                                        │
│  │   Go     │ │   QA     │   │                                        │
│  │          │ │ Testing  │   │                                        │
│  │ ○ IDLE   │ │ ● BUSY   │   │                                        │
│  └──────────┘ └──────────┘   │                                        │
│                               │  FILES CHANGED                         │
│  ┌──────────┐ ┌──────────┐   │  + auth/models/user.py       (+42)    │
│  │  Docs    │ │  SDD     │   │  ~ auth/routes/login.py      (+8-3)   │
│  │          │ │ Planner  │   │  + tests/test_auth.py         (+67)   │
│  │ ○ IDLE   │ │ ○ IDLE   │   │                                        │
│  └──────────┘ └──────────┘   │                                        │
├───────────────────────────────┴─────────────────────────────────────────┤
│  2 agents active  │  3 tasks completed  │  DB: ~/.claude/agenthq.db    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Installation

### Method 1: Claude Code Plugin (recommended)

```bash
claude plugin install juani/agent-hq
```

### Method 2: Git Clone

```bash
git clone https://github.com/juani/agent-hq.git ~/.claude/agent-hq
cd ~/.claude/agent-hq
bash install.sh
```

### Method 3: Manual Build

```bash
git clone https://github.com/juani/agent-hq.git ~/.claude/agent-hq
cd ~/.claude/agent-hq
go build -o ~/.local/bin/agenthq ./cmd/agenthq
go build -o ~/.local/bin/agenthq-mcp ./cmd/agenthq-mcp
cp agents/*.md ~/.claude/agents/
```

### Prerequisites

- **Go 1.22+** -- [go.dev/dl](https://go.dev/dl/)
- **sqlite3** (optional) -- for manual DB operations; the MCP server creates the DB automatically

## Quick Start

1. **Install** using any method above.

2. **Open two terminal panes** (Zellij, Tmux, or two tabs):

   ```
   # Pane 1: Claude Code
   claude

   # Pane 2: Dashboard
   agenthq
   ```

3. **Ask Claude Code to delegate work.** Use `agent_spawn` to launch a specialized agent:

   ```
   "Spawn a python-backend agent to implement the auth module"
   ```

   The orchestrator calls `agent_spawn` with the profile, task description, and model. Agent HQ registers the agent, resolves the profile, checks concurrency limits, and returns the profile content for the sub-agent.

4. **Watch the dashboard** update as agents start, work, and complete tasks.

5. **For parallel work**, use `agent_spawn_batch` to launch multiple agents at once with dependency edges:

   ```
   "Spawn python-backend for the API and frontend for the UI, then qa-testing depending on both"
   ```

## Agent Profiles

| Agent | File | Specialization |
|-------|------|----------------|
| **Python Backend** | `python-backend.md` | FastAPI, Clean Architecture, async, Pydantic v2 |
| **Frontend** | `frontend.md` | React 19, Zustand 5, Tailwind 4, PWA |
| **Go** | `go.md` | Backend Go, TUIs, CLI tooling |
| **SQL & Data** | `sql-data.md` | SQLAlchemy 2.0, Alembic, query optimization |
| **QA & Testing** | `qa-testing.md` | pytest, vitest, testing-library, edge cases |
| **Documentation** | `docs.md` | API docs, READMEs, ADRs, changelogs |
| **SDD Planner** | `sdd-planner.md` | Proposals, specs, designs, task breakdowns |
| **DevOps** | `devops.md` | Docker, CI/CD, Redis, WebSocket infra |
| **Security** | `security.md` | Auth, RBAC, OWASP, headers, rate limiting |
| **Reviewer** | `reviewer.md` | Code review, quality gates, approve/reject |
| **Architect** | `architect.md` | Architecture validation, layer boundaries |
| **Git** | `git.md` | Branches, conventional commits, PRs |
| **Research** | `research.md` | Web search, library evaluation, tradeoffs |
| **Self Improver** | `self-improver.md` | Maintains and evolves Agent HQ itself |

Profiles live in `internal/profiles/agents/` (embedded at build time) and follow a standard format:

```markdown
---
name: Agent Name
role: What this agent does
skills: [skill-a, skill-b]
---

## Identity
## Rules
## Workflow
## Output Contract
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `Tab` | Switch between offices and task log panels |
| `h` / `l` | Navigate between offices |
| `j` / `k` | Scroll within a panel |
| `Enter` | View agent detail |
| `Esc` | Back to main view |
| `?` | Toggle help overlay |
| `q` | Quit |

## MCP Tools Reference

The `agenthq-mcp` server exposes 22 tools to Claude Code, organized in three tiers.

### Core Tools

| Tool | Description | Key Params |
|------|-------------|------------|
| `agent_register` | Register a new agent with status=running | `id`, `profile`, `task`, `parent_task?`, `model?` |
| `agent_complete` | Mark an agent as completed or failed | `id`, `status` (completed/failed), `result_summary?` |
| `agent_log_activity` | Log an activity for an agent | `agent_id`, `action`, `detail?`, `file_path?` |
| `agent_status` | Get status of all agents, optionally filtered by profile | `profile?` |
| `agent_list_profiles` | List all available agent profiles with descriptions | -- |
| `agent_get_profile` | Get the full markdown content of a profile | `name` |

### Tier 1 -- Spawn and Tokens

| Tool | Description | Key Params |
|------|-------------|------------|
| `agent_spawn` | Prepare an agent for launch with concurrency checks, profile resolution, and optional DAG dependencies | `id`, `profile`, `task`, `model?`, `timeout_ms?`, `depends_on?` |
| `agent_spawn_batch` | Spawn multiple agents at once with concurrency validation | `agents[]`, `parent_task` |
| `agent_update_tokens` | Record token usage for an agent | `agent_id`, `token_input`, `token_output` |
| `agent_cost` | Get cost summary for a task group or single agent | `parent_task?` or `agent_id?` |

### Tier 2 -- DAG, Coordination, and Safety

| Tool | Description | Key Params |
|------|-------------|------------|
| `dag_define` | Define dependency graph edges between agents | `parent_task`, `edges[]` ({from, to}) |
| `dag_next` | Get next agents ready to run based on dependency graph | `parent_task` |
| `artifact_put` | Store an output artifact for an agent | `agent_id`, `key`, `value` |
| `artifact_get` | Get artifacts from an agent's dependencies | `agent_id`, `parent_task` |
| `file_lock_check` | Pre-spawn conflict detection for file locks | `files[]`, `agent_id?` |
| `file_lock_manage` | Acquire or release file locks for an agent | `action` (acquire/release/release_all), `agent_id`, `files?` |
| `gate_define` | Define a quality gate for a pipeline phase | `parent_task`, `phase`, `gate_name`, `command` |
| `gate_report` | Report a quality gate result and get overall status | `gate_id`, `status` (passed/failed), `output?` |

### Tier 3 -- Pipelines

| Tool | Description | Key Params |
|------|-------------|------------|
| `pipeline_create` | Create a reusable pipeline template | `id`, `name`, `definition` (JSON) |
| `pipeline_run` | Start a pipeline execution | `pipeline_id`, `run_id` |
| `pipeline_status` | Get pipeline run status with step details | `run_id` |
| `pipeline_history` | Get past pipeline runs | `pipeline_id`, `limit?` |

## DAG and Dependencies

Agent HQ supports directed acyclic graphs (DAGs) to define execution order between agents. When agent B depends on agent A, B will not be marked as ready until A completes.

**Define dependencies at spawn time:**

```json
{
  "tool": "agent_spawn",
  "args": {
    "id": "agt-tests",
    "profile": "qa-testing",
    "task": "Write tests for the auth module",
    "parent_task": "implement-auth",
    "depends_on": ["agt-api", "agt-models"]
  }
}
```

**Or define edges explicitly:**

```json
{
  "tool": "dag_define",
  "args": {
    "parent_task": "implement-auth",
    "edges": [
      { "from": "agt-api", "to": "agt-tests" },
      { "from": "agt-models", "to": "agt-tests" }
    ]
  }
}
```

**Query what's ready to run:**

```json
{
  "tool": "dag_next",
  "args": { "parent_task": "implement-auth" }
}
// Returns: { "ready_agents": ["agt-api", "agt-models"], "running_count": 0, "max_agents": 3 }
```

Agents can pass context to dependents via artifacts (`artifact_put` / `artifact_get`).

## Pipelines

Pipelines are reusable multi-phase workflows. Define a template once, run it many times.

```json
{
  "tool": "pipeline_create",
  "args": {
    "id": "feature-pipeline",
    "name": "Standard Feature Pipeline",
    "definition": {
      "phases": [
        { "name": "plan", "profile": "sdd-planner" },
        { "name": "implement", "profile": "python-backend" },
        { "name": "test", "profile": "qa-testing" },
        { "name": "review", "profile": "reviewer" },
        { "name": "docs", "profile": "docs" }
      ]
    }
  }
}
```

Run a pipeline:

```json
{
  "tool": "pipeline_run",
  "args": { "pipeline_id": "feature-pipeline", "run_id": "run-auth-001" }
}
```

Track progress with `pipeline_status` and view past executions with `pipeline_history`. Pipeline runs track per-step and aggregate token usage automatically.

## Quality Gates

Quality gates are checkpoints that must pass before a pipeline can proceed. Define gates per phase, run them, and report results.

```json
{
  "tool": "gate_define",
  "args": {
    "parent_task": "implement-auth",
    "phase": "test",
    "gate_name": "unit-tests-pass",
    "command": "pytest tests/ -x",
    "required": true
  }
}
```

After running the gate check, report the result:

```json
{
  "tool": "gate_report",
  "args": {
    "gate_id": 1,
    "status": "passed",
    "output": "42 tests passed, 0 failed"
  }
}
```

The response includes the overall gate status: which gates are still pending, which failed, and whether all required gates have passed.

## Token Tracking

Every agent tracks input and output token consumption. Update tokens as agents work:

```json
{
  "tool": "agent_update_tokens",
  "args": { "agent_id": "agt-api", "token_input": 15000, "token_output": 3200 }
}
```

Get cost reports aggregated by task group:

```json
{
  "tool": "agent_cost",
  "args": { "parent_task": "implement-auth" }
}
// Returns per-agent breakdown + totals
```

Token usage is also rolled up into pipeline runs automatically.

## Architecture

```
┌──────────────┐         ┌──────────────────────────┐
│  Claude Code │────────>│  agenthq-mcp             │
│  (orchestr.) │  MCP    │  (Go binary)             │
└──────────────┘  proto  │                          │
                         │  22 MCP tools:           │
                         │  Core, Spawn, DAG,       │
                         │  Pipelines, Gates,       │
                         │  File Locks, Artifacts   │
                         └────────────┬─────────────┘
                                      │ SQLite
                                      v
                         ┌──────────────────────────┐
                         │  agenthq.db              │
                         │                          │
                         │  agents, activity_log,   │
                         │  files_changed, dag_edges,│
                         │  artifacts, file_locks,  │
                         │  quality_gates, pipelines,│
                         │  pipeline_runs,          │
                         │  pipeline_steps          │
                         └────────────┬─────────────┘
                                      │ reads
                                      v
                         ┌──────────────────────────┐
                         │  agenthq (TUI)           │
                         │  Bubble Tea              │
                         │                          │
                         │  Polls DB every 500ms,   │
                         │  renders live dashboard  │
                         │  with model badges and   │
                         │  token counters          │
                         └──────────────────────────┘
```

The MCP server writes to SQLite. The TUI reads from SQLite. They share no state other than the database file, which makes the system simple and reliable.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENTHQ_DB` | `~/.claude/agenthq.db` | Path to the SQLite database |
| `AGENTHQ_REFRESH` | `500ms` | TUI refresh interval |
| `AGENTHQ_MAX_AGENTS` | `3` | Maximum concurrent agents per parent task |

## Creating Custom Profiles

1. Create a new file in `~/.claude/agents/`:

   ```bash
   touch ~/.claude/agents/my-agent.md
   ```

2. Follow the profile format:

   ```markdown
   ---
   name: My Custom Agent
   role: Describe what this agent specializes in
   skills: [relevant-skill-1, relevant-skill-2]
   ---

   ## Identity
   You are a specialist in X. You have deep expertise in...

   ## Rules
   - Always do X before Y
   - Never commit without tests

   ## Workflow
   1. Understand the task
   2. Analyze existing code
   3. Implement changes
   4. Verify correctness

   ## Output Contract
   Return: status, summary, files changed, issues found
   ```

3. The orchestrator will pick up the new profile automatically on the next agent launch.

## Contributing

### Project Structure

```
cmd/
  agenthq/          # TUI dashboard entry point
  agenthq-mcp/      # MCP server entry point
internal/
  db/               # SQLite layer (schema.sql, sqlite.go, write.go)
  mcp/              # MCP protocol + tool handlers (tools.go, handlers)
  tui/              # Bubble Tea UI (app.go, office.go, detail.go)
  model/            # Data types (types.go)
  profiles/         # Embedded agent profiles
    agents/         # 14 .md profile files
```

### How to Add a New MCP Tool

1. Add the tool definition to `ListTools()` in `internal/mcp/tools.go`
2. Add a `case` for the tool name in `CallTool()` in the same file
3. Implement the handler method on `ToolHandler`
4. Add any required DB methods in `internal/db/`

### How to Add a New Agent Profile

Drop a `.md` file in `internal/profiles/agents/` following the frontmatter format (name, role, skills). Rebuild the binary to embed it.

### How to Add New DB Tables

1. Add the `CREATE TABLE` statement to `internal/db/schema.sql`
2. Add the corresponding Go type in `internal/model/types.go`
3. Add query/write methods in `internal/db/sqlite.go` and `internal/db/write.go`

### How to Update the TUI

- `internal/tui/app.go` -- polling loop and main model
- `internal/tui/office.go` -- office grid rendering
- `internal/tui/detail.go` -- agent detail view

### Self Improver

The `self-improver` agent profile exists specifically for AI-assisted improvements to Agent HQ itself. It has knowledge of the codebase structure and conventions.

## Constraints

- **Max 3 concurrent agents** per parent task (configurable via `AGENTHQ_MAX_AGENTS`)
- **No recursion** -- sub-agents cannot launch other sub-agents
- **Ephemeral agents** -- they start, work, and terminate; persistence lives in Engram, not in the agent
- **Pure Go SQLite** -- no CGO dependency; uses `modernc.org/sqlite`
- **File locks are advisory** -- they prevent conflicts between cooperating agents but are not OS-level locks

## Uninstall

```bash
cd ~/.claude/agent-hq
bash uninstall.sh
```

Or manually:

```bash
rm -f ~/.local/bin/agenthq ~/.local/bin/agenthq-mcp
rm -rf ~/.claude/agents/    # removes profiles
rm -f ~/.claude/agenthq.db  # removes database
```

## License

[MIT](LICENSE)
