# SKAgent v2.0 - Advanced AI Agent Framework

![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

SKAgent è un framework avanzato per agenti AI con interfaccia grafica moderna, modalità headless, API REST, server MCP e integrazione con project manager.

## 🚀 Caratteristiche Principali

### ✨ Interfaccia Grafica Moderna
- **Dashboard Interattivo**: Panoramica completa degli agenti e del sistema
- **Terminal Mode**: Interfaccia terminale avanzata con supporto temi
- **Gestione Temi**: 4 temi predefiniti (Dark, Light, Solarized, Neon)
- **UI Responsive**: Adattamento automatico alle dimensioni della finestra
- **Animazioni Fluide**: Transizioni e animazioni personalizzabili

### 🖥️ Modalità Headless
- **Esecuzione Server**: Modalità completamente headless per ambienti di produzione
- **Gestione PID**: Controllo dei processi con file PID
- **Logging Avanzato**: Logging strutturato con livelli configurabili
- **Auto-start**: Avvio automatico dei servizi configurabile
- **Signal Handling**: Gestione优雅 degli shutdown

### 🌐 API REST Completa
- **Endpoints Completi**: `/agents`, `/tasks`, `/tools`, `/system`, `/project`
- **Documentazione Integrata**: Endpoint discovery automatico
- **CORS Support**: Supporto completo per applicazioni web
- **Rate Limiting**: Protezione da sovraccarico
- **Health Checks**: Monitoraggio dello stato del sistema

### 🔗 Server MCP (Model Context Protocol)
- **Strumenti Integrati**: Gestione agenti, task, e sistema
- **Protocollo Standard**: Compatibile con MCP v2024-11-05
- **Tool Registry**: Sistema di registrazione dinamica degli strumenti
- **Interoperabilità**: Integrazione con altri sistemi MCP

### 📋 Integrazione Project Manager
- **Task Management**: Gestione completa dei task di progetto
- **Agent Assignment**: Assegnazione automatica e manuale degli agenti
- **AI Recommendations**: Raccomandazioni AI-powered per l'assegnazione
- **Status Tracking**: Monitoraggio stato task e progressi
- **Webhook Support**: Notifiche automatiche per cambiamenti di stato

## 📦 Installazione

### Prerequisiti
- Go 1.21 o superiore
- Git

### Installazione Rapida
```bash
git clone https://github.com/biodoia/skagent.git
cd skagent
go build -o skagent ./cmd/skagent
./skagent --help
```

### Installazione con Go Modules
```bash
go mod tidy
go build -o skagent ./cmd/skagent
```

## 🎯 Modalità di Utilizzo

### 1. Modalità Interattiva (Default)
```bash
./skagent
```
Avvia l'interfaccia grafica completa con dashboard e terminal.

### 2. Modalità Headless
```bash
./skagent --mode headless --port 8080 --daemon
```
Avvia il server in modalità headless per ambienti di produzione.

### 3. Modalità Server
```bash
./skagent --mode server --host 0.0.0.0 --port 8080
```
Avvia tutti i servizi (API REST + MCP) per accesso remoto.

### 4. Modalità Setup
```bash
./skagent --mode setup
```
Avvia il wizard di configurazione iniziale.

## ⚙️ Configurazione

### Configurazione Base
Il file di configurazione viene creato automaticamente in `~/.config/skagent/config.json`:

```json
{
  "version": "2.0.0",
  "default_provider": "openrouter",
  "api": {
    "host": "localhost",
    "port": 8080,
    "enable_cors": true
  },
  "mcp": {
    "host": "localhost", 
    "port": 8081
  },
  "headless": {
    "enabled": true,
    "auto_start": false,
    "max_agents": 10
  },
  "theme": {
    "name": "dark",
    "auto_save": true,
    "font_size": 14
  },
  "project": {
    "enabled": false,
    "api_key": "",
    "base_url": ""
  }
}
```

### Temi Disponibili
- **Dark**: Tema scuro con colori catppuccin
- **Light**: Tema chiaro per ambienti luminosi
- **Solarized Dark**: Tema scuro solare
- **Neon**: Tema neon per un look futuristico

## 🔌 API REST Endpoints

### Agent Management
- `GET /agents` - Lista tutti gli agenti
- `POST /agents` - Crea un nuovo agente
- `GET /agents/{id}` - Dettagli di un agente
- `PUT /agents/{id}` - Aggiorna un agente
- `DELETE /agents/{id}` - Elimina un agente
- `POST /agents/{id}/start` - Avvia un agente
- `POST /agents/{id}/stop` - Ferma un agente

### Task Management
- `GET /tasks` - Lista tutti i task
- `POST /tasks` - Crea un nuovo task
- `GET /tasks/{id}` - Dettagli di un task
- `PUT /tasks/{id}` - Aggiorna un task
- `DELETE /tasks/{id}` - Cancella un task

### Project Manager Integration
- `GET /project/tasks` - Task del progetto
- `POST /project/tasks` - Crea task progetto
- `GET /project/assignments` - Assegnazioni task
- `POST /project/assignments` - Assegna task ad agente
- `GET /project/agents` - Agenti disponibili
- `POST /project/recommend` - Raccomandazioni AI

### System
- `GET /health` - Health check
- `GET /status` - Status completo sistema
- `GET /system/config` - Configurazione sistema
- `POST /system/shutdown` - Shutdown graceful

## 🔧 MCP Server

### Endpoints Disponibili
- `GET /health` - Health check MCP
- `GET /tools` - Lista strumenti disponibili
- `GET /tools/{name}` - Dettagli strumento
- `POST /tools/{name}/call` - Chiama strumento
- `GET /agents` - Lista agenti
- `GET /capabilities` - Capacità server

### Strumenti Integrati
- `list_agents` - Lista agenti con filtri
- `get_agent` - Dettagli agente specifico
- `start_agent` / `stop_agent` - Controllo ciclo vita
- `create_task` - Creazione task
- `get_task_status` - Status task
- `get_system_status` - Status sistema
- `list_project_tasks` - Task progetto
- `assign_task_to_agent` - Assegnazione task
- `recommend_agents` - Raccomandazioni AI

## 🎨 Interfaccia Grafica

### Dashboard
- Panoramica sistema con statistiche
- Lista agenti con status real-time
- Monitoraggio task attivi
- Quick actions per operazioni comuni

### Terminal Mode
- Terminale interattivo completo
- Command history e auto-completion
- Output streaming in tempo reale
- Modalità multiple (interactive, batch, server)

### Settings & Themes
- Gestione temi integrata
- Personalizzazione interfaccia
- Esportazione/importazione configurazioni
- Preview temi in tempo reale

## 🏗️ Architettura

### Componenti Core
```
┌─────────────────────────────────────────┐
│                 CLI                     │
├─────────────────────────────────────────┤
│               TUI Layer                 │
│  ┌─────────────┬─────────────┬────────┐ │
│  │  Dashboard  │  Terminal   │Settings│ │
│  └─────────────┴─────────────┴────────┘ │
├─────────────────────────────────────────┤
│               Core Engine               │
│  ┌─────────────┬─────────────┬────────┐ │
│  │   Agents    │   Tasks     │ Tools  │ │
│  └─────────────┴─────────────┴────────┘ │
├─────────────────────────────────────────┤
│              Server Layer               │
│  ┌─────────────┬─────────────┬────────┐ │
│  │  REST API   │   MCP API   │ Project│ │
│  └─────────────┴─────────────┴────────┘ │
└─────────────────────────────────────────┘
```

### Flusso Dati
1. **User Input** → TUI/CLI/API
2. **Command Processing** → Core Engine
3. **Agent Dispatch** → Agent Registry
4. **Task Execution** → Tool System
5. **Result Processing** → Response Handler

## 🔒 Sicurezza

### Autenticazione
- Supporto API key per REST API
- Token-based authentication per MCP
- Configurazione CORS per web clients
- Rate limiting per prevenire abuse

### Configurazione Sicura
- File config con permessi 0600
- API keys crittografate in storage
- Environment variables support
- Configurazione ambiente-specifica

## 📊 Monitoraggio

### Metrics Disponibili
- Status agenti (active/idle/offline)
- Task completion rate
- System resource usage
- API request metrics
- Agent performance metrics

### Logging
- Structured logging con livelli
- Rotazione automatica log
- Integration con external loggers
- Real-time log streaming

## 🚀 Deployment

### Docker (Raccomandato)
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o skagent ./cmd/skagent

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/skagent /usr/local/bin/
CMD ["skagent", "--mode", "server", "--host", "0.0.0.0"]
```

### Systemd Service
```ini
[Unit]
Description=SKAgent AI Framework
After=network.target

[Service]
Type=simple
User=skagent
ExecStart=/usr/local/bin/skagent --mode server --host 0.0.0.0
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: skagent
spec:
  replicas: 3
  selector:
    matchLabels:
      app: skagent
  template:
    metadata:
      labels:
        app: skagent
    spec:
      containers:
      - name: skagent
        image: skagent:latest
        args: ["--mode", "server", "--host", "0.0.0.0"]
        ports:
        - containerPort: 8080
        - containerPort: 8081
```

## 🧪 Testing

### Unit Tests
```bash
go test ./internal/agents/... -v
go test ./internal/server/... -v
go test ./internal/tools/... -v
```

### Integration Tests
```bash
go test ./integration/... -v
```

### Performance Tests
```bash
go test -bench=. ./benchmarks/...
```

## 📈 Roadmap v2.1

- [ ] Plugin system per agenti custom
- [ ] WebSocket support per real-time updates
- [ ] Database integration (PostgreSQL/MySQL)
- [ ] Distributed agent coordination
- [ ] Advanced analytics dashboard
- [ ] Kubernetes operator
- [ ] GitOps integration
- [ ] Multi-tenant support

## 🤝 Contributing

1. Fork del repository
2. Crea feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add AmazingFeature'`)
4. Push al branch (`git push origin feature/AmazingFeature`)
5. Apri Pull Request

### Development Setup
```bash
git clone https://github.com/biodoia/skagent.git
cd skagent
go mod tidy
go test ./...
go build ./...
```

## 📝 License

Questo progetto è licenziato sotto licenza MIT - vedere file [LICENSE](LICENSE) per dettagli.

## 🙏 Acknowledgments

- [BubbleTea](https://github.com/charmbracelet/bubbletea) per l'interfaccia TUI
- [Chi](https://github.com/go-chi/chi) per il routing HTTP
- [LipGloss](https://github.com/charmbracelet/lipgloss) per lo styling
- Community Go per i packages opensource

---

**SKAgent v2.0** - Il futuro degli agenti AI è qui! 🚀