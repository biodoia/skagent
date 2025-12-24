package tools

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestWebSearchTool_CanHandle(t *testing.T) {
	tool := NewWebSearchTool()

	tests := []struct {
		intent   string
		expected bool
	}{
		{"search for golang tutorials", true},
		{"find best practices", true},
		{"look up documentation", true},
		{"create a file", false},
		{"delete something", false},
		{"google this topic", true},
		{"web search for info", true},
	}

	for _, tt := range tests {
		t.Run(tt.intent, func(t *testing.T) {
			result := tool.CanHandle(tt.intent)
			if result != tt.expected {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.intent, result, tt.expected)
			}
		})
	}
}

func TestWebSearchTool_ExtractSearchTerms(t *testing.T) {
	tests := []struct {
		query    string
		minTerms int
	}{
		{"search for golang cli tools", 2},
		{"find the best dotfiles manager", 2},
		{"look up kubernetes documentation", 1},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			terms := extractSearchTerms(tt.query)
			if len(terms) < tt.minTerms {
				t.Errorf("extractSearchTerms(%q) returned %d terms, want at least %d", 
					tt.query, len(terms), tt.minTerms)
			}
		})
	}
}

func TestGitHubTool_CanHandle(t *testing.T) {
	tool := NewGitHubTool("")

	tests := []struct {
		intent   string
		expected bool
	}{
		{"create a github repo", true},
		{"clone repository", true},
		{"create new issue", true},
		{"list pull requests", true},
		{"search the web", false},
		{"make a plan", false},
	}

	for _, tt := range tests {
		t.Run(tt.intent, func(t *testing.T) {
			result := tool.CanHandle(tt.intent)
			if result != tt.expected {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.intent, result, tt.expected)
			}
		})
	}
}

func TestSpecKitTool_CanHandle(t *testing.T) {
	tool := NewSpecKitTool("")

	tests := []struct {
		intent   string
		expected bool
	}{
		{"create a spec", true},
		{"make a plan", true},
		{"generate tasks", true},
		{"implement the feature", true},
		{"search the web", false},
		{"clone a repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.intent, func(t *testing.T) {
			result := tool.CanHandle(tt.intent)
			if result != tt.expected {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.intent, result, tt.expected)
			}
		})
	}
}

func TestToolManager(t *testing.T) {
	tm := NewToolManager()
	tm.AddTool(NewSpecKitTool(""))
	tm.AddTool(NewGitHubTool(""))
	tm.AddTool(NewWebSearchTool())

	t.Run("GetTool", func(t *testing.T) {
		tool := tm.GetTool("speckit")
		if tool == nil {
			t.Error("GetTool('speckit') returned nil")
		}
		if tool.Name() != "speckit" {
			t.Errorf("GetTool('speckit').Name() = %q, want 'speckit'", tool.Name())
		}
	})

	t.Run("FindTool", func(t *testing.T) {
		tool := tm.FindTool("search for something")
		if tool == nil {
			t.Error("FindTool for search intent returned nil")
		}
		if tool.Name() != "websearch" {
			t.Errorf("FindTool returned %q, want 'websearch'", tool.Name())
		}
	})

	t.Run("CanHandle", func(t *testing.T) {
		if !tm.CanHandle("create a github repo") {
			t.Error("CanHandle should return true for github intent")
		}
		if tm.CanHandle("unknown random thing xyz") {
			t.Error("CanHandle should return false for unknown intent")
		}
	})
}

func TestExtractArg(t *testing.T) {
	tests := []struct {
		input    string
		keyword  string
		expected string
	}{
		{"init myproject", "init", "myproject"},
		{"create new-repo", "create", "new-repo"},
		{"clone github.com/user/repo", "clone", "github.com/user/repo"},
		{"no keyword here", "init", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractArg(tt.input, tt.keyword)
			if result != tt.expected {
				t.Errorf("extractArg(%q, %q) = %q, want %q", 
					tt.input, tt.keyword, result, tt.expected)
			}
		})
	}
}

func TestExtractQuotedArg(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`create issue "Fix the bug"`, "Fix the bug"},
		{`new issue 'Add feature'`, "Add feature"},
		{"no quotes here", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractQuotedArg(tt.input)
			if result != tt.expected {
				t.Errorf("extractQuotedArg(%q) = %q, want %q", 
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestWebSearch_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tool := NewWebSearchTool()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := tool.Execute(ctx, "search github golang cli framework")
	if err != nil {
		t.Fatalf("WebSearch failed: %v", err)
	}

	if !strings.Contains(result, "Found") && !strings.Contains(result, "repository") {
		t.Errorf("Expected search results, got: %s", result[:min(100, len(result))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
