package project

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// NewWebhookServer creates a new webhook server
func NewWebhookServer(manager *Manager, port int) *WebhookServer {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: manager.createWebhookHandler(),
	}
	
	return &WebhookServer{
		manager: manager,
		server:  server,
	}
}

// Start starts the webhook server
func (ws *WebhookServer) Start() error {
	go func() {
		log.Printf("Starting webhook server on %s", ws.server.Addr)
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Webhook server error: %v", err)
		}
	}()
	return nil
}

// Stop stops the webhook server
func (ws *WebhookServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ws.server.Shutdown(ctx)
}

// createWebhookHandler creates the HTTP handler for webhooks
func (m *Manager) createWebhookHandler() http.Handler {
	mux := http.NewServeMux()
	
	// Handle webhook events
	mux.HandleFunc("/webhook", m.handleWebhook)
	
	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	
	return mux
}

// handleWebhook handles incoming webhook events
func (m *Manager) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		m.logger.Printf("Failed to decode webhook event: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	m.logger.Printf("Received webhook event: %s", event.Type)
	
	// Process different event types
	switch event.Type {
	case "task.created":
		m.handleTaskCreated(event)
	case "task.updated":
		m.handleTaskUpdated(event)
	case "task.assigned":
		m.handleTaskAssigned(event)
	default:
		m.logger.Printf("Unknown webhook event type: %s", event.Type)
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
}

// WebhookEvent represents a webhook event from the project manager
type WebhookEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// handleTaskCreated handles new task creation events
func (m *Manager) handleTaskCreated(event WebhookEvent) {
	taskData, ok := event.Data["task"].(map[string]interface{})
	if !ok {
		m.logger.Printf("Invalid task data in event")
		return
	}
	
	// Convert to Task struct
	taskJSON, err := json.Marshal(taskData)
	if err != nil {
		m.logger.Printf("Failed to marshal task data: %v", err)
		return
	}
	
	var task Task
	if err := json.Unmarshal(taskJSON, &task); err != nil {
		m.logger.Printf("Failed to unmarshal task: %v", err)
		return
	}
	
	// Store task
	m.taskMutex.Lock()
	m.tasks[task.ID] = &task
	m.taskMutex.Unlock()
	
	// Auto-assign if enabled
	if m.config.AutoAssign && task.Assignee == "" {
		m.autoAssignTask(&task)
	}
	
	m.logger.Printf("Processed new task: %s", task.Title)
}

// handleTaskUpdated handles task update events
func (m *Manager) handleTaskUpdated(event WebhookEvent) {
	taskID, ok := event.Data["task_id"].(string)
	if !ok {
		m.logger.Printf("Invalid task_id in update event")
		return
	}
	
	// Update task if we have it
	m.taskMutex.Lock()
	defer m.taskMutex.Unlock()
	
	if task, exists := m.tasks[taskID]; exists {
		// Update task fields based on event data
		if status, ok := event.Data["status"].(string); ok {
			task.Status = status
		}
		if assignee, ok := event.Data["assignee"].(string); ok {
			task.Assignee = assignee
		}
		
		m.logger.Printf("Updated task %s", taskID)
	}
}

// handleTaskAssigned handles task assignment events
func (m *Manager) handleTaskAssigned(event WebhookEvent) {
	taskID, ok := event.Data["task_id"].(string)
	if !ok {
		m.logger.Printf("Invalid task_id in assignment event")
		return
	}
	
	agentID, ok := event.Data["agent_id"].(string)
	if !ok {
		m.logger.Printf("Invalid agent_id in assignment event")
		return
	}
	
	// Create or update assignment
	m.taskMutex.Lock()
	defer m.taskMutex.Unlock()
	
	assignment := &TaskAssignment{
		TaskID:      taskID,
		AgentID:     agentID,
		AssignedAt:  time.Now(),
		Status:      "assigned",
	}
	
	m.assignments[taskID] = assignment
	
	// Start execution if task is ready
	if task, exists := m.tasks[taskID]; exists && task.Status == "todo" {
		go m.executeTask(assignment)
	}
	
	m.logger.Printf("Processed assignment: task %s -> agent %s", taskID, agentID)
}