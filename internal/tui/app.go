package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sergio/skagent/internal/ai"
	"github.com/sergio/skagent/internal/config"
	"github.com/sergio/skagent/internal/tools"
)

// RequestTimeout for AI and tool operations
const RequestTimeout = 60 * time.Second

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			Background(lipgloss.Color("#1E1E2E")).
			Padding(0, 1)

	userStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")).
			Bold(true)

	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))

	systemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9E2AF")).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89B4FA")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			Italic(true)

	providerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CBA6F7")).
			Bold(true)
)

// Message types for tea.Msg
type Message struct {
	Role    string
	Content string
}

type aiResponseMsg struct {
	response string
	err      error
}

type toolResultMsg struct {
	tool   string
	result string
	err    error
}

// Model is the main application model
type Model struct {
	messages    []Message
	history     []ai.Message // AI conversation history
	input       textinput.Model
	viewport    viewport.Model
	spinner     spinner.Model
	provider    ai.Provider
	config      *config.Config
	tools       *tools.ToolManager
	autonomous  bool
	loading     bool
	width       int
	height      int
	ready       bool
}

// InitialModel creates the initial application state with default config
func InitialModel() Model {
	cfg := config.DefaultConfig()
	return initialModelWithConfig(cfg)
}

// InitialModelWithConfig creates the initial application state with custom config
func initialModelWithConfig(cfg *config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "Describe your project idea... (type /help for commands)"
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 70

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))

	// Initialize tool manager with all tools
	tm := tools.NewToolManager()
	tm.AddTool(tools.NewSpecKitTool(""))
	tm.AddTool(tools.NewGitHubTool(""))
	tm.AddTool(tools.NewWebSearchTool())

	// Create AI provider
	var provider ai.Provider
	var err error
	if cfg != nil {
		provider, err = ai.CreateProvider(cfg)
		if err != nil {
			// Will show error in UI
			provider = nil
		}
	}

	return Model{
		messages:   []Message{},
		history:    []ai.Message{},
		input:      ti,
		spinner:    sp,
		provider:   provider,
		config:     cfg,
		tools:      tm,
		autonomous: false,
		loading:    false,
		ready:      false,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.loading {
				m.loading = false
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			if m.input.Value() != "" && !m.loading {
				userInput := m.input.Value()
				m.input.Reset()

				// Handle commands
				if strings.HasPrefix(userInput, "/") {
					return m.handleCommand(userInput)
				}

				// Add user message
				m.messages = append(m.messages, Message{
					Role:    "user",
					Content: userInput,
				})
				m.history = append(m.history, ai.Message{
					Role:    "user",
					Content: userInput,
				})
				m.loading = true

				// Process based on mode
				if m.autonomous {
					return m, m.processAutonomous(userInput)
				}
				return m, m.processInteractive(userInput)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 3
		footerHeight := 4

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}
		m.input.Width = msg.Width - 4
		m.viewport.SetContent(m.renderMessages())

	case aiResponseMsg:
		m.loading = false
		if msg.err != nil {
			m.messages = append(m.messages, Message{
				Role:    "error",
				Content: fmt.Sprintf("Error: %v", msg.err),
			})
		} else {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: msg.response,
			})
			m.history = append(m.history, ai.Message{
				Role:    "assistant",
				Content: msg.response,
			})
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case toolResultMsg:
		if msg.err != nil {
			m.messages = append(m.messages, Message{
				Role:    "error",
				Content: fmt.Sprintf("Tool %s error: %v", msg.tool, msg.err),
			})
		} else {
			m.messages = append(m.messages, Message{
				Role:    "system",
				Content: fmt.Sprintf("[%s] %s", msg.tool, msg.result),
			})
		}
		m.viewport.SetContent(m.renderMessages())

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update input
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	// Update viewport
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

func (m Model) handleCommand(cmd string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(cmd)
	command := strings.ToLower(parts[0])

	switch command {
	case "/auto", "/autonomous":
		m.autonomous = !m.autonomous
		status := "disabled"
		if m.autonomous {
			status = "enabled"
		}
		m.messages = append(m.messages, Message{
			Role:    "system",
			Content: fmt.Sprintf("Autonomous mode %s", status),
		})

	case "/clear":
		m.messages = []Message{}
		m.history = []ai.Message{}

	case "/provider":
		if m.config != nil {
			providerName := "unknown"
			if m.provider != nil {
				providerName = m.provider.Name()
			}
			model := m.config.GetActiveProvider().Model
			m.messages = append(m.messages, Message{
				Role:    "system",
				Content: fmt.Sprintf("Current provider: %s\nModel: %s", providerName, model),
			})
		}

	case "/models":
		var sb strings.Builder
		sb.WriteString("Available free models on OpenRouter:\n\n")
		for i, model := range config.OpenRouterFreeModels {
			if model.Recommended {
				sb.WriteString("⭐ ")
			} else {
				sb.WriteString("   ")
			}
			sb.WriteString(fmt.Sprintf("%d. %s (%dk ctx)\n", i+1, model.Name, model.ContextLength/1000))
		}
		m.messages = append(m.messages, Message{
			Role:    "system",
			Content: sb.String(),
		})

	case "/help":
		m.messages = append(m.messages, Message{
			Role:    "system",
			Content: helpText(),
		})

	case "/quit", "/exit":
		return m, tea.Quit

	default:
		m.messages = append(m.messages, Message{
			Role:    "error",
			Content: fmt.Sprintf("Unknown command: %s\nType /help for available commands", cmd),
		})
	}
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
	return m, nil
}

func helpText() string {
	return `
╭─────────────────────────────────────────────╮
│           SkAgent Commands                   │
╰─────────────────────────────────────────────╯

  /auto      Toggle autonomous mode
  /provider  Show current AI provider
  /models    List available free models
  /clear     Clear conversation
  /help      Show this help
  /quit      Exit application

╭─────────────────────────────────────────────╮
│           SpecKit Workflow                   │
╰─────────────────────────────────────────────╯

In autonomous mode, the agent will:
1. Analyze your project idea
2. Search for best practices
3. Generate SpecKit specifications
4. Create technical plans
5. Break down into tasks

Just describe your project idea to get started!

╭─────────────────────────────────────────────╮
│           Keyboard Shortcuts                 │
╰─────────────────────────────────────────────╯

  Enter      Send message
  Ctrl+C     Exit
  Esc        Cancel/Exit
  ↑/↓        Scroll messages`
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Header with provider info
	header := titleStyle.Render("🚀 SkAgent")

	providerInfo := ""
	if m.provider != nil {
		providerInfo = providerStyle.Render(fmt.Sprintf(" [%s]", m.provider.Name()))
	}

	modeIndicator := ""
	if m.autonomous {
		modeIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")).
			Bold(true).
			Render(" AUTO")
	}
	header += providerInfo + modeIndicator

	// Error if no provider
	if m.provider == nil {
		return fmt.Sprintf("%s\n\n%s\n\nRun 'skagent setup' to configure an AI provider.",
			header,
			errorStyle.Render("⚠ No AI provider configured"))
	}

	// Status line
	model := ""
	if m.config != nil {
		model = m.config.GetActiveProvider().Model
		if model != "" {
			// Shorten model name for display
			parts := strings.Split(model, "/")
			if len(parts) > 1 {
				model = parts[1]
			}
			if len(model) > 30 {
				model = model[:27] + "..."
			}
		}
	}
	status := statusStyle.Render(fmt.Sprintf("Model: %s | Messages: %d | /help for commands", model, len(m.messages)))

	// Loading indicator
	loadingIndicator := ""
	if m.loading {
		loadingIndicator = fmt.Sprintf(" %s Thinking...", m.spinner.View())
	}

	// Input box
	inputBox := inputStyle.Render(m.input.View()) + loadingIndicator

	return fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
		header,
		m.viewport.View(),
		inputBox,
		status,
	)
}

func (m Model) renderMessages() string {
	var sb strings.Builder

	if len(m.messages) == 0 {
		sb.WriteString(systemStyle.Render("Welcome! Describe your project idea or type /help for commands.\n"))
	}

	for _, msg := range m.messages {
		var styled string
		switch msg.Role {
		case "user":
			styled = userStyle.Render("You: ") + msg.Content
		case "assistant":
			styled = assistantStyle.Render("Agent: ") + msg.Content
		case "system":
			styled = systemStyle.Render("System: ") + msg.Content
		case "error":
			styled = errorStyle.Render("Error: ") + msg.Content
		default:
			styled = msg.Content
		}
		sb.WriteString(styled + "\n\n")
	}

	return sb.String()
}

func (m Model) processInteractive(input string) tea.Cmd {
	return func() tea.Msg {
		if m.provider == nil {
			return aiResponseMsg{err: fmt.Errorf("no AI provider configured")}
		}

		systemPrompt := ai.SystemPrompt + "\n\n" + ai.SpecKitDocs

		response, err := m.provider.Complete(context.Background(), m.history, systemPrompt)
		return aiResponseMsg{response: response, err: err}
	}
}

func (m Model) processAutonomous(input string) tea.Cmd {
	return func() tea.Msg {
		if m.provider == nil {
			return aiResponseMsg{err: fmt.Errorf("no AI provider configured")}
		}

		// In autonomous mode, we add extra context
		prompt := fmt.Sprintf(`You are in AUTONOMOUS mode. The user wants to create a project:

"%s"

Analyze this idea and provide:
1. A clear summary of what will be built
2. Key requirements and features
3. Suggested tech stack based on the requirements
4. A step-by-step plan using the SpecKit workflow:
   - SPECIFY: Define what and why
   - PLAN: Technical blueprint
   - TASKS: Atomic work items
   - IMPLEMENT: Build with TDD

Be proactive and thorough. Start generating specifications immediately.`, input)

		// Replace last user message with enhanced prompt
		history := make([]ai.Message, len(m.history)-1)
		copy(history, m.history[:len(m.history)-1])
		history = append(history, ai.Message{Role: "user", Content: prompt})

		systemPrompt := ai.SystemPrompt + "\n\n" + ai.SpecKitDocs

		response, err := m.provider.Complete(context.Background(), history, systemPrompt)
		return aiResponseMsg{response: response, err: err}
	}
}

// Run starts the TUI application with default config
func Run() error {
	return RunWithConfig(nil)
}

// RunWithConfig starts the TUI application with custom config
func RunWithConfig(cfg *config.Config) error {
	var m Model
	if cfg != nil {
		m = initialModelWithConfig(cfg)
	} else {
		m = InitialModel()
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
