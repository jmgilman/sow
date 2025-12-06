package logformat

import (
	"fmt"
	"strings"
	"time"
)

const (
	// MaxContentLength is the maximum length for tool output content.
	// Longer content is truncated with an indicator.
	MaxContentLength = 1000

	// MaxThinkingLength is the maximum length for thinking/text blocks.
	MaxThinkingLength = 500

	// LineWidth is the width of separator lines.
	LineWidth = 80
)

// Formatter formats stream-json events into human-readable output.
type Formatter struct {
	// MaxContentLen controls truncation of tool output (0 = use default).
	MaxContentLen int

	// MaxThinkingLen controls truncation of thinking text (0 = use default).
	MaxThinkingLen int
}

// NewFormatter creates a Formatter with default settings.
func NewFormatter() *Formatter {
	return &Formatter{
		MaxContentLen:  MaxContentLength,
		MaxThinkingLen: MaxThinkingLength,
	}
}

// Format formats a single event into human-readable lines.
// Returns empty string for events that don't need to be shown.
func (f *Formatter) Format(event *Event) string {
	switch event.Type {
	case EventTypeSystem:
		if event.Subtype == "init" {
			return f.formatInit(event)
		}
		return ""
	case EventTypeAssistant:
		return f.formatAssistant(event)
	case EventTypeUser:
		return f.formatToolResult(event)
	case EventTypeResult:
		return f.formatResult(event)
	default:
		return ""
	}
}

// formatInit formats the session initialization event.
func (f *Formatter) formatInit(event *Event) string {
	var b strings.Builder

	b.WriteString(doubleLine())
	b.WriteString("SESSION STARTED\n")
	b.WriteString(doubleLine())
	b.WriteString(fmt.Sprintf("  Session ID: %s\n", event.SessionID))
	b.WriteString(fmt.Sprintf("  Model: %s\n", event.Model))
	if event.CWD != "" {
		b.WriteString(fmt.Sprintf("  Working Dir: %s\n", event.CWD))
	}
	if event.ClaudeCodeVersion != "" {
		b.WriteString(fmt.Sprintf("  Claude Code: v%s\n", event.ClaudeCodeVersion))
	}
	if event.PermissionMode != "" && event.PermissionMode != "default" {
		b.WriteString(fmt.Sprintf("  Permission Mode: %s\n", event.PermissionMode))
	}
	b.WriteString("\n")

	return b.String()
}

// formatAssistant formats assistant messages (thinking and tool calls).
func (f *Formatter) formatAssistant(event *Event) string {
	if event.Message == nil {
		return ""
	}

	blocks, err := event.Message.ParseContentBlocks()
	if err != nil || len(blocks) == 0 {
		return ""
	}

	var b strings.Builder

	for _, block := range blocks {
		switch block.Type {
		case "text":
			if text := strings.TrimSpace(block.Text); text != "" {
				b.WriteString(singleLine())
				b.WriteString("THINKING\n")
				b.WriteString(singleLine())
				b.WriteString(f.indent(f.truncate(text, f.maxThinking()), "  "))
				b.WriteString("\n\n")
			}
		case "tool_use":
			b.WriteString(f.formatToolCall(&block))
		}
	}

	return b.String()
}

// formatToolCall formats a tool invocation.
func (f *Formatter) formatToolCall(block *ContentBlock) string {
	var b strings.Builder

	b.WriteString(singleLine())
	b.WriteString(fmt.Sprintf("TOOL: %s\n", block.Name))
	b.WriteString(singleLine())

	input, _ := block.ParseToolInput()
	if input != nil {
		f.writeToolInput(&b, block.Name, input, block.Input)
	}
	b.WriteString("\n")

	return b.String()
}

// writeToolInput writes formatted tool input details to the builder.
func (f *Formatter) writeToolInput(b *strings.Builder, toolName string, input *ToolInput, rawInput []byte) {
	switch toolName {
	case "Bash":
		f.writeBashInput(b, input)
	case "Read":
		fmt.Fprintf(b, "  File: %s\n", input.FilePath)
	case "Write":
		f.writeWriteInput(b, input)
	case "Edit":
		f.writeEditInput(b, input)
	case "Glob", "Grep":
		f.writeSearchInput(b, input)
	case "TodoWrite":
		f.writeTodoInput(b, input)
	default:
		fmt.Fprintf(b, "  Input: %s\n", f.truncateLine(string(rawInput), 200))
	}
}

func (f *Formatter) writeBashInput(b *strings.Builder, input *ToolInput) {
	if input.Description != "" {
		fmt.Fprintf(b, "  Description: %s\n", input.Description)
	}
	if input.Command != "" {
		fmt.Fprintf(b, "  Command: %s\n", f.truncateLine(input.Command, 200))
	}
}

func (f *Formatter) writeWriteInput(b *strings.Builder, input *ToolInput) {
	fmt.Fprintf(b, "  File: %s\n", input.FilePath)
	if input.Content != "" {
		lines := strings.Count(input.Content, "\n") + 1
		fmt.Fprintf(b, "  Content: %d lines\n", lines)
	}
}

func (f *Formatter) writeEditInput(b *strings.Builder, input *ToolInput) {
	fmt.Fprintf(b, "  File: %s\n", input.FilePath)
	if input.OldString != "" {
		fmt.Fprintf(b, "  Replacing: %s\n", f.truncateLine(input.OldString, 100))
	}
}

func (f *Formatter) writeSearchInput(b *strings.Builder, input *ToolInput) {
	fmt.Fprintf(b, "  Pattern: %s\n", input.Pattern)
	if input.Path != "" {
		fmt.Fprintf(b, "  Path: %s\n", input.Path)
	}
}

func (f *Formatter) writeTodoInput(b *strings.Builder, input *ToolInput) {
	if len(input.Todos) == 0 {
		return
	}
	b.WriteString("  Todos:\n")
	for _, todo := range input.Todos {
		status := statusIcon(todo.Status)
		fmt.Fprintf(b, "    %s %s\n", status, todo.Content)
	}
}

// formatToolResult formats the result of a tool execution.
func (f *Formatter) formatToolResult(event *Event) string {
	if event.Message == nil {
		return ""
	}

	blocks, err := event.Message.ParseContentBlocks()
	if err != nil || len(blocks) == 0 {
		return ""
	}

	var b strings.Builder

	for _, block := range blocks {
		if block.Type != "tool_result" {
			continue
		}

		status := "OK"
		if block.IsError {
			status = "ERROR"
		}

		b.WriteString(singleLine())
		b.WriteString(fmt.Sprintf("RESULT [%s]\n", status))
		b.WriteString(singleLine())

		// Prefer tool_use_result if available (has structured stdout/stderr)
		if event.ToolUseResult != nil {
			if event.ToolUseResult.Stderr != "" {
				b.WriteString("  STDERR:\n")
				b.WriteString(f.indent(f.truncate(event.ToolUseResult.Stderr, f.maxContent()), "    "))
				b.WriteString("\n")
			}
			if event.ToolUseResult.Stdout != "" {
				output := f.truncate(event.ToolUseResult.Stdout, f.maxContent())
				b.WriteString(f.indent(output, "  "))
				b.WriteString("\n")
			}
			if event.ToolUseResult.Interrupted {
				b.WriteString("  [INTERRUPTED]\n")
			}
		} else if block.Content != "" {
			output := f.truncate(block.Content, f.maxContent())
			b.WriteString(f.indent(output, "  "))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatResult formats the final session result.
func (f *Formatter) formatResult(event *Event) string {
	var b strings.Builder

	b.WriteString(doubleLine())
	if event.IsError {
		b.WriteString("SESSION FAILED\n")
	} else {
		b.WriteString("SESSION COMPLETE\n")
	}
	b.WriteString(doubleLine())

	// Duration
	duration := formatDuration(event.DurationMS)
	apiDuration := formatDuration(event.DurationAPIMS)
	b.WriteString(fmt.Sprintf("  Duration: %s (API: %s)\n", duration, apiDuration))

	// Turns
	b.WriteString(fmt.Sprintf("  Turns: %d\n", event.NumTurns))

	// Cost
	if event.TotalCostUSD > 0 {
		b.WriteString(fmt.Sprintf("  Cost: $%.2f\n", event.TotalCostUSD))
	}

	b.WriteString("\n")

	// Final result
	if event.Result != "" {
		b.WriteString("  FINAL RESULT:\n")
		b.WriteString(f.indent(event.Result, "  "))
		b.WriteString("\n")
	}

	return b.String()
}

// Helper functions

func (f *Formatter) maxContent() int {
	if f.MaxContentLen > 0 {
		return f.MaxContentLen
	}
	return MaxContentLength
}

func (f *Formatter) maxThinking() int {
	if f.MaxThinkingLen > 0 {
		return f.MaxThinkingLen
	}
	return MaxThinkingLength
}

func (f *Formatter) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + fmt.Sprintf("\n  ... [truncated, %d total chars]", len(s))
}

func (f *Formatter) truncateLine(s string, maxLen int) string {
	// Replace newlines with spaces for single-line display
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (f *Formatter) indent(s string, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

func doubleLine() string {
	return strings.Repeat("=", LineWidth) + "\n"
}

func singleLine() string {
	return strings.Repeat("-", LineWidth) + "\n"
}

func statusIcon(status string) string {
	switch status {
	case "completed":
		return "[x]"
	case "in_progress":
		return "[>]"
	case "pending":
		return "[ ]"
	default:
		return "[?]"
	}
}

func formatDuration(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	if d < time.Second {
		return fmt.Sprintf("%dms", ms)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}
