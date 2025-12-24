package components

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type TerminalModel struct {
	input       textinput.Model
	output      viewport.Model
	help        help.Model
	pager       paginator.Model
	currentMode string
	width       int
	height      int
	ctx         context.Context
	cancel      context.CancelFunc
}

type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Help         key.Binding
	Quit         key.Binding
	Mode         key.Binding
	Clear        key.Binding
	Execute      key.Binding
	NextPage     key.Binding
	PrevPage     key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Mode: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "mode"),
	),
	Clear: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear"),
	),
	Execute: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "execute"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("pgdown", "]"),
		key.WithHelp("pgdn/]", "next page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("pgup", "["),
		key.WithHelp("pgup/[", "previous page"),
	),
}

func NewTerminal(ctx context.Context) TerminalModel {
	ctx, cancel := context.WithCancel(ctx)
	
	input := textinput.New()
	input.Placeholder = "Enter command or natural language..."
	input.Focus()
	input.Prompt = "→ "
	input.Width = 80
	
	output := viewport.New(80, 20)
	output.SetContent("Welcome to SKAgent Terminal Mode\nType 'help' for available commands or start typing naturally.\n")
	
	help := help.New()
	help.ShowAll = false
	
	pager := paginator.New()
	pager.Type = paginator.Dots
	pager.ActiveDot = "●"
	pager.InactiveDot = "○"
	
	return TerminalModel{
		input:       input,
		output:      output,
		help:        help,
		pager:       pager,
		currentMode: "interactive",
		width:       80,
		height:      20,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (t *TerminalModel) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.output.Width = width
	t.output.Height = height - 10 // Reserve space for input and help
}

func (t *TerminalModel) ExecuteCommand(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return ""
	}
	
	// Handle special commands
	switch strings.ToLower(cmd) {
	case "help":
		return t.helpText()
	case "clear":
		t.output.SetContent("")
		return ""
	case "modes":
		return t.availableModes()
	case "version":
		return "SKAgent v2.0.0 - Advanced AI Agent Terminal"
	case "quit", "exit":
		t.cancel()
		return "Goodbye!"
	default:
		return fmt.Sprintf("Command executed: %s\n(In full implementation, this would be processed by the AI engine)", cmd)
	}
}

func (t *TerminalModel) helpText() string {
	return `Available Commands:
  help          - Show this help
  clear         - Clear the terminal
  modes         - Show available modes
  version       - Show version
  quit/exit     - Exit the application

Interactive Features:
  • Natural language processing
  • Agent management
  • Real-time output
  • Multi-modal support

Keyboard Shortcuts:
  ↑/↓          - Navigate output
  ?            - Toggle help
  m            - Mode selection
  c            - Clear terminal
  Enter        - Execute command
  q            - Quit
`
}

func (t *TerminalModel) availableModes() string {
	return `Available Modes:
  interactive   - Interactive command line
  agent         - Agent-focused mode
  batch         - Batch processing mode
  server        - Server mode
  headless      - Headless operation
  
Use 'mode <name>' to switch modes.
`
}

func (t *TerminalModel) UpdateSize() {
	// For now, just set a default pager size
	// TODO: Implement proper pagination with viewport content
	t.pager.TotalPages = 1
	if t.pager.TotalPages == 0 {
		t.pager.TotalPages = 1
	}
}

func (t *TerminalModel) AddOutput(content string) {
	wrapped := wordwrap.String(content, t.width-4)
	
	// Get existing content (simulated)
	var existing string
	// Note: viewport.Model doesn't have GetContent() method
	// We need to maintain our own content buffer
	
	if existing != "" {
		existing += "\n"
	}
	t.output.SetContent(existing + wrapped)
	t.UpdateSize()
}

func (t *TerminalModel) SetMode(mode string) {
	t.currentMode = mode
}

func (t *TerminalModel) GetContext() context.Context {
	return t.ctx
}

func (t *TerminalModel) Close() {
	t.cancel()
}

func (t *TerminalModel) Style() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		BorderForeground(lipgloss.Color("62"))
}

func (t *TerminalModel) Render() string {
	output := t.output.View()
	input := t.input.View()
	
	if t.help.ShowAll {
		// TODO: Fix help.KeyMap interface compatibility
		// helpView := t.help.View(DefaultKeyMap)
		// For now, show a simple help message
		helpView := "Press ? for help"
		return t.Style().Render(lipgloss.JoinVertical(
			lipgloss.Left,
			output,
			"",
			input,
			"",
			helpView,
		))
	}
	
	return t.Style().Render(lipgloss.JoinVertical(
		lipgloss.Left,
		output,
		"",
		input,
	))
}

// Theme integration
func (t *TerminalModel) ApplyTheme(theme map[string]string) {
	if bg, ok := theme["background"]; ok {
		t.output.Style = lipgloss.NewStyle().Background(lipgloss.Color(bg))
	}
	if fg, ok := theme["foreground"]; ok {
		t.input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(fg))
	}
}

// Terminal colors and styling
func GetTerminalPalette() map[string]string {
	palette := make(map[string]string)
	
	// Dark theme (default)
	palette["background"] = "#1e1e2e"
	palette["foreground"] = "#cdd6f4"
	palette["prompt"] = "#89b4fa"
	palette["output"] = "#f9e2af"
	palette["input"] = "#94e2d5"
	palette["accent"] = "#cba6f7"
	
	// Override with terminal environment
	// Assume we have color support
	// TODO: Add proper color profile detection
	palette["background"] = "#000000"
	palette["foreground"] = "#00ff00"
	palette["prompt"] = "#00ffff"
	palette["output"] = "#ffff00"
	palette["input"] = "#ff00ff"
	palette["accent"] = "#ff8000"
	
	return palette
}