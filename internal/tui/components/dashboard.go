package components

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type AgentInfo struct {
	ID          string
	Name        string
	Status      string
	Type        string
	LastActive  time.Time
	TasksDone   int
	SuccessRate float64
}

type DashboardModel struct {
	agents       []AgentInfo
	selectedRow  int
	table        table.Model
	search       textinput.Model
	ctx          context.Context
	width        int
	height       int
}

func NewDashboard(ctx context.Context) DashboardModel {
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 12},
		{Title: "Type", Width: 15},
		{Title: "Last Active", Width: 18},
		{Title: "Tasks", Width: 8},
		{Title: "Success %", Width: 10},
	}
	
	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
	)
	
	search := textinput.New()
	search.Placeholder = "Search agents..."
	search.Prompt = "🔍 "
	
	return DashboardModel{
		agents:      []AgentInfo{},
		table:       t,
		search:      search,
		ctx:         ctx,
		selectedRow: 0,
	}
}

func (d *DashboardModel) SetAgents(agents []AgentInfo) {
	d.agents = agents
	d.refreshTable()
}

func (d *DashboardModel) refreshTable() {
	rows := make([]table.Row, 0, len(d.agents))
	
	// Sort by last active
	sorted := make([]AgentInfo, len(d.agents))
	copy(sorted, d.agents)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastActive.After(sorted[j].LastActive)
	})
	
	for _, agent := range sorted {
		statusIcon := getStatusIcon(agent.Status)
		lastActive := formatRelativeTime(agent.LastActive)
		successRate := strconv.Itoa(int(agent.SuccessRate)) + "%"
		
		row := table.Row{
			agent.ID,
			agent.Name,
			statusIcon + " " + agent.Status,
			agent.Type,
			lastActive,
			strconv.Itoa(agent.TasksDone),
			successRate,
		}
		rows = append(rows, row)
	}
	
	d.table.SetRows(rows)
}

func (d *DashboardModel) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.table.SetWidth(width - 4)
	d.table.SetHeight(height - 10)
}

func (d *DashboardModel) UpdateTable() {
	d.refreshTable()
}

func (d *DashboardModel) GetSelectedAgent() *AgentInfo {
	if d.selectedRow < len(d.agents) {
		// Account for sorting
		sorted := make([]AgentInfo, len(d.agents))
		copy(sorted, d.agents)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].LastActive.After(sorted[j].LastActive)
		})
		
		if d.selectedRow < len(sorted) {
			return &sorted[d.selectedRow]
		}
	}
	return nil
}

func getStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "active", "running", "online":
		return "🟢"
	case "idle", "waiting":
		return "🟡"
	case "offline", "stopped":
		return "🔴"
	case "error", "failed":
		return "❌"
	default:
		return "⚪"
	}
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	
	if diff < time.Minute {
		return "Just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return strconv.Itoa(minutes) + "m ago"
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return strconv.Itoa(hours) + "h ago"
	} else {
		days := int(diff.Hours() / 24)
		return strconv.Itoa(days) + "d ago"
	}
}

func (d *DashboardModel) Render() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("89b4fa")).
		Render("🤖 Agent Dashboard")
	
	stats := d.renderStats()
	search := d.search.View()
	table := d.table.View()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		stats,
		"",
		"Search:",
		search,
		"",
		table,
	)
}

func (d *DashboardModel) renderStats() string {
	active := 0
	idle := 0
	offline := 0
	totalTasks := 0
	
	for _, agent := range d.agents {
		switch strings.ToLower(agent.Status) {
		case "active", "running", "online":
			active++
		case "idle", "waiting":
			idle++
		case "offline", "stopped":
			offline++
		}
		totalTasks += agent.TasksDone
	}
	
	statsText := []string{
		"📊 Statistics:",
		"  Active: " + strconv.Itoa(active),
		"  Idle: " + strconv.Itoa(idle),
		"  Offline: " + strconv.Itoa(offline),
		"  Total Tasks: " + strconv.Itoa(totalTasks),
	}
	
	return strings.Join(statsText, "\n")
}

func (d *DashboardModel) FilterAgents(query string) []AgentInfo {
	if query == "" {
		return d.agents
	}
	
	query = strings.ToLower(query)
	var filtered []AgentInfo
	
	for _, agent := range d.agents {
		if strings.Contains(strings.ToLower(agent.Name), query) ||
		   strings.Contains(strings.ToLower(agent.Type), query) ||
		   strings.Contains(strings.ToLower(agent.Status), query) {
			filtered = append(filtered, agent)
		}
	}
	
	return filtered
}

func (d *DashboardModel) ApplyTheme(theme map[string]string) {
	// Apply theme colors to table
	style := lipgloss.NewStyle()
	
	if headerBg, ok := theme["header_background"]; ok {
		style = style.Background(lipgloss.Color(headerBg))
	}
	
	if headerFg, ok := theme["header_foreground"]; ok {
		style = style.Foreground(lipgloss.Color(headerFg))
	}
}