# Roadmap & Optimization Plan: Skagent v2

This document outlines the refactoring and feature additions required to move `skagent` from Alpha to Beta/Production.

## Phase 1: Core Architecture Hardening (The "Brain" Fix)

### 1.1 Implement Recursive Tool Loop (ReAct Pattern)
**Objective:** Enable the agent to autonomous execute tools and reason on results.
- **Refactor `internal/core/engine.go`**:
  - Change `Process` to a loop that continues while `StopReason == "tool_use"`.
  - Implement a `ToolExecutor` interface that maps `ToolCall.Name` to actual Go functions.
  - Append `ToolResult` messages back to the conversation history.

### 1.2 Switch to Streaming API
**Objective:** Improve UX and prevent timeouts on long tasks.
- **Refactor `internal/ai/client.go`**:
  - Replace `client.Messages.New` with `client.Messages.NewStreaming`.
  - Update `Engine` to return a `chan string` or `chan Event` instead of a static string.
  - Update TUI and MCP server to consume the stream.

### 1.3 Persistence Layer (SQLite)
**Objective:** Save state across restarts.
- **New Package `internal/storage`**:
  - Implement `Repository` interface.
  - Use `mattn/go-sqlite3` or `modernc.org/sqlite` (pure Go).
  - Schema: `sessions`, `messages`, `agents`, `tasks`.

---

## Phase 2: MCP Server Maturity

### 2.1 Real Implementations for Mock Tools
**Objective:** Make the MCP server functional.
- **Task Management:** Implement a real Task Queue (in SQLite).
- **Agent Recommendation:** Implement basic vector similarity (using a lightweight embedding model or simple keyword matching) instead of random mocks.
- **Project Integration:** Connect `list_project_tasks` to a real provider (e.g., GitHub Issues API or local file scanning).

### 2.2 Standardize Error Responses
- Implement JSON-RPC 2.0 compliant error codes for the MCP server.

---

## Phase 3: Advanced Features

### 3.1 Context Management
- Implement "Sliding Window" context: Keep system prompt + last N messages.
- Implement "Summary" context: Summarize older messages to save tokens.

### 3.2 TUI Implementation
- Connect the `bubbletea` model to the new Streaming Engine.
- specific views for:
  - Chat (Streaming text)
  - Tool Log (Show what the agent is doing in background)
  - Task List (Kanban style)

---

## Suggested Libraries for Refactoring
- **Database:** `gorm.io/gorm` (ORM) or `sqlx` (Raw SQL helper).
- **Logging:** `log/slog` (Go 1.21+ standard).
- **CLI/TUI:** Keep `charmbracelet/bubbletea` and `cobra` (if needing complex flags).
- **Embeddings (Optional):** `github.com/tmc/langchaingo` (for future RAG features).
