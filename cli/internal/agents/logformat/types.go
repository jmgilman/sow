// Package logformat provides parsing and formatting for Claude Code stream-json output.
//
// Claude Code's --output-format stream-json produces newline-delimited JSON events
// that describe the agent's actions. This package parses those events and formats
// them into human-readable log lines for easier comprehension.
//
// Event types:
//   - system (init): Session metadata (model, session_id, cwd, tools)
//   - assistant: Model output (text thinking, tool_use calls)
//   - user: Tool results (stdout, stderr, errors)
//   - result: Final summary (duration, cost, turns, outcome)
package logformat

import (
	"encoding/json"
	"fmt"
)

// EventType represents the type of stream-json event.
type EventType string

// Event type constants for stream-json events.
const (
	EventTypeSystem    EventType = "system"
	EventTypeAssistant EventType = "assistant"
	EventTypeUser      EventType = "user"
	EventTypeResult    EventType = "result"
)

// Event represents a parsed stream-json event.
// This is the common structure across all event types.
type Event struct {
	Type      EventType `json:"type"`
	Subtype   string    `json:"subtype,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	UUID      string    `json:"uuid,omitempty"`

	// For system events
	CWD              string   `json:"cwd,omitempty"`
	Model            string   `json:"model,omitempty"`
	PermissionMode   string   `json:"permissionMode,omitempty"`
	ClaudeCodeVersion string  `json:"claude_code_version,omitempty"`
	Tools            []string `json:"tools,omitempty"`

	// For assistant/user events
	Message *Message `json:"message,omitempty"`

	// For user events (tool results)
	ToolUseResult *ToolUseResult `json:"tool_use_result,omitempty"`

	// For result events
	IsError       bool    `json:"is_error,omitempty"`
	DurationMS    int64   `json:"duration_ms,omitempty"`
	DurationAPIMS int64   `json:"duration_api_ms,omitempty"`
	NumTurns      int     `json:"num_turns,omitempty"`
	Result        string  `json:"result,omitempty"`
	TotalCostUSD  float64 `json:"total_cost_usd,omitempty"`
}

// Message represents the message content in assistant/user events.
type Message struct {
	Role    string          `json:"role,omitempty"`
	Content json.RawMessage `json:"content,omitempty"` // Can be string or array
	Model   string          `json:"model,omitempty"`
	ID      string          `json:"id,omitempty"`
	Usage   *Usage          `json:"usage,omitempty"`
}

// ContentBlock represents a single content block in a message.
type ContentBlock struct {
	Type string `json:"type"` // "text" or "tool_use" or "tool_result"

	// For text blocks
	Text string `json:"text,omitempty"`

	// For tool_use blocks
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// For tool_result blocks
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// ToolInput represents common tool input structures.
type ToolInput struct {
	// Bash tool
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`

	// Read tool
	FilePath string `json:"file_path,omitempty"`

	// Edit tool
	OldString string `json:"old_string,omitempty"`
	NewString string `json:"new_string,omitempty"`

	// Write tool
	Content string `json:"content,omitempty"`

	// Glob/Grep tools
	Pattern string `json:"pattern,omitempty"`
	Path    string `json:"path,omitempty"`

	// TodoWrite tool
	Todos []TodoItem `json:"todos,omitempty"`
}

// TodoItem represents a todo item from the TodoWrite tool.
type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"activeForm"`
}

// ToolUseResult contains the result of a tool execution.
type ToolUseResult struct {
	Stdout      string `json:"stdout,omitempty"`
	Stderr      string `json:"stderr,omitempty"`
	Interrupted bool   `json:"interrupted,omitempty"`
	IsImage     bool   `json:"isImage,omitempty"`
}

// Usage contains token usage information.
type Usage struct {
	InputTokens              int `json:"input_tokens,omitempty"`
	OutputTokens             int `json:"output_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

// ParseEvent parses a single JSON line into an Event.
func ParseEvent(line []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(line, &event); err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}
	return &event, nil
}

// ParseContentBlocks parses the message content into ContentBlocks.
// The content can be a string or an array of blocks.
func (m *Message) ParseContentBlocks() ([]ContentBlock, error) {
	if m == nil || len(m.Content) == 0 {
		return nil, nil
	}

	// Try parsing as array first (most common)
	var blocks []ContentBlock
	if err := json.Unmarshal(m.Content, &blocks); err == nil {
		return blocks, nil
	}

	// Try parsing as string (less common)
	var text string
	if err := json.Unmarshal(m.Content, &text); err == nil {
		return []ContentBlock{{Type: "text", Text: text}}, nil
	}

	return nil, nil
}

// ParseToolInput parses the input JSON into a ToolInput struct.
func (cb *ContentBlock) ParseToolInput() (*ToolInput, error) {
	if len(cb.Input) == 0 {
		return nil, nil
	}
	var input ToolInput
	if err := json.Unmarshal(cb.Input, &input); err != nil {
		return nil, fmt.Errorf("failed to parse tool input: %w", err)
	}
	return &input, nil
}
