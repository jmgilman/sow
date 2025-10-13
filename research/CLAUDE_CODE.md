# Claude Code Feature Research

Research conducted: 2025-10-12

This document summarizes the capabilities available in Claude Code for building agentic workflows.

## Documentation Links

- [Plugins](https://docs.claude.com/en/docs/claude-code/plugins)
- [Sub-agents](https://docs.claude.com/en/docs/claude-code/sub-agents)
- [Hooks Guide](https://docs.claude.com/en/docs/claude-code/hooks-guide)
- [Model Context Protocol (MCP)](https://docs.claude.com/en/docs/claude-code/mcp)
- [Slash Commands](https://docs.claude.com/en/docs/claude-code/slash-commands)
- [Agent SDK](https://docs.claude.com/en/api/agent-sdk/overview)

---

## 1. Plugins

**Purpose**: Extensible tool for sharing custom commands, agents, hooks, and MCP servers across projects and teams.

### Capabilities
- Create custom slash commands
- Define specialized agents
- Add event handling hooks
- Integrate external tools via MCP
- Automate workflows
- Share functionality across teams

### Structure
```
my-plugin/
├── .claude-plugin/
│   └── plugin.json     # Metadata (name, description, version, author)
├── commands/           # Custom slash commands
├── agents/             # Custom agents
└── hooks/              # Event handlers
```

### Installation Methods
- Interactive menu
- Direct CLI commands
- Repository-level configuration
- Marketplace discovery

### Key Features
- Semantic versioning support
- Local marketplace for development
- Can be distributed via marketplaces

### Requirements
- Must follow specific structural guidelines
- Requires Claude Code installation

---

## 2. Sub-agents

**Purpose**: Specialized AI assistants with focused expertise that operate in separate context windows.

### Capabilities
- Specific domain expertise
- Separate context preservation
- Custom system prompts
- Configurable tool permissions
- Reusable across projects
- Can choose specific models (inherit, sonnet, opus, haiku)

### Configuration Location
`.claude/agents/` directory

### Configuration Format (YAML)
```yaml
---
name: code-reviewer
description: "Expert code review specialist"
tools: Read, Grep, Glob, Bash
model: inherit
---
System prompt goes here...
```

### Creation Methods
1. Interactive `/agents` command
2. Manual file creation in `.claude/agents/`
3. CLI configuration with `--agents` flag

### Invocation
- Automatic delegation based on task context
- Explicit invocation via user request

### Performance Notes
- Helps preserve main conversation context
- May add slight latency during context gathering

---

## 3. Hooks

**Purpose**: User-defined shell commands that execute at specific points in Claude Code's workflow.

### Hook Events
1. **PreToolUse** - Before tool calls
2. **PostToolUse** - After tool calls
3. **UserPromptSubmit** - When user submits a prompt
4. **Notification** - On notifications
5. **Stop** - When session stops
6. **SubagentStop** - When subagent stops
7. **PreCompact** - Before context compaction
8. **SessionStart** - At session start
9. **SessionEnd** - At session end

### Capabilities
- Block or modify tool calls
- Customize notifications
- Automatically format code
- Implement custom permissions
- Log and track commands
- Provide automated feedback

### Configuration Location
`hooks.json`

### Configuration Process
1. Use `/hooks` slash command
2. Select hook event
3. Add matcher (e.g., "Bash" or "*")
4. Define hook command
5. Choose storage location (user or project settings)

### Use Cases
- Auto-format files after editing
- Fix markdown formatting
- Send desktop notifications
- Protect sensitive files
- Log shell commands

### Security Considerations
- Runs with current environment credentials
- Potential for data exfiltration
- Always review implementations before registering

---

## 4. Model Context Protocol (MCP)

**Purpose**: Open-source standard for AI-tool integrations with external tools, databases, and APIs.

### Capabilities
- Connect to hundreds of tools and services
- Implement features from issue trackers
- Analyze monitoring data
- Query databases
- Integrate designs
- Automate workflows

### Server Types
1. Remote HTTP servers
2. Remote SSE servers (deprecated)
3. Local stdio servers

### Configuration Scopes
- **Local**: Project-specific, private configuration
- **Project**: Shared team configuration
- **User**: Available across multiple projects

### Configuration Location
`mcp.json`

### Installation Commands
```bash
# Add HTTP server
claude mcp add --transport http <name> <url>

# Add stdio server
claude mcp add --transport stdio <name> <command>
```

### Authentication
- Supports OAuth 2.0
- Can authenticate via `/mcp` command

### Features
- Environment variable expansion
- Resource referencing with @ mentions
- Slash commands for quick actions
- Enterprise configuration options

### Limitations
- Output warnings for large tool responses (default 10,000 tokens)
- Requires trust management

---

## 5. Slash Commands

**Purpose**: Interactive commands that control Claude's behavior during sessions.

### Types
1. **Built-in Commands** - Predefined (`/clear`, `/help`, `/review`, `/model`, `/config`)
2. **Custom Commands** - User-defined in markdown files
3. **Plugin Commands** - Distributed through plugins, can be namespaced

### Configuration Location
`.claude/commands/` directory

### File Format
Markdown files with optional frontmatter

### Example with Frontmatter
```markdown
---
allowed-tools: Bash(git add:*), Bash(git commit:*)
description: Create a git commit
---
Create a git commit with message: $ARGUMENTS
```

### Capabilities
- Support argument passing
- Execute bash commands
- Reference files
- Include frontmatter configuration
- Support thinking mode and extended reasoning

### Configuration Options (Frontmatter)
- Specify allowed tools
- Set argument hints
- Define description
- Choose specific model
- Control model invocation

### Organization
- Can use subdirectories for organization
- Command names derived from file paths

---

## 6. Agent SDK

**Purpose**: Development toolkit for building custom AI agents powered by Claude.

### Languages Supported
- TypeScript
- Python

### Installation
```bash
npm install @anthropic-ai/claude-agent-sdk
```

### Key Capabilities
- Automatic context management
- Rich tool ecosystem (file operations, code execution, web search)
- Advanced permissions control
- Production-ready features (error handling, session management)
- Optimized Claude model integration
- Automatic compaction and context management
- Fine-grained tool permissions

### Integration with Claude Code
- Supports file system configurations
- Enables subagents, custom hooks, and slash commands
- Supports MCP for external service integration

### Authentication Options
- Claude API key
- Amazon Bedrock
- Google Vertex AI

### Use Cases
- SRE diagnostic tools
- Security review bots
- Oncall engineering assistants
- Code review agents
- Legal contract reviewers
- Finance analysis assistants
- Customer support agents
- Content creation tools

---

## Summary of File Locations

Claude Code uses the `.claude/` directory for configuration:

```
.claude/
├── .claude-plugin/
│   └── plugin.json          # Plugin metadata
├── agents/
│   └── *.yaml               # Sub-agent definitions
├── commands/
│   └── *.md                 # Slash command definitions
├── hooks.json               # Hook configurations
└── mcp.json                 # MCP server configurations
```

All features are designed to work together and can be distributed as a cohesive plugin package.
