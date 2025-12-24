# Code Review & Technical Specifications: Skagent

**Date:** December 24, 2025
**Version:** Analysis of current `main` branch
**Target System:** Linux (Go Environment)

---

## 1. Executive Summary
`skagent` is a modular AI agent framework written in Go, designed to operate via a Terminal User Interface (TUI) or as a Model Context Protocol (MCP) server. It leverages the Anthropic API (Claude) for intelligence.

**Current Status:** Alpha/Prototype. The architectural skeleton is robust and idiomatic, but core logic regarding state persistence, AI streaming, and recursive tool execution is partial or mocked.

---

## 2. Technical Architecture

### 2.1 Technology Stack
- **Language:** Go (Golang) 1.22+
- **AI Provider:** Anthropic (`anthropics/anthropic-sdk-go`)
- **Web Server:** `go-chi/chi` (Lightweight, idiomatic router)
- **TUI:** `charmbracelet/bubbletea` (Elm architecture for terminal)
- **Config:** JSON-based with environment variable overrides
- **Protocol:** Model Context Protocol (MCP) for tool/agent exposure

### 2.2 Directory Structure
- `cmd/skagent`: Entry point. Orchestrates mode selection (Interactive vs Server).
- `internal/core`: Contains the `Engine` (central logic) and `Session` management.
- `internal/ai`: Wrapper around Anthropic SDK.
- `internal/server/mcp`: Implementation of the MCP server and tool definitions.
- `internal/agents`: Agent registry and state management.

---

## 3. Detailed Specifications (Current Implementation)

### 3.1 Core Engine (`internal/core/engine.go`)
- **Session Management:** In-memory `map[string]*Session`.
  - *Risk:* Data loss on restart. Not scalable across multiple instances.
- **Processing Logic:**
  - `Process(ctx, sessionID, input)`: Handles user input, appends to history, calls AI.
  - **Tool Handling Gap:** The engine creates `ToolCall` structures but currently lacks a recursive loop to:
    1. Execute the tool.
    2. Feed the result back to the LLM.
    3. Get the final response.
    *Current behavior looks like a single-turn request.*

### 3.2 AI Client (`internal/ai/client.go`)
- **Model:** Defaults to `claude-3-5-sonnet-20240620` (inferred).
- **Interaction:** Uses standard `Messages.New` (Unary call).
  - *Missing:* `NewStreaming` implementation for real-time feedback.
- **Context:** Maintains a simple history slice. No advanced context window management (e.g., sliding window or summarization).

### 3.3 MCP Server (`internal/server/mcp/mcp_server.go`)
- **Endpoints:**
  - `/health`, `/info`, `/capabilities` (System info)
  - `/tools`, `/tools/{name}/call` (Tool execution)
  - `/agents`, `/agents/{id}/execute` (Agent control)
- **Implemented Tools:**
  - `list_agents`, `get_agent`, `start/stop_agent` (Registry interaction)
  - `create_task`, `get_task_status` (Mocked logic)
  - `recommend_agents` (Mocked logic)
- **Concurrency:** Uses `sync.RWMutex` for thread safety on the map-based registry.

---

## 4. Code Review Findings & Critical Gaps

### High Priority
1.  **No Persistence:** Sessions and Agent states are stored in Go maps.
    - *Impact:* Critical. Cannot be used for long-running tasks or deployed production systems.
2.  **Broken Tool Loop:** The `Engine` detects tools but does not seem to execute them automatically and return results to the LLM in the same turn.
    - *Impact:* The "Agent" capabilities are severely limited to just "talking" about tools rather than using them.
3.  **No Streaming:** The UI waits for the full response.
    - *Impact:* Poor UX for long generation tasks (typical of coding agents).

### Medium Priority
1.  **Mock Data:** Many MCP tools return hardcoded or mock data (`mock project tasks`, `mock recommendations`).
2.  **Error Handling:** Basic error wrapping. No structured error types for API failures vs Internal failures.
3.  **Configuration:** Hardcoded API keys in some incomplete sections (though `config` package exists, usage isn't uniform).

### Low Priority
1.  **TUI Completeness:** The TUI code exists but `main.go` suggests it's full of placeholders.
2.  **Logging:** Uses standard `log`. Should move to `slog` (structured logging) for production observability.
