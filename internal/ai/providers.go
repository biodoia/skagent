package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/biodoia/skagent/internal/config"
)

// Provider interface for different AI backends
type Provider interface {
	Complete(ctx context.Context, messages []Message, systemPrompt string) (string, error)
	Name() string
}

// OpenRouterProvider uses OpenRouter's API for free models
type OpenRouterProvider struct {
	apiKey  string
	model   string
	baseURL string
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(cfg config.ProviderConfig) *OpenRouterProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	return &OpenRouterProvider{
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		baseURL: baseURL,
	}
}

func (p *OpenRouterProvider) Name() string { return "OpenRouter" }

func (p *OpenRouterProvider) Complete(ctx context.Context, messages []Message, systemPrompt string) (string, error) {
	// Build request body
	var reqMessages []map[string]string

	// Add system prompt
	if systemPrompt != "" {
		reqMessages = append(reqMessages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	// Add conversation messages
	for _, msg := range messages {
		reqMessages = append(reqMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": reqMessages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/biodoia/skagent")
	req.Header.Set("X-Title", "SkAgent")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return result.Choices[0].Message.Content, nil
}

// GenericOpenAIProvider works with OpenAI-compatible APIs (DeepSeek, Kimi, GLM, etc.)
type GenericOpenAIProvider struct {
	name    string
	apiKey  string
	model   string
	baseURL string
}

// NewGenericOpenAIProvider creates a provider for OpenAI-compatible APIs
func NewGenericOpenAIProvider(name string, cfg config.ProviderConfig, defaultModel string) *GenericOpenAIProvider {
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}
	return &GenericOpenAIProvider{
		name:    name,
		apiKey:  cfg.APIKey,
		model:   model,
		baseURL: cfg.BaseURL,
	}
}

func (p *GenericOpenAIProvider) Name() string { return p.name }

func (p *GenericOpenAIProvider) Complete(ctx context.Context, messages []Message, systemPrompt string) (string, error) {
	var reqMessages []map[string]string

	if systemPrompt != "" {
		reqMessages = append(reqMessages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	for _, msg := range messages {
		reqMessages = append(reqMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": reqMessages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return result.Choices[0].Message.Content, nil
}

// CLIProvider uses CLI tools like gemini, codex
type CLIProvider struct {
	name    string
	command string
	args    []string
}

// NewGeminiCLIProvider creates a provider using Gemini CLI
func NewGeminiCLIProvider() *CLIProvider {
	return &CLIProvider{
		name:    "Gemini CLI",
		command: "gemini",
		args:    []string{"chat"},
	}
}

// NewCodexCLIProvider creates a provider using Codex CLI
func NewCodexCLIProvider() *CLIProvider {
	return &CLIProvider{
		name:    "Codex CLI",
		command: "codex",
		args:    []string{},
	}
}

func (p *CLIProvider) Name() string { return p.name }

func (p *CLIProvider) Complete(ctx context.Context, messages []Message, systemPrompt string) (string, error) {
	// Build prompt from messages
	var prompt strings.Builder

	if systemPrompt != "" {
		prompt.WriteString("System: ")
		prompt.WriteString(systemPrompt)
		prompt.WriteString("\n\n")
	}

	for _, msg := range messages {
		prompt.WriteString(msg.Role)
		prompt.WriteString(": ")
		prompt.WriteString(msg.Content)
		prompt.WriteString("\n\n")
	}

	// Run CLI command
	args := append(p.args, prompt.String())
	cmd := exec.CommandContext(ctx, p.command, args...)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("CLI error: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ClaudeMaxProvider uses Claude Code's OAuth authentication
type ClaudeMaxProvider struct {
	// Uses the existing Claude Code authentication
}

// NewClaudeMaxProvider creates a provider using Claude Max subscription
func NewClaudeMaxProvider() *ClaudeMaxProvider {
	return &ClaudeMaxProvider{}
}

func (p *ClaudeMaxProvider) Name() string { return "Claude Max" }

func (p *ClaudeMaxProvider) Complete(ctx context.Context, messages []Message, systemPrompt string) (string, error) {
	// Build prompt
	var prompt strings.Builder

	if systemPrompt != "" {
		prompt.WriteString(systemPrompt)
		prompt.WriteString("\n\n")
	}

	for _, msg := range messages {
		prompt.WriteString(msg.Content)
		prompt.WriteString("\n")
	}

	// Use claude CLI with the prompt
	cmd := exec.CommandContext(ctx, "claude", "-p", prompt.String())

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Claude CLI error: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CreateProvider creates the appropriate provider based on configuration
func CreateProvider(cfg *config.Config) (Provider, error) {
	providerCfg := cfg.GetActiveProvider()

	switch cfg.DefaultProvider {
	case config.ProviderOpenRouter:
		if providerCfg.APIKey == "" {
			return nil, fmt.Errorf("OpenRouter API key not configured")
		}
		return NewOpenRouterProvider(providerCfg), nil

	case config.ProviderClaudeMax:
		return NewClaudeMaxProvider(), nil

	case config.ProviderGeminiCLI:
		return NewGeminiCLIProvider(), nil

	case config.ProviderCodex:
		return NewCodexCLIProvider(), nil

	case config.ProviderKimi:
		if providerCfg.APIKey == "" {
			return nil, fmt.Errorf("Kimi API key not configured")
		}
		return NewGenericOpenAIProvider("Kimi", providerCfg, "moonshot-v1-8k"), nil

	case config.ProviderGLM:
		if providerCfg.APIKey == "" {
			return nil, fmt.Errorf("GLM API key not configured")
		}
		return NewGenericOpenAIProvider("GLM", providerCfg, "glm-4"), nil

	case config.ProviderDeepSeek:
		if providerCfg.APIKey == "" {
			return nil, fmt.Errorf("DeepSeek API key not configured")
		}
		return NewGenericOpenAIProvider("DeepSeek", providerCfg, "deepseek-chat"), nil

	case config.ProviderMinimax:
		if providerCfg.APIKey == "" {
			return nil, fmt.Errorf("Minimax API key not configured")
		}
		return NewGenericOpenAIProvider("Minimax", providerCfg, "abab5.5-chat"), nil

	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.DefaultProvider)
	}
}
