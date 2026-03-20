# Agent HQ -- Manual Completo

## Que es

Agent HQ es un sistema de sub-agentes especializados para Claude Code con un dashboard TUI en tiempo real. Transforma la forma en que Claude Code trabaja: en vez de un agente generico haciendo todo, hay **14 agentes especializados** con perfiles, skills y roles definidos, coordinados por un orquestador central. Incluye grafos de dependencias (DAG), pipelines reutilizables, quality gates, file locking y tracking de tokens.

## Componentes

```
~/.claude/
├── agent-hq/                    # Proyecto Go
│   ├── cmd/
│   │   ├── agenthq/             # Entry point TUI Dashboard
│   │   └── agenthq-mcp/         # Entry point MCP Server
│   ├── internal/
│   │   ├── tui/                 # Bubbletea UI (oficinas, logs, detalle)
│   │   ├── db/                  # SQLite queries (pure Go, sin CGO)
│   │   ├── mcp/                 # MCP protocol + 22 tool handlers
│   │   ├── model/               # Tipos: Agent, Activity, DAGEdge, Pipeline, etc.
│   │   └── profiles/            # Perfiles embebidos
│   │       └── agents/          # 14 archivos .md de perfiles
│   ├── plugin.json              # Descriptor del plugin para Claude Code
│   └── install.sh               # Script de instalacion
├── agents/                      # Agent Profiles instalados (~/.claude/agents/)
└── agenthq.db                   # SQLite DB (se crea automaticamente)
```

## Los 14 Agentes

| Agente | Archivo | Especialidad |
|--------|---------|-------------|
| **Python Backend** | `python-backend.md` | FastAPI, Clean Arch, async, Pydantic v2 |
| **Frontend** | `frontend.md` | React 19, Zustand 5, Tailwind 4, PWA |
| **Go** | `go.md` | Backend Go, TUIs, herramientas CLI |
| **SQL & Data** | `sql-data.md` | SQLAlchemy 2.0, Alembic, queries |
| **QA & Testing** | `qa-testing.md` | pytest, vitest, testing-library, edge cases |
| **Documentacion** | `docs.md` | API docs, READMEs, ADRs, changelogs |
| **SDD Planner** | `sdd-planner.md` | Proposals, specs, disenos, task breakdowns |
| **DevOps** | `devops.md` | Docker, CI/CD, Redis, WebSocket infra |
| **Security** | `security.md` | Auth, RBAC, OWASP, headers, rate limiting |
| **Reviewer** | `reviewer.md` | Code review, quality gate, approve/reject |
| **Architect** | `architect.md` | Validacion arquitectonica, layer boundaries |
| **Git** | `git.md` | Branches, conventional commits, PRs |
| **Research** | `research.md` | Web search, evaluacion de libs, tradeoffs |
| **Self Improver** | `self-improver.md` | Mantiene y evoluciona Agent HQ |

### Estructura de cada perfil

```markdown
---
name: Nombre del Agente
role: Descripcion del rol
skills: [lista de skills que se cargan]
---

## Identity      -> Quien es, su expertise
## Rules         -> Reglas que DEBE seguir
## Workflow      -> Pasos que sigue para cada tarea
## Output Contract -> Que devuelve cuando termina
```

## Instalacion

### Requisitos
- Go 1.22+ ([go.dev/dl](https://go.dev/dl/))
- sqlite3 CLI (opcional -- la DB se crea automaticamente)
- Zellij o Tmux (para split de paneles)

### Build e instalacion

```bash
cd ~/.claude/agent-hq
bash install.sh
```

O manualmente:

```bash
cd ~/.claude/agent-hq
go build -o bin/agenthq ./cmd/agenthq
go build -o bin/agenthq-mcp ./cmd/agenthq-mcp
cp bin/* ~/.local/bin/
```

## Uso

### Setup del terminal

```
┌─────────────────────────┬──────────────────────────────┐
│                         │                              │
│   Claude Code           │   agenthq                   │
│                         │                              │
│   Panel izquierdo       │   Panel derecho              │
│   (donde hablas)        │   (dashboard en vivo)        │
│                         │                              │
└─────────────────────────┴──────────────────────────────┘
```

**Con Zellij:**
```bash
# Terminal 1
claude

# Ctrl+P N (nuevo panel a la derecha)
agenthq
```

**Con Tmux:**
```bash
# Terminal 1
tmux
claude

# Ctrl+B % (split vertical)
agenthq
```

### Keybindings del Dashboard

| Tecla | Accion |
|-------|--------|
| `Tab` | Cambiar entre panel de oficinas y task log |
| `h` / `l` | Navegar entre oficinas |
| `j` / `k` | Navegar dentro del panel |
| `Enter` | Ver detalle de un agente |
| `Esc` | Volver a la vista principal |
| `?` | Mostrar/ocultar ayuda |
| `q` | Salir |

## Como funciona el flujo

```
1. Le decis a Claude: "Implementa el auth"

2. Claude (orquestador):
   ├── Llama agent_spawn con profile="python-backend", task="..."
   ├── Agent HQ verifica concurrencia (max 3 agentes)
   ├── Resuelve el perfil y lo devuelve en la respuesta
   └── Registra el agente en SQLite con status=running

3. El agente trabaja:
   ├── agent_log_activity para registrar cada accion
   ├── artifact_put para guardar resultados intermedios
   └── file_lock_manage para evitar conflictos de archivos

4. El agente termina:
   ├── agent_update_tokens con el consumo final
   └── agent_complete con status=completed/failed

5. La TUI (agenthq):
   └── Lee SQLite cada 500ms -> renderiza el estado
```

## Las 22 MCP Tools

### Core (gestion basica de agentes)

| Tool | Que hace | Params clave |
|------|----------|-------------|
| `agent_register` | Registra un agente con status=running | `id`, `profile`, `task` |
| `agent_complete` | Marca un agente como completado o fallido | `id`, `status` |
| `agent_log_activity` | Loguea una actividad del agente | `agent_id`, `action` |
| `agent_status` | Estado de todos los agentes | `profile?` |
| `agent_list_profiles` | Lista los perfiles disponibles | -- |
| `agent_get_profile` | Devuelve el markdown completo de un perfil | `name` |

### Tier 1 -- Spawn y Tokens

| Tool | Que hace | Params clave |
|------|----------|-------------|
| `agent_spawn` | Lanza un agente con chequeo de concurrencia, resolucion de perfil y dependencias DAG | `id`, `profile`, `task`, `depends_on?` |
| `agent_spawn_batch` | Lanza multiples agentes de una | `agents[]`, `parent_task` |
| `agent_update_tokens` | Registra consumo de tokens | `agent_id`, `token_input`, `token_output` |
| `agent_cost` | Reporte de costo por tarea o agente | `parent_task?` o `agent_id?` |

### Tier 2 -- DAG, Coordinacion y Seguridad

| Tool | Que hace | Params clave |
|------|----------|-------------|
| `dag_define` | Define aristas de dependencia entre agentes | `parent_task`, `edges[]` |
| `dag_next` | Devuelve los agentes listos para correr | `parent_task` |
| `artifact_put` | Guarda un artefacto de salida de un agente | `agent_id`, `key`, `value` |
| `artifact_get` | Obtiene artefactos de las dependencias | `agent_id`, `parent_task` |
| `file_lock_check` | Detecta conflictos de file locks antes de spawnear | `files[]` |
| `file_lock_manage` | Adquiere o libera locks de archivos | `action`, `agent_id`, `files?` |
| `gate_define` | Define un quality gate para una fase | `parent_task`, `phase`, `gate_name`, `command` |
| `gate_report` | Reporta resultado de un gate | `gate_id`, `status` |

### Tier 3 -- Pipelines

| Tool | Que hace | Params clave |
|------|----------|-------------|
| `pipeline_create` | Crea un template de pipeline reutilizable | `id`, `name`, `definition` |
| `pipeline_run` | Ejecuta un pipeline | `pipeline_id`, `run_id` |
| `pipeline_status` | Estado del pipeline con detalle por step | `run_id` |
| `pipeline_history` | Historial de ejecuciones de un pipeline | `pipeline_id`, `limit?` |

## DAG y Dependencias

El DAG te permite definir orden de ejecucion entre agentes. Si el agente B depende de A, B no corre hasta que A complete.

```json
{
  "tool": "agent_spawn",
  "args": {
    "id": "agt-tests",
    "profile": "qa-testing",
    "task": "Tests para el auth module",
    "parent_task": "implement-auth",
    "depends_on": ["agt-api", "agt-models"]
  }
}
```

Consultar que esta listo para correr:

```json
{
  "tool": "dag_next",
  "args": { "parent_task": "implement-auth" }
}
```

Los agentes pasan contexto a sus dependientes via artefactos (`artifact_put` / `artifact_get`).

## Pipelines

Los pipelines son workflows multi-fase reutilizables. Definis un template una vez, lo corres muchas veces.

```json
{
  "tool": "pipeline_create",
  "args": {
    "id": "feature-pipeline",
    "name": "Feature Pipeline Estandar",
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

Cada ejecucion trackea tokens por step y en total automaticamente.

## Quality Gates

Los quality gates son checkpoints que deben pasar antes de avanzar en un pipeline.

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

Despues de ejecutar el gate, reportas el resultado con `gate_report`. La respuesta incluye el estado general: que gates estan pendientes, cuales fallaron, y si todos los requeridos pasaron.

## Tracking de Tokens

Cada agente trackea consumo de tokens de input y output. Se actualizan con `agent_update_tokens` y se consultan con `agent_cost`. Si el agente pertenece a un pipeline run, los tokens se acumulan automaticamente en el run.

## Base de datos

SQLite en `~/.claude/agenthq.db` con 10 tablas:

- **agents** -- cada instancia de agente (id, profile, task, status, model, tokens, timestamps)
- **activity_log** -- que hace cada agente (tool calls, reads, writes)
- **files_changed** -- archivos creados/modificados/eliminados por agente
- **dag_edges** -- aristas de dependencia entre agentes
- **artifacts** -- artefactos key-value producidos por agentes
- **file_locks** -- locks de archivos para deteccion de conflictos
- **quality_gates** -- checkpoints de calidad por fase
- **pipelines** -- templates de pipelines reutilizables
- **pipeline_runs** -- ejecuciones de pipelines
- **pipeline_steps** -- steps individuales dentro de un pipeline run

## Configuracion

| Variable | Default | Descripcion |
|----------|---------|-------------|
| `AGENTHQ_DB` | `~/.claude/agenthq.db` | Path a la base de datos SQLite |
| `AGENTHQ_REFRESH` | `500ms` | Intervalo de refresh de la TUI |
| `AGENTHQ_MAX_AGENTS` | `3` | Maximo de agentes concurrentes por parent task |

## Personalizacion

### Agregar un nuevo agente

1. Crear `~/.claude/agents/mi-agente.md` siguiendo el formato de frontmatter
2. Definir Identity, Rules, Workflow, Output Contract
3. El orquestador lo detecta automaticamente

### Modificar un agente

Edita el archivo `.md` correspondiente. Los cambios aplican en el proximo agente que se lance.

### Agregar un perfil embebido

Dropealo en `internal/profiles/agents/` y rebuildealo para que se embeba en el binario.

## Contribuir

### Estructura del proyecto

```
cmd/
  agenthq/          # Entry point del TUI dashboard
  agenthq-mcp/      # Entry point del MCP server
internal/
  db/               # Capa SQLite (schema.sql, sqlite.go, write.go)
  mcp/              # Protocolo MCP + handlers de tools (tools.go)
  tui/              # Bubble Tea UI (app.go, office.go, detail.go)
  model/            # Tipos de datos (types.go)
  profiles/         # Perfiles embebidos
    agents/         # 14 archivos .md
```

### Como agregar una nueva MCP tool

1. Agregar la definicion a `ListTools()` en `internal/mcp/tools.go`
2. Agregar un `case` en `CallTool()` en el mismo archivo
3. Implementar el handler como metodo de `ToolHandler`
4. Agregar los metodos de DB necesarios en `internal/db/`

### Como agregar un nuevo perfil de agente

Dropea un `.md` en `internal/profiles/agents/` con el frontmatter (name, role, skills). Rebuildealo.

### Como agregar nuevas tablas de DB

1. Agregar el `CREATE TABLE` en `internal/db/schema.sql`
2. Agregar el tipo Go en `internal/model/types.go`
3. Agregar metodos de query/write en `internal/db/sqlite.go` y `internal/db/write.go`

### Como actualizar la TUI

- `internal/tui/app.go` -- polling loop y modelo principal
- `internal/tui/office.go` -- rendering del grid de oficinas
- `internal/tui/detail.go` -- vista de detalle del agente

### Self Improver

El perfil `self-improver` existe para mejoras asistidas por IA al propio Agent HQ. Conoce la estructura del codebase y las convenciones.

## Limitaciones

- **Max 3 agentes** en paralelo por tarea (configurable via `AGENTHQ_MAX_AGENTS`)
- **Sin recursion** -- los sub-agentes no pueden lanzar otros sub-agentes
- **Agentes efimeros** -- nacen, trabajan, mueren. La persistencia esta en Engram, no en el agente
- **Pure Go SQLite** -- sin dependencia CGO, usa `modernc.org/sqlite`
- **File locks son advisory** -- previenen conflictos entre agentes cooperantes, no son locks a nivel OS

## Desinstalar

```bash
cd ~/.claude/agent-hq
bash uninstall.sh
```

O manualmente:

```bash
rm -f ~/.local/bin/agenthq ~/.local/bin/agenthq-mcp
rm -rf ~/.claude/agents/
rm -f ~/.claude/agenthq.db
```
