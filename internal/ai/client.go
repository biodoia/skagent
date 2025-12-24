package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// SpecKitDocs contains the embedded documentation for SpecKit
const SpecKitDocs = `
# GitHub Spec-Kit Commands

## Core Commands
- /speckit.constitution: Establish governance principles
- /speckit.specify: Define requirements and specifications
- /speckit.plan: Create technical implementation plan
- /speckit.tasks: Generate actionable task list
- /speckit.implement: Execute implementation

## Optional Commands
- /speckit.clarify: Refine under-specified areas
- /speckit.analyze: Validate consistency
- /speckit.checklist: Verify requirements quality

## Workflow
1. SPECIFY -> Define what and why
2. PLAN -> Technical blueprint
3. TASKS -> Atomic work items
4. IMPLEMENT -> Build with TDD

## The Nine Articles (Constitution)
1. Library-First: Features as standalone libraries
2. CLI Mandate: All functionality via CLI
3. Test-First: No implementation without tests
4. Simplicity: Max 3 projects per implementation
5. Anti-Abstraction: Use frameworks directly
6. Integration-First: Test in realistic environments
`

// SystemPrompt is the core system prompt for the AI agent
const SystemPrompt = `You are an expert AI agent for spec-driven development using GitHub Spec-Kit.

You are PROACTIVE: You propose solutions, search for alternatives, and suggest improvements.
You can work in AUTONOMOUS mode: When activated, proceed without confirmations.

Your capabilities:
1. SpecKit: Initialize projects, create specifications, plans, and tasks
2. GitHub: Create repositories, issues, projects via gh CLI
3. Web Search: Find best practices and references
4. Plandex: Delegate detailed planning when needed

When given a project idea, you:
1. Analyze the idea and identify key requirements
2. Search for similar projects and best practices
3. Generate comprehensive specifications using SpecKit
4. Create a technical plan with architecture decisions
5. Break down into atomic, testable tasks
6. Optionally create a GitHub repository

%s

Always follow the spec-driven workflow: SPECIFY -> PLAN -> TASKS -> IMPLEMENT
`

// Config holds the AI client configuration
type Config struct {
	APIKey      string
	Model       anthropic.Model
	MaxTokens   int64
	Temperature float64
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		APIKey:      os.Getenv("ANTHROPIC_API_KEY"),
		Model:       anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}

// Client wraps the Anthropic API client
type Client struct {
	client        anthropic.Client
	config        Config
	history       []anthropic.MessageParam
	simpleHistory []Message // Keep a simple copy for GetHistory
}

// Message represents a conversation message for external use
type Message struct {
	Role    string
	Content string
}

// NewClient creates a new AI client with default configuration
func NewClient() *Client {
	return NewClientWithConfig(DefaultConfig())
}

// NewClientWithConfig creates a new AI client with custom configuration
func NewClientWithConfig(config Config) *Client {
	var opts []option.RequestOption
	if config.APIKey != "" {
		opts = append(opts, option.WithAPIKey(config.APIKey))
	}

	return &Client{
		client:        anthropic.NewClient(opts...),
		config:        config,
		history:       []anthropic.MessageParam{},
		simpleHistory: []Message{},
	}
}

// Complete sends a message and returns the AI response
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	// Add user message to history
	c.history = append(c.history, anthropic.NewUserMessage(
		anthropic.NewTextBlock(prompt),
	))
	c.simpleHistory = append(c.simpleHistory, Message{Role: "user", Content: prompt})

	// Create the message request
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.config.Model,
		MaxTokens: c.config.MaxTokens,
		System: []anthropic.TextBlockParam{
			{Text: fmt.Sprintf(SystemPrompt, SpecKitDocs)},
		},
		Messages: c.history,
	})

	if err != nil {
		return "", fmt.Errorf("failed to complete: %w", err)
	}

	// Extract text response
	var response string
	for _, block := range message.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			response += textBlock.Text
		}
	}

	// Add assistant response to history
	c.history = append(c.history, message.ToParam())
	c.simpleHistory = append(c.simpleHistory, Message{Role: "assistant", Content: response})

	return response, nil
}

// CompleteWithTools sends a message and can use tools
func (c *Client) CompleteWithTools(ctx context.Context, prompt string, tools []Tool) (string, []ToolCall, error) {
	// Add user message to history
	c.history = append(c.history, anthropic.NewUserMessage(
		anthropic.NewTextBlock(prompt),
	))
	c.simpleHistory = append(c.simpleHistory, Message{Role: "user", Content: prompt})

	// Build tools for API
	apiTools := c.buildTools(tools)

	// Create the message request with tools
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.config.Model,
		MaxTokens: c.config.MaxTokens,
		System: []anthropic.TextBlockParam{
			{Text: fmt.Sprintf(SystemPrompt, SpecKitDocs)},
		},
		Messages: c.history,
		Tools:    apiTools,
	})

	if err != nil {
		return "", nil, fmt.Errorf("failed to complete with tools: %w", err)
	}

	// Extract response and tool calls
	var response string
	var toolCalls []ToolCall

	for _, block := range message.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			response += variant.Text
		case anthropic.ToolUseBlock:
			toolCalls = append(toolCalls, ToolCall{
				ID:    variant.ID,
				Name:  variant.Name,
				Input: fmt.Sprintf("%v", variant.Input),
			})
		}
	}

	// Add assistant response to history
	c.history = append(c.history, message.ToParam())
	c.simpleHistory = append(c.simpleHistory, Message{Role: "assistant", Content: response})

	return response, toolCalls, nil
}

// ClearHistory clears the conversation history
func (c *Client) ClearHistory() {
	c.history = []anthropic.MessageParam{}
	c.simpleHistory = []Message{}
}

// GetHistory returns the conversation history as simple messages
func (c *Client) GetHistory() []Message {
	return c.simpleHistory
}

// Tool represents a tool that the AI can use
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// ToolCall represents a tool call made by the AI
type ToolCall struct {
	ID    string
	Name  string
	Input string
}

// buildTools converts internal tools to API format
func (c *Client) buildTools(tools []Tool) []anthropic.ToolUnionParam {
	var apiTools []anthropic.ToolUnionParam

	for _, tool := range tools {
		apiTools = append(apiTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: tool.InputSchema,
				},
			},
		})
	}

	return apiTools
}
