package components

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ThemeItem implements list.Item for theme selection
type ThemeItem struct {
	name string
}

func (t ThemeItem) FilterValue() string {
	return t.name
}

func (t ThemeItem) Title() string {
	return t.name
}

func (t ThemeItem) Description() string {
	return fmt.Sprintf("Theme: %s", t.name)
}

type Theme struct {
	Name             string            `json:"name"`
	Colors           map[string]string `json:"colors"`
	FontSize         int               `json:"font_size"`
	ShowAnimations   bool              `json:"show_animations"`
	CompactMode      bool              `json:"compact_mode"`
	AutoSave         bool              `json:"auto_save"`
}

type SettingsModel struct {
	themeList       list.Model
	fontSizeInput   textinput.Model
	animationsInput textinput.Model
	compactInput    textinput.Model
	themes          []Theme
	currentTheme    Theme
	configPath      string
	ctx             context.Context
	changed         bool
}

func NewSettings(ctx context.Context) SettingsModel {
	// Get config directory
	configDir := getConfigDir()
	
	themeList := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 10)
	themeList.Title = "Available Themes"
	
	fontSizeInput := textinput.New()
	fontSizeInput.Placeholder = "14"
	fontSizeInput.Width = 10
	
	animationsInput := textinput.New()
	animationsInput.Placeholder = "true"
	animationsInput.Width = 10
	
	compactInput := textinput.New()
	compactInput.Placeholder = "false"
	compactInput.Width = 10
	
	// Load default themes
	themes := loadDefaultThemes()
	
	// Populate theme list
	items := make([]list.Item, len(themes))
	for i, theme := range themes {
		items[i] = ThemeItem{name: theme.Name}
	}
	themeList.SetItems(items)
	
	return SettingsModel{
		themeList:       themeList,
		fontSizeInput:   fontSizeInput,
		animationsInput: animationsInput,
		compactInput:    compactInput,
		themes:          themes,
		currentTheme:    themes[0], // Default to first theme
		configPath:      configDir,
		ctx:             ctx,
		changed:         false,
	}
}

func (s *SettingsModel) Init() {
	// Set current values
	s.fontSizeInput.SetValue(fmt.Sprintf("%d", s.currentTheme.FontSize))
	s.animationsInput.SetValue(fmt.Sprintf("%v", s.currentTheme.ShowAnimations))
	s.compactInput.SetValue(fmt.Sprintf("%v", s.currentTheme.CompactMode))
	
	// Load saved settings
	s.loadSettings()
}

func (s *SettingsModel) SaveSettings() error {
	if !s.changed {
		return nil
	}
	
	s.currentTheme.FontSize = 14 // Parse from input
	if val := s.fontSizeInput.Value(); val != "" {
		if size, err := fmt.Sscanf(val, "%d", &s.currentTheme.FontSize); err == nil && size == 1 {
			// Successfully parsed
		}
	}
	
	// Parse animations
	if val := s.animationsInput.Value(); val != "" {
		s.currentTheme.ShowAnimations = (val == "true" || val == "1")
	}
	
	// Parse compact mode
	if val := s.compactInput.Value(); val != "" {
		s.currentTheme.CompactMode = (val == "true" || val == "1")
	}
	
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(s.configPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Save to file
	configFile := filepath.Join(s.configPath, "settings.json")
	data, err := json.MarshalIndent(s.currentTheme, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}
	
	s.changed = false
	return nil
}

func (s *SettingsModel) loadSettings() error {
	configFile := filepath.Join(s.configPath, "settings.json")
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No settings file, use defaults
			return nil
		}
		return fmt.Errorf("failed to read settings file: %w", err)
	}
	
	var loadedTheme Theme
	if err := json.Unmarshal(data, &loadedTheme); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}
	
	s.currentTheme = loadedTheme
	s.updateUI()
	return nil
}

func (s *SettingsModel) updateUI() {
	s.fontSizeInput.SetValue(fmt.Sprintf("%d", s.currentTheme.FontSize))
	s.animationsInput.SetValue(fmt.Sprintf("%v", s.currentTheme.ShowAnimations))
	s.compactInput.SetValue(fmt.Sprintf("%v", s.currentTheme.CompactMode))
	
	// Select theme in list
	for i, theme := range s.themes {
		if theme.Name == s.currentTheme.Name {
			// Select item in list (would be implemented with list navigation)
			s.themeList.Select(i)
			break
		}
	}
}

func (s *SettingsModel) ApplyTheme(themeName string) {
	for _, theme := range s.themes {
		if theme.Name == themeName {
			s.currentTheme = theme
			s.changed = true
			break
		}
	}
}

func (s *SettingsModel) GetCurrentTheme() Theme {
	return s.currentTheme
}

func (s *SettingsModel) MarkChanged() {
	s.changed = true
}

func (s *SettingsModel) Render() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("89b4fa")).
		Render("⚙️ Settings & Themes")
	
	themeSection := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("f9e2af")).
		Render("🎨 Appearance")
	
	displaySection := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("f9e2af")).
		Render("📺 Display")
	
	themeView := s.themeList.View()
	fontSizeView := fmt.Sprintf("Font Size: %s", s.fontSizeInput.View())
	animationsView := fmt.Sprintf("Enable Animations: %s", s.animationsInput.View())
	compactView := fmt.Sprintf("Compact Mode: %s", s.compactInput.View())
	
	saveStatus := ""
	if s.changed {
		saveStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("f38ba8")).
			Render("● Unsaved changes")
	} else {
		saveStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("a6e3a1")).
			Render("✓ All changes saved")
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		themeSection,
		themeView,
		"",
		displaySection,
		fontSizeView,
		animationsView,
		compactView,
		"",
		saveStatus,
	)
}

func getConfigDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "skagent")
	}
	return ".config/skagent"
}

func loadDefaultThemes() []Theme {
	return []Theme{
		{
			Name: "Dark Terminal",
			Colors: map[string]string{
				"background":     "#1e1e2e",
				"foreground":     "#cdd6f4",
				"primary":        "#89b4fa",
				"secondary":      "#f9e2af",
				"accent":         "#cba6f7",
				"success":        "#a6e3a1",
				"warning":        "#f9e2af",
				"error":          "#f38ba8",
				"header_background": "#313244",
				"header_foreground": "#cdd6f4",
			},
			FontSize:        14,
			ShowAnimations:  true,
			CompactMode:     false,
			AutoSave:        true,
		},
		{
			Name: "Light",
			Colors: map[string]string{
				"background":     "#ffffff",
				"foreground":     "#1c1c1c",
				"primary":        "#0066cc",
				"secondary":      "#666666",
				"accent":         "#9900cc",
				"success":        "#00aa00",
				"warning":        "#cc8800",
				"error":          "#cc0000",
				"header_background": "#f0f0f0",
				"header_foreground": "#1c1c1c",
			},
			FontSize:        16,
			ShowAnimations:  false,
			CompactMode:     true,
			AutoSave:        true,
		},
		{
			Name: "Solarized Dark",
			Colors: map[string]string{
				"background":     "#002b36",
				"foreground":     "#839496",
				"primary":        "#268bd2",
				"secondary":      "#b58900",
				"accent":         "#d33682",
				"success":        "#859900",
				"warning":        "#b58900",
				"error":          "#dc322f",
				"header_background": "#073642",
				"header_foreground": "#839496",
			},
			FontSize:        13,
			ShowAnimations:  true,
			CompactMode:     false,
			AutoSave:        true,
		},
		{
			Name: "Neon",
			Colors: map[string]string{
				"background":     "#000000",
				"foreground":     "#00ff00",
				"primary":        "#00ffff",
				"secondary":      "#ffff00",
				"accent":         "#ff00ff",
				"success":        "#00ff00",
				"warning":        "#ffff00",
				"error":          "#ff0000",
				"header_background": "#111111",
				"header_foreground": "#00ff00",
			},
			FontSize:        12,
			ShowAnimations:  true,
			CompactMode:     false,
			AutoSave:        true,
		},
	}
}

func (s *SettingsModel) ExportTheme(filename string) error {
	data, err := json.MarshalIndent(s.currentTheme, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

func (s *SettingsModel) ImportTheme(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return err
	}
	
	s.themes = append(s.themes, theme)
	s.changed = true
	return nil
}