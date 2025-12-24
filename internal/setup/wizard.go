package setup

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sergio/skagent/internal/config"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")).
			Bold(true).
			PaddingLeft(2)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			Italic(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))
)

// Step represents a setup step
type Step int

const (
	StepWelcome Step = iota
	StepSelectProvider
	StepConfigureProvider
	StepSelectModel
	StepTestConnection
	StepComplete
)

// ProviderItem represents a selectable provider
type ProviderItem struct {
	provider    config.Provider
	name        string
	description string
	authType    string
	available   bool
}

func (i ProviderItem) FilterValue() string { return i.name }
func (i ProviderItem) Title() string       { return i.name }
func (i ProviderItem) Description() string { return i.description }

// ModelItem represents a selectable model
type ModelItem struct {
	model       config.FreeModel
}

func (i ModelItem) FilterValue() string { return i.model.Name }
func (i ModelItem) Title() string       { return i.model.Name }
func (i ModelItem) Description() string {
	return fmt.Sprintf("%s | %dk context | %s", i.model.Provider, i.model.ContextLength/1000, i.model.Description)
}

// Model is the setup wizard model
type Model struct {
	step            Step
	config          *config.Config
	selectedProvider config.Provider
	providerList    list.Model
	modelList       list.Model
	textInput       textinput.Model
	inputLabel      string
	err             error
	width           int
	height          int
	testResult      string
	testing         bool
}

// NewWizard creates a new setup wizard
func NewWizard() Model {
	// Create provider list
	providers := []list.Item{
		ProviderItem{
			provider:    config.ProviderOpenRouter,
			name:        "🌐 OpenRouter (Free Models)",
			description: "35+ free models including Qwen Coder, DeepSeek R1, Llama 3.3 70B",
			authType:    "api_key",
			available:   true,
		},
		ProviderItem{
			provider:    config.ProviderClaudeMax,
			name:        "🔮 Claude Max (OAuth)",
			description: "Use your Claude Max subscription - no API costs",
			authType:    "oauth",
			available:   checkClaudeMaxAvailable(),
		},
		ProviderItem{
			provider:    config.ProviderGeminiCLI,
			name:        "🔷 Gemini CLI",
			description: "Use Google's Gemini via CLI - free tier available",
			authType:    "cli",
			available:   checkGeminiCLIAvailable(),
		},
		ProviderItem{
			provider:    config.ProviderCodex,
			name:        "🟢 OpenAI Codex CLI",
			description: "Use OpenAI Codex via CLI",
			authType:    "cli",
			available:   checkCodexAvailable(),
		},
		ProviderItem{
			provider:    config.ProviderKimi,
			name:        "🌙 Kimi (Moonshot)",
			description: "Moonshot AI's Kimi model - free tier available",
			authType:    "api_key",
			available:   true,
		},
		ProviderItem{
			provider:    config.ProviderGLM,
			name:        "🇨🇳 GLM-4 (Zhipu)",
			description: "Zhipu's GLM-4 - Chinese-optimized with free tier",
			authType:    "api_key",
			available:   true,
		},
		ProviderItem{
			provider:    config.ProviderDeepSeek,
			name:        "🔍 DeepSeek",
			description: "DeepSeek models - very affordable",
			authType:    "api_key",
			available:   true,
		},
		ProviderItem{
			provider:    config.ProviderMinimax,
			name:        "📦 Minimax",
			description: "Minimax models with coding focus",
			authType:    "api_key",
			available:   true,
		},
	}

	providerDelegate := list.NewDefaultDelegate()
	providerList := list.New(providers, providerDelegate, 60, 15)
	providerList.Title = "Select AI Provider"
	providerList.SetShowHelp(false)

	// Create model list with free models
	var modelItems []list.Item
	for _, m := range config.OpenRouterFreeModels {
		modelItems = append(modelItems, ModelItem{model: m})
	}

	modelDelegate := list.NewDefaultDelegate()
	modelList := list.New(modelItems, modelDelegate, 60, 15)
	modelList.Title = "Select Model"
	modelList.SetShowHelp(false)

	// Text input for API keys
	ti := textinput.New()
	ti.Placeholder = "Enter your API key..."
	ti.CharLimit = 200
	ti.Width = 50

	return Model{
		step:         StepWelcome,
		config:       config.DefaultConfig(),
		providerList: providerList,
		modelList:    modelList,
		textInput:    ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "esc":
			if m.step > StepWelcome {
				m.step--
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.providerList.SetSize(msg.Width-4, msg.Height-10)
		m.modelList.SetSize(msg.Width-4, msg.Height-10)
	}

	// Update appropriate component based on step
	var cmd tea.Cmd
	switch m.step {
	case StepSelectProvider:
		m.providerList, cmd = m.providerList.Update(msg)
	case StepConfigureProvider:
		m.textInput, cmd = m.textInput.Update(msg)
	case StepSelectModel:
		m.modelList, cmd = m.modelList.Update(msg)
	}

	return m, cmd
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case StepWelcome:
		m.step = StepSelectProvider
		return m, nil

	case StepSelectProvider:
		if item, ok := m.providerList.SelectedItem().(ProviderItem); ok {
			m.selectedProvider = item.provider

			// Configure provider
			providerCfg := config.ProviderConfig{
				Enabled:  true,
				AuthType: item.authType,
			}

			switch item.provider {
			case config.ProviderOpenRouter:
				providerCfg.BaseURL = "https://openrouter.ai/api/v1"
				m.inputLabel = "OpenRouter API Key (get free at openrouter.ai):"
				m.textInput.Placeholder = "sk-or-v1-..."
				m.textInput.Focus()
				m.step = StepConfigureProvider

			case config.ProviderClaudeMax:
				// OAuth login
				m.step = StepConfigureProvider
				m.inputLabel = "Starting Claude Max OAuth login..."
				return m, m.startClaudeOAuth()

			case config.ProviderGeminiCLI:
				// Check if already logged in
				if checkGeminiCLIAvailable() {
					m.step = StepTestConnection
				} else {
					m.inputLabel = "Run 'gemini auth login' first, then press Enter"
					m.step = StepConfigureProvider
				}

			case config.ProviderKimi:
				providerCfg.BaseURL = "https://api.moonshot.cn/v1"
				m.inputLabel = "Kimi API Key (get at platform.moonshot.cn):"
				m.textInput.Placeholder = "sk-..."
				m.textInput.Focus()
				m.step = StepConfigureProvider

			case config.ProviderGLM:
				providerCfg.BaseURL = "https://open.bigmodel.cn/api/paas/v4"
				m.inputLabel = "GLM API Key (get at open.bigmodel.cn):"
				m.textInput.Placeholder = "..."
				m.textInput.Focus()
				m.step = StepConfigureProvider

			case config.ProviderDeepSeek:
				providerCfg.BaseURL = "https://api.deepseek.com/v1"
				m.inputLabel = "DeepSeek API Key (get at platform.deepseek.com):"
				m.textInput.Placeholder = "sk-..."
				m.textInput.Focus()
				m.step = StepConfigureProvider

			case config.ProviderMinimax:
				providerCfg.BaseURL = "https://api.minimax.chat/v1"
				m.inputLabel = "Minimax API Key:"
				m.textInput.Placeholder = "..."
				m.textInput.Focus()
				m.step = StepConfigureProvider
			}

			m.config.Providers[item.provider] = providerCfg
			m.config.DefaultProvider = item.provider
		}
		return m, nil

	case StepConfigureProvider:
		// Save API key
		if m.textInput.Value() != "" {
			cfg := m.config.Providers[m.selectedProvider]
			cfg.APIKey = m.textInput.Value()
			m.config.Providers[m.selectedProvider] = cfg
		}

		// If OpenRouter, go to model selection
		if m.selectedProvider == config.ProviderOpenRouter {
			m.step = StepSelectModel
		} else {
			m.step = StepTestConnection
		}
		return m, nil

	case StepSelectModel:
		if item, ok := m.modelList.SelectedItem().(ModelItem); ok {
			cfg := m.config.Providers[m.selectedProvider]
			cfg.Model = item.model.ID
			m.config.Providers[m.selectedProvider] = cfg
		}
		m.step = StepTestConnection
		return m, nil

	case StepTestConnection:
		// Save and complete
		if err := m.config.Save(); err != nil {
			m.err = err
			return m, nil
		}
		m.step = StepComplete
		return m, nil

	case StepComplete:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	// Header
	s.WriteString(titleStyle.Render("🚀 SkAgent Setup Wizard"))
	s.WriteString("\n\n")

	switch m.step {
	case StepWelcome:
		s.WriteString(subtitleStyle.Render("Welcome to SkAgent!"))
		s.WriteString("\n\n")
		s.WriteString("SkAgent is an AI-powered spec-driven development assistant.\n")
		s.WriteString("Let's configure your AI provider.\n\n")
		s.WriteString("Available options:\n")
		s.WriteString(itemStyle.Render("• OpenRouter - 35+ FREE models (no cost!)\n"))
		s.WriteString(itemStyle.Render("• Claude Max - Use your subscription\n"))
		s.WriteString(itemStyle.Render("• Gemini/Codex CLI - Free tiers available\n"))
		s.WriteString(itemStyle.Render("• Kimi, GLM, DeepSeek, Minimax - Free/cheap options\n"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter to continue, Ctrl+C to quit"))

	case StepSelectProvider:
		s.WriteString(subtitleStyle.Render("Step 1: Select AI Provider"))
		s.WriteString("\n")
		s.WriteString(m.providerList.View())
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("↑/↓ to navigate, Enter to select, Esc to go back"))

	case StepConfigureProvider:
		s.WriteString(subtitleStyle.Render("Step 2: Configure Provider"))
		s.WriteString("\n\n")
		s.WriteString(m.inputLabel)
		s.WriteString("\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Enter to continue, Esc to go back"))

	case StepSelectModel:
		s.WriteString(subtitleStyle.Render("Step 3: Select Model"))
		s.WriteString("\n")
		s.WriteString(descStyle.Render("Recommended models are marked with ⭐"))
		s.WriteString("\n")
		s.WriteString(m.modelList.View())
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("↑/↓ to navigate, Enter to select, / to filter"))

	case StepTestConnection:
		s.WriteString(subtitleStyle.Render("Step 4: Testing Connection"))
		s.WriteString("\n\n")
		if m.testing {
			s.WriteString("Testing connection...\n")
		} else if m.err != nil {
			s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
			s.WriteString("\n")
		} else {
			s.WriteString(successStyle.Render("✓ Configuration saved!"))
			s.WriteString("\n\n")
			s.WriteString(fmt.Sprintf("Provider: %s\n", m.selectedProvider))
			cfg := m.config.Providers[m.selectedProvider]
			if cfg.Model != "" {
				s.WriteString(fmt.Sprintf("Model: %s\n", cfg.Model))
			}
		}
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Press Enter to finish"))

	case StepComplete:
		s.WriteString(successStyle.Render("🎉 Setup Complete!"))
		s.WriteString("\n\n")
		s.WriteString("Your configuration has been saved.\n\n")
		s.WriteString("You can now run skagent to start using the assistant.\n")
		s.WriteString("Use /help in the app to see available commands.\n\n")
		s.WriteString(helpStyle.Render("Press Enter or Ctrl+C to exit"))
	}

	return s.String()
}

func (m Model) startClaudeOAuth() tea.Cmd {
	return func() tea.Msg {
		// This would integrate with Claude Code's OAuth
		// For now, return a message
		return nil
	}
}

// Helper functions to check CLI availability
func checkClaudeMaxAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func checkGeminiCLIAvailable() bool {
	_, err := exec.LookPath("gemini")
	return err == nil
}

func checkCodexAvailable() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

// Run starts the setup wizard
func Run() (*config.Config, error) {
	p := tea.NewProgram(NewWizard(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return nil, err
	}

	wizard := m.(Model)
	return wizard.config, nil
}

// NeedsSetup checks if setup is required
func NeedsSetup() bool {
	return !config.Exists()
}
