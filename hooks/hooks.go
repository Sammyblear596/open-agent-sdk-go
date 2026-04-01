package hooks

import (
	"context"
	"fmt"
	"strings"
)

// HookEvent represents when a hook fires.
type HookEvent string

const (
	HookPreToolUse   HookEvent = "PreToolUse"
	HookPostToolUse  HookEvent = "PostToolUse"
	HookPostSampling HookEvent = "PostSampling"
	HookStop         HookEvent = "Stop"
)

// HookFn is a function that runs as a hook.
// Returns an error message to block the action, or empty string to allow.
type HookFn func(ctx context.Context, toolName string, input map[string]interface{}) (string, error)

// HookRule defines when a hook should fire.
type HookRule struct {
	// Matcher is a tool name pattern (e.g., "Bash", "Edit|Write", "*")
	Matcher string `json:"matcher"`
	// Hooks are the functions to run
	Hooks []HookFn `json:"-"`
}

// HookConfig holds all hook definitions.
type HookConfig struct {
	PreToolUse   []HookRule `json:"PreToolUse,omitempty"`
	PostToolUse  []HookRule `json:"PostToolUse,omitempty"`
	PostSampling []HookRule `json:"PostSampling,omitempty"`
	Stop         []HookRule `json:"Stop,omitempty"`
}

// HookProgress represents progress from a hook execution.
type HookProgress struct {
	Event         HookEvent `json:"event"`
	HookName      string    `json:"hookName"`
	ToolName      string    `json:"toolName"`
	StatusMessage string    `json:"statusMessage,omitempty"`
	Blocked       bool      `json:"blocked,omitempty"`
}

// HookResult is the result of running hooks.
type HookResult struct {
	Blocked  bool           `json:"blocked"`
	Message  string         `json:"message,omitempty"`
	Progress []HookProgress `json:"progress,omitempty"`
}

// Manager handles hook execution.
type Manager struct {
	config HookConfig
}

// NewManager creates a new hook manager.
func NewManager(config HookConfig) *Manager {
	return &Manager{config: config}
}

// RunPreToolUse runs pre-tool-use hooks. Returns result with block status.
func (m *Manager) RunPreToolUse(ctx context.Context, toolName string, input map[string]interface{}) (*HookResult, error) {
	return m.runHooks(ctx, HookPreToolUse, m.config.PreToolUse, toolName, input)
}

// RunPostToolUse runs post-tool-use hooks.
func (m *Manager) RunPostToolUse(ctx context.Context, toolName string, input map[string]interface{}) (*HookResult, error) {
	return m.runHooks(ctx, HookPostToolUse, m.config.PostToolUse, toolName, input)
}

// RunPostSampling runs post-sampling hooks after API response.
func (m *Manager) RunPostSampling(ctx context.Context, toolName string, input map[string]interface{}) (*HookResult, error) {
	return m.runHooks(ctx, HookPostSampling, m.config.PostSampling, toolName, input)
}

// RunStop runs stop hooks at end of conversation.
func (m *Manager) RunStop(ctx context.Context, toolName string, input map[string]interface{}) (*HookResult, error) {
	return m.runHooks(ctx, HookStop, m.config.Stop, toolName, input)
}

func (m *Manager) runHooks(ctx context.Context, event HookEvent, rules []HookRule, toolName string, input map[string]interface{}) (*HookResult, error) {
	result := &HookResult{}

	for _, rule := range rules {
		if !matchesTool(rule.Matcher, toolName) {
			continue
		}
		for i, hook := range rule.Hooks {
			progress := HookProgress{
				Event:    event,
				HookName: fmt.Sprintf("%s_hook_%d", rule.Matcher, i),
				ToolName: toolName,
			}

			msg, err := hook(ctx, toolName, input)
			if err != nil {
				return nil, fmt.Errorf("hook error (%s): %w", progress.HookName, err)
			}
			if msg != "" {
				progress.Blocked = true
				progress.StatusMessage = msg
				result.Blocked = true
				result.Message = msg
			}
			result.Progress = append(result.Progress, progress)

			if result.Blocked {
				return result, nil
			}
		}
	}
	return result, nil
}

// HasHooks returns true if any hooks are configured.
func (m *Manager) HasHooks() bool {
	return len(m.config.PreToolUse) > 0 ||
		len(m.config.PostToolUse) > 0 ||
		len(m.config.PostSampling) > 0 ||
		len(m.config.Stop) > 0
}

// matchesTool checks if a matcher pattern matches a tool name.
func matchesTool(matcher, toolName string) bool {
	if matcher == "*" {
		return true
	}

	parts := strings.Split(matcher, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == toolName {
			return true
		}
		if strings.Contains(part, "*") {
			if strings.HasPrefix(part, "*") && strings.HasSuffix(toolName, strings.TrimPrefix(part, "*")) {
				return true
			}
			if strings.HasSuffix(part, "*") && strings.HasPrefix(toolName, strings.TrimSuffix(part, "*")) {
				return true
			}
		}
		// MCP prefix matching: "mcp__server" matches "mcp__server__tool"
		if strings.HasPrefix(part, "mcp__") && strings.HasPrefix(toolName, part+"__") {
			return true
		}
	}

	return false
}
