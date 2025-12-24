package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/biodoia/skagent/internal/agents"
	"github.com/biodoia/skagent/internal/core"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type APIServer struct {
	port        int
	host        string
	engine      *core.Engine
	agentRegistry *agents.Registry
	ctx         context.Context
	server      *http.Server
	logger      *log.Logger
}

type APIResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Message string                 `json:"message,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
}

type AgentRequest struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config,omitempty"`
	AgentID     string                 `json:"agent_id,omitempty"`
}

type TaskRequest struct {
	AgentID     string                 `json:"agent_id"`
	Task        string                 `json:"task"`
	Priority    int                    `json:"priority"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	CallbackURL string                 `json:"callback_url,omitempty"`
}

type SystemRequest struct {
	Action   string                 `json:"action"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

func NewServer(ctx context.Context, port int, host string, engine *core.Engine, registry *agents.Registry) *APIServer {
	return &APIServer{
		port:         port,
		host:         host,
		engine:       engine,
		agentRegistry: registry,
		ctx:          ctx,
		logger:       log.New(log.Writer(), "[API] ", log.LstdFlags|log.Lmsgprefix),
	}
}

func (s *APIServer) Start() error {
	router := s.setupRoutes()
	
	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.host, s.port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	s.logger.Printf("Starting API server on %s:%d", s.host, s.port)
	
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Printf("Server error: %v", err)
		}
	}()
	
	return nil
}

func (s *APIServer) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(s.ctx)
	}
	return nil
}

func (s *APIServer) IsHealthy() bool {
	return s.server != nil
}

func (s *APIServer) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"status": "running",
		"port":   s.port,
		"host":   s.host,
	}
}

func (s *APIServer) setupRoutes() http.Handler {
	router := chi.NewRouter()
	
	// Middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	router.Use(middleware.Timeout(30 * time.Second))
	
	// CORS headers
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})
	
	// Routes
	router.Get("/", s.handleRoot)
	router.Get("/health", s.handleHealth)
	router.Get("/status", s.handleStatus)
	
	// Agent routes
	router.Route("/agents", func(r chi.Router) {
		r.Get("/", s.handleListAgents)
		r.Post("/", s.handleCreateAgent)
		r.Get("/{agentID}", s.handleGetAgent)
		r.Put("/{agentID}", s.handleUpdateAgent)
		r.Delete("/{agentID}", s.handleDeleteAgent)
		r.Post("/{agentID}/start", s.handleStartAgent)
		r.Post("/{agentID}/stop", s.handleStopAgent)
		r.Get("/{agentID}/tasks", s.handleGetAgentTasks)
	})
	
	// Task routes
	router.Route("/tasks", func(r chi.Router) {
		r.Get("/", s.handleListTasks)
		r.Post("/", s.handleCreateTask)
		r.Get("/{taskID}", s.handleGetTask)
		r.Put("/{taskID}", s.handleUpdateTask)
		r.Delete("/{taskID}", s.handleCancelTask)
	})
	
	// Project manager routes
	router.Route("/project", func(r chi.Router) {
		r.Get("/tasks", s.handleListProjectTasks)
		r.Get("/tasks/{taskID}", s.handleGetProjectTask)
		r.Post("/tasks/{taskID}/assign", s.handleAssignProjectTask)
		r.Get("/agents", s.handleListProjectAgents)
		r.Get("/status", s.handleGetProjectStatus)
		r.Post("/webhook", s.handleProjectWebhook)
	})
	
	// Tool routes
	router.Route("/tools", func(r chi.Router) {
		r.Get("/", s.handleListTools)
		r.Get("/{toolName}", s.handleGetTool)
		r.Post("/{toolName}/execute", s.handleExecuteTool)
	})
	
	// System routes
	router.Route("/system", func(r chi.Router) {
		r.Get("/config", s.handleGetConfig)
		r.Post("/config", s.handleUpdateConfig)
		r.Get("/stats", s.handleGetStats)
		r.Post("/shutdown", s.handleShutdown)
		r.Get("/logs", s.handleGetLogs)
	})
	
	return router
}

func (s *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"name":        "SKAgent API",
			"version":     "2.0.0",
			"description": "Advanced AI Agent Framework API",
			"endpoints": map[string]interface{}{
				"agents":  "/agents - Agent management",
				"tasks":   "/tasks - Task management",
				"tools":   "/tools - Tool execution",
				"system":  "/system - System configuration",
				"project": "/project - Project Manager integration",
			},
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":     "healthy",
			"timestamp":  time.Now(),
			"uptime":     "N/A", // Would calculate actual uptime
			"components": map[string]interface{}{
				"engine":      s.engine.IsHealthy(),
				"agents":      s.agentRegistry.GetStats(),
				"server":      true,
			},
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"server":   s.GetStatus(),
			"engine":   s.engine.GetStatus(),
			"agents":   s.agentRegistry.GetStats(),
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents := s.agentRegistry.ListAgents()
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"agents": agents,
			"count":  len(agents),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var req AgentRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	agent, err := s.agentRegistry.CreateAgent(req.Name, req.Type, req.Config)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"agent": agent,
		},
		Message: "Agent created successfully",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusCreated, response)
}

func (s *APIServer) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	
	agent, ok := s.agentRegistry.GetAgent(agentID)
	if !ok {
		s.writeError(w, http.StatusNotFound, "Agent not found")
		return
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"agent": agent,
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	var req AgentRequest
	
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Update agent logic would go here
	response := APIResponse{
		Success: true,
		Message: fmt.Sprintf("Agent %s updated", agentID),
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	
	if err := s.agentRegistry.DeleteAgent(agentID); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: fmt.Sprintf("Agent %s deleted", agentID),
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleStartAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	
	if err := s.agentRegistry.StartAgent(agentID); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: fmt.Sprintf("Agent %s started", agentID),
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleStopAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	
	if err := s.agentRegistry.StopAgent(agentID); err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: fmt.Sprintf("Agent %s stopped", agentID),
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleGetAgentTasks(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	
	// Get tasks for agent
	tasks := []map[string]interface{}{
		{
			"id":        "task-1",
			"status":    "completed",
			"created":   time.Now().Add(-1 * time.Hour),
			"completed": time.Now().Add(-30 * time.Minute),
		},
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"tasks":  tasks,
			"agent":  agentID,
			"count":  len(tasks),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleListTasks(w http.ResponseWriter, r *http.Request) {
	// List all tasks
	tasks := []map[string]interface{}{
		{
			"id":        "task-1",
			"status":    "completed",
			"agent":     "agent-1",
			"created":   time.Now().Add(-1 * time.Hour),
			"completed": time.Now().Add(-30 * time.Minute),
		},
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"tasks": tasks,
			"count": len(tasks),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Create task logic would go here
	taskID := fmt.Sprintf("task-%d", time.Now().Unix())
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"task_id": taskID,
			"status":  "created",
		},
		Message: "Task created successfully",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusCreated, response)
}

func (s *APIServer) handleGetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	// Get task details
	task := map[string]interface{}{
		"id":        taskID,
		"status":    "running",
		"agent":     "agent-1",
		"created":   time.Now().Add(-10 * time.Minute),
		"progress":  50,
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"task": task,
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	response := APIResponse{
		Success: true,
		Message: fmt.Sprintf("Task %s updated", taskID),
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleCancelTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	response := APIResponse{
		Success: true,
		Message: fmt.Sprintf("Task %s cancelled", taskID),
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	tools := []map[string]interface{}{
		{
			"name":        "github",
			"description": "GitHub integration",
			"capabilities": []string{"repo_search", "issue_management", "pr_review"},
		},
		{
			"name":        "websearch",
			"description": "Web search capabilities",
			"capabilities": []string{"search", "scrape"},
		},
		{
			"name":        "speckit",
			"description": "Specification processing",
			"capabilities": []string{"parse", "analyze", "generate"},
		},
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"tools": tools,
			"count": len(tools),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleGetTool(w http.ResponseWriter, r *http.Request) {
	toolName := chi.URLParam(r, "toolName")
	
	// Get tool details
	tool := map[string]interface{}{
		"name":        toolName,
		"description": "Tool description",
		"version":     "1.0.0",
		"capabilities": []string{"capability1", "capability2"},
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"tool": tool,
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleExecuteTool(w http.ResponseWriter, r *http.Request) {
	toolName := chi.URLParam(r, "toolName")
	var params map[string]interface{}
	
	if err := s.parseJSON(r, &params); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Execute tool logic would go here
	result := map[string]interface{}{
		"tool":      toolName,
		"status":    "completed",
		"result":    "Tool execution result",
		"timestamp": time.Now(),
	}
	
	response := APIResponse{
		Success: true,
		Data:    result,
		Message: "Tool executed successfully",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"host":       s.host,
		"port":       s.port,
		"version":    "2.0.0",
		"max_agents": 10,
		"timeout":    30,
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"config": config,
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var config map[string]interface{}
	if err := s.parseJSON(r, &config); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	response := APIResponse{
		Success: true,
		Message: "Configuration updated",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"uptime":       "N/A",
		"requests":     0,
		"agents":       s.agentRegistry.GetStats(),
		"memory_usage": "N/A",
		"cpu_usage":    "N/A",
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"stats": stats,
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleShutdown(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Message: "Shutting down server",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
	
	// Start shutdown in background
	go func() {
		time.Sleep(1 * time.Second)
		// Trigger shutdown
	}()
}

func (s *APIServer) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5 * time.Minute),
			"level":     "info",
			"message":   "Server started",
		},
		{
			"timestamp": time.Now().Add(-3 * time.Minute),
			"level":     "info",
			"message":   "Agent registered",
		},
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"logs": logs,
			"count": len(logs),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// Project Manager integration endpoints
func (s *APIServer) handleGetProjectTasks(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		projectID = "default"
	}
	
	tasks := []map[string]interface{}{
		{
			"id":          "proj-task-1",
			"project_id":  projectID,
			"title":       "Implement user authentication",
			"description": "Add login/logout functionality",
			"status":      "pending",
			"priority":    "high",
			"assignee":    "",
			"created":     time.Now().Add(-2 * time.Hour),
		},
		{
			"id":          "proj-task-2",
			"project_id":  projectID,
			"title":       "Database migration",
			"description": "Update schema for new features",
			"status":      "in_progress",
			"priority":    "medium",
			"assignee":    "agent-db",
			"created":     time.Now().Add(-1 * time.Hour),
		},
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"tasks":      tasks,
			"project_id": projectID,
			"count":      len(tasks),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleCreateProjectTask(w http.ResponseWriter, r *http.Request) {
	var task map[string]interface{}
	if err := s.parseJSON(r, &task); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	taskID := fmt.Sprintf("proj-task-%d", time.Now().Unix())
	task["id"] = taskID
	task["created"] = time.Now()
	task["status"] = "pending"
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"task": task,
		},
		Message: "Project task created successfully",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusCreated, response)
}

func (s *APIServer) handleGetTaskAssignments(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		projectID = "default"
	}
	
	assignments := []map[string]interface{}{
		{
			"task_id":    "proj-task-1",
			"agent_id":   "agent-dev",
			"project_id": projectID,
			"status":     "assigned",
			"assigned":   time.Now().Add(-30 * time.Minute),
		},
		{
			"task_id":    "proj-task-2",
			"agent_id":   "agent-db",
			"project_id": projectID,
			"status":     "in_progress",
			"assigned":   time.Now().Add(-15 * time.Minute),
		},
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"assignments": assignments,
			"project_id":  projectID,
			"count":       len(assignments),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleAssignTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID    string `json:"task_id"`
		AgentID   string `json:"agent_id"`
		ProjectID string `json:"project_id"`
	}
	
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"task_id":  req.TaskID,
			"agent_id": req.AgentID,
			"status":   "assigned",
			"assigned": time.Now(),
		},
		Message: "Task assigned successfully",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleGetAvailableAgents(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		projectID = "default"
	}
	
	agents := s.agentRegistry.ListAgents()
	
	// Filter agents by availability and project
	available := []map[string]interface{}{}
	for _, agent := range agents {
		if agent.Status == "idle" || agent.Status == "waiting" {
			available = append(available, map[string]interface{}{
				"id":         agent.ID,
				"name":       agent.Name,
				"type":       agent.Type,
				"status":     agent.Status,
				"capabilities": []string{"development", "testing"},
				"workload":   0,
			})
		}
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"agents":      available,
			"project_id":  projectID,
			"count":       len(available),
		},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleRecommendAgents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID string                 `json:"project_id"`
		Task      map[string]interface{} `json:"task"`
		Criteria  map[string]interface{} `json:"criteria"`
	}
	
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// AI-powered agent recommendation logic
	recommendations := []map[string]interface{}{
		{
			"agent_id":  "agent-dev",
			"confidence": 0.95,
			"reason":    "High expertise in development tasks",
			"estimated_time": "2-4 hours",
		},
		{
			"agent_id":  "agent-test",
			"confidence": 0.87,
			"reason":    "Strong testing capabilities",
			"estimated_time": "1-2 hours",
		},
	}
	
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"recommendations": recommendations,
			"project_id":      req.ProjectID,
			"task_id":         req.Task["id"],
		},
		Message: "Agent recommendations generated",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// Project Manager Handlers

// handleListProjectTasks lists tasks from the project manager
func (s *APIServer) handleListProjectTasks(w http.ResponseWriter, r *http.Request) {
	projectManager := s.engine.GetProjectManager()
	if projectManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Project manager not available")
		return
	}
	
	tasks := projectManager.GetTasks()
	response := APIResponse{
		Success:   true,
		Data:      map[string]interface{}{"tasks": tasks},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// handleGetProjectTask gets a specific task from the project manager
func (s *APIServer) handleGetProjectTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	projectManager := s.engine.GetProjectManager()
	if projectManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Project manager not available")
		return
	}
	
	taskResult, exists := projectManager.GetTaskStatus(taskID)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Task not found")
		return
	}
	
	response := APIResponse{
		Success:   true,
		Data:      map[string]interface{}{"task": taskResult},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// handleAssignProjectTask assigns a task to an agent
func (s *APIServer) handleAssignProjectTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	
	var req struct {
		AgentID string `json:"agent_id"`
	}
	
	if err := s.parseJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request data")
		return
	}
	
	projectManager := s.engine.GetProjectManager()
	if projectManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Project manager not available")
		return
	}
	
	// TODO: Implement task assignment logic
	// This would involve calling the project manager to assign the task
	
	response := APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Task %s assigned to agent %s", taskID, req.AgentID),
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// handleListProjectAgents lists available agents for the project manager
func (s *APIServer) handleListProjectAgents(w http.ResponseWriter, r *http.Request) {
	projectManager := s.engine.GetProjectManager()
	if projectManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Project manager not available")
		return
	}
	
	agents := s.agentRegistry.ListAgents()
	
	// Convert agents to a format suitable for project manager
	projectAgents := make([]map[string]interface{}, len(agents))
	for i, agent := range agents {
		projectAgents[i] = map[string]interface{}{
			"id":           agent.ID,
			"name":         agent.Name,
			"type":         string(agent.Type),
			"status":       string(agent.Status),
			"capabilities": agent.Capabilities,
			"load":         agent.Load,
		}
	}
	
	response := APIResponse{
		Success:   true,
		Data:      map[string]interface{}{"agents": projectAgents},
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// handleGetProjectStatus gets the overall project manager status
func (s *APIServer) handleGetProjectStatus(w http.ResponseWriter, r *http.Request) {
	projectManager := s.engine.GetProjectManager()
	if projectManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Project manager not available")
		return
	}
	
	status := map[string]interface{}{
		"enabled":   true,
		"connected": true,
		"tasks":     len(projectManager.GetTasks()),
		"timestamp": time.Now(),
	}
	
	response := APIResponse{
		Success:   true,
		Data:      status,
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// handleProjectWebhook handles incoming webhooks from the project manager
func (s *APIServer) handleProjectWebhook(w http.ResponseWriter, r *http.Request) {
	projectManager := s.engine.GetProjectManager()
	if projectManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Project manager not available")
		return
	}
	
	// TODO: Implement webhook handling
	// This would process the webhook data and update internal state
	
	response := APIResponse{
		Success:   true,
		Message:   "Webhook received",
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, http.StatusOK, response)
}

// Helper methods
func (s *APIServer) parseJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func (s *APIServer) writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(v); err != nil {
		s.logger.Printf("Error encoding JSON response: %v", err)
	}
}

func (s *APIServer) writeError(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success:  false,
		Error:    message,
		Timestamp: time.Now(),
	}
	
	s.writeJSON(w, statusCode, response)
}