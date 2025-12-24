package project

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Task represents a task from the project manager
type Task struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"` // low, medium, high, critical
	Status      string                 `json:"status"`   // todo, in_progress, done, blocked
	Assignee    string                 `json:"assignee"`
	Labels      []string               `json:"labels"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
}

// TaskAssignment represents the assignment of a task to an agent
type TaskAssignment struct {
	TaskID   string `json:"task_id"`
	AgentID  string `json:"agent_id"`
	AssignedBy string `json:"assigned_by"`
	AssignedAt time.Time `json:"assigned_at"`
	Status   string `json:"status"` // assigned, in_progress, completed, failed
	Result   string `json:"result,omitempty"`
}

// AgentInfo represents an available agent
type AgentInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Capabilities []string         `json:"capabilities"`
	Status      string            `json:"status"` // active, busy, offline
	Load        int               `json:"load"`   // 0-100
	Metadata    map[string]interface{} `json:"metadata"`
}

// Client represents a project manager client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	ctx        context.Context
}

// NewClient creates a new project manager client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetContext sets the context for the client
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// GetTasks retrieves tasks from the project manager
func (c *Client) GetTasks(ctx context.Context, filters map[string]interface{}) ([]Task, error) {
	req, err := c.newRequest(ctx, "GET", "/api/v1/tasks", nil)
	if err != nil {
		return nil, err
	}

	// Add query parameters for filtering
	q := req.URL.Query()
	for key, value := range filters {
		if strValue, ok := value.(string); ok {
			q.Set(key, strValue)
		}
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tasks: %s", resp.Status)
	}

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTask retrieves a specific task by ID
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	req, err := c.newRequest(ctx, "GET", fmt.Sprintf("/api/v1/tasks/%s", taskID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get task %s: %s", taskID, resp.Status)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, err
	}

	return &task, nil
}

// UpdateTaskStatus updates the status of a task
func (c *Client) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	update := map[string]string{
		"status": status,
	}

	req, err := c.newRequest(ctx, "PATCH", fmt.Sprintf("/api/v1/tasks/%s", taskID), update)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update task %s: %s", taskID, resp.Status)
	}

	return nil
}

// AssignTask assigns a task to an agent
func (c *Client) AssignTask(ctx context.Context, taskID, agentID string) error {
	assignment := TaskAssignment{
		TaskID:     taskID,
		AgentID:    agentID,
		AssignedAt: time.Now(),
		Status:     "assigned",
	}

	req, err := c.newRequest(ctx, "POST", "/api/v1/task-assignments", assignment)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to assign task %s: %s", taskID, resp.Status)
	}

	return nil
}

// GetAgents retrieves available agents from the project manager
func (c *Client) GetAgents(ctx context.Context) ([]AgentInfo, error) {
	req, err := c.newRequest(ctx, "GET", "/api/v1/agents", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get agents: %s", resp.Status)
	}

	var agents []AgentInfo
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil, err
	}

	return agents, nil
}

// CreateWebhook creates a webhook for real-time task updates
func (c *Client) CreateWebhook(ctx context.Context, callbackURL string) error {
	webhook := map[string]string{
		"url": callbackURL,
		"events": "task.created,task.updated,task.assigned",
	}

	req, err := c.newRequest(ctx, "POST", "/api/v1/webhooks", webhook)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create webhook: %s", resp.Status)
	}

	return nil
}

// newRequest creates a new HTTP request with proper headers
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var url string
	if c.baseURL[len(c.baseURL)-1] == '/' {
		url = c.baseURL[:len(c.baseURL)-1] + path
	} else {
		url = c.baseURL + path
	}

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return req, nil
}