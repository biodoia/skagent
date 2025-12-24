package tools

import (
	"context"
	"fmt"
)

// Tool interface for all tool implementations
type Tool interface {
	// Name returns the tool's identifier
	Name() string
	// Description returns a human-readable description
	Description() string
	// Execute runs the tool with the given input
	Execute(ctx context.Context, input string) (string, error)
	// CanHandle returns true if this tool can handle the given intent
	CanHandle(intent string) bool
}

// ToolManager manages a collection of tools
type ToolManager struct {
	tools []Tool
}

// NewToolManager creates a new tool manager
func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: []Tool{},
	}
}

// AddTool registers a new tool
func (tm *ToolManager) AddTool(tool Tool) {
	tm.tools = append(tm.tools, tool)
}

// GetTool returns a tool by name
func (tm *ToolManager) GetTool(name string) Tool {
	for _, tool := range tm.tools {
		if tool.Name() == name {
			return tool
		}
	}
	return nil
}

// ListTools returns all registered tools
func (tm *ToolManager) ListTools() []Tool {
	return tm.tools
}

// CanHandle checks if any tool can handle the given intent
func (tm *ToolManager) CanHandle(intent string) bool {
	for _, tool := range tm.tools {
		if tool.CanHandle(intent) {
			return true
		}
	}
	return false
}

// FindTool returns the first tool that can handle the intent
func (tm *ToolManager) FindTool(intent string) Tool {
	for _, tool := range tm.tools {
		if tool.CanHandle(intent) {
			return tool
		}
	}
	return nil
}

// Execute finds and runs the appropriate tool for the intent
func (tm *ToolManager) Execute(ctx context.Context, intent string, input string) (string, error) {
	tool := tm.FindTool(intent)
	if tool == nil {
		return "", fmt.Errorf("no tool can handle intent: %s", intent)
	}
	return tool.Execute(ctx, input)
}

// ExecuteByName runs a specific tool by name
func (tm *ToolManager) ExecuteByName(ctx context.Context, name string, input string) (string, error) {
	tool := tm.GetTool(name)
	if tool == nil {
		return "", fmt.Errorf("tool not found: %s", name)
	}
	return tool.Execute(ctx, input)
}

// GetToolDescriptions returns a map of tool names to descriptions
func (tm *ToolManager) GetToolDescriptions() map[string]string {
	descriptions := make(map[string]string)
	for _, tool := range tm.tools {
		descriptions[tool.Name()] = tool.Description()
	}
	return descriptions
}
