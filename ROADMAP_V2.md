# SkAgent v2.0 - Piano di Sviluppo Avanzato

## 🎯 Obiettivi

1. **Nuova TUI accattivante** con temi, animazioni, pannelli multipli
2. **Modalità Headless** per automazione e CI/CD
3. **Server MCP** (Model Context Protocol) per integrazione con Claude, Cursor, etc.
4. **Server ACP** (Agent Communication Protocol) per comunicazione inter-agent
5. **API REST** per integrazioni esterne
6. **Project Manager Integration** per task assignment e agent orchestration

---

## 📐 Nuova Architettura

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           SKAGENT v2.0                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │   TUI v2    │  │  Headless   │  │  REST API   │  │  MCP/ACP    │    │
│  │  (Bubble)   │  │    Mode     │  │   Server    │  │   Server    │    │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘    │
│         │                │                │                │            │
│  ┌──────┴────────────────┴────────────────┴────────────────┴──────┐    │
│  │                        CORE ENGINE                              │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │    │
│  │  │ AI Engine│  │  Tools   │  │  Agents  │  │  Config  │        │    │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                    │                                     │
│  ┌─────────────────────────────────┴─────────────────────────────────┐  │
│  │                    PROJECT MANAGER BRIDGE                          │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │  Linear  │  │  Jira    │  │  GitHub  │  │ Plane.so │          │  │
│  │  │  Issues  │  │  Cloud   │  │ Projects │  │          │          │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘          │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 🎨 1. Nuova TUI Accattivante

### Features
- **Layout multi-pannello**: Chat, Tasks, Agents, Logs
- **Temi**: Catppuccin, Dracula, Nord, Tokyo Night, Gruvbox, Custom
- **Animazioni**: Spinner, progress bars, typing effect
- **Dashboard**: Status agents, task progress, metrics
- **Split views**: Verticale/orizzontale
- **Tabs**: Multiple conversations
- **Notifiche**: Toast notifications
- **Syntax highlighting**: Code blocks colorati

### Struttura
```
internal/
├── tui/
│   ├── app.go           # Main TUI app
│   ├── views/
│   │   ├── chat.go      # Chat view
│   │   ├── dashboard.go # Dashboard view
│   │   ├── tasks.go     # Task manager view
│   │   ├── agents.go    # Agent status view
│   │   └── logs.go      # Logs view
│   ├── components/
│   │   ├── message.go   # Message bubble
│   │   ├── taskcard.go  # Task card
│   │   ├── agentcard.go # Agent status card
│   │   ├── progress.go  # Progress bars
│   │   └── notify.go    # Notifications
│   ├── themes/
│   │   ├── theme.go     # Theme interface
│   │   ├── catppuccin.go
│   │   ├── dracula.go
│   │   ├── nord.go
│   │   ├── tokyo.go
│   │   └── custom.go
│   └── layout/
│       ├── split.go     # Split layout
│       ├── tabs.go      # Tab container
│       └── flex.go      # Flexbox layout
```

---

## 🤖 2. Modalità Headless

### Uso
```bash
# Esegui singolo comando
skagent --headless "crea un progetto CLI per gestire dotfiles"

# Esegui da file
skagent --headless --input tasks.json --output results.json

# Pipeline mode
echo "analizza questo codice" | skagent --headless --pipe

# Watch mode (daemon)
skagent --headless --watch --port 9999
```

### Features
- Output JSON strutturato
- Exit codes significativi
- Logging configurabile
- Stdin/stdout/stderr separati
- Batch processing

---

## 🔌 3. Server MCP (Model Context Protocol)

### Endpoint Tools Esposti
```json
{
  "tools": [
    {
      "name": "skagent_plan",
      "description": "Create a project plan using SpecKit methodology",
      "inputSchema": {
        "type": "object",
        "properties": {
          "idea": {"type": "string"},
          "style": {"type": "string", "enum": ["minimal", "detailed", "enterprise"]}
        }
      }
    },
    {
      "name": "skagent_task",
      "description": "Execute a specific task",
      "inputSchema": {...}
    },
    {
      "name": "skagent_agent_status",
      "description": "Get status of running agents"
    },
    {
      "name": "skagent_assign_task",
      "description": "Assign task to specific agent"
    }
  ]
}
```

### Uso
```bash
# Start MCP server
skagent serve --mcp --port 3000

# In Claude/Cursor config
{
  "mcpServers": {
    "skagent": {
      "command": "skagent",
      "args": ["serve", "--mcp"]
    }
  }
}
```

---

## 🔗 4. Server ACP (Agent Communication Protocol)

### Protocollo
- WebSocket per real-time
- Pub/Sub per task broadcast
- Request/Response per queries
- Event streaming per status updates

### Messaggi
```json
// Task Assignment
{
  "type": "task.assign",
  "task_id": "task-123",
  "agent_id": "coder-1",
  "payload": {
    "title": "Implement login API",
    "spec": "..."
  }
}

// Status Update
{
  "type": "agent.status",
  "agent_id": "coder-1",
  "status": "working",
  "progress": 45,
  "current_task": "task-123"
}

// Task Complete
{
  "type": "task.complete",
  "task_id": "task-123",
  "result": {...}
}
```

---

## 🌐 5. API REST

### Endpoints

```
# Health & Info
GET  /api/v1/health
GET  /api/v1/info

# Conversations
POST /api/v1/chat
GET  /api/v1/chat/:id
DELETE /api/v1/chat/:id

# Tasks
GET  /api/v1/tasks
POST /api/v1/tasks
GET  /api/v1/tasks/:id
PUT  /api/v1/tasks/:id
DELETE /api/v1/tasks/:id

# Agents
GET  /api/v1/agents
POST /api/v1/agents/:id/assign
GET  /api/v1/agents/:id/status

# Project Manager Sync
POST /api/v1/sync/linear
POST /api/v1/sync/github
POST /api/v1/sync/jira

# Webhooks
POST /api/v1/webhooks/task-update
POST /api/v1/webhooks/agent-complete
```

---

## 📋 6. Project Manager Integration

### Supporto
- **Linear** - GraphQL API
- **GitHub Projects** - GraphQL API v2
- **Jira Cloud** - REST API
- **Plane.so** - REST API
- **Notion** - API
- **Trello** - REST API

### Flusso
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Linear    │────▶│   SkAgent   │────▶│   Agents    │
│   Issues    │     │   Bridge    │     │  (Workers)  │
└─────────────┘     └──────┬──────┘     └──────┬──────┘
                           │                    │
                           │◀───────────────────┘
                           │  Status Updates
                           ▼
                    ┌─────────────┐
                    │   Linear    │
                    │   Update    │
                    └─────────────┘
```

### Configurazione
```yaml
# ~/.config/skagent/projects.yaml
projects:
  - name: my-project
    manager: linear
    team_id: "TEAM-123"
    api_key: "${LINEAR_API_KEY}"
    sync:
      interval: 5m
      auto_assign: true
    agents:
      - id: coder
        type: coding
        labels: ["backend", "api"]
      - id: reviewer
        type: review
        labels: ["review", "security"]
      - id: docs
        type: documentation
        labels: ["docs", "readme"]
```

---

## 📁 Nuova Struttura Directory

```
skagent/
├── cmd/
│   └── skagent/
│       └── main.go
├── internal/
│   ├── core/              # Core engine
│   │   ├── engine.go
│   │   ├── session.go
│   │   └── context.go
│   ├── ai/                # AI providers
│   ├── agents/            # Agent definitions
│   │   ├── agent.go
│   │   ├── coder.go
│   │   ├── reviewer.go
│   │   ├── planner.go
│   │   └── registry.go
│   ├── tools/             # Tools
│   ├── tui/               # New TUI
│   │   ├── app.go
│   │   ├── views/
│   │   ├── components/
│   │   ├── themes/
│   │   └── layout/
│   ├── headless/          # Headless mode
│   │   ├── runner.go
│   │   ├── batch.go
│   │   └── output.go
│   ├── server/            # Servers
│   │   ├── rest/
│   │   │   ├── server.go
│   │   │   ├── handlers/
│   │   │   └── middleware/
│   │   ├── mcp/
│   │   │   ├── server.go
│   │   │   ├── tools.go
│   │   │   └── protocol.go
│   │   └── acp/
│   │       ├── server.go
│   │       ├── pubsub.go
│   │       └── protocol.go
│   ├── projects/          # Project managers
│   │   ├── manager.go
│   │   ├── linear.go
│   │   ├── github.go
│   │   ├── jira.go
│   │   └── sync.go
│   ├── config/
│   └── storage/           # Persistence
│       ├── sqlite.go
│       ├── tasks.go
│       └── sessions.go
├── api/                   # OpenAPI specs
│   └── openapi.yaml
├── themes/                # Theme files
│   ├── catppuccin.json
│   ├── dracula.json
│   └── ...
└── examples/
    ├── headless/
    ├── mcp-config/
    └── project-config/
```

---

## 🚀 Piano di Implementazione

### Fase 1: Core Refactoring (Week 1)
- [ ] Separare core engine dalla TUI
- [ ] Creare agent registry
- [ ] Implementare session management
- [ ] Aggiungere storage SQLite

### Fase 2: Nuova TUI (Week 2)
- [ ] Sistema temi
- [ ] Layout multi-pannello
- [ ] Dashboard view
- [ ] Task manager view
- [ ] Componenti riutilizzabili

### Fase 3: Headless Mode (Week 2)
- [ ] CLI runner
- [ ] JSON output
- [ ] Batch processing
- [ ] Pipe mode

### Fase 4: REST API (Week 3)
- [ ] Server HTTP con Fiber/Echo
- [ ] Endpoints CRUD
- [ ] Authentication
- [ ] OpenAPI documentation

### Fase 5: MCP Server (Week 3)
- [ ] Protocol implementation
- [ ] Tool definitions
- [ ] Stdio transport
- [ ] Testing con Claude

### Fase 6: ACP Server (Week 4)
- [ ] WebSocket server
- [ ] Pub/Sub system
- [ ] Agent communication
- [ ] Status broadcasting

### Fase 7: Project Managers (Week 4)
- [ ] Linear integration
- [ ] GitHub Projects integration
- [ ] Sync engine
- [ ] Auto-assignment logic

---

## 📦 Nuove Dipendenze

```go
require (
    // TUI
    github.com/charmbracelet/bubbletea v2
    github.com/charmbracelet/lipgloss v2
    github.com/charmbracelet/bubbles v2
    github.com/charmbracelet/glamour   // Markdown rendering
    
    // Server
    github.com/gofiber/fiber/v2        // REST API
    github.com/gorilla/websocket       // WebSocket
    
    // Storage
    github.com/mattn/go-sqlite3        // SQLite
    github.com/jmoiron/sqlx            // SQL helpers
    
    // Config
    github.com/spf13/viper             // Config management
    github.com/spf13/cobra             // CLI framework
    
    // Project Managers
    github.com/shurcooL/graphql        // GraphQL client
    
    // Utils
    github.com/rs/zerolog              // Logging
    github.com/google/uuid             // UUIDs
)
```

---

## 🎮 Nuovi Comandi CLI

```bash
# TUI Mode (default)
skagent

# Headless Mode
skagent run "create a todo app"
skagent run --file tasks.json
skagent run --pipe < input.txt

# Server Mode
skagent serve                    # All servers
skagent serve --rest             # REST only
skagent serve --mcp              # MCP only
skagent serve --acp              # ACP only
skagent serve --port 8080

# Config
skagent config                   # Interactive config
skagent config set theme dracula
skagent config set provider openrouter
skagent config show

# Projects
skagent projects list
skagent projects sync linear
skagent projects add --manager linear --team TEAM-123

# Agents
skagent agents list
skagent agents status
skagent agents assign task-123 coder-1

# Themes
skagent themes list
skagent themes preview dracula
skagent themes set tokyo-night
```

---

## ⚡ Quick Start (Dopo Implementazione)

```bash
# 1. Install
go install github.com/biodoia/skagent@latest

# 2. Setup
skagent config

# 3. Connect to Linear
skagent projects add --manager linear

# 4. Start in TUI mode
skagent

# 5. Or start servers
skagent serve --rest --mcp --port 8080
```
