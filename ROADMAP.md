# Agent HQ Roadmap

## Vision

Agent HQ should become the **operating system layer for Claude Code multi-agent workflows** — the thing that makes you say "how was I running agents without this?" It is not a framework or a platform; it is a thin, fast control plane that makes sub-agents reliable, observable, cost-aware, and recoverable. The north star: you launch a complex multi-agent pipeline, close your laptop, come back, and everything either finished successfully or paused at a clear decision point with a full audit trail.

## Current State (v0.2)

Agent HQ today provides: (1) an MCP server exposing ~20 tools for agent lifecycle, DAG dependencies, artifacts, file locks, quality gates, and pipelines; (2) a Bubbletea TUI dashboard that polls SQLite and renders agent offices, activity logs, and token counts; (3) 13 embedded agent profiles with role-specific system prompts; (4) basic token tracking per agent and per pipeline run. The architecture is intentionally simple — MCP writes to SQLite, TUI reads from SQLite, no shared state. This is a solid foundation, but it is mostly a **passive tracker**. It records what happens but does not actively prevent problems, recover from failures, or learn from past runs.

## Planned Improvements

### Phase 1: Smart Orchestration (v0.3)

#### 1.1 Agent Heartbeat & Timeout Detection
- **Problem**: An agent can hang silently — stuck in a loop, waiting for input, or crashed. The orchestrator has no idea. The TUI shows "BUSY" forever. The user wastes time and tokens.
- **Solution**: Add a `last_heartbeat` column to `agents`. Expose an `agent_heartbeat` MCP tool that agents call periodically (or the orchestrator calls when logging activity). A background goroutine in the MCP server checks for agents whose last heartbeat exceeds `timeout_ms` and auto-transitions them to `failed` with reason `timeout`. The TUI shows a warning icon for agents approaching timeout (>80% of limit).
- **Effort**: S
- **Impact**: High

#### 1.2 Smart Model Selection
- **Problem**: Every agent defaults to Sonnet. A simple file rename does not need Sonnet. An architecture decision deserves Opus. The user either wastes money or gets subpar results.
- **Solution**: Add a `model_hint` field to agent profiles (e.g., `model: haiku` for git agent, `model: opus` for architect). The `agent_spawn` tool reads the profile's model hint and uses it as default, but the orchestrator can override per-spawn. Add a `task_complexity` optional field to `agent_spawn` — if provided, the MCP server applies a simple mapping: `simple -> haiku`, `standard -> sonnet`, `complex -> opus`. This is a suggestion, not enforcement — the orchestrator always has final say.
- **Effort**: S
- **Impact**: High

#### 1.3 Budget Alerts & Hard Limits
- **Problem**: A pipeline can burn through $10+ of tokens before anyone notices. There is no circuit breaker.
- **Solution**: Add `budget_limit_tokens` to `pipeline_runs` (optional). The `agent_update_tokens` handler checks cumulative usage against the limit after each update. If usage exceeds 80%, return a warning in the response. If it exceeds 100%, return an error and set the pipeline status to `paused`. Add an `agent_budget` MCP tool that returns current spend vs. limit with a percentage. The TUI shows a budget bar in the footer when a pipeline is running.
- **Effort**: M
- **Impact**: High

#### 1.4 Diff Preview Before Merge
- **Problem**: Multiple agents edit files in parallel. You cannot see what each agent changed until after the fact. If an agent makes a bad change, there is no easy rollback.
- **Solution**: Add a `workspace_snapshot` MCP tool that runs `git stash create` (or `git diff HEAD`) before a pipeline starts and stores the stash ref in the pipeline run. Add a `files_diff` MCP tool that returns the git diff for files changed by a specific agent (using the `files_changed` table to know which files). The TUI detail view gets a "Diff" tab that shows the actual changes. On agent failure, the orchestrator can use `git checkout -- <files>` to revert that agent's changes only.
- **Effort**: M
- **Impact**: High

#### 1.5 Profile Inheritance
- **Problem**: Many profiles share common rules (error handling, output format, engram integration). Updating a shared convention means editing 13 files.
- **Solution**: Support a `base` field in profile frontmatter: `base: _base.md`. The `profiles.Get()` function resolves inheritance by reading the base profile and prepending it to the specialized content. A `_base.md` profile contains shared rules (always log activity, always save to engram, output contract format). Profiles can chain: `go.md` -> `_backend-base.md` -> `_base.md`. Max depth of 3 to prevent cycles.
- **Effort**: S
- **Impact**: Medium

#### 1.6 Prompt Templates (Agent Templates)
- **Problem**: The orchestrator has to manually compose the full prompt for each sub-agent every time: profile + task + context + engram references + skill paths. This is error-prone and verbose.
- **Solution**: Add a `templates/` directory with pre-built prompt templates for common tasks: `add-endpoint.md`, `write-tests.md`, `refactor-module.md`, `fix-bug.md`. Each template has placeholders: `{{task}}`, `{{files}}`, `{{profile}}`, `{{engram_context}}`. Add a `template_resolve` MCP tool that takes a template name + variables and returns the fully composed prompt. The orchestrator calls this once, gets the prompt, and passes it to `claude --print` or `task`/`delegate`.
- **Effort**: M
- **Impact**: Medium

### Phase 2: Session Resilience (v0.4)

#### 2.1 Session Handoff (Compaction Recovery)
- **Problem**: When context compacts or a session dies, all in-flight state is lost. The orchestrator does not know which agents were running, what phase the pipeline was in, or what artifacts were produced.
- **Solution**: Add a `session_state` table that stores a JSON snapshot of the current orchestration state: active pipeline, running agents, pending DAG nodes, current phase. The MCP server updates this on every state transition (agent start, complete, fail, pipeline phase change). Add a `session_recover` MCP tool that returns the last known state. On session restart, the orchestrator calls `session_recover`, gets the snapshot, and resumes from where it left off. Combined with engram's `mem_context`, this gives full recovery.
- **Effort**: M
- **Impact**: High

#### 2.2 Workspace Snapshots
- **Problem**: If an agent corrupts files, there is no automatic rollback. The user has to manually `git stash` before risky operations.
- **Solution**: Add a `snapshot_create` MCP tool that runs `git stash push -m "agenthq-{pipeline_run_id}"` before a pipeline starts. Add `snapshot_restore` that runs `git stash pop` on the matching stash. Add `snapshot_diff` that shows what changed since the snapshot. The pipeline_create definition supports an `auto_snapshot: true` option. On pipeline failure, the TUI shows a "Restore snapshot?" option.
- **Effort**: M
- **Impact**: High

#### 2.3 Agent Retry with Context
- **Problem**: When an agent fails, `max_retries` exists but a retry without context about WHY it failed will likely fail again the same way.
- **Solution**: On agent failure, the `agent_complete` handler saves the `result_summary` (which should contain the error) as an artifact with key `_failure_reason`. When `agent_spawn` detects a retry (same task, same profile, `retry_count > 0`), it automatically includes the previous failure artifact in the response so the orchestrator can pass it to the retried agent's prompt: "Previous attempt failed because: {reason}. Avoid this approach."
- **Effort**: S
- **Impact**: High

#### 2.4 Agent Replay
- **Problem**: Debugging a failed agent is hard. You cannot re-run it with the same inputs to reproduce the issue.
- **Solution**: Add an `agent_inputs` table that stores the full spawn parameters (profile content, task, model, artifacts received) for each agent. Add an `agent_replay` MCP tool that takes an agent ID and returns the exact spawn parameters so the orchestrator can re-launch it identically. The TUI detail view gets a "Replay" action. This is also useful for A/B testing different models on the same task.
- **Effort**: M
- **Impact**: Medium

### Phase 3: Intelligence Layer (v0.5)

#### 3.1 Cross-Session Learning (Performance Analytics)
- **Problem**: Every session starts from scratch. You do not know which profiles work best for which tasks, which model gives the best cost/quality ratio, or which pipelines tend to fail.
- **Solution**: Add an `agent_outcomes` table that tracks: agent_id, profile, model, task_category (extracted from task text via simple keyword matching), duration, token_cost, success/failure, retry_count. Add a `stats_profile` MCP tool that returns performance stats per profile: avg duration, success rate, avg tokens, best model. Add `stats_pipeline` for pipeline-level analytics. The TUI gets a "Stats" view (accessible via `s` key) showing a leaderboard of profiles by success rate and efficiency.
- **Effort**: M
- **Impact**: Medium

#### 3.2 Engram Auto-Integration
- **Problem**: Sub-agents are supposed to save discoveries to engram, but this is a convention that depends on the orchestrator remembering to include the instruction. If forgotten, knowledge is lost.
- **Solution**: Bake engram integration into the base profile (`_base.md` from 1.5). The `agent_complete` handler, when status is `completed`, automatically calls engram's `mem_save` with the agent's result_summary, profile, and task as a structured observation. Topic key: `agent-hq/runs/{agent_id}`. This creates an automatic audit trail in engram without relying on sub-agent cooperation. The orchestrator can later search engram for past solutions to similar tasks.
- **Effort**: S
- **Impact**: High

#### 3.3 Smart Task Routing
- **Problem**: The orchestrator has to manually pick which profile to use for each task. Sometimes the mapping is obvious (write tests -> qa-testing), sometimes it is ambiguous (refactor auth module — is that python-backend or security?).
- **Solution**: Add a `route_task` MCP tool that takes a task description and returns a ranked list of recommended profiles with confidence scores. The ranking uses: (1) keyword matching against profile role descriptions, (2) file path patterns (`.py` files -> python-backend, `.tsx` -> frontend), (3) historical success rate from `agent_outcomes` for similar tasks. This is a SUGGESTION, not an auto-router — the orchestrator makes the final call.
- **Effort**: M
- **Impact**: Medium

#### 3.4 Adaptive Prompting
- **Problem**: Profile prompts are static. A profile that consistently fails on certain task types keeps failing the same way.
- **Solution**: Add a `prompt_patches` table that stores profile-specific amendments learned from failures. When a task fails and is manually fixed, the orchestrator (or user) can call `prompt_patch_add` with the profile name and a rule like "When modifying auth routes, always check middleware order first." The `agent_spawn` handler appends active patches to the profile content before returning it. Patches have an `active` flag so they can be disabled.
- **Effort**: M
- **Impact**: Medium

### Phase 4: Ecosystem (v0.6)

#### 4.1 Notification Support
- **Problem**: Long pipelines run in the background. You switch to another terminal or take a break. There is no way to know when it finishes without checking the TUI.
- **Solution**: Add a `notify` MCP tool that sends a desktop notification via `notify-send` (Linux), `osascript` (macOS), or `powershell` (Windows). The `pipeline_run` handler auto-notifies on completion or failure. The TUI also supports a `--notify` flag for agent completion events. Keep it simple — just system notifications, no webhooks or Slack integrations initially.
- **Effort**: S
- **Impact**: Medium

#### 4.2 Pipeline Library (Shareable Pipelines)
- **Problem**: Users create useful pipelines (e.g., "add REST endpoint with tests and docs") but cannot share them.
- **Solution**: Add a `pipeline_export` MCP tool that serializes a pipeline definition to a standalone YAML file in `~/.claude/agent-hq/pipelines/`. Add `pipeline_import` that loads from a file or URL. The YAML format includes: name, description, phases with profile + task template + model hint + quality gates. Ship 3-5 built-in pipelines: `add-feature`, `bug-fix`, `refactor`, `add-tests`, `code-review`.
- **Effort**: M
- **Impact**: Medium

#### 4.3 Profile Marketplace (Community Profiles)
- **Problem**: Building good agent profiles requires trial and error. There is no way to benefit from others' work.
- **Solution**: Support a `profiles_remote` directory structure on GitHub. Add a `profile_install` MCP tool that downloads a profile from a GitHub repo URL into `~/.claude/agents/`. Add `profile_list_remote` that fetches an index. Start with a curated `agent-hq-profiles` repo. Each profile includes: the .md file, a README with example tasks and expected outputs, and performance benchmarks (tokens, success rate).
- **Effort**: L
- **Impact**: Medium

#### 4.4 TUI Enhancements
- **Problem**: The TUI is functional but basic. It does not show pipeline progress, budget status, or allow interactive actions.
- **Solution**: Add these views/features incrementally: (1) Pipeline view — Gantt-style timeline showing phases and their agents; (2) Budget bar — real-time token cost in the footer with color coding (green/yellow/red); (3) Interactive actions — press `r` on a failed agent to trigger replay, `x` to cancel a running agent; (4) Filter/search — press `/` to filter agents by profile, status, or task keyword; (5) Log streaming — live tail of agent activity instead of polling snapshots.
- **Effort**: L
- **Impact**: Medium

## Priority Matrix

| Improvement | Phase | Effort | Impact | Priority |
|-------------|-------|--------|--------|----------|
| Agent Heartbeat | 1 | S | High | P0 |
| Smart Model Selection | 1 | S | High | P0 |
| Budget Alerts | 1 | M | High | P0 |
| Session Handoff | 2 | M | High | P0 |
| Workspace Snapshots | 2 | M | High | P1 |
| Agent Retry with Context | 2 | S | High | P1 |
| Engram Auto-Integration | 3 | S | High | P1 |
| Diff Preview | 1 | M | High | P1 |
| Profile Inheritance | 1 | S | Medium | P2 |
| Notification Support | 4 | S | Medium | P2 |
| Prompt Templates | 1 | M | Medium | P2 |
| Agent Replay | 2 | M | Medium | P2 |
| Smart Task Routing | 3 | M | Medium | P3 |
| Cross-Session Learning | 3 | M | Medium | P3 |
| Adaptive Prompting | 3 | M | Medium | P3 |
| Pipeline Library | 4 | M | Medium | P3 |
| TUI Enhancements | 4 | L | Medium | P3 |
| Profile Marketplace | 4 | L | Medium | P4 |

## Non-Goals

Things Agent HQ should NOT try to be:

- **Not a process manager**. It does not spawn OS processes or manage Claude Code instances. The orchestrator (Claude Code itself) does that. Agent HQ is the control plane, not the data plane.
- **Not a Kubernetes for AI agents**. No service mesh, no load balancing, no auto-scaling. Keep it simple.
- **Not a prompt engineering framework**. Profiles are markdown files with conventions. No DSL, no prompt compiler, no chain-of-thought templates.
- **Not a replacement for engram**. Agent HQ tracks operational state (who is running, what they changed). Engram tracks knowledge (what was learned, what decisions were made). They complement each other.
