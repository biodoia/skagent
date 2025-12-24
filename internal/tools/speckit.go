package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout for CLI commands
const DefaultTimeout = 30 * time.Second

// SpecKitTool wraps GitHub Spec-Kit commands
type SpecKitTool struct {
	docsPath string
	timeout  time.Duration
}

// NewSpecKitTool creates a new SpecKit tool
func NewSpecKitTool(docsPath string) *SpecKitTool {
	return &SpecKitTool{
		docsPath: docsPath,
		timeout:  DefaultTimeout,
	}
}

// Name returns the tool identifier
func (s *SpecKitTool) Name() string {
	return "speckit"
}

// Description returns tool description
func (s *SpecKitTool) Description() string {
	return "GitHub Spec Kit for spec-driven development. Commands: init, constitution, specify, plan, tasks, implement"
}

// CanHandle checks if this tool can handle the intent
func (s *SpecKitTool) CanHandle(intent string) bool {
	lower := strings.ToLower(intent)
	keywords := []string{"spec", "plan", "task", "constitution", "implement", "specify"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// Execute runs the appropriate spec-kit command
func (s *SpecKitTool) Execute(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)

	// Add timeout to context if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	switch {
	case strings.Contains(lower, "init"):
		return s.executeInit(ctx, input)
	case strings.Contains(lower, "constitution"):
		return s.executeCommand(ctx, "/speckit.constitution")
	case strings.Contains(lower, "specify"):
		return s.executeCommand(ctx, "/speckit.specify")
	case strings.Contains(lower, "plan"):
		return s.executeCommand(ctx, "/speckit.plan")
	case strings.Contains(lower, "tasks"):
		return s.executeCommand(ctx, "/speckit.tasks")
	case strings.Contains(lower, "implement"):
		return s.executeCommand(ctx, "/speckit.implement")
	default:
		return "", fmt.Errorf("unknown spec-kit command in input: %s", input)
	}
}

func (s *SpecKitTool) executeInit(ctx context.Context, input string) (string, error) {
	// Extract project name from input
	projectName := extractArg(input, "init")
	if projectName == "" {
		return "", fmt.Errorf("project name not found in input")
	}

	cmd := exec.CommandContext(ctx, "specify", "init", projectName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", s.timeout)
		}
		return "", fmt.Errorf("failed to execute specify init: %w\n%s", err, output)
	}

	return string(output), nil
}

func (s *SpecKitTool) executeCommand(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "specify", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", s.timeout)
		}
		// If specify command doesn't exist, provide helpful message
		if strings.Contains(err.Error(), "executable file not found") {
			return fmt.Sprintf("[SpecKit] Command '%s' would execute here.\nNote: 'specify' CLI not found in PATH. Install it or use manual spec-driven workflow.", command), nil
		}
		return "", fmt.Errorf("failed to execute %s: %w\n%s", command, err, output)
	}

	return string(output), nil
}

// extractArg extracts the argument following a keyword
func extractArg(input, keyword string) string {
	parts := strings.Fields(input)
	for i, part := range parts {
		if strings.EqualFold(part, keyword) && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
