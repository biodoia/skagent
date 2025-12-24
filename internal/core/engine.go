package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/biodoia/skagent/internal/ai"
	"github.com/biodoia/skagent/internal/agents"
	"github.com/biodoia/skagent/internal/config"
	"github.com/biodoia/skagent/internal/project"
	"github.com/biodoia/skagent/internal/tools"
)

// Engine is the core processing engine
type Engine struct {
	config         *config.Config
	provider       ai.Provider
	tools          *tools.ToolManager
	agentRegistry  *agents.Registry
	projectManager *project.Manager
	sessions       map[string]*Session
	mu             sync.RWMutex
}

// Session represents a conversation session
type Session struct {
	ID        string       `json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Messages  []Message    `json:"messages"`
	Metadata  SessionMeta  `json:"metadata"`
}

// SessionMeta contains session metadata
type SessionMeta struct {
	Title       string            `json:"title,omitempty"`
	Autonomous  bool              `json:"autonomous"`
	AgentID     string            `json:"agent_id,omitempty"`
	ProjectID   string            `json:"project_id,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// Message represents a conversation message
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // user, assistant, system, tool
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Metadata  MsgMeta   `json:"metadata,omitempty"`
}

// ToolCall represents a tool invocation
type ToolCall struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Input  string `json:"input"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// MsgMeta contains message metadata
type MsgMeta struct {
	Model    string `json:"model,omitempty"`
	Tokens   int    `json:"tokens,omitempty"`
	Duration int64  `json:"duration_ms,omitempty"`
}

// NewEngine creates a new engine instance
func NewEngine(ctx context.Context, cfg *config.Config, agentRegistry *agents.Registry) (*Engine, error) {
	provider, err := ai.CreateProvider(cfg)
	if err != nil {
		return nil, err
	}

	tm := tools.NewToolManager()
	tm.AddTool(tools.NewSpecKitTool(""))
	tm.AddTool(tools.NewGitHubTool(""))
	tm.AddTool(tools.NewWebSearchTool())

	engine := &Engine{
		config:        cfg,
		provider:      provider,
		tools:         tm,
		agentRegistry: agentRegistry,
		sessions:      make(map[string]*Session),
	}

	// Initialize project manager if enabled
	if cfg.IsProjectEnabled() {
		projectClient := project.NewClient(cfg.Project.BaseURL, cfg.Project.APIKey)
		projectManager := project.NewManager(projectClient, agentRegistry, cfg.GetProjectConfig())
		engine.projectManager = projectManager
	}

	return engine, nil
}

// CreateSession creates a new conversation session
func (e *Engine) CreateSession() *Session {
	session := &Session{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []Message{},
		Metadata:  SessionMeta{},
	}

	e.mu.Lock()
	e.sessions[session.ID] = session
	e.mu.Unlock()

	return session
}

// GetSession returns a session by ID
func (e *Engine) GetSession(id string) (*Session, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	session, ok := e.sessions[id]
	return session, ok
}

// ListSessions returns all sessions
func (e *Engine) ListSessions() []*Session {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	sessions := make([]*Session, 0, len(e.sessions))
	for _, s := range e.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// DeleteSession removes a session
func (e *Engine) DeleteSession(id string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if _, ok := e.sessions[id]; ok {
		delete(e.sessions, id)
		return true
	}
	return false
}

// ProcessInput handles user input and returns response
type ProcessResult struct {
	Response   string     `json:"response"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	Error      error      `json:"-"`
	TokensUsed int        `json:"tokens_used,omitempty"`
	Duration   int64      `json:"duration_ms"`
}

// Process handles a user message in a session
func (e *Engine) Process(ctx context.Context, sessionID, input string) (*ProcessResult, error) {
	session, ok := e.GetSession(sessionID)
	if !ok {
		return nil, ErrSessionNotFound
	}

	start := time.Now()

	// Add user message
	userMsg := Message{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   input,
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, userMsg)
	session.UpdatedAt = time.Now()

	// Convert to AI messages
	aiMessages := make([]ai.Message, len(session.Messages))
	for i, msg := range session.Messages {
		aiMessages[i] = ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Get system prompt
	systemPrompt := e.buildSystemPrompt(session)

	// Call AI provider
	response, err := e.provider.Complete(ctx, aiMessages, systemPrompt)
	if err != nil {
		return &ProcessResult{Error: err}, err
	}

	// Add assistant message
	assistantMsg := Message{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
		Metadata: MsgMeta{
			Duration: time.Since(start).Milliseconds(),
		},
	}
	session.Messages = append(session.Messages, assistantMsg)
	session.UpdatedAt = time.Now()

	return &ProcessResult{
		Response: response,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

// ProcessAutonomous handles autonomous mode processing
func (e *Engine) ProcessAutonomous(ctx context.Context, sessionID, input string) (*ProcessResult, error) {
	session, ok := e.GetSession(sessionID)
	if !ok {
		return nil, ErrSessionNotFound
	}

	session.Metadata.Autonomous = true

	// Enhanced prompt for autonomous mode
	enhancedInput := buildAutonomousPrompt(input)

	return e.Process(ctx, sessionID, enhancedInput)
}

func (e *Engine) buildSystemPrompt(session *Session) string {
	prompt := ai.SystemPrompt + "\n\n" + ai.SpecKitDocs

	if session.Metadata.Autonomous {
		prompt += "\n\nYou are in AUTONOMOUS mode. Be proactive and thorough. Execute tasks without asking for confirmation."
	}

	return prompt
}

func buildAutonomousPrompt(input string) string {
	return `Analyze this request and provide a comprehensive response:

"` + input + `"

Include:
1. Clear summary of what will be done
2. Key requirements and features
3. Suggested approach
4. Step-by-step execution plan

Be proactive and start working immediately.`
}

// Tools returns the tool manager
func (e *Engine) Tools() *tools.ToolManager {
	return e.tools
}

// Provider returns the AI provider
func (e *Engine) Provider() ai.Provider {
	return e.provider
}

// Config returns the configuration
func (e *Engine) Config() *config.Config {
	return e.config
}

// Errors
var (
	ErrSessionNotFound = NewError("session not found")
)

// Error represents an engine error
type Error struct {
	message string
}

func NewError(msg string) *Error {
	return &Error{message: msg}
}

func (e *Error) Error() string {
	return e.message
}

// IsHealthy returns true if the engine is healthy
func (e *Engine) IsHealthy() bool {
	return e.provider != nil && e.tools != nil
}

// GetStatus returns the current status of the engine
func (e *Engine) GetStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return map[string]interface{}{
		"status":    "running",
		"sessions":  len(e.sessions),
		"healthy":   e.IsHealthy(),
		"timestamp": time.Now(),
	}
}

// Start initializes the engine
func (e *Engine) Start() error {
	// Start project manager if enabled
	if e.projectManager != nil {
		if err := e.projectManager.Start(); err != nil {
			return fmt.Errorf("failed to start project manager: %w", err)
		}
	}
	return nil
}

// Stop gracefully shuts down the engine
func (e *Engine) Stop() error {
	// Stop project manager if enabled
	if e.projectManager != nil {
		if err := e.projectManager.Stop(); err != nil {
			return fmt.Errorf("failed to stop project manager: %w", err)
		}
	}
	return nil
}

// GetProjectManager returns the project manager instance
func (e *Engine) GetProjectManager() *project.Manager {
	return e.projectManager
}
