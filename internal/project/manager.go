package project

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/biodoia/skagent/internal/agents"
	"github.com/biodoia/skagent/internal/config"
)

// Manager orchestrates project manager integration
type Manager struct {
	client       *Client
	agentRegistry *agents.Registry
	config       config.ProjectConfig
	logger       *log.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	
	// Task tracking
	tasks        map[string]*Task
	assignments  map[string]*TaskAssignment
	taskMutex    sync.RWMutex
	
	// Webhook handling
	webhookServer *WebhookServer
}

// AssignRule defines automatic task assignment rules
// TODO: Implement when config structure supports it
/*
type AssignRule struct {
	Category   string            `json:"category"`
	Keywords   []string          `json:"keywords"`
	AgentType  string            `json:"agent_type"`
	Priority   int               `json:"priority"`
	Conditions map[string]string `json:"conditions"`
}
*/

// WebhookServer handles incoming webhooks from project manager
type WebhookServer struct {
	manager *Manager
	server  *http.Server
}

// TaskAssignmentResult represents the result of task assignment
type TaskAssignmentResult struct {
	AssignmentID string                 `json:"assignment_id"`
	TaskID       string                 `json:"task_id"`
	AgentID      string                 `json:"agent_id"`
	Status       string                 `json:"status"`
	Result       map[string]interface{} `json:"result"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// NewManager creates a new project manager
func NewManager(client *Client, agentRegistry *agents.Registry, config config.ProjectConfig) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	
	m := &Manager{
		client:       client,
		agentRegistry: agentRegistry,
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		tasks:        make(map[string]*Task),
		assignments:  make(map[string]*TaskAssignment),
		logger:       log.New(os.Stdout, "[PROJECT] ", log.LstdFlags|log.Lmsgprefix),
	}
	
	client.SetContext(ctx)
	
	return m
}

// Start starts the project manager integration
func (m *Manager) Start() error {
	m.logger.Printf("Starting project manager integration...")
	
	if !m.config.Enabled {
		m.logger.Println("Project manager integration disabled")
		return nil
	}
	
	// Start webhook server
	const webhookPort = 8082 // Default webhook port
	
	if webhookPort > 0 {
		m.webhookServer = NewWebhookServer(m, webhookPort)
		if err := m.webhookServer.Start(); err != nil {
			m.logger.Printf("Failed to start webhook server: %v", err)
		} else {
			m.logger.Printf("Webhook server started on port %d", webhookPort)
			
			// Register webhook with project manager
			if err := m.client.CreateWebhook(m.ctx, fmt.Sprintf("http://localhost:%d/webhook", webhookPort)); err != nil {
				m.logger.Printf("Failed to register webhook: %v", err)
			}
		}
	}
	
	// Start polling for tasks
	m.wg.Add(1)
	go m.taskPoller()
	
	m.logger.Println("Project manager integration started")
	return nil
}

// Stop stops the project manager integration
func (m *Manager) Stop() error {
	m.logger.Println("Stopping project manager integration...")
	
	m.cancel()
	
	if m.webhookServer != nil {
		m.webhookServer.Stop()
	}
	
	// Wait for background goroutines
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.wg.Wait()
	}()
	
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		m.logger.Println("Timeout waiting for background tasks")
	}
	
	return nil
}

// taskPoller periodically polls for new tasks
func (m *Manager) taskPoller() {
	defer m.wg.Done()
	
	// Convert PollInterval from seconds to duration
	pollInterval := time.Duration(m.config.PollInterval) * time.Second
	if pollInterval == 0 {
		pollInterval = 30 * time.Second // Default 30 seconds
	}
	
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	
	// Initial load
	m.loadTasks()
	
	for {
		select {
		case <-ticker.C:
			m.loadTasks()
		case <-m.ctx.Done():
			return
		}
	}
}

// loadTasks loads tasks from the project manager
func (m *Manager) loadTasks() {
	filters := map[string]interface{}{
		"status": "todo",
	}
	
	// TODO: Add task categories support when available in config
	// if len(m.config.TaskCategories) > 0 {
	// 	filters["categories"] = m.config.TaskCategories
	// }
	
	tasks, err := m.client.GetTasks(m.ctx, filters)
	if err != nil {
		m.logger.Printf("Failed to load tasks: %v", err)
		return
	}
	
	m.taskMutex.Lock()
	defer m.taskMutex.Unlock()
	
	// Update tasks
	for _, task := range tasks {
		m.tasks[task.ID] = &task
		
		// Auto-assign if enabled
		if m.config.AutoAssign && task.Assignee == "" {
			m.autoAssignTask(&task)
		}
	}
	
	m.logger.Printf("Loaded %d tasks", len(m.tasks))
}

// autoAssignTask automatically assigns a task to an appropriate agent
func (m *Manager) autoAssignTask(task *Task) {
	agentID := m.findBestAgent(task)
	if agentID == "" {
		m.logger.Printf("No suitable agent found for task %s", task.ID)
		return
	}
	
	// Assign task
	if err := m.client.AssignTask(m.ctx, task.ID, agentID); err != nil {
		m.logger.Printf("Failed to assign task %s to agent %s: %v", task.ID, agentID, err)
		return
	}
	
	// Track assignment
	assignment := &TaskAssignment{
		TaskID:     task.ID,
		AgentID:    agentID,
		AssignedAt: time.Now(),
		Status:     "assigned",
	}
	
	m.assignments[task.ID] = assignment
	
	// Start task execution
	go m.executeTask(assignment)
	
	m.logger.Printf("Auto-assigned task %s to agent %s", task.ID, agentID)
}

// findBestAgent finds the best agent for a task based on capabilities and load
func (m *Manager) findBestAgent(task *Task) string {
	// Get available agents
	agents := m.agentRegistry.ListAgents()
	
	bestAgent := ""
	bestScore := 0.0
	
	for _, agent := range agents {
		if agent.Status != "active" {
			continue
		}
		
		// Calculate compatibility score
		score := m.calculateAgentScore(task, agent)
		
		if score > bestScore {
			bestScore = score
			bestAgent = agent.ID
		}
	}
	
	return bestAgent
}

// calculateAgentScore calculates how well an agent fits a task
func (m *Manager) calculateAgentScore(task *Task, agent *agents.Agent) float64 {
	score := 0.0
	
	// Check task keywords against agent capabilities
	for _, keyword := range extractKeywords(task.Title + " " + task.Description) {
		for _, capability := range agent.Capabilities {
			if containsString(capability, keyword) {
				score += 1.0
			}
		}
	}
	
	// Factor in agent load (prefer less loaded agents)
	loadFactor := 1.0 - (float64(agent.Load) / 100.0)
	score *= loadFactor
	
	// Apply assignment rules
	// TODO: Add auto assign rules support when available in config
	// for _, rule := range m.config.AutoAssignRules {
	// 	if m.taskMatchesRule(task, rule) {
	// 		if rule.AgentType == string(agent.Type) {
	// 			score += float64(rule.Priority)
	// 		}
	// 	}
	// }
	
	return score
}

// taskMatchesRule checks if a task matches an assignment rule
// TODO: Implement when AssignRule type is available
/*
func (m *Manager) taskMatchesRule(task *Task, rule AssignRule) bool {
	// Placeholder implementation
	return false
}
*/

// executeTask executes a task with the assigned agent
func (m *Manager) executeTask(assignment *TaskAssignment) {
	m.logger.Printf("Starting execution of task %s with agent %s", assignment.TaskID, assignment.AgentID)
	
	// Get task details
	task, err := m.client.GetTask(m.ctx, assignment.TaskID)
	if err != nil {
		m.logger.Printf("Failed to get task %s: %v", assignment.TaskID, err)
		return
	}
	
	// Update task status
	if err := m.client.UpdateTaskStatus(m.ctx, assignment.TaskID, "in_progress"); err != nil {
		m.logger.Printf("Failed to update task status: %v", err)
	}
	
	// Execute with agent
	result := m.executeWithAgent(assignment.AgentID, task)
	
	// Update assignment with result
	assignment.Status = result.Status
	if result.CompletedAt != nil {
		// Update task status in project manager
		if result.Status == "completed" {
			m.client.UpdateTaskStatus(m.ctx, assignment.TaskID, "done")
		} else {
			m.client.UpdateTaskStatus(m.ctx, assignment.TaskID, "blocked")
		}
	}
	
	m.logger.Printf("Task %s execution completed with status: %s", assignment.TaskID, result.Status)
}

// executeWithAgent executes a task using the specified agent
func (m *Manager) executeWithAgent(agentID string, task *Task) *TaskAssignmentResult {
	result := &TaskAssignmentResult{
		AssignmentID: fmt.Sprintf("%s-%s", agentID, task.ID),
		TaskID:       task.ID,
		AgentID:      agentID,
		Status:       "in_progress",
		StartedAt:    time.Now(),
		Result:       make(map[string]interface{}),
	}
	
	// Get the agent
	agent, exists := m.agentRegistry.GetAgent(agentID)
	if !exists {
		result.Status = "failed"
		result.Result["error"] = "Agent not found"
		now := time.Now()
		result.CompletedAt = &now
		return result
	}
	
	// Execute task (this is a simplified version)
	// In a real implementation, this would call the agent's Execute method
	output, err := m.simulateTaskExecution(task, agent)
	
	if err != nil {
		result.Status = "failed"
		result.Result["error"] = err.Error()
	} else {
		result.Status = "completed"
		result.Result["output"] = output
	}
	
	now := time.Now()
	result.CompletedAt = &now
	
	return result
}

// simulateTaskExecution simulates task execution (placeholder)
func (m *Manager) simulateTaskExecution(task *Task, agent *agents.Agent) (string, error) {
	// This is a placeholder implementation
	// In a real implementation, this would call the agent's actual execution method
	
	m.logger.Printf("Simulating execution of task '%s' with agent '%s'", task.Title, agent.Name)
	
	// Simulate some work
	time.Sleep(2 * time.Second)
	
	// Simple simulation based on task type
	if containsString(task.Title, "code") || containsString(task.Title, "develop") {
		return "Generated code successfully", nil
	} else if containsString(task.Title, "test") {
		return "Ran tests and reported results", nil
	} else if containsString(task.Title, "review") {
		return "Reviewed code and provided feedback", nil
	} else {
		return "Task completed successfully", nil
	}
}

// GetTaskStatus returns the status of a task
func (m *Manager) GetTaskStatus(taskID string) (*TaskAssignmentResult, bool) {
	m.taskMutex.RLock()
	defer m.taskMutex.RUnlock()
	
	assignment, exists := m.assignments[taskID]
	if !exists {
		return nil, false
	}
	
	// Convert to result format
	result := &TaskAssignmentResult{
		TaskID:    assignment.TaskID,
		AgentID:   assignment.AgentID,
		Status:    assignment.Status,
		StartedAt: assignment.AssignedAt,
	}
	
	if assignment.Status == "completed" || assignment.Status == "failed" {
		now := time.Now()
		result.CompletedAt = &now
	}
	
	return result, true
}

// GetTasks returns all tracked tasks
func (m *Manager) GetTasks() map[string]*Task {
	m.taskMutex.RLock()
	defer m.taskMutex.RUnlock()
	
	// Return a copy to avoid race conditions
	tasks := make(map[string]*Task)
	for id, task := range m.tasks {
		tasks[id] = task
	}
	
	return tasks
}

// Helper functions
func extractKeywords(text string) []string {
	// Simple keyword extraction (in a real implementation, this would be more sophisticated)
	keywords := []string{}
	words := splitWords(text)
	
	for _, word := range words {
		if len(word) > 3 && !isStopWord(word) {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

func containsString(text, substring string) bool {
	return len(text) >= len(substring) && 
		   (text == substring || 
		    len(text) > len(substring) && 
		    (text[:len(substring)] == substring || 
		     text[len(text)-len(substring):] == substring ||
		     containsSubstring(text, substring)))
}

func containsSubstring(text, substring string) bool {
	for i := 0; i <= len(text)-len(substring); i++ {
		if text[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}

func splitWords(text string) []string {
	// Simple word splitting (in a real implementation, this would be more sophisticated)
	words := []string{}
	current := ""
	
	for _, char := range text {
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		words = append(words, current)
	}
	
	return words
}

func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true, "but": true,
		"not": true, "you": true, "all": true, "can": true, "had": true,
		"her": true, "was": true, "one": true, "our": true, "out": true,
		"day": true, "get": true, "has": true, "him": true, "his": true,
		"how": true, "its": true, "may": true, "new": true, "now": true,
		"old": true, "see": true, "two": true, "way": true, "who": true,
		"boy": true, "did": true, "she": true, "use": true, "with": true,
		"this": true, "that": true, "from": true, "they": true, "have": true,
		"will": true, "would": true, "there": true, "their": true,
	}
	
	return stopWords[word]
}