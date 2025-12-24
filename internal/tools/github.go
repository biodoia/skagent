package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitHubTool provides GitHub operations via gh CLI
type GitHubTool struct {
	mcpEndpoint string
	timeout     time.Duration
}

// NewGitHubTool creates a new GitHub tool
func NewGitHubTool(mcpEndpoint string) *GitHubTool {
	return &GitHubTool{
		mcpEndpoint: mcpEndpoint,
		timeout:     DefaultTimeout,
	}
}

// Name returns the tool identifier
func (g *GitHubTool) Name() string {
	return "github"
}

// Description returns tool description
func (g *GitHubTool) Description() string {
	return "GitHub operations: create repos, clone, manage issues, pull requests"
}

// CanHandle checks if this tool can handle the intent
func (g *GitHubTool) CanHandle(intent string) bool {
	lower := strings.ToLower(intent)
	keywords := []string{"github", "repo", "repository", "clone", "issue", "pr", "pull request"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// Execute runs the appropriate gh command
func (g *GitHubTool) Execute(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)

	// Add timeout to context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	switch {
	case strings.Contains(lower, "create") || strings.Contains(lower, "new repo"):
		return g.createRepo(ctx, input)
	case strings.Contains(lower, "clone"):
		return g.cloneRepo(ctx, input)
	case strings.Contains(lower, "issue"):
		return g.manageIssue(ctx, input)
	case strings.Contains(lower, "pr") || strings.Contains(lower, "pull request"):
		return g.managePR(ctx, input)
	case strings.Contains(lower, "list"):
		return g.listRepos(ctx)
	default:
		return "", fmt.Errorf("unknown github command in input: %s", input)
	}
}

func (g *GitHubTool) createRepo(ctx context.Context, input string) (string, error) {
	// Extract repo name
	repoName := extractArg(input, "create")
	if repoName == "" {
		repoName = extractArg(input, "new")
	}
	if repoName == "" {
		return "", fmt.Errorf("repo name not found in input")
	}

	// Determine visibility
	visibility := "--private"
	if strings.Contains(strings.ToLower(input), "public") {
		visibility = "--public"
	}

	cmd := exec.CommandContext(ctx, "gh", "repo", "create", repoName, visibility, "--confirm")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", g.timeout)
		}
		return "", fmt.Errorf("failed to create repo: %w\n%s", err, output)
	}

	return fmt.Sprintf("Repository '%s' created successfully!\n%s", repoName, string(output)), nil
}

func (g *GitHubTool) cloneRepo(ctx context.Context, input string) (string, error) {
	repoURL := extractArg(input, "clone")
	if repoURL == "" {
		return "", fmt.Errorf("repo URL not found in input")
	}

	cmd := exec.CommandContext(ctx, "gh", "repo", "clone", repoURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", g.timeout)
		}
		return "", fmt.Errorf("failed to clone repo: %w\n%s", err, output)
	}

	return fmt.Sprintf("Repository cloned successfully!\n%s", string(output)), nil
}

func (g *GitHubTool) manageIssue(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)

	if strings.Contains(lower, "create") || strings.Contains(lower, "new") {
		title := extractQuotedArg(input)
		if title == "" {
			title = "New Issue"
		}
		cmd := exec.CommandContext(ctx, "gh", "issue", "create", "--title", title)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to create issue: %w\n%s", err, output)
		}
		return string(output), nil
	}

	if strings.Contains(lower, "list") {
		cmd := exec.CommandContext(ctx, "gh", "issue", "list")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to list issues: %w\n%s", err, output)
		}
		return string(output), nil
	}

	return "", fmt.Errorf("unknown issue command")
}

func (g *GitHubTool) managePR(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)

	if strings.Contains(lower, "create") || strings.Contains(lower, "new") {
		cmd := exec.CommandContext(ctx, "gh", "pr", "create", "--fill")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to create PR: %w\n%s", err, output)
		}
		return string(output), nil
	}

	if strings.Contains(lower, "list") {
		cmd := exec.CommandContext(ctx, "gh", "pr", "list")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to list PRs: %w\n%s", err, output)
		}
		return string(output), nil
	}

	return "", fmt.Errorf("unknown PR command")
}

func (g *GitHubTool) listRepos(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "repo", "list", "--limit", "20")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list repos: %w\n%s", err, output)
	}
	return string(output), nil
}

// extractQuotedArg extracts content within quotes (double or single)
func extractQuotedArg(input string) string {
	// Try double quotes first
	if start := strings.Index(input, "\""); start != -1 {
		end := strings.Index(input[start+1:], "\"")
		if end != -1 {
			return input[start+1 : start+1+end]
		}
	}
	// Try single quotes
	if start := strings.Index(input, "'"); start != -1 {
		end := strings.Index(input[start+1:], "'")
		if end != -1 {
			return input[start+1 : start+1+end]
		}
	}
	return ""
}
