# Agent HQ — Manual Completo

## Qué es

Agent HQ es un sistema de sub-agentes especializados para Claude Code con un dashboard TUI en tiempo real. Transforma la forma en que Claude Code trabaja: en vez de un agente genérico haciendo todo, hay **13 agentes especializados** con perfiles, skills y roles definidos, coordinados por un orquestador central.

## Componentes

```
~/.claude/
├── agent-hq/              # Proyecto Go — TUI Dashboard
│   ├── cmd/agenthq/       # Entry point
│   ├── internal/
│   │   ├── tui/           # Bubbletea UI (oficinas, logs, detalle)
│   │   ├── db/            # SQLite queries (pure Go, sin CGO)
│   │   └── model/         # Tipos: Agent, Activity, FileChange
│   ├── schema.sql         # Schema de la DB
│   └── Makefile           # build, install, run, clean
├── agents/                # 13 Agent Profiles (.md)
├── hooks/                 # 4 Hook Scripts (.sh)
└── agenthq.db             # SQLite DB (se crea automáticamente)
```

## Los 13 Agentes

| Agente | Archivo | Especialidad | Skills Inyectadas |
|--------|---------|-------------|-------------------|
| **Python Backend** | `python-backend.md` | FastAPI, Clean Arch, async, Pydantic v2 | fastapi-clean-arch, jwt-auth-rbac, sqlalchemy-multitenant |
| **Frontend** | `frontend.md` | React 19, Zustand 5, Tailwind 4, PWA | react19-zustand, tailwind-dark-theme, pwa-workbox |
| **Go** | `go.md` | Backend Go, TUIs, herramientas CLI | go-testing |
| **SQL & Data** | `sql-data.md` | SQLAlchemy 2.0, Alembic, queries | sqlalchemy-multitenant |
| **QA & Testing** | `qa-testing.md` | pytest, vitest, testing-library, edge cases | go-testing |
| **Documentación** | `docs.md` | API docs, READMEs, ADRs, changelogs | — |
| **SDD Planner** | `sdd-planner.md` | Proposals, specs, diseños, task breakdowns | sdd-* |
| **DevOps** | `devops.md` | Docker, CI/CD, Redis, WebSocket infra | redis-patterns, websocket-gateway |
| **Security** | `security.md` | Auth, RBAC, OWASP, headers, rate limiting | jwt-auth-rbac, redis-patterns |
| **Reviewer** | `reviewer.md` | Code review, quality gate, approve/reject | — |
| **Architect** | `architect.md` | Validación arquitectónica, layer boundaries | — |
| **Git** | `git.md` | Branches, conventional commits, PRs | — |
| **Research** | `research.md` | Web search, evaluación de libs, tradeoffs | — |

### Estructura de cada perfil

```markdown
---
name: Nombre del Agente
role: Descripción del rol
skills: [lista de skills que se cargan]
---

## Identity      → Quién es, su expertise
## Rules         → Reglas que DEBE seguir
## Workflow      → Pasos que sigue para cada tarea
## Output Contract → Qué devuelve cuando termina
```

## Instalación

### Requisitos
- Go 1.25+ (se auto-actualiza con `go mod tidy`)
- sqlite3 CLI (`sudo apt-get install sqlite3`)
- Zellij o Tmux (para split de paneles)

### Build e instalación

```bash
# Build
cd ~/.claude/agent-hq
make build

# Instalar en PATH
make install
# → binario en ~/.local/bin/agenthq

# Inicializar la DB
~/.claude/hooks/init-db.sh
```

## Uso

### Setup del terminal

```
┌─────────────────────────┬──────────────────────────────┐
│                         │                              │
│   Claude Code           │   agenthq                   │
│                         │                              │
│   Panel izquierdo       │   Panel derecho              │
│   (donde hablás)        │   (dashboard en vivo)        │
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

| Tecla | Acción |
|-------|--------|
| `Tab` | Cambiar entre panel de oficinas y task log |
| `h` / `l` | Navegar entre oficinas |
| `j` / `k` | Navegar dentro del panel |
| `Enter` | Ver detalle de un agente |
| `Esc` | Volver a la vista principal |
| `?` | Mostrar/ocultar ayuda |
| `q` | Salir |

## Cómo funciona el flujo

```
1. Vos le decís a Claude: "Implementá el auth"

2. Claude (orquestador):
   ├── Lee el perfil de ~/.claude/agents/python-backend.md
   ├── Inyecta el perfil + skills en el prompt del sub-agente
   ├── Lanza el agente con la tool Agent
   │
   └── Hook on-agent-start.sh → INSERT en SQLite

3. El agente trabaja:
   └── Hook on-tool-call.sh → INSERT actividad en SQLite

4. El agente termina:
   └── Hook on-agent-end.sh → UPDATE en SQLite

5. La TUI (agenthq):
   └── Lee SQLite cada 500ms → renderiza el estado
```

## Hooks

Los hooks son scripts bash que capturan eventos de los agentes y escriben en SQLite.

| Hook | Cuándo se ejecuta | Args |
|------|--------------------|------|
| `init-db.sh` | Primera vez / DB no existe | — |
| `on-agent-start.sh` | Cuando se lanza un agente | `agent_id profile task [parent_task]` |
| `on-agent-end.sh` | Cuando un agente termina | `agent_id status [result_summary]` |
| `on-tool-call.sh` | Cada tool call de un agente | `agent_id action detail [file_path]` |

### Configuración en settings.json

Los hooks se conectan en `~/.claude/settings.json` (pendiente de configurar por el usuario según el formato de hooks de Claude Code).

## Base de datos

SQLite en `~/.claude/agenthq.db` con 3 tablas:

- **agents** — cada instancia de agente (id, profile, task, status, timestamps)
- **activity_log** — qué hace cada agente (tool calls, reads, writes)
- **files_changed** — archivos creados/modificados/eliminados por agente

## Personalización

### Agregar un nuevo agente

1. Crear `~/.claude/agents/mi-agente.md` siguiendo el formato de frontmatter
2. Definir Identity, Rules, Workflow, Output Contract
3. El orquestador lo detecta automáticamente

### Modificar un agente

Editá el archivo `.md` correspondiente. Los cambios aplican en el próximo agente que se lance (no persisten entre instancias).

### Agregar skills a un agente

Agregá el nombre de la skill al array `skills` en el frontmatter:
```yaml
skills: [fastapi-clean-arch, mi-nueva-skill]
```

## Limitaciones actuales

- **Max 3 agentes** en paralelo por tarea (decisión de diseño para no quemar tokens)
- **Sin recursión** — los sub-agentes no pueden lanzar otros sub-agentes
- **Agentes efímeros** — nacen, trabajan, mueren. La persistencia está en Engram, no en el agente
- **Hooks necesitan sqlite3 CLI** — la TUI usa driver Go puro, pero los hooks usan el CLI
