package config

// configTemplate is the default configuration template written by 'sow config init'.
// This template is shared between the init and edit commands.
var configTemplate = `# Sow Agent Configuration
# Location: ~/.config/sow/config.yaml
#
# This file configures which AI CLI tools handle which agent roles.
# If this file doesn't exist, all agents use Claude Code by default.
#
# Configuration priority:
#   1. Environment variables (SOW_AGENTS_IMPLEMENTER=cursor)
#   2. This config file
#   3. Built-in defaults (Claude Code)

agents:
  # Executor definitions
  executors:
    claude-code:
      type: "claude"
      settings:
        yolo_mode: false    # Set true to skip permission prompts
        # model: "sonnet"   # or "opus", "haiku"

    # Uncomment to enable Cursor
    # cursor:
    #   type: "cursor"
    #   settings:
    #     yolo_mode: false

    # Uncomment to enable Windsurf
    # windsurf:
    #   type: "windsurf"
    #   settings:
    #     yolo_mode: false

  # Bindings: which executor handles which agent role
  bindings:
    orchestrator: "claude-code"
    implementer: "claude-code"
    architect: "claude-code"
    reviewer: "claude-code"
    planner: "claude-code"
    researcher: "claude-code"
    decomposer: "claude-code"
`
