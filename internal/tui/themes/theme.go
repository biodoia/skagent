package themes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme and styling for the TUI
type Theme struct {
	Name        string      `json:"name"`
	Author      string      `json:"author,omitempty"`
	Description string      `json:"description,omitempty"`
	Colors      ThemeColors `json:"colors"`
}

// ThemeColors contains all color definitions
type ThemeColors struct {
	// Base colors
	Background    string `json:"background"`
	Foreground    string `json:"foreground"`
	Primary       string `json:"primary"`
	Secondary     string `json:"secondary"`
	Accent        string `json:"accent"`
	
	// Semantic colors
	Success       string `json:"success"`
	Warning       string `json:"warning"`
	Error         string `json:"error"`
	Info          string `json:"info"`
	
	// UI elements
	Border        string `json:"border"`
	BorderFocused string `json:"border_focused"`
	Selection     string `json:"selection"`
	Muted         string `json:"muted"`
	
	// Chat colors
	UserMessage      string `json:"user_message"`
	AssistantMessage string `json:"assistant_message"`
	SystemMessage    string `json:"system_message"`
	
	// Syntax highlighting
	Keyword    string `json:"keyword"`
	String     string `json:"string"`
	Number     string `json:"number"`
	Comment    string `json:"comment"`
	Function   string `json:"function"`
}

// Styles contains pre-computed lipgloss styles
type Styles struct {
	// App
	App           lipgloss.Style
	Header        lipgloss.Style
	Footer        lipgloss.Style
	
	// Panels
	Panel         lipgloss.Style
	PanelFocused  lipgloss.Style
	PanelTitle    lipgloss.Style
	
	// Messages
	UserMessage      lipgloss.Style
	AssistantMessage lipgloss.Style
	SystemMessage    lipgloss.Style
	ErrorMessage     lipgloss.Style
	
	// Input
	Input         lipgloss.Style
	InputFocused  lipgloss.Style
	Placeholder   lipgloss.Style
	
	// Components
	Button        lipgloss.Style
	ButtonActive  lipgloss.Style
	Tab           lipgloss.Style
	TabActive     lipgloss.Style
	Badge         lipgloss.Style
	Progress      lipgloss.Style
	
	// Status
	StatusOnline  lipgloss.Style
	StatusOffline lipgloss.Style
	StatusBusy    lipgloss.Style
	
	// Text
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	Body          lipgloss.Style
	Muted         lipgloss.Style
	Bold          lipgloss.Style
	Italic        lipgloss.Style
	Code          lipgloss.Style
}

// ThemeManager handles theme loading and switching
type ThemeManager struct {
	current    *Theme
	styles     *Styles
	themes     map[string]*Theme
	themesPath string
}

// NewThemeManager creates a new theme manager
func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		themes: make(map[string]*Theme),
	}
	
	// Register built-in themes
	tm.registerBuiltinThemes()
	
	// Set default theme
	tm.SetTheme("catppuccin-mocha")
	
	return tm
}

// registerBuiltinThemes adds all built-in themes
func (tm *ThemeManager) registerBuiltinThemes() {
	tm.themes["catppuccin-mocha"] = CatppuccinMocha()
	tm.themes["catppuccin-latte"] = CatppuccinLatte()
	tm.themes["dracula"] = Dracula()
	tm.themes["nord"] = Nord()
	tm.themes["tokyo-night"] = TokyoNight()
	tm.themes["gruvbox-dark"] = GruvboxDark()
	tm.themes["one-dark"] = OneDark()
	tm.themes["solarized-dark"] = SolarizedDark()
	tm.themes["monokai"] = Monokai()
	tm.themes["github-dark"] = GitHubDark()
}

// ListThemes returns all available theme names
func (tm *ThemeManager) ListThemes() []string {
	names := make([]string, 0, len(tm.themes))
	for name := range tm.themes {
		names = append(names, name)
	}
	return names
}

// GetTheme returns a theme by name
func (tm *ThemeManager) GetTheme(name string) (*Theme, bool) {
	theme, ok := tm.themes[name]
	return theme, ok
}

// CurrentTheme returns the current theme
func (tm *ThemeManager) CurrentTheme() *Theme {
	return tm.current
}

// Styles returns the current styles
func (tm *ThemeManager) Styles() *Styles {
	return tm.styles
}

// SetTheme switches to a new theme
func (tm *ThemeManager) SetTheme(name string) error {
	theme, ok := tm.themes[name]
	if !ok {
		return fmt.Errorf("theme not found: %s", name)
	}
	
	tm.current = theme
	tm.styles = tm.buildStyles(theme)
	return nil
}

// LoadCustomTheme loads a theme from a JSON file
func (tm *ThemeManager) LoadCustomTheme(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read theme file: %w", err)
	}
	
	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return fmt.Errorf("failed to parse theme: %w", err)
	}
	
	// Use filename as theme name if not specified
	if theme.Name == "" {
		theme.Name = filepath.Base(path)
	}
	
	tm.themes[theme.Name] = &theme
	return nil
}

// buildStyles creates lipgloss styles from a theme
func (tm *ThemeManager) buildStyles(t *Theme) *Styles {
	c := t.Colors
	
	return &Styles{
		// App
		App: lipgloss.NewStyle().
			Background(lipgloss.Color(c.Background)).
			Foreground(lipgloss.Color(c.Foreground)),
			
		Header: lipgloss.NewStyle().
			Background(lipgloss.Color(c.Primary)).
			Foreground(lipgloss.Color(c.Background)).
			Bold(true).
			Padding(0, 1),
			
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Muted)).
			Padding(0, 1),
		
		// Panels
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(c.Border)).
			Padding(1, 2),
			
		PanelFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(c.BorderFocused)).
			Padding(1, 2),
			
		PanelTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Primary)).
			Bold(true),
		
		// Messages
		UserMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.UserMessage)).
			Bold(true),
			
		AssistantMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.AssistantMessage)),
			
		SystemMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.SystemMessage)).
			Italic(true),
			
		ErrorMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Error)).
			Bold(true),
		
		// Input
		Input: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(c.Border)).
			Padding(0, 1),
			
		InputFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(c.Primary)).
			Padding(0, 1),
			
		Placeholder: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Muted)),
		
		// Components
		Button: lipgloss.NewStyle().
			Background(lipgloss.Color(c.Secondary)).
			Foreground(lipgloss.Color(c.Foreground)).
			Padding(0, 2),
			
		ButtonActive: lipgloss.NewStyle().
			Background(lipgloss.Color(c.Primary)).
			Foreground(lipgloss.Color(c.Background)).
			Bold(true).
			Padding(0, 2),
			
		Tab: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Muted)).
			Padding(0, 2),
			
		TabActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Primary)).
			Bold(true).
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color(c.Primary)),
			
		Badge: lipgloss.NewStyle().
			Background(lipgloss.Color(c.Accent)).
			Foreground(lipgloss.Color(c.Background)).
			Padding(0, 1),
			
		Progress: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Primary)),
		
		// Status
		StatusOnline: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Success)),
			
		StatusOffline: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Muted)),
			
		StatusBusy: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Warning)),
		
		// Text
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Primary)).
			Bold(true),
			
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Secondary)),
			
		Body: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Foreground)),
			
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Muted)),
			
		Bold: lipgloss.NewStyle().
			Bold(true),
			
		Italic: lipgloss.NewStyle().
			Italic(true),
			
		Code: lipgloss.NewStyle().
			Background(lipgloss.Color(c.Selection)).
			Foreground(lipgloss.Color(c.Accent)).
			Padding(0, 1),
	}
}
