# Hooks and Integrations

**Last Updated**: 2025-10-12
**Status**: Comprehensive Architecture Documentation

---

## Table of Contents

- [Overview](#overview)
- [Hooks System](#hooks-system)
  - [What Are Hooks?](#what-are-hooks)
  - [Available Hook Events](#available-hook-events)
  - [Hook Configuration](#hook-configuration)
  - [SessionStart Hook (MVP)](#sessionstart-hook-mvp)
  - [Other Hook Examples](#other-hook-examples)
  - [Hook Use Cases](#hook-use-cases)
- [MCP Integrations](#mcp-integrations)
  - [What is MCP?](#what-is-mcp)
  - [MCP Server Types](#mcp-server-types)
  - [Configuration Scopes](#configuration-scopes)
  - [Common Integrations](#common-integrations)
  - [Authentication](#authentication)
- [Security Considerations](#security-considerations)
- [Plugin Metadata](#plugin-metadata)
- [Creating Custom Hooks](#creating-custom-hooks)
- [Creating Custom MCP Integrations](#creating-custom-mcp-integrations)
- [Related Documentation](#related-documentation)

---

## Overview

`sow` supports two types of extensibility:

1. **Hooks** - User-defined shell commands that execute at specific points in Claude Code's workflow
2. **MCP Integrations** - Connections to external tools, databases, and APIs via Model Context Protocol

Both systems allow you to customize and extend `sow` behavior without modifying core code.

---

## Hooks System

### What Are Hooks?

Hooks are shell commands that run automatically at specific lifecycle events in Claude Code. They enable:

- **Automation** - Run tasks automatically (formatting, validation)
- **Notifications** - Alert external systems or users
- **Customization** - Modify behavior at key points
- **Logging** - Track actions and events
- **Permissions** - Implement custom access controls

**Configuration Location**: `hooks.json`

### Available Hook Events

`sow` uses hooks at various points in the workflow:

| Hook Event | When It Fires | Use Cases |
|------------|---------------|-----------|
| **SessionStart** | When Claude Code session starts | Provide context, check versions, show project status |
| **SessionEnd** | When Claude Code session ends | Save state, cleanup, notifications |
| **PreToolUse** | Before any tool is executed | Block dangerous operations, validate inputs, log commands |
| **PostToolUse** | After any tool completes | Auto-format code, validate outputs, sync state |
| **UserPromptSubmit** | When user submits a prompt | Log user requests, pre-process inputs |
| **Notification** | When Claude Code shows notification | Desktop notifications, external alerts |
| **Stop** | When execution stops | Save progress, cleanup resources |
| **SubagentStop** | When a subagent stops | Collect results, update parent context |
| **PreCompact** | Before context compaction | Save important context, log conversation summary |

### Hook Configuration

**File Format**: `hooks.json`

**Structure**:
```json
{
  "HookEventName": {
    "matcher": "pattern",
    "command": "shell command to execute"
  }
}
```

**Matcher Patterns**:
- `"*"` - Match all events
- `"Bash"` - Match specific tool (e.g., Bash tool calls)
- `"Edit"` - Match Edit tool calls
- `"Write"` - Match Write tool calls
- Regular expressions for complex matching

**Example Configuration**:
```json
{
  "SessionStart": {
    "matcher": "*",
    "command": "sow session-info"
  },
  "PostToolUse": {
    "matcher": "Edit",
    "command": "prettier --write $FILE"
  },
  "PreCompact": {
    "matcher": "*",
    "command": "sow save-context"
  }
}
```

### SessionStart Hook (MVP)

The `SessionStart` hook is the **primary hook for MVP**. It provides immediate context when Claude Code starts.

**Purpose**:
- Show current project status
- Detect version mismatches
- Remind users of available commands
- Orient the orchestrator

**Configuration** (`hooks.json`):
```json
{
  "SessionStart": {
    "matcher": "*",
    "command": "sow session-info"
  }
}
```

**CLI Implementation** (`sow session-info`):

```bash
#!/bin/bash

# Check if sow repository
if [ ! -d ".sow" ]; then
  echo "⚠️  Not a sow repository"
  echo "💡 Use /init to set up sow"
  exit 0
fi

# Read versions
STRUCT_VERSION=$(yq .sow_structure_version .sow/.version)
PLUGIN_VERSION=$(cat .claude/.plugin-version)

echo "📋 You are in a sow-enabled repository"

# Check for project
if [ -d ".sow/project" ]; then
  PROJECT_NAME=$(yq .project.name .sow/project/state.yaml)
  BRANCH=$(git branch --show-current)
  echo "🚀 Active project: $PROJECT_NAME (branch: $BRANCH)"
  echo "📂 Use /continue to resume work"
else
  echo "💡 No active project. Use /start-project <name> to begin"
fi

echo ""

# Version mismatch check
if [ "$STRUCT_VERSION" != "$PLUGIN_VERSION" ]; then
  echo "⚠️  Version mismatch detected!"
  echo "   Repository structure: $STRUCT_VERSION"
  echo "   Plugin version: $PLUGIN_VERSION"
  echo ""
  echo "💡 Run /migrate to upgrade your repository structure"
  echo "   Migration path: $STRUCT_VERSION → $PLUGIN_VERSION"
  echo "   Review changes: https://github.com/your-org/sow/blob/main/CHANGELOG.md"
  exit 0
fi

echo "✓ Versions aligned (v$PLUGIN_VERSION)"
echo ""
echo "📖 Available commands:"
echo "   /start-project <name> - Create new project"
echo "   /continue - Resume existing project"
echo "   /cleanup - Delete project before merge"
echo "   /sync - Update sinks and repos"
```

**Benefits**:
- Low overhead (just read files)
- Immediate context on session start
- Proactive version management
- Helpful command reminders

**Example Output** (with active project):
```
📋 You are in a sow-enabled repository
🚀 Active project: Add authentication (branch: feat/add-auth)
📂 Use /continue to resume work

✓ Versions aligned (v0.2.0)

📖 Available commands:
   /start-project <name> - Create new project
   /continue - Resume existing project
   /cleanup - Delete project before merge
   /sync - Update sinks and repos
```

**Example Output** (version mismatch):
```
📋 You are in a sow-enabled repository
💡 No active project. Use /start-project <name> to begin

⚠️  Version mismatch detected!
   Repository structure: 0.1.0
   Plugin version: 0.2.0

💡 Run /migrate to upgrade your repository structure
   Migration path: 0.1.0 → 0.2.0
   Review changes: https://github.com/your-org/sow/blob/main/CHANGELOG.md
```

### Other Hook Examples

#### PostToolUse - Auto-formatting

**Purpose**: Automatically format code after editing

**Configuration**:
```json
{
  "PostToolUse": {
    "matcher": "Edit",
    "command": "format-file.sh $FILE"
  }
}
```

**Script** (`format-file.sh`):
```bash
#!/bin/bash
FILE=$1

# Determine file type and format accordingly
case "$FILE" in
  *.py)
    black "$FILE"
    isort "$FILE"
    ;;
  *.js|*.ts|*.jsx|*.tsx)
    prettier --write "$FILE"
    ;;
  *.go)
    gofmt -w "$FILE"
    ;;
  *.md)
    prettier --write "$FILE"
    ;;
esac

echo "✓ Formatted $FILE"
```

**Benefits**:
- Consistent formatting
- No manual formatting needed
- Follows project conventions

#### PreToolUse - Permission Checking

**Purpose**: Prevent dangerous operations on sensitive files

**Configuration**:
```json
{
  "PreToolUse": {
    "matcher": "Write",
    "command": "check-permissions.sh $FILE"
  }
}
```

**Script** (`check-permissions.sh`):
```bash
#!/bin/bash
FILE=$1

# Block writes to sensitive files
SENSITIVE_FILES=(.env .env.production credentials.json secrets.yaml)

for PATTERN in "${SENSITIVE_FILES[@]}"; do
  if [[ "$FILE" == *"$PATTERN"* ]]; then
    echo "❌ Error: Cannot modify sensitive file: $FILE"
    exit 1
  fi
done

echo "✓ Permission granted for $FILE"
exit 0
```

**Benefits**:
- Protects sensitive data
- Prevents accidental leaks
- Custom security policies

#### Notification - Desktop Alerts

**Purpose**: Send desktop notification when work completes

**Configuration**:
```json
{
  "Notification": {
    "matcher": "*",
    "command": "notify-desktop.sh \"$MESSAGE\""
  }
}
```

**Script** (`notify-desktop.sh`):
```bash
#!/bin/bash
MESSAGE=$1

# macOS
if command -v osascript &> /dev/null; then
  osascript -e "display notification \"$MESSAGE\" with title \"sow\""
fi

# Linux
if command -v notify-send &> /dev/null; then
  notify-send "sow" "$MESSAGE"
fi
```

**Benefits**:
- Stay informed when away from terminal
- Multi-tasking friendly
- Cross-platform

#### PreCompact - Save Context

**Purpose**: Save important context before compaction

**Configuration**:
```json
{
  "PreCompact": {
    "matcher": "*",
    "command": "sow save-context"
  }
}
```

**CLI Implementation** (`sow save-context`):
```bash
#!/bin/bash

# Only save if project exists
if [ ! -d ".sow/project" ]; then
  exit 0
fi

# Extract and save key decisions from conversation
# (This would involve more sophisticated logic)
echo "💾 Saving context before compaction..."

# Log that compaction occurred
echo "Context compacted at $(date)" >> .sow/project/log.md

exit 0
```

**Benefits**:
- Preserves important context
- Tracks conversation compaction events
- Helps with debugging

### Hook Use Cases

#### Workflow Automation

- Auto-format code after editing
- Run linters after file changes
- Update indexes after adding sinks
- Regenerate documentation

#### Quality Gates

- Block commits without tests
- Enforce code review checklist
- Validate YAML/JSON syntax
- Check for sensitive data

#### Integration

- Send notifications to Slack
- Update issue tracker status
- Post metrics to monitoring
- Trigger CI/CD pipelines

#### Logging and Auditing

- Log all bash commands
- Track tool usage
- Record user prompts
- Audit file modifications

#### Context Management

- Save context before compaction
- Load context on session start
- Sync state with external systems
- Preserve conversation summaries

---

## MCP Integrations

### What is MCP?

**Model Context Protocol (MCP)** is an open-source standard for AI-tool integrations. It allows Claude Code to connect to:

- External tools and services
- Databases
- APIs
- Issue trackers
- Monitoring systems
- Documentation platforms

**Configuration Location**: `mcp.json`

**Key Benefit**: Extends agent capabilities beyond built-in tools

### MCP Server Types

#### 1. Remote HTTP Servers

Connect to HTTP-based MCP servers:

```bash
claude mcp add --transport http github-api https://mcp.github.com
```

**Use Cases**:
- Public APIs
- Managed services
- Cloud integrations

#### 2. Local Stdio Servers

Run MCP servers as local processes:

```bash
claude mcp add --transport stdio git-server "node /path/to/git-mcp-server/index.js"
```

**Use Cases**:
- Local tools
- Custom integrations
- Development/testing

#### 3. Remote SSE Servers (Deprecated)

Server-Sent Events transport (deprecated, use HTTP instead)

### Configuration Scopes

MCP servers can be configured at different levels:

| Scope | Location | Purpose |
|-------|----------|---------|
| **Local** | Project-specific, not committed | Personal, private integrations |
| **Project** | `.claude/mcp.json` (committed) | Shared team integrations |
| **User** | `~/.config/claude/mcp.json` | Available across all projects |

**Recommendation**: Use project-level configuration for team-shared integrations (committed to git).

### Common Integrations

#### GitHub Integration

**Purpose**: Read issues, create PRs, comment on reviews

**Setup**:
```bash
# Add GitHub MCP server
claude mcp add --transport http github https://mcp.github.com

# Authenticate
/mcp auth github
```

**Capabilities**:
- Read issue descriptions
- Create pull requests
- Add PR comments
- Check CI status
- List repository files

**Example Usage**:
```
Orchestrator: Creating PR for completed work...

Uses GitHub MCP to:
1. Create PR with title and description
2. Add reviewers
3. Link to related issues
4. Comment with project summary
```

#### Jira/Linear Integration

**Purpose**: Sync tasks, update issue status

**Setup**:
```bash
# Add Jira MCP server
claude mcp add --transport http jira https://mcp.jira.com

# Authenticate
/mcp auth jira
```

**Capabilities**:
- Read issue details
- Update issue status
- Add comments
- Create subtasks
- Link related issues

**Example Usage**:
```
Orchestrator: Syncing with Jira...

1. Read Jira issue PROJ-123
2. Create sow project from issue description
3. Update Jira status to "In Progress"
4. Post updates to Jira as work progresses
```

#### Monitoring/Observability

**Purpose**: Query logs, check alerts, analyze metrics

**Setup**:
```bash
# Add Datadog MCP server
claude mcp add --transport http datadog https://mcp.datadog.com
```

**Capabilities**:
- Query logs
- Check alert status
- Fetch metrics
- Analyze performance
- Investigate incidents

**Example Usage**:
```
Worker (in discovery phase):
- Query production logs for error patterns
- Check recent alert history
- Identify performance bottlenecks
- Report findings in design document
```

#### Documentation Platforms

**Purpose**: Search Confluence, Notion, internal wikis

**Setup**:
```bash
# Add Confluence MCP server
claude mcp add --transport http confluence https://mcp.confluence.com
```

**Capabilities**:
- Search documentation
- Read pages
- Extract context
- Link to references

**Example Usage**:
```
Architect agent:
- Searches Confluence for existing auth patterns
- Reads company security guidelines
- Incorporates findings into design document
```

### Authentication

MCP servers may require authentication:

**OAuth 2.0**:
```bash
# Authenticate via CLI
/mcp auth <server-name>

# Follow OAuth flow in browser
```

**API Keys**:
```json
{
  "mcpServers": {
    "github": {
      "transport": "http",
      "url": "https://mcp.github.com",
      "headers": {
        "Authorization": "Bearer ${GITHUB_TOKEN}"
      }
    }
  }
}
```

**Environment Variables**:
- Store credentials in environment variables
- Reference in `mcp.json` via `${VAR_NAME}`
- Never commit credentials to git

### MCP Configuration Example

**File**: `.claude/mcp.json`

```json
{
  "mcpServers": {
    "github": {
      "transport": "http",
      "url": "https://api.github.com/mcp",
      "headers": {
        "Authorization": "Bearer ${GITHUB_TOKEN}"
      }
    },
    "jira": {
      "transport": "http",
      "url": "https://your-domain.atlassian.net/mcp",
      "headers": {
        "Authorization": "Bearer ${JIRA_TOKEN}"
      }
    },
    "git-local": {
      "transport": "stdio",
      "command": "node",
      "args": ["/path/to/git-mcp-server/index.js"]
    }
  }
}
```

### Resource Referencing

MCP resources can be referenced with @ mentions:

```
@github/issues/123 - Reference GitHub issue
@confluence/page/456 - Reference Confluence page
@jira/PROJ-789 - Reference Jira ticket
```

**Example**:
```
User: "Implement @jira/PROJ-123"

Orchestrator:
- Fetches Jira issue details via MCP
- Extracts requirements
- Creates project with context from Jira
```

### Slash Commands via MCP

MCP servers can provide slash commands:

```
/github pr create - Create PR via GitHub MCP
/jira update PROJ-123 - Update Jira issue
/datadog logs query - Query Datadog logs
```

---

## Security Considerations

### Hooks Security

**Risk**: Hooks run with your environment credentials and can:
- Access sensitive files
- Exfiltrate data
- Execute arbitrary code
- Modify system state

**Best Practices**:

1. **Review Before Registering**:
   - Always review hook scripts before adding them
   - Understand what each command does
   - Verify sources of third-party hooks

2. **Principle of Least Privilege**:
   - Grant minimum required permissions
   - Use restricted shell environments when possible
   - Avoid running hooks as root

3. **Input Validation**:
   - Validate all inputs in hook scripts
   - Sanitize file paths
   - Escape shell arguments

4. **Audit Trail**:
   - Log hook executions
   - Track what commands ran
   - Monitor for suspicious activity

5. **Team Coordination**:
   - Commit hook configurations to git (project scope)
   - Review hook changes in PRs
   - Document hook purposes

### MCP Security

**Risk**: MCP servers can:
- Access external APIs with your credentials
- Read and write data
- Execute operations on your behalf

**Best Practices**:

1. **Authentication Management**:
   - Use environment variables for credentials
   - Never commit credentials to git
   - Rotate credentials regularly
   - Use OAuth when available

2. **Permission Scoping**:
   - Grant minimum required permissions
   - Use read-only tokens when possible
   - Limit API scope to necessary operations

3. **Trust Management**:
   - Only use MCP servers from trusted sources
   - Review MCP server code if available
   - Monitor MCP server activity
   - Set output limits (default: 10,000 tokens)

4. **Network Security**:
   - Use HTTPS for remote servers
   - Verify SSL certificates
   - Use VPN for internal services

5. **Data Privacy**:
   - Be aware of what data MCP servers can access
   - Avoid sending sensitive data to third-party servers
   - Review MCP server privacy policies

---

## Plugin Metadata

**File**: `.claude-plugin/plugin.json`

**Purpose**: Metadata for Claude Code plugin distribution

**Schema**:
```json
{
  "name": "sow",
  "version": "0.2.0",
  "description": "AI-powered system of work for software engineering",
  "author": {
    "name": "sow contributors",
    "email": "maintainers@example.com"
  },
  "homepage": "https://github.com/your-org/sow",
  "repository": "https://github.com/your-org/sow",
  "license": "MIT",
  "keywords": ["productivity", "workflow", "agents", "project-management"],
  "engines": {
    "claude-code": ">=1.0.0"
  }
}
```

**Fields**:

| Field | Required | Purpose |
|-------|----------|---------|
| `name` | Yes | Plugin identifier (kebab-case) |
| `version` | Yes | Semantic version (MAJOR.MINOR.PATCH) |
| `description` | Yes | Short description |
| `author` | Yes | Author information (name, email) |
| `homepage` | No | Plugin homepage URL |
| `repository` | No | Source code repository URL |
| `license` | No | License identifier (e.g., MIT, Apache-2.0) |
| `keywords` | No | Search keywords for marketplace |
| `engines` | No | Required Claude Code version |

**Distribution**:
- Plugin bundles all `.claude/` contents
- Distributed via Claude Code marketplace
- Users install via `/plugin install sow@marketplace`
- Automatic updates when new versions released

---

## Creating Custom Hooks

### Step 1: Identify Hook Point

Determine which hook event fits your need:
- Automation → PostToolUse
- Validation → PreToolUse
- Notifications → Notification
- Context → SessionStart, PreCompact

### Step 2: Write Hook Script

Create executable script:

```bash
#!/bin/bash
# File: .claude/hooks/my-hook.sh

# Access hook variables
FILE=$1
TOOL=$2
RESULT=$3

# Implement logic
echo "Running custom hook for $FILE"

# Exit with appropriate code
# 0 = success
# non-zero = failure (blocks operation for PreToolUse)
exit 0
```

### Step 3: Configure Hook

Add to `hooks.json`:

```json
{
  "PostToolUse": {
    "matcher": "Edit",
    "command": ".claude/hooks/my-hook.sh $FILE $TOOL $RESULT"
  }
}
```

### Step 4: Test Hook

```bash
# Trigger the hook by performing matching action
# e.g., Edit a file to trigger PostToolUse hook

# Check hook output in Claude Code
```

### Step 5: Commit Configuration

```bash
# Commit hooks.json (shared with team)
git add hooks.json

# Optionally commit hook scripts
git add .claude/hooks/my-hook.sh

git commit -m "feat: add custom hook for..."
```

---

## Creating Custom MCP Integrations

### Step 1: Choose MCP Server

- Use existing MCP server from marketplace
- Or build custom MCP server for proprietary tools

### Step 2: Configure Server

Add to `.claude/mcp.json`:

```json
{
  "mcpServers": {
    "my-service": {
      "transport": "http",
      "url": "https://api.myservice.com/mcp",
      "headers": {
        "Authorization": "Bearer ${MY_SERVICE_TOKEN}"
      }
    }
  }
}
```

### Step 3: Set Environment Variables

```bash
# Add to ~/.bashrc or ~/.zshrc
export MY_SERVICE_TOKEN="your-token-here"

# Reload shell
source ~/.bashrc
```

### Step 4: Test Integration

```bash
# Restart Claude Code
exit
claude

# Test MCP server
/mcp list  # Should show "my-service"

# Use in conversation
"Query my-service for ..."
```

### Step 5: Document Integration

Add to project documentation:

```markdown
## My Service Integration

This project uses My Service via MCP for [purpose].

### Setup

1. Obtain API token from My Service
2. Set environment variable: `export MY_SERVICE_TOKEN=...`
3. Restart Claude Code

### Usage

Reference My Service resources: `@my-service/resource-id`
```

---

## Related Documentation

- **[COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md)** - Slash commands and skills
- **[AGENTS.md](./AGENTS.md)** - Agent system and capabilities
- **[USER_GUIDE.md](./USER_GUIDE.md)** - Day-to-day workflows
- **[DISTRIBUTION.md](./DISTRIBUTION.md)** - Plugin packaging and versioning
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - CLI command reference
- **[../research/CLAUDE_CODE.md](../research/CLAUDE_CODE.md)** - Claude Code hooks and MCP documentation
