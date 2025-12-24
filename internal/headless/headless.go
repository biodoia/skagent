package headless

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/biodoia/skagent/internal/agents"
	"github.com/biodoia/skagent/internal/config"
	"github.com/biodoia/skagent/internal/core"
	"github.com/biodoia/skagent/internal/server/mcp"
	"github.com/biodoia/skagent/internal/server/rest"
)

type HeadlessMode struct {
	engine       *core.Engine
	agentRegistry *agents.Registry
	mcpServer    *mcp.Server
	restServer   *rest.APIServer
	config       *config.Config
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	logger       *log.Logger
}

type Command struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Command     string                 `json:"command"`
	Params      map[string]interface{} `json:"params"`
	AgentID     string                 `json:"agent_id,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	CallbackURL string                 `json:"callback_url,omitempty"`
}

type CommandResult struct {
	ID        string                 `json:"id"`
	Status    string                 `json:"status"`
	Result    map[string]interface{} `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

func NewHeadless(configPath string) (*HeadlessMode, error) {
	// Load configuration
	config, err := loadHeadlessConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	// Set up context
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create logger
	logger := log.New(os.Stdout, "[HEADLESS] ", log.LstdFlags|log.Lmsgprefix)
	
	// Initialize agent registry
	agentRegistry := agents.NewRegistry(ctx)
	
	// Initialize core components
	engine, err := core.NewEngine(ctx, config, agentRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}
	
	// Initialize servers
	mcpServer := mcp.NewServer(ctx, agentRegistry)
	restServer := rest.NewServer(ctx, config.API.Port, config.API.Host, engine, agentRegistry)
	
	return &HeadlessMode{
		engine:        engine,
		agentRegistry: agentRegistry,
		mcpServer:     mcpServer,
		restServer:    restServer,
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
	}, nil
}

func (h *HeadlessMode) Start() error {
	h.logger.Printf("Starting SKAgent in headless mode on %s:%d", h.config.API.Host, h.config.API.Port)
	
	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Create PID file if configured
	if h.config.Headless.PidFile != "" {
		if err := h.createPidFile(); err != nil {
			return fmt.Errorf("failed to create PID file: %w", err)
		}
		defer h.removePidFile()
	}
	
	// Set runtime options
	if h.config.Headless.Profile {
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(1)
	}
	
	if h.config.Headless.MaxProcs > 0 {
		runtime.GOMAXPROCS(h.config.Headless.MaxProcs)
	}
	
	// Start core engine
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		if err := h.engine.Start(); err != nil {
			h.logger.Printf("Engine error: %v", err)
		}
	}()
	
	// Start MCP server
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		if err := h.mcpServer.Start(); err != nil {
			h.logger.Printf("MCP server error: %v", err)
		}
	}()
	
	// Start REST API server
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		if err := h.restServer.Start(); err != nil {
			h.logger.Printf("REST server error: %v", err)
		}
	}()
	
	// Auto-start agents if configured
	if h.config.Headless.AutoStart {
		h.startDefaultAgents()
	}
	
	h.logger.Println("Headless mode started successfully")
	
	// Wait for shutdown signal
	<-sigChan
	
	h.logger.Println("Received shutdown signal, stopping services...")
	h.Stop()
	
	return nil
}

func (h *HeadlessMode) Stop() error {
	h.logger.Println("Stopping headless mode...")
	
	h.cancel()
	
	// Stop servers
	if h.restServer != nil {
		h.restServer.Stop()
	}
	
	if h.mcpServer != nil {
		h.mcpServer.Stop()
	}
	
	// Stop engine
	if h.engine != nil {
		h.engine.Stop()
	}
	
	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		defer close(done)
		h.wg.Wait()
	}()
	
	select {
	case <-done:
		h.logger.Println("All services stopped successfully")
	case <-time.After(30 * time.Second):
		h.logger.Println("Timeout waiting for services to stop")
	}
	
	return nil
}

func (h *HeadlessMode) ExecuteCommand(cmd Command) CommandResult {
	startTime := time.Now()
	
	result := CommandResult{
		ID:        cmd.ID,
		Timestamp: startTime,
	}
	
	// Set timeout
	timeout := cmd.Timeout
	if timeout == 0 {
		timeout = time.Duration(h.config.Headless.Timeout) * time.Second
	}
	
	ctx, cancel := context.WithTimeout(h.ctx, timeout)
	defer cancel()
	
	switch cmd.Type {
	case "agent":
		result = h.executeAgentCommand(ctx, cmd)
	case "tool":
		result = h.executeToolCommand(ctx, cmd)
	case "system":
		result = h.executeSystemCommand(ctx, cmd)
	default:
		result.Status = "error"
		result.Error = fmt.Sprintf("unknown command type: %s", cmd.Type)
	}
	
	result.Duration = time.Since(startTime)
	return result
}

func (h *HeadlessMode) executeAgentCommand(ctx context.Context, cmd Command) CommandResult {
	agentID := cmd.AgentID
	if agentID == "" {
		agentID = "default"
	}
	
	switch cmd.Command {
	case "list":
		agents := h.agentRegistry.ListAgents()
		return CommandResult{
			ID:        cmd.ID,
			Status:    "success",
			Result:    map[string]interface{}{"agents": agents},
			Timestamp: time.Now(),
		}
		
	case "start":
		if err := h.agentRegistry.StartAgent(agentID); err != nil {
			return CommandResult{
				ID:        cmd.ID,
				Status:    "error",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
		return CommandResult{
			ID:        cmd.ID,
			Status:    "success",
			Result:    map[string]interface{}{"message": fmt.Sprintf("Agent %s started", agentID)},
			Timestamp: time.Now(),
		}
		
	case "stop":
		if err := h.agentRegistry.StopAgent(agentID); err != nil {
			return CommandResult{
				ID:        cmd.ID,
				Status:    "error",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
		return CommandResult{
			ID:        cmd.ID,
			Status:    "success",
			Result:    map[string]interface{}{"message": fmt.Sprintf("Agent %s stopped", agentID)},
			Timestamp: time.Now(),
		}
		
	default:
		return CommandResult{
			ID:        cmd.ID,
			Status:    "error",
			Error:     fmt.Sprintf("unknown agent command: %s", cmd.Command),
			Timestamp: time.Now(),
		}
	}
}

func (h *HeadlessMode) executeToolCommand(ctx context.Context, cmd Command) CommandResult {
	// Tool execution would be implemented here
	// This is a simplified version
	return CommandResult{
		ID:        cmd.ID,
		Status:    "success",
		Result:    map[string]interface{}{"message": "Tool execution simulated"},
		Timestamp: time.Now(),
	}
}

func (h *HeadlessMode) executeSystemCommand(ctx context.Context, cmd Command) CommandResult {
	switch cmd.Command {
	case "status":
		return h.getSystemStatus()
	case "config":
		return h.getSystemConfig()
	case "health":
		return h.healthCheck()
	default:
		return CommandResult{
			ID:        cmd.ID,
			Status:    "error",
			Error:     fmt.Sprintf("unknown system command: %s", cmd.Command),
			Timestamp: time.Now(),
		}
	}
}

func (h *HeadlessMode) getSystemStatus() CommandResult {
	status := map[string]interface{}{
		"uptime":        time.Since(time.Now()), // This would be actual uptime
		"agents":        h.agentRegistry.GetStats(),
		"rest_server":   h.restServer.GetStatus(),
		"mcp_server":    h.mcpServer.GetStatus(),
		"engine":        h.engine.GetStatus(),
		"memory_usage":  getMemoryUsage(),
		"cpu_usage":     getCPUUsage(),
	}
	
	return CommandResult{
		ID:        "status",
		Status:    "success",
		Result:    status,
		Timestamp: time.Now(),
	}
}

func (h *HeadlessMode) getSystemConfig() CommandResult {
	config := map[string]interface{}{
		"host":         h.config.API.Host,
		"port":         h.config.API.Port,
		"max_agents":   h.config.Headless.MaxAgents,
		"timeout":      h.config.Headless.Timeout,
		"log_level":    h.config.Headless.LogLevel,
		"auto_start":   h.config.Headless.AutoStart,
		"daemon":       h.config.Headless.Enabled,
	}
	
	return CommandResult{
		ID:        "config",
		Status:    "success",
		Result:    config,
		Timestamp: time.Now(),
	}
}

func (h *HeadlessMode) healthCheck() CommandResult {
	health := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now(),
		"version":    "2.0.0",
		"components": map[string]interface{}{
			"engine":      h.engine.IsHealthy(),
			"rest_server": h.restServer.IsHealthy(),
			"mcp_server":  h.mcpServer.IsHealthy(),
		},
	}
	
	return CommandResult{
		ID:        "health",
		Status:    "success",
		Result:    health,
		Timestamp: time.Now(),
	}
}

func (h *HeadlessMode) startDefaultAgents() {
	// Start default agents based on configuration
	h.logger.Println("Starting default agents...")
	
	// This would initialize default agents based on config
	// For now, it's a placeholder
}

func (h *HeadlessMode) createPidFile() error {
	dir := filepath.Dir(h.config.Headless.PidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	pid := os.Getpid()
	return os.WriteFile(h.config.Headless.PidFile, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

func (h *HeadlessMode) removePidFile() {
	if h.config.Headless.PidFile != "" {
		os.Remove(h.config.Headless.PidFile)
	}
}

func loadHeadlessConfig(configPath string) (*config.Config, error) {
	cfg := config.DefaultConfig()
	
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}
	
	// Try to load from file
	if data, err := os.ReadFile(configPath); err == nil {
		// Parse JSON config
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}
	
	return cfg, nil
}

func getDefaultConfigPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "skagent", "headless.json")
	}
	return "headless.json"
}

func getMemoryUsage() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"alloc":       m.Alloc,
		"sys":         m.Sys,
		"heap_alloc":  m.HeapAlloc,
		"heap_sys":    m.HeapSys,
		"num_gc":      m.NumGC,
		"pause_total": m.PauseTotalNs,
	}
}

func getCPUUsage() map[string]interface{} {
	// This would implement actual CPU usage tracking
	return map[string]interface{}{
		"goroutines": runtime.NumGoroutine(),
		"cgo_calls":  runtime.NumCgoCall(),
	}
}

// Utility functions for CLI integration
func RunHeadless(configPath string, daemon bool) error {
	mode, err := NewHeadless(configPath)
	if err != nil {
		return err
	}
	
	if daemon {
		// Daemon mode would detach from terminal
		return mode.Start()
	}
	
	// Interactive headless mode
	return runInteractiveHeadless(mode)
}

func runInteractiveHeadless(mode *HeadlessMode) error {
	fmt.Println("Headless mode interactive shell. Type 'help' for commands.")
	
	for {
		fmt.Print("skagent> ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}
		
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		
		if input == "quit" || input == "exit" {
			break
		}
		
		// Process command
		cmd := Command{
			ID:          fmt.Sprintf("cmd-%d", time.Now().Unix()),
			Type:        "system",
			Command:     input,
			Timeout:     10 * time.Second,
		}
		
		result := mode.ExecuteCommand(cmd)
		fmt.Printf("Result: %+v\n", result)
	}
	
	return mode.Stop()
}