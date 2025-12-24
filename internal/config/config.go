package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Provider represents an AI provider type
type Provider string

const (
	ProviderOpenRouter   Provider = "openrouter"
	ProviderClaudeMax    Provider = "claude_max"
	ProviderGeminiCLI    Provider = "gemini_cli"
	ProviderCodex        Provider = "codex"
	ProviderMinimax      Provider = "minimax"
	ProviderKimi         Provider = "kimi"
	ProviderGLM          Provider = "glm"
	ProviderDeepSeek     Provider = "deepseek"
	ProviderLocal        Provider = "local"
)

// FreeModel represents a free model available on OpenRouter
type FreeModel struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ContextLength int    `json:"context_length"`
	Provider      string `json:"provider"`
	Description   string `json:"description"`
	Recommended   bool   `json:"recommended"`
}

// OpenRouterFreeModels contains all available free models
var OpenRouterFreeModels = []FreeModel{
	// Coding-focused models (recommended)
	{ID: "qwen/qwen3-coder:free", Name: "Qwen3 Coder 480B", ContextLength: 262000, Provider: "Qwen", Description: "Best for coding tasks", Recommended: true},
	{ID: "mistralai/devstral-2512:free", Name: "Mistral Devstral", ContextLength: 262144, Provider: "Mistral", Description: "Developer-focused model", Recommended: true},
	{ID: "kwaipilot/kat-coder-pro:free", Name: "KAT-Coder Pro", ContextLength: 256000, Provider: "Kwaipilot", Description: "Professional coding assistant", Recommended: true},
	{ID: "deepseek/deepseek-r1-0528:free", Name: "DeepSeek R1", ContextLength: 163840, Provider: "DeepSeek", Description: "Reasoning model for complex tasks", Recommended: true},

	// Large context models
	{ID: "google/gemini-2.0-flash-exp:free", Name: "Gemini 2.0 Flash", ContextLength: 1048576, Provider: "Google", Description: "1M context, fast responses", Recommended: true},
	{ID: "xiaomi/mimo-v2-flash:free", Name: "MiMo V2 Flash", ContextLength: 262144, Provider: "Xiaomi", Description: "Large context, fast"},
	{ID: "nvidia/nemotron-3-nano-30b-a3b:free", Name: "Nemotron 3 Nano 30B", ContextLength: 256000, Provider: "NVIDIA", Description: "NVIDIA's efficient model"},

	// General purpose models
	{ID: "meta-llama/llama-3.3-70b-instruct:free", Name: "Llama 3.3 70B", ContextLength: 131072, Provider: "Meta", Description: "Best open-source general model", Recommended: true},
	{ID: "meta-llama/llama-3.1-405b-instruct:free", Name: "Llama 3.1 405B", ContextLength: 131072, Provider: "Meta", Description: "Largest open model"},
	{ID: "nousresearch/hermes-3-llama-3.1-405b:free", Name: "Hermes 3 405B", ContextLength: 131072, Provider: "Nous", Description: "Fine-tuned for instructions"},
	{ID: "z-ai/glm-4.5-air:free", Name: "GLM 4.5 Air", ContextLength: 131072, Provider: "Z.AI", Description: "Chinese-optimized model"},
	{ID: "moonshotai/kimi-k2:free", Name: "Kimi K2", ContextLength: 32768, Provider: "MoonshotAI", Description: "Moonshot's latest model"},

	// Thinking/reasoning models
	{ID: "allenai/olmo-3.1-32b-think:free", Name: "Olmo 3.1 32B Think", ContextLength: 65536, Provider: "AllenAI", Description: "Thinking-focused model"},
	{ID: "tngtech/tng-r1t-chimera:free", Name: "TNG R1T Chimera", ContextLength: 163840, Provider: "TNG", Description: "Reasoning chimera"},
	{ID: "alibaba/tongyi-deepresearch-30b-a3b:free", Name: "Tongyi DeepResearch", ContextLength: 131072, Provider: "Alibaba", Description: "Research-focused"},

	// Smaller/faster models
	{ID: "mistralai/mistral-small-3.1-24b-instruct:free", Name: "Mistral Small 3.1", ContextLength: 128000, Provider: "Mistral", Description: "Fast and capable"},
	{ID: "google/gemma-3-27b-it:free", Name: "Gemma 3 27B", ContextLength: 131072, Provider: "Google", Description: "Google's efficient model"},
	{ID: "qwen/qwen3-4b:free", Name: "Qwen3 4B", ContextLength: 40960, Provider: "Qwen", Description: "Small and fast"},
	{ID: "meta-llama/llama-3.2-3b-instruct:free", Name: "Llama 3.2 3B", ContextLength: 131072, Provider: "Meta", Description: "Tiny but capable"},
	{ID: "mistralai/mistral-7b-instruct:free", Name: "Mistral 7B", ContextLength: 32768, Provider: "Mistral", Description: "Classic efficient model"},

	// Uncensored/creative
	{ID: "cognitivecomputations/dolphin-mistral-24b-venice-edition:free", Name: "Venice Uncensored", ContextLength: 32768, Provider: "CogComp", Description: "Uncensored model"},
}

// ProviderConfig holds configuration for a specific provider
type ProviderConfig struct {
	Enabled   bool              `json:"enabled"`
	APIKey    string            `json:"api_key,omitempty"`
	BaseURL   string            `json:"base_url,omitempty"`
	Model     string            `json:"model,omitempty"`
	AuthType  string            `json:"auth_type,omitempty"` // "api_key", "oauth", "cli"
	ExtraArgs map[string]string `json:"extra_args,omitempty"`
}

// Config holds the complete application configuration
type Config struct {
	Version         string                    `json:"version"`
	DefaultProvider Provider                  `json:"default_provider"`
	Providers       map[Provider]ProviderConfig `json:"providers"`
	SpecKitPath     string                    `json:"speckit_path,omitempty"`
	GitHubUser      string                    `json:"github_user,omitempty"`
	Autonomous      bool                      `json:"autonomous_default"`
	Theme           string                    `json:"theme"`
}

// DefaultConfig returns a new configuration with defaults
func DefaultConfig() *Config {
	return &Config{
		Version:         "1.0.0",
		DefaultProvider: ProviderOpenRouter,
		Providers: map[Provider]ProviderConfig{
			ProviderOpenRouter: {
				Enabled:  true,
				BaseURL:  "https://openrouter.ai/api/v1",
				Model:    "qwen/qwen3-coder:free", // Best free coding model
				AuthType: "api_key",
			},
			ProviderClaudeMax: {
				Enabled:  false,
				AuthType: "oauth",
			},
			ProviderGeminiCLI: {
				Enabled:  false,
				AuthType: "cli",
			},
			ProviderCodex: {
				Enabled:  false,
				AuthType: "cli",
			},
		},
		Theme: "catppuccin",
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "skagent", "config.json"), nil
}

// Load loads configuration from disk
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No config yet
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves configuration to disk
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// Exists checks if a config file exists
func Exists() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// GetActiveProvider returns the configuration for the default provider
func (c *Config) GetActiveProvider() ProviderConfig {
	if cfg, ok := c.Providers[c.DefaultProvider]; ok {
		return cfg
	}
	return ProviderConfig{}
}
