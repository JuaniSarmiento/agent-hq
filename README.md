# Agent HQ

A specialized sub-agent system with a real-time TUI dashboard for [Claude Code](https://docs.anthropic.com/en/docs/claude-code). Instead of one generalist agent doing everything, Agent HQ gives you **13 specialized agents** — each with defined roles, skills, and workflows — coordinated through an MCP server and monitored via a terminal dashboard.

Think of it as a virtual engineering floor: each agent sits in their "office" with their expertise, and you watch them work in real time from the command line.

```
┌─────────────────────────────────────────────────────────────────────┐
│  Agent HQ                                            ?: help  q: quit │
├──────────────────────────────┬──────────────────────────────────────┤
│  OFFICES                     │  TASK LOG                           │
│                              │                                     │
│  ┌─────────┐ ┌─────────┐    │  14:02:31  python   Read auth/mod.. │
│  │ Python  │ │Frontend │    │  14:02:33  python   Write models/u..│
│  │ Backend │ │         │    │  14:02:35  qa       Read tests/tes..│
│  │ ● BUSY  │ │ ○ IDLE  │    │  14:02:36  python   Edit routes/au..│
│  └─────────┘ └─────────┘    │  14:02:38  qa       Write tests/te..│
│  ┌─────────┐ ┌─────────┐    │  14:02:40  python   ✓ Completed    │
│  │   Go    │ │   QA    │    │  14:02:41  qa       ✓ Completed    │
│  │         │ │Testing  │    │                                     │
│  │ ○ IDLE  │ │ ● BUSY  │    │                                     │
│  └─────────┘ └─────────┘    │                                     │
│                              │                                     │
│  ┌─────────┐ ┌─────────┐    │  FILES CHANGED                     │
│  │  Docs   │ │  SDD    │    │  + auth/models/user.py       (+42) │
│  │         │ │Planner  │    │  ~ auth/routes/login.py      (+8-3)│
│  │ ○ IDLE  │ │ ○ IDLE  │    │  + tests/test_auth.py        (+67) │
│  └─────────┘ └─────────┘    │                                     │
├──────────────────────────────┴──────────────────────────────────────┤
│  2 agents active  │  3 tasks completed  │  DB: ~/.claude/agenthq.db │
└─────────────────────────────────────────────────────────────────────┘
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

- **Go 1.22+** — [go.dev/dl](https://go.dev/dl/)
- **sqlite3** (optional) — for manual DB operations; the MCP server creates the DB automatically

## Quick Start

1. **Install** using any method above.

2. **Open two terminal panes** (Zellij, Tmux, or two tabs):

   ```
   # Pane 1: Claude Code
   claude

   # Pane 2: Dashboard
   agenthq
   ```

3. **Ask Claude Code to delegate work.** The orchestrator picks the right agent profile, launches a sub-agent, and Agent HQ tracks everything in real time.

4. **Watch the dashboard** update as agents start, work, and complete tasks.

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

Profiles live in `~/.claude/agents/` and follow a standard format:

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

## MCP Tools

The `agenthq-mcp` server exposes these tools to Claude Code:

| Tool | Description |
|------|-------------|
| `agent_start` | Register a new agent instance (profile, task, parent) |
| `agent_end` | Mark an agent as completed or failed with result summary |
| `agent_status` | Get current status of a specific agent |
| `agent_list` | List all agents, optionally filtered by status or parent task |
| `activity_log` | Record an activity entry (tool call, file read/write, search) |
| `activity_feed` | Get recent activity across all agents |
| `files_changed` | Record a file change (create, modify, delete) with line counts |
| `files_list` | List all files changed by an agent or across all agents |
| `dashboard_summary` | Get aggregated stats for the TUI (counts, active agents, recent activity) |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENTHQ_DB` | `~/.claude/agenthq.db` | Path to the SQLite database |
| `AGENTHQ_REFRESH` | `500ms` | TUI refresh interval |
| `AGENTHQ_MAX_AGENTS` | `3` | Maximum concurrent agents per parent task |

## Architecture

```
┌──────────────┐         ┌──────────────────┐
│  Claude Code │────────>│  agenthq-mcp     │
│  (orchestr.) │  MCP    │  (Go binary)     │
└──────────────┘  proto  │                  │
                         │  Tools:          │
                         │  - agent_start   │
                         │  - agent_end     │
                         │  - activity_log  │
                         │  - files_changed │
                         │  - ...           │
                         └────────┬─────────┘
                                  │ SQLite
                                  ▼
                         ┌──────────────────┐
                         │  agenthq.db      │
                         │                  │
                         │  agents          │
                         │  activity_log    │
                         │  files_changed   │
                         └────────┬─────────┘
                                  │ reads
                                  ▼
                         ┌──────────────────┐
                         │  agenthq         │
                         │  (TUI - Bubble   │
                         │   Tea)           │
                         │                  │
                         │  Polls DB every  │
                         │  500ms, renders  │
                         │  live dashboard  │
                         └──────────────────┘
```

The MCP server writes to SQLite. The TUI reads from SQLite. They share no state other than the database file, which makes the system simple and reliable.

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

## Constraints

- **Max 3 concurrent agents** per parent task (configurable via `AGENTHQ_MAX_AGENTS`)
- **No recursion** — sub-agents cannot launch other sub-agents
- **Ephemeral agents** — they start, work, and terminate; persistence lives in Engram, not in the agent
- **Pure Go SQLite** — no CGO dependency; the TUI uses `modernc.org/sqlite`

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
