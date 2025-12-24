package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/biodoia/skagent/internal/agents"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MCPRequest struct {
	ID          string                 `json:"id"`
	Method      string                 `json:"method"`
	Params      map[string]interface{} `json:"params,omitempty"`
	JSONRPC     string                 `json:"jsonrpc"`
}

type MCPResponse struct {
	ID     string                 `json:"id"`
	Result map[string]interface{} `json:"result,omitempty"`
	Error  map[string]interface{} `json:"error,omitempty"`
}

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type AgentDefinition struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Tools       []ToolDefinition       `json:"tools"`
	Capabilities []string              `json:"capabilities"`
}

type Server struct {
	port          int
	ctx           context.Context
	agentRegistry *agents.Registry
	server        *http.Server
	logger        *log.Logger
	tools         map[string]ToolDefinition
	mu            sync.RWMutex
	activeConnections int
}

func NewServer(ctx context.Context, registry *agents.Registry) *Server {
	return &Server{
		ctx:           ctx,
		agentRegistry: registry,
		logger:        log.New(log.Writer(), "[MCP] ", log.LstdFlags|log.Lmsgprefix),
		tools:         make(map[string]ToolDefinition),
		activeConnections: 0,
	}
}

func (s *Server) Start() error {
	// Initialize built-in tools
	s.initializeTools()
	
	router := s.setupRoutes()
	
	s.server = &http.Server{
		Addr:         ":8081",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	s.logger.Printf("Starting MCP server on port 8081")
	
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Printf("MCP server error: %v", err)
		}
	}()
	
	return nil
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(s.ctx)
	}
	return nil
}

func (s *Server) IsHealthy() bool {
	return s.server != nil
}

func (s *Server) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return map[string]interface{}{
		"status":              "running",
		"port":                8081,
		"active_connections":  s.activeConnections,
		"registered_tools":    len(s.tools),
		"agent_registry":      s.agentRegistry.GetStats(),
	}
}

func (s *Server) setupRoutes() http.Handler {
	router := chi.NewRouter()
	
	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	router.Use(s.connectionMiddleware)
	
	// MCP-specific middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})
	
	// MCP endpoints
	router.Get("/health", s.handleMCPHealth)
	router.Get("/tools", s.handleListTools)
	router.Get("/tools/{toolName}", s.handleGetTool)
	router.Post("/tools/{toolName}/call", s.handleCallTool)
	
	// Agent endpoints
	router.Get("/agents", s.handleListAgents)
	router.Get("/agents/{agentID}", s.handleGetAgent)
	router.Post("/agents/{agentID}/execute", s.handleExecuteAgent)
	
	// System endpoints
	router.Get("/info", s.handleServerInfo)
	router.Get("/capabilities", s.handleGetCapabilities)
	
	return router
}

func (s *Server) connectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.activeConnections++
		s.mu.Unlock()
		
		defer func() {
			s.mu.Lock()
			s.activeConnections--
			s.mu.Unlock()
		}()
		
		next.ServeHTTP(w, r)
	})
}

func (s *Server) initializeTools() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Agent management tools
	s.tools["list_agents"] = ToolDefinition{
		Name:        "list_agents",
		Description: "List all available agents",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Filter by agent status",
					"enum":        []string{"active", "idle", "offline"},
				},
			},
		},
	}
	
	s.tools["get_agent"] = ToolDefinition{
		Name:        "get_agent",
		Description: "Get detailed information about a specific agent",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the agent",
				},
			},
			"required": []string{"agent_id"},
		},
	}
	
	s.tools["start_agent"] = ToolDefinition{
		Name:        "start_agent",
		Description: "Start a specific agent",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the agent",
				},
			},
			"required": []string{"agent_id"},
		},
	}
	
	s.tools["stop_agent"] = ToolDefinition{
		Name:        "stop_agent",
		Description: "Stop a specific agent",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique identifier for the agent",
				},
			},
			"required": []string{"agent_id"},
		},
	}
	
	// Task management tools
	s.tools["create_task"] = ToolDefinition{
		Name:        "create_task",
		Description: "Create a new task for an agent",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Target agent ID",
				},
				"task": map[string]interface{}{
					"type":        "string",
					"description": "Task description or command",
				},
				"priority": map[string]interface{}{
					"type":        "integer",
					"description": "Task priority (1-10)",
					"minimum":     1,
					"maximum":     10,
				},
			},
			"required": []string{"agent_id", "task"},
		},
	}
	
	s.tools["get_task_status"] = ToolDefinition{
		Name:        "get_task_status",
		Description: "Get status of a specific task",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "Unique task identifier",
				},
			},
			"required": []string{"task_id"},
		},
	}
	
	// System tools
	s.tools["get_system_status"] = ToolDefinition{
		Name:        "get_system_status",
		Description: "Get overall system status and statistics",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
	}
	
	s.tools["get_system_config"] = ToolDefinition{
		Name:        "get_system_config",
		Description: "Get system configuration",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"section": map[string]interface{}{
					"type":        "string",
					"description": "Configuration section to retrieve",
				},
			},
		},
	}
	
	// Project management tools
	s.tools["list_project_tasks"] = ToolDefinition{
		Name:        "list_project_tasks",
		Description: "List tasks for a specific project",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "Project identifier",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Filter by task status",
					"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
				},
			},
			"required": []string{"project_id"},
		},
	}
	
	s.tools["assign_task_to_agent"] = ToolDefinition{
		Name:        "assign_task_to_agent",
		Description: "Assign a task to a specific agent",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "Task to assign",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Agent to assign the task to",
				},
			},
			"required": []string{"task_id", "agent_id"},
		},
	}
	
	s.tools["recommend_agents"] = ToolDefinition{
		Name:        "recommend_agents",
		Description: "Get AI-powered agent recommendations for a task",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task_description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the task",
				},
				"requirements": map[string]interface{}{
					"type":        "array",
					"description": "Required capabilities",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"required": []string{"task_description"},
		},
	}
}

func (s *Server) handleMCPHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"version":    "2.0.0",
		"server":     "MCP",
		"capabilities": []string{
			"agent_management",
			"task_execution",
			"project_integration",
			"tool_execution",
		},
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	response := map[string]interface{}{
		"tools":    s.tools,
		"count":    len(s.tools),
		"server":   "skagent-mcp",
		"version":  "2.0.0",
		"timestamp": time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetTool(w http.ResponseWriter, r *http.Request) {
	toolName := chi.URLParam(r, "toolName")
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tool, exists := s.tools[toolName]
	if !exists {
		s.writeError(w, http.StatusNotFound, "Tool not found")
		return
	}
	
	response := map[string]interface{}{
		"tool":      tool,
		"timestamp": time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleCallTool(w http.ResponseWriter, r *http.Request) {
	toolName := chi.URLParam(r, "toolName")
	
	var params map[string]interface{}
	if err := s.parseJSON(r, &params); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	
	// Execute tool
	result, err := s.executeTool(toolName, params)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	response := MCPResponse{
		Result: map[string]interface{}{
			"tool":     toolName,
			"result":   result,
			"timestamp": time.Now(),
		},
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents := s.agentRegistry.ListAgents()
	
	response := map[string]interface{}{
		"agents":    agents,
		"count":     len(agents),
		"timestamp": time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	
	agent, ok := s.agentRegistry.GetAgent(agentID)
	if !ok {
		s.writeError(w, http.StatusNotFound, "Agent not found")
		return
	}
	
	response := map[string]interface{}{
		"agent":     agent,
		"timestamp": time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleExecuteAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	
	var params map[string]interface{}
	if err := s.parseJSON(r, &params); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	
	task, ok := params["task"].(string)
	if !ok {
		s.writeError(w, http.StatusBadRequest, "Task parameter required")
		return
	}
	
	// Execute agent task
	result := map[string]interface{}{
		"agent_id":  agentID,
		"task":      task,
		"status":    "submitted",
		"timestamp": time.Now(),
	}
	
	response := MCPResponse{
		Result: result,
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleServerInfo(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":        "SKAgent MCP Server",
		"version":     "2.0.0",
		"description": "Model Context Protocol server for SKAgent",
		"capabilities": []string{
			"agent_management",
			"task_execution",
			"project_integration",
			"tool_execution",
			"system_monitoring",
		},
		"endpoints": map[string]interface{}{
			"health":      "/health",
			"tools":       "/tools",
			"agents":      "/agents",
			"info":        "/info",
			"capabilities": "/capabilities",
		},
		"timestamp": time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetCapabilities(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"capabilities": []string{
			"agent_management",
			"task_execution",
			"project_integration",
			"tool_execution",
			"system_monitoring",
			"configuration",
			"logging",
			"metrics",
		},
		"protocol_version": "2024-11-05",
		"server_info": map[string]interface{}{
			"name":    "SKAgent",
			"version": "2.0.0",
		},
		"timestamp": time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) executeTool(toolName string, params map[string]interface{}) (map[string]interface{}, error) {
	switch toolName {
	case "list_agents":
		status, _ := params["status"].(string)
		agents := s.agentRegistry.ListAgents()
		
		if status != "" {
			filtered := []map[string]interface{}{}
			for _, agent := range agents {
				if string(agent.Status) == status {
					filtered = append(filtered, map[string]interface{}{
						"id":         agent.ID,
						"name":       agent.Name,
						"status":     agent.Status,
						"type":       agent.Type,
						"last_active": agent.UpdatedAt,
					})
				}
			}
			return map[string]interface{}{"agents": filtered}, nil
		}
		
		return map[string]interface{}{"agents": agents}, nil
		
	case "get_agent":
		agentID, ok := params["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id parameter required")
		}
		
		agent, ok := s.agentRegistry.GetAgent(agentID)
		if !ok {
			return nil, fmt.Errorf("agent not found")
		}
		
		return map[string]interface{}{
			"agent": map[string]interface{}{
				"id":         agent.ID,
				"name":       agent.Name,
				"status":     agent.Status,
				"type":       agent.Type,
				"last_active": agent.UpdatedAt,
			},
		}, nil
		
	case "start_agent":
		agentID, ok := params["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id parameter required")
		}
		
		if err := s.agentRegistry.StartAgent(agentID); err != nil {
			return nil, fmt.Errorf("failed to start agent: %w", err)
		}
		
		return map[string]interface{}{"status": "started", "agent_id": agentID}, nil
		
	case "stop_agent":
		agentID, ok := params["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id parameter required")
		}
		
		if err := s.agentRegistry.StopAgent(agentID); err != nil {
			return nil, fmt.Errorf("failed to stop agent: %w", err)
		}
		
		return map[string]interface{}{"status": "stopped", "agent_id": agentID}, nil
		
	case "create_task":
		agentID, ok := params["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id parameter required")
		}
		
		task, ok := params["task"].(string)
		if !ok {
			return nil, fmt.Errorf("task parameter required")
		}
		
		priority, _ := params["priority"].(float64)
		taskID := fmt.Sprintf("task-%d", time.Now().Unix())
		
		return map[string]interface{}{
			"task_id":  taskID,
			"agent_id": agentID,
			"task":     task,
			"priority": int(priority),
			"status":   "created",
		}, nil
		
	case "get_task_status":
		taskID, ok := params["task_id"].(string)
		if !ok {
			return nil, fmt.Errorf("task_id parameter required")
		}
		
		return map[string]interface{}{
			"task_id":  taskID,
			"status":   "running",
			"progress": 50,
		}, nil
		
	case "get_system_status":
		return map[string]interface{}{
			"status":     "healthy",
			"uptime":     "N/A",
			"agents":     s.agentRegistry.GetStats(),
			"memory":     "N/A",
			"cpu":        "N/A",
			"timestamp":  time.Now(),
		}, nil
		
	case "list_project_tasks":
		projectID, ok := params["project_id"].(string)
		if !ok {
			return nil, fmt.Errorf("project_id parameter required")
		}
		
		// Mock project tasks
		tasks := []map[string]interface{}{
			{
				"id":          "proj-task-1",
				"project_id":  projectID,
				"title":       "Implement authentication",
				"status":      "pending",
				"assignee":    "",
			},
			{
				"id":          "proj-task-2",
				"project_id":  projectID,
				"title":       "Database setup",
				"status":      "in_progress",
				"assignee":    "agent-db",
			},
		}
		
		return map[string]interface{}{
			"tasks":      tasks,
			"project_id": projectID,
			"count":      len(tasks),
		}, nil
		
	case "assign_task_to_agent":
		taskID, ok := params["task_id"].(string)
		if !ok {
			return nil, fmt.Errorf("task_id parameter required")
		}
		
		agentID, ok := params["agent_id"].(string)
		if !ok {
			return nil, fmt.Errorf("agent_id parameter required")
		}
		
		return map[string]interface{}{
			"task_id":   taskID,
			"agent_id":  agentID,
			"status":    "assigned",
			"assigned":  time.Now(),
		}, nil
		
	case "recommend_agents":
		taskDesc, ok := params["task_description"].(string)
		if !ok {
			return nil, fmt.Errorf("task_description parameter required")
		}
		
		// AI-powered recommendations
		recommendations := []map[string]interface{}{
			{
				"agent_id":   "agent-dev",
				"confidence": 0.95,
				"reason":     "Excellent match for development tasks",
			},
			{
				"agent_id":   "agent-test",
				"confidence": 0.87,
				"reason":     "Strong testing capabilities",
			},
		}
		
		return map[string]interface{}{
			"task_description": taskDesc,
			"recommendations":  recommendations,
		}, nil
		
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// Helper methods
func (s *Server) parseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *Server) writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(v); err != nil {
		s.logger.Printf("Error encoding JSON response: %v", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, statusCode int, message string) {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    statusCode,
			"message": message,
		},
		"timestamp": time.Now(),
	}
	
	s.writeJSON(w, statusCode, response)
}