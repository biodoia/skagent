package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AgentType defines the type of agent
type AgentType string

const (
	AgentTypeCoder      AgentType = "coder"
	AgentTypeReviewer   AgentType = "reviewer"
	AgentTypePlanner    AgentType = "planner"
	AgentTypeDocumenter AgentType = "documenter"
	AgentTypeTester     AgentType = "tester"
	AgentTypeGeneral    AgentType = "general"
)

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	StatusIdle     AgentStatus = "idle"
	StatusWorking  AgentStatus = "working"
	StatusPaused   AgentStatus = "paused"
	StatusError    AgentStatus = "error"
	StatusOffline  AgentStatus = "offline"
)

// Agent represents an AI agent instance
type Agent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         AgentType         `json:"type"`
	Status       AgentStatus       `json:"status"`
	Description  string            `json:"description,omitempty"`
	Labels       []string          `json:"labels,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Load         int               `json:"load,omitempty"` // 0-100
	Config       AgentConfig       `json:"config"`
	Stats        AgentStats        `json:"stats"`
	CurrentTask  *Task             `json:"current_task,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Meta         map[string]string `json:"meta,omitempty"`
	mu           sync.RWMutex
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	Provider       string   `json:"provider,omitempty"`
	Model          string   `json:"model,omitempty"`
	SystemPrompt   string   `json:"system_prompt,omitempty"`
	MaxConcurrent  int      `json:"max_concurrent"`
	Timeout        int      `json:"timeout_seconds"`
	AutoAssign     bool     `json:"auto_assign"`
	PreferredTasks []string `json:"preferred_tasks,omitempty"`
}

// AgentStats tracks agent performance metrics
type AgentStats struct {
	TasksCompleted int       `json:"tasks_completed"`
	TasksFailed    int       `json:"tasks_failed"`
	TotalTime      int64     `json:"total_time_ms"`
	AvgTime        int64     `json:"avg_time_ms"`
	LastActive     time.Time `json:"last_active"`
	SuccessRate    float64   `json:"success_rate"`
}

// Task represents a work item for an agent
type Task struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Priority    TaskPriority      `json:"priority"`
	Status      TaskStatus        `json:"status"`
	AssignedTo  string            `json:"assigned_to,omitempty"`
	Labels      []string          `json:"labels,omitempty"`
	ProjectID   string            `json:"project_id,omitempty"`
	ExternalID  string            `json:"external_id,omitempty"` // ID from project manager
	Source      string            `json:"source,omitempty"`      // linear, github, jira
	Result      *TaskResult       `json:"result,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

// TaskPriority defines task priority levels
type TaskPriority int

const (
	PriorityLow TaskPriority = iota
	PriorityMedium
	PriorityHigh
	PriorityUrgent
)

// TaskStatus represents task state
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusQueued     TaskStatus = "queued"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskResult holds the result of a completed task
type TaskResult struct {
	Success   bool      `json:"success"`
	Output    string    `json:"output,omitempty"`
	Error     string    `json:"error,omitempty"`
	Artifacts []string  `json:"artifacts,omitempty"` // file paths, URLs, etc.
	Duration  int64     `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
}

// Registry manages all agents
type Registry struct {
	agents map[string]*Agent
	tasks  map[string]*Task
	mu     sync.RWMutex
	ctx    context.Context
}

// NewRegistry creates a new agent registry
func NewRegistry(ctx context.Context) *Registry {
	return &Registry{
		agents: make(map[string]*Agent),
		tasks:  make(map[string]*Task),
		ctx:    ctx,
	}
}

// RegisterAgent adds a new agent to the registry
func (r *Registry) RegisterAgent(agent *Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = time.Now()
	agent.Status = StatusIdle
	
	r.agents[agent.ID] = agent
}

// GetAgent returns an agent by ID
func (r *Registry) GetAgent(id string) (*Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, ok := r.agents[id]
	return agent, ok
}

// ListAgents returns all registered agents
func (r *Registry) ListAgents() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	agents := make([]*Agent, 0, len(r.agents))
	for _, a := range r.agents {
		agents = append(agents, a)
	}
	return agents
}

// GetAgentsByType returns agents of a specific type
func (r *Registry) GetAgentsByType(agentType AgentType) []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var agents []*Agent
	for _, a := range r.agents {
		if a.Type == agentType {
			agents = append(agents, a)
		}
	}
	return agents
}

// GetIdleAgents returns all idle agents
func (r *Registry) GetIdleAgents() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var agents []*Agent
	for _, a := range r.agents {
		if a.Status == StatusIdle {
			agents = append(agents, a)
		}
	}
	return agents
}

// CreateTask creates a new task
func (r *Registry) CreateTask(task *Task) *Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.Status = TaskStatusPending
	
	r.tasks[task.ID] = task
	return task
}

// GetTask returns a task by ID
func (r *Registry) GetTask(id string) (*Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.tasks[id]
	return task, ok
}

// ListTasks returns all tasks
func (r *Registry) ListTasks() []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tasks := make([]*Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// GetPendingTasks returns all pending tasks
func (r *Registry) GetPendingTasks() []*Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var tasks []*Task
	for _, t := range r.tasks {
		if t.Status == TaskStatusPending || t.Status == TaskStatusQueued {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

// AssignTask assigns a task to an agent
func (r *Registry) AssignTask(taskID, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	task, ok := r.tasks[taskID]
	if !ok {
		return ErrTaskNotFound
	}
	
	agent, ok := r.agents[agentID]
	if !ok {
		return ErrAgentNotFound
	}
	
	if agent.Status != StatusIdle {
		return ErrAgentBusy
	}
	
	task.AssignedTo = agentID
	task.Status = TaskStatusInProgress
	now := time.Now()
	task.StartedAt = &now
	task.UpdatedAt = now
	
	agent.Status = StatusWorking
	agent.CurrentTask = task
	agent.UpdatedAt = now
	
	return nil
}

// CompleteTask marks a task as completed
func (r *Registry) CompleteTask(taskID string, result *TaskResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	task, ok := r.tasks[taskID]
	if !ok {
		return ErrTaskNotFound
	}
	
	now := time.Now()
	task.Status = TaskStatusCompleted
	task.CompletedAt = &now
	task.UpdatedAt = now
	task.Result = result
	
	// Update agent stats
	if task.AssignedTo != "" {
		if agent, ok := r.agents[task.AssignedTo]; ok {
			agent.Status = StatusIdle
			agent.CurrentTask = nil
			agent.Stats.TasksCompleted++
			agent.Stats.LastActive = now
			if result != nil {
				agent.Stats.TotalTime += result.Duration
				agent.Stats.AvgTime = agent.Stats.TotalTime / int64(agent.Stats.TasksCompleted)
				if agent.Stats.TasksCompleted > 0 {
					agent.Stats.SuccessRate = float64(agent.Stats.TasksCompleted) / 
						float64(agent.Stats.TasksCompleted+agent.Stats.TasksFailed)
				}
			}
			agent.UpdatedAt = now
		}
	}
	
	return nil
}

// AutoAssign finds and assigns idle agents to pending tasks
func (r *Registry) AutoAssign(ctx context.Context) (assigned int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, task := range r.tasks {
		if task.Status != TaskStatusPending {
			continue
		}
		
		// Find matching idle agent
		for _, agent := range r.agents {
			if agent.Status != StatusIdle || !agent.Config.AutoAssign {
				continue
			}
			
			// Check if agent handles this type of task
			if matchesLabels(agent.Labels, task.Labels) {
				now := time.Now()
				task.AssignedTo = agent.ID
				task.Status = TaskStatusQueued
				task.UpdatedAt = now
				
				agent.Status = StatusWorking
				agent.CurrentTask = task
				agent.UpdatedAt = now
				assigned++
				break
			}
		}
	}
	
	return assigned
}

// matchesLabels checks if agent can handle task based on labels
func matchesLabels(agentLabels, taskLabels []string) bool {
	if len(agentLabels) == 0 {
		return true // Agent handles any task
	}
	
	for _, al := range agentLabels {
		for _, tl := range taskLabels {
			if al == tl {
				return true
			}
		}
	}
	return false
}

// DefaultAgents creates the default set of agents
func DefaultAgents() []*Agent {
	return []*Agent{
		{
			Name:        "Coder",
			Type:        AgentTypeCoder,
			Description: "Writes and refactors code",
			Labels:      []string{"code", "implement", "refactor", "fix"},
			Config: AgentConfig{
				AutoAssign:     true,
				MaxConcurrent:  1,
				Timeout:        300,
				PreferredTasks: []string{"implement", "code", "fix"},
			},
		},
		{
			Name:        "Reviewer",
			Type:        AgentTypeReviewer,
			Description: "Reviews code and suggests improvements",
			Labels:      []string{"review", "security", "quality"},
			Config: AgentConfig{
				AutoAssign:     true,
				MaxConcurrent:  2,
				Timeout:        180,
				PreferredTasks: []string{"review", "analyze"},
			},
		},
		{
			Name:        "Planner",
			Type:        AgentTypePlanner,
			Description: "Creates plans and breaks down tasks",
			Labels:      []string{"plan", "design", "architecture"},
			Config: AgentConfig{
				AutoAssign:     true,
				MaxConcurrent:  1,
				Timeout:        120,
				PreferredTasks: []string{"plan", "specify", "design"},
			},
		},
		{
			Name:        "Documenter",
			Type:        AgentTypeDocumenter,
			Description: "Writes documentation and READMEs",
			Labels:      []string{"docs", "readme", "documentation"},
			Config: AgentConfig{
				AutoAssign:     true,
				MaxConcurrent:  2,
				Timeout:        180,
				PreferredTasks: []string{"document", "readme"},
			},
		},
	}
}

// Errors
var (
	ErrAgentNotFound = &AgentError{message: "agent not found"}
	ErrTaskNotFound  = &AgentError{message: "task not found"}
	ErrAgentBusy     = &AgentError{message: "agent is busy"}
)

type AgentError struct {
	message string
}

func (e *AgentError) Error() string {
	return e.message
}

// GetStats returns statistics about the registry
func (r *Registry) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var totalTasks, completedTasks, failedTasks int
	var activeAgents, idleAgents int
	
	for _, agent := range r.agents {
		switch agent.Status {
		case StatusIdle:
			idleAgents++
		case StatusWorking, StatusPaused:
			activeAgents++
		}
	}
	
	for _, task := range r.tasks {
		totalTasks++
		switch task.Status {
		case TaskStatusCompleted:
			completedTasks++
		case TaskStatusFailed:
			failedTasks++
		}
	}
	
	return map[string]interface{}{
		"total_agents":   len(r.agents),
		"active_agents":  activeAgents,
		"idle_agents":    idleAgents,
		"total_tasks":    totalTasks,
		"completed_tasks": completedTasks,
		"failed_tasks":   failedTasks,
	}
}

// StartAgent starts a specific agent
func (r *Registry) StartAgent(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	agent, ok := r.agents[agentID]
	if !ok {
		return ErrAgentNotFound
	}
	
	agent.Status = StatusIdle
	agent.UpdatedAt = time.Now()
	return nil
}

// StopAgent stops a specific agent
func (r *Registry) StopAgent(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	agent, ok := r.agents[agentID]
	if !ok {
		return ErrAgentNotFound
	}
	
	agent.Status = StatusOffline
	agent.UpdatedAt = time.Now()
	return nil
}

// CreateAgent creates a new agent with given parameters
func (r *Registry) CreateAgent(name, agentType string, config map[string]interface{}) (*Agent, error) {
	agent := &Agent{
		Name:        name,
		Type:        AgentType(agentType),
		Status:      StatusIdle,
		Description: fmt.Sprintf("Agent of type %s", agentType),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config: AgentConfig{
			AutoAssign:     true,
			MaxConcurrent:  1,
			Timeout:        300,
		},
	}
	
	// Apply config overrides
	if cfg, ok := config["auto_assign"].(bool); ok {
		agent.Config.AutoAssign = cfg
	}
	
	r.RegisterAgent(agent)
	return agent, nil
}

// DeleteAgent removes an agent from the registry
func (r *Registry) DeleteAgent(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.agents[agentID]; !ok {
		return ErrAgentNotFound
	}
	
	delete(r.agents, agentID)
	return nil
}
