# Distribution

**Last Updated**: 2025-10-12
**Status**: Comprehensive Architecture Documentation

---

## Table of Contents

- [Overview](#overview)
- [Plugin Packaging](#plugin-packaging)
  - [Package Structure](#package-structure)
  - [Plugin Metadata](#plugin-metadata)
  - [Version Files](#version-files)
- [Semantic Versioning](#semantic-versioning)
  - [Version Format](#version-format)
  - [Version Tracking](#version-tracking)
  - [Version Compatibility](#version-compatibility)
- [Installation Flow](#installation-flow)
  - [First-Time Setup](#first-time-setup)
  - [Plugin Installation](#plugin-installation)
  - [Repository Initialization](#repository-initialization)
  - [CLI Installation](#cli-installation)
- [Upgrade Workflow](#upgrade-workflow)
  - [Checking for Updates](#checking-for-updates)
  - [Upgrading the Plugin](#upgrading-the-plugin)
  - [Version Mismatch Detection](#version-mismatch-detection)
  - [Running Migrations](#running-migrations)
- [Migration System](#migration-system)
  - [Migration Files](#migration-files)
  - [Migration Execution](#migration-execution)
  - [Sequential Migrations](#sequential-migrations)
  - [Rollback Procedures](#rollback-procedures)
- [CLI Distribution](#cli-distribution)
  - [Binary Distribution](#binary-distribution)
  - [Installation Methods](#installation-methods)
  - [Version Alignment](#version-alignment)
- [Marketplace Publishing](#marketplace-publishing)
  - [Publishing Process](#publishing-process)
  - [Marketplace Listing](#marketplace-listing)
  - [Updating Releases](#updating-releases)
- [Version Check Implementation](#version-check-implementation)
  - [SessionStart Hook](#sessionstart-hook)
  - [Version Detection](#version-detection)
  - [Migration Prompts](#migration-prompts)
- [Best Practices](#best-practices)
  - [For Plugin Developers](#for-plugin-developers)
  - [For Users](#for-users)
- [Troubleshooting](#troubleshooting)
  - [Common Issues](#common-issues)
  - [Recovery Procedures](#recovery-procedures)
- [Related Documentation](#related-documentation)

---

## Overview

`sow` is distributed as a **Claude Code Plugin** with an **optional CLI** for enhanced functionality.

**Distribution Components**:
- **Claude Code Plugin** (required) - Agents, commands, hooks, migrations
- **CLI Binary** (optional) - Fast operations, sink/repo management

**Versioning Strategy**:
- Plugin uses semantic versioning (MAJOR.MINOR.PATCH)
- Repository structure tracks version independently
- Migrations bridge version gaps when structure changes

---

## Plugin Packaging

### Package Structure

The plugin bundles all execution layer components:

```
sow-plugin/
├── .claude-plugin/
│   └── plugin.json              # Metadata with version
│
├── .plugin-version               # Plugin version (for runtime access)
│
├── agents/
│   ├── orchestrator.md
│   ├── architect.md
│   ├── implementer.md
│   ├── integration-tester.md
│   ├── reviewer.md
│   └── documenter.md
│
├── commands/
│   ├── workflows/
│   │   ├── init.md              # Bootstrap repository
│   │   ├── start-project.md
│   │   ├── continue.md
│   │   ├── cleanup.md
│   │   ├── migrate.md           # Version migration
│   │   └── sync.md
│   └── skills/
│       ├── architect/
│       ├── implementer/
│       ├── integration-tester/
│       ├── reviewer/
│       └── documenter/
│
├── hooks.json                    # SessionStart hook for version checking
│
├── migrations/                   # Migration instructions
│   ├── 0.1.0-to-0.2.0.md
│   └── 0.2.0-to-0.3.0.md
│
└── README.md                     # Installation instructions
```

**What Gets Installed**:
- All agents (`.md` files)
- All commands (workflows and skills)
- Hooks configuration
- Migration files
- Plugin metadata

**What's NOT Included**:
- `.sow/` data layer (created by `/init`)
- User-specific configurations
- Project state
- Installed sinks or repos

### Plugin Metadata

**File**: `.claude-plugin/plugin.json`

**Purpose**: Metadata for Claude Code plugin system

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

**Required Fields**:
- `name` - Plugin identifier (kebab-case)
- `version` - Semantic version
- `description` - Short description
- `author` - Author information

**Optional Fields**:
- `homepage` - Plugin homepage URL
- `repository` - Source code repository
- `license` - License identifier (MIT, Apache-2.0, etc.)
- `keywords` - Search keywords for marketplace
- `engines` - Required Claude Code version

### Version Files

Two version files track versions:

**1. Plugin Version** (`.plugin-version`):
```
0.2.0
```
- Simple text file
- Plugin's current version
- Read by orchestrator at runtime
- Used for version mismatch detection

**2. Repository Structure Version** (`.sow/.version`):
```yaml
sow_structure_version: 0.2.0
plugin_version: 0.2.0
last_migrated: 2025-10-12T16:30:00Z
initialized: 2025-10-12T14:00:00Z
```
- Tracks repository structure version
- Created by `/init`
- Updated by `/migrate`
- Committed to git

**Why Two Files?**:
- Plugin version = what's installed
- Structure version = what repository uses
- Allows detection of version mismatches
- Guides migration process

---

## Semantic Versioning

### Version Format

**Format**: `MAJOR.MINOR.PATCH`

**Example**: `0.2.0`, `1.3.5`, `2.0.0`

### Version Meanings

**MAJOR** (Breaking Changes):
- Incompatible structure changes requiring migration
- Agent behavior changes that break existing workflows
- Removed commands or significant API changes
- Example: `0.9.0` → `1.0.0`

**MINOR** (New Features):
- New agents, commands, or hooks
- New phases or phase behaviors
- Enhanced functionality (backward compatible)
- Optional structure changes with migration
- Example: `0.1.0` → `0.2.0`

**PATCH** (Bug Fixes):
- Bug fixes in agent prompts
- Documentation updates
- Performance improvements
- No structure changes, no migration needed
- Example: `0.1.0` → `0.1.1`

**Examples**:
```
# Patch release (no migration)
0.1.0 → 0.1.1: Fixed typo in architect agent prompt

# Minor release (optional migration)
0.1.0 → 0.2.0: Added context/ directory for projects

# Major release (required migration)
1.0.0 → 2.0.0: Complete restructure of phase system
```

### Version Tracking

**Plugin Version**:
- Stored in `plugin.json` and `.plugin-version`
- Distributed with plugin
- Read by orchestrator at runtime

**Repository Structure Version**:
- Stored in `.sow/.version` (committed to git)
- Tracks what structure version repository uses
- Updated by migrations
- Team shares via git

**Version Flow**:
```
1. Plugin v0.2.0 installed
2. Repository initialized → .sow/.version shows 0.2.0
3. Plugin upgraded to v0.3.0
4. SessionStart detects mismatch (repo: 0.2.0, plugin: 0.3.0)
5. User runs /migrate
6. .sow/.version updated to 0.3.0
7. Versions aligned
```

### Version Compatibility

**Compatible Versions**:
- Plugin `0.2.x` works with structure `0.2.y` (patch differences OK)
- Plugin `0.2.0` prompts migration if structure is `0.1.x`
- Major version bumps always require migration

**Incompatible Scenarios**:
```
Plugin 0.2.0 + Structure 0.1.0 → Migration needed
Plugin 1.0.0 + Structure 0.9.0 → Migration required
Plugin 0.2.1 + Structure 0.2.0 → Compatible (patch diff)
```

**Migration Decision Matrix**:

| Scenario | Migration Required? |
|----------|---------------------|
| Same MAJOR.MINOR, different PATCH | No |
| Different MINOR, same MAJOR | Maybe (check release notes) |
| Different MAJOR | Yes |

---

## Installation Flow

### First-Time Setup

**Prerequisites**:
- Claude Code installed
- Git repository (existing or new)

### Plugin Installation

**Via Marketplace**:
```bash
# Add sow marketplace (one-time)
/plugin marketplace add your-org/sow-marketplace

# Install plugin
/plugin install sow@sow-marketplace
```

**Via Git URL**:
```bash
/plugin install https://github.com/your-org/sow
```

**With Specific Version**:
```bash
/plugin install sow@sow-marketplace --version 0.2.0
```

**Restart Required**:
```bash
exit
claude
```

Changes take effect after restart.

### Repository Initialization

After plugin installed, initialize repository:

```
/init
```

**Actions**:
1. Check prerequisites (git repository, not already initialized)
2. Create `.sow/` structure:
   - `.sow/knowledge/` (with `overview.md` template)
   - `.sow/sinks/` (with empty `index.json`)
   - `.sow/repos/` (with empty `index.json`)
   - `.sow/.version` (tracks version 0.2.0)
3. Create `.gitignore` entries for `.sow/sinks/` and `.sow/repos/`
4. Commit structure to git
5. Offer optional CLI installation

**Success Output**:
```
✓ Checking prerequisites...
✓ Creating .sow/ structure...
  - .sow/knowledge/ (with overview.md template)
  - .sow/sinks/ (with index.json)
  - .sow/repos/ (with index.json)
  - .sow/.version (tracking version 0.2.0)

✓ Creating .gitignore entries...
  - .sow/sinks/
  - .sow/repos/

✓ Committing structure to git...
  [main abc1234] Initialize sow (v0.2.0)

🚀 Optional: Install sow CLI for enhanced functionality?
   [Details about CLI installation...]

[y/n]:
```

### CLI Installation

**Optional** but provides faster operations:

**Download**:
```bash
# macOS
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos -o sow
chmod +x sow
mv sow ~/.local/bin/sow

# Linux
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-linux -o sow
chmod +x sow
mv sow ~/.local/bin/sow

# Windows
# Download sow-windows.exe and add to PATH
```

**Verify**:
```bash
sow --version
# sow 0.2.0
```

**Benefits**:
- Fast logging (used by agents)
- Sink management commands
- Repository management
- Validation utilities

---

## Upgrade Workflow

### Checking for Updates

**Refresh Marketplace**:
```bash
/plugin marketplace update sow-marketplace
```

**List Installed Plugins**:
```bash
/plugin
# Shows: sow (installed: 0.1.0, available: 0.2.0)
```

**Check Release Notes**:
- Visit GitHub releases page
- Review CHANGELOG.md
- Understand what's changing

### Upgrading the Plugin

**Step 1: Upgrade Plugin**

```bash
/plugin install sow@sow-marketplace
```

This reinstalls to get latest version.

Or specify explicit version:
```bash
/plugin install sow@sow-marketplace --version 0.2.0
```

**Step 2: Restart Claude Code**

Required for plugin changes to take effect:
```bash
exit
claude
```

**Step 3: Version Mismatch Detected**

SessionStart hook automatically detects mismatch.

### Version Mismatch Detection

When Claude Code starts after plugin upgrade:

```
📋 You are in a sow-enabled repository

⚠️  Version mismatch detected!
   Repository structure: 0.1.0
   Plugin version: 0.2.0

💡 Run /migrate to upgrade your repository structure

   Migration path: 0.1.0 → 0.2.0
   Review changes: https://github.com/your-org/sow/blob/main/CHANGELOG.md#020

[Do you want to migrate now? y/n]:
```

**Automatic Detection**:
- SessionStart hook runs `sow session-info`
- Reads `.sow/.version` (structure version)
- Reads `.claude/.plugin-version` (plugin version)
- Compares versions
- Prompts if mismatch

### Running Migrations

**Command**: `/migrate`

**Process**:
```
/migrate

🔍 Analyzing versions...
   Current: 0.1.0
   Target: 0.2.0

📋 Migration: 0.1.0 → 0.2.0

Changes in this migration:
- Add .sow/project/context/ directory for project-specific context
- Convert state.json to state.yaml (YAML format)
- Add iteration field to all task states
- Update hooks.json with SessionStart hook

🚀 Applying migration...

✓ Created .sow/project/context/ (if project exists)
✓ Converted state.json to state.yaml
✓ Added iteration field to 3 tasks
✓ Updated .sow/.version
✓ Committed changes

Migration complete! Your repository is now at v0.2.0

📝 Changelog: https://github.com/your-org/sow/blob/main/CHANGELOG.md#020
```

**What Happens**:
1. Read versions (current from `.sow/.version`, target from `.plugin-version`)
2. Build migration chain if skipping versions
3. Parse migration file(s) from `.claude/migrations/`
4. Execute automated steps (create dirs, convert files, update state)
5. Update `.sow/.version`
6. Commit changes: `chore: migrate sow structure to v<target>`
7. Report completion

---

## Migration System

### Migration Files

**Location**: `.claude/migrations/` (in plugin)

**Naming**: `<from>-to-<to>.md`

**Examples**:
- `0.1.0-to-0.2.0.md`
- `0.2.0-to-0.3.0.md`
- `1.0.0-to-2.0.0.md`

**Format**: Markdown with structured sections

**Example** (`0.1.0-to-0.2.0.md`):

```markdown
# Migration: 0.1.0 → 0.2.0

## Summary

This migration adds project context support and converts JSON state files to YAML.

**Breaking Changes**: None (backward compatible)

## Changes

1. **Add project context directory**
   - Creates `.sow/project/context/` for storing project-specific context
   - Includes `overview.md`, `decisions.md`, `memories.md` templates

2. **Convert state format**
   - Renames `state.json` to `state.yaml`
   - Converts content from JSON to YAML format
   - Preserves all existing data

3. **Add iteration tracking**
   - Adds `iteration: 1` field to all task state files
   - Enables worker attempt tracking

4. **Update hooks**
   - Adds SessionStart hook for version checking
   - Updates hooks.json structure

## Automated Steps

The `/migrate` command will automatically:

1. Check if `.sow/project/` exists
2. If yes:
   - Create `context/` directory with templates
   - Convert `state.json` to `state.yaml`
   - Add iteration field to all task states
3. Update `.sow/.version`:
   - `sow_structure_version: 0.2.0`
   - `plugin_version: 0.2.0`
   - `last_migrated: <timestamp>`
4. Commit changes with message: "chore: migrate sow structure to v0.2.0"

## Manual Steps

None required - all changes are automated.

## Rollback

If issues occur:

```bash
git revert HEAD  # Revert migration commit
/plugin install sow@sow-marketplace --version 0.1.0  # Reinstall old version
# Restart Claude Code
```

## Testing

After migration:
1. Verify `.sow/.version` shows correct version
2. Check project state files (if any) are valid YAML
3. Try `/continue` if project exists
4. SessionStart hook should no longer show version mismatch

## Support

Issues? Report at: https://github.com/your-org/sow/issues
```

### Migration Execution

**`/migrate` Command Logic**:

1. **Read Versions**:
   - Current: `.sow/.version` → `sow_structure_version`
   - Target: `.plugin-version` → plugin version

2. **Build Migration Chain**:
   - If current `0.1.0` and target `0.3.0`:
     - Apply: `0.1.0-to-0.2.0.md`
     - Then: `0.2.0-to-0.3.0.md`

3. **Parse Migration File**:
   - Read `.claude/migrations/<version-pair>.md`
   - Extract automated steps section
   - Understand what needs to be done

4. **Apply Changes**:
   - Create directories
   - Convert files
   - Update state files
   - Update `.sow/.version`

5. **Commit**:
   - Stage all changes
   - Commit: `chore: migrate sow structure to v<target>`

6. **Report**:
   - Show what changed
   - Provide changelog link
   - Confirm completion

### Sequential Migrations

**Scenario**: User on `0.1.0`, latest is `0.3.0`

**Process**:
```
User: /migrate

Orchestrator:
🔍 Multiple migrations required: 0.1.0 → 0.3.0

Migration path:
  0.1.0 → 0.2.0 (add context support)
  0.2.0 → 0.3.0 (add parallel task support)

Apply all migrations? [y/n]

User: y

🚀 Applying 0.1.0 → 0.2.0...
✓ Complete

🚀 Applying 0.2.0 → 0.3.0...
✓ Complete

All migrations applied successfully!
Your repository is now at v0.3.0
```

**Benefits**:
- Handles version skipping
- Applies all intermediate migrations
- Ensures consistent state
- Single command for multiple upgrades

### Rollback Procedures

**If Migration Fails**:

1. **Check Git Status**:
   ```bash
   git status
   # See partial changes
   ```

2. **Restore State**:
   ```bash
   # Option 1: Rollback commit (if committed)
   git revert HEAD

   # Option 2: Discard changes (if not committed)
   git restore .
   ```

3. **Reinstall Old Version**:
   ```bash
   /plugin install sow@sow-marketplace --version 0.1.0
   # Restart Claude Code
   ```

4. **Try Again** or **Report Issue**

**If Migration Succeeds but Causes Issues**:

1. **Revert Commit**:
   ```bash
   git revert HEAD
   ```

2. **Downgrade Plugin**:
   ```bash
   /plugin install sow@sow-marketplace --version 0.1.0
   # Restart Claude Code
   ```

3. **Report Issue** with:
   - Old version
   - New version
   - Error details
   - Steps to reproduce

---

## CLI Distribution

### Binary Distribution

**GitHub Releases**:
```
https://github.com/your-org/sow/releases/
├── v0.2.0/
│   ├── sow-macos          (macOS binary)
│   ├── sow-linux          (Linux binary)
│   └── sow-windows.exe    (Windows binary)
```

**Platforms**:
- macOS (Intel and Apple Silicon)
- Linux (x86_64)
- Windows (x86_64)

### Installation Methods

**macOS/Linux**:
```bash
# Download
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos -o sow

# Make executable
chmod +x sow

# Move to PATH
mv sow ~/.local/bin/sow

# Verify
sow --version
```

**Windows**:
```powershell
# Download sow-windows.exe
# Add to PATH
# Verify
sow --version
```

**Via Package Managers** (future):
```bash
# Homebrew (macOS/Linux)
brew install your-org/tap/sow

# Apt (Debian/Ubuntu)
apt install sow

# Chocolatey (Windows)
choco install sow
```

### Version Alignment

**Important**: CLI version should match plugin version

**Check Versions**:
```bash
# CLI version
sow --version
# sow 0.2.0

# Plugin version
cat .claude/.plugin-version
# 0.2.0
```

**During `/init`**:
- Orchestrator detects plugin version
- Offers matching CLI version for download
- Provides platform-specific instructions

**Manual Upgrade**:
```bash
# Check latest release
# https://github.com/your-org/sow/releases/latest

# Download matching version
curl -L https://github.com/your-org/sow/releases/download/v0.3.0/sow-macos -o sow
chmod +x sow
mv sow ~/.local/bin/sow

# Verify
sow --version
# sow 0.3.0
```

---

## Marketplace Publishing

### Publishing Process

**Prerequisites**:
1. GitHub repository with plugin structure
2. Valid `plugin.json` with version
3. README.md with installation instructions
4. Tagged release matching version

**Steps**:

**1. Prepare Release**:
```bash
# Update version in plugin.json
# Update .plugin-version
# Update CHANGELOG.md
# Commit changes

# Tag version
git tag v0.2.0
git push origin v0.2.0
```

**2. Create GitHub Release**:
- Go to repository releases page
- Create new release for tag `v0.2.0`
- Include CHANGELOG.md excerpt
- Attach CLI binaries (if applicable)

**3. Create Marketplace Listing**:

Create `marketplace.json` in marketplace repo:
```json
{
  "plugins": [
    {
      "name": "sow",
      "description": "AI-powered system of work for software engineering",
      "repository": "https://github.com/your-org/sow",
      "version": "0.2.0",
      "author": "sow contributors",
      "homepage": "https://github.com/your-org/sow",
      "tags": ["productivity", "workflow", "agents", "project-management"]
    }
  ]
}
```

**4. Update Marketplace**:
```bash
cd sow-marketplace
# Update marketplace.json
git add marketplace.json
git commit -m "Update sow to v0.2.0"
git push
```

**5. Users Can Install**:
```bash
/plugin marketplace update sow-marketplace
/plugin install sow@sow-marketplace
```

### Marketplace Listing

**Required Fields**:
- `name` - Plugin identifier
- `description` - Short description
- `repository` - Git repository URL
- `version` - Current version
- `author` - Author name

**Optional Fields**:
- `homepage` - Plugin homepage
- `tags` - Search keywords
- `license` - License identifier

### Updating Releases

**For New Versions**:

1. **Update Plugin Files**:
   ```json
   // plugin.json
   {
     "version": "0.3.0"
   }
   ```
   ```
   // .plugin-version
   0.3.0
   ```

2. **Add Migration** (if structure changes):
   - Create `.claude/migrations/0.2.0-to-0.3.0.md`

3. **Update CHANGELOG.md**:
   ```markdown
   ## [0.3.0] - 2025-10-15

   ### Added
   - Parallel task execution support
   - New `parallel` flag in task state

   ### Changed
   - Orchestrator can spawn multiple workers simultaneously

   ### Migration
   - Run `/migrate` to upgrade from 0.2.0
   ```

4. **Commit and Tag**:
   ```bash
   git commit -m "chore: bump version to 0.3.0"
   git tag v0.3.0
   git push origin v0.3.0
   ```

5. **Create GitHub Release**:
   - Tag: v0.3.0
   - Release notes from CHANGELOG.md
   - Attach CLI binaries

6. **Update Marketplace**:
   - Update `marketplace.json` version to `0.3.0`
   - Push marketplace changes

---

## Version Check Implementation

### SessionStart Hook

**Configuration** (`hooks.json`):
```json
{
  "SessionStart": {
    "matcher": "*",
    "command": "sow session-info"
  }
}
```

### Version Detection

**CLI Command**: `sow session-info`

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

### Migration Prompts

**Automatic Prompt**:
```
⚠️  Version mismatch detected!
   Repository structure: 0.1.0
   Plugin version: 0.2.0

💡 Run /migrate to upgrade your repository structure

   Migration path: 0.1.0 → 0.2.0
   Review changes: https://github.com/your-org/sow/blob/main/CHANGELOG.md#020

[Do you want to migrate now? y/n]:
```

**User Response**:
- `y` → Run `/migrate` immediately
- `n` → Skip for now (prompted again next session)

---

## Best Practices

### For Plugin Developers

1. **Always Include Migration Files**:
   - Even if "no changes needed", document it
   - Helps users understand what changed

2. **Test Migrations Thoroughly**:
   - Create test repo at old version
   - Run migration
   - Verify all scenarios (with/without project, etc.)

3. **Semantic Versioning Discipline**:
   - Be conservative with major bumps
   - Document breaking changes clearly
   - Patch versions never require migration

4. **Changelog Maintenance**:
   - Update CHANGELOG.md with every release
   - Link from migration files
   - Follow "Keep a Changelog" format

5. **Version Alignment**:
   - Keep CLI version aligned with plugin version
   - Test CLI + plugin together
   - Offer matching CLI during `/init`

6. **Rollback Support**:
   - Test downgrade paths
   - Document rollback procedures
   - Support at least one version back

### For Users

1. **Check for Updates Regularly**:
   ```bash
   /plugin marketplace update sow-marketplace
   /plugin
   ```

2. **Read Changelogs Before Upgrading**:
   - Visit release page
   - Understand what's changing
   - Check for breaking changes

3. **Backup Before Major Upgrades**:
   ```bash
   git branch backup-before-sow-upgrade
   ```

4. **Run Migration Promptly**:
   - Don't work across version mismatches
   - Run `/migrate` when prompted
   - Verify migration success

5. **Keep CLI Aligned**:
   - Upgrade CLI when upgrading plugin
   - Check versions match
   - Download from releases page

6. **Report Issues**:
   - If migration fails, report with:
     - Old version
     - New version
     - Error message
     - Repository state

---

## Troubleshooting

### Common Issues

#### Version Mismatch Won't Clear

**Symptoms**: After `/migrate`, still shows version mismatch

**Solution**:
```bash
# Check .sow/.version manually
cat .sow/.version

# Should show matching versions
# If not, manually update and commit
yq -i '.sow_structure_version = "0.2.0"' .sow/.version
yq -i '.plugin_version = "0.2.0"' .sow/.version
git add .sow/.version
git commit -m "fix: update sow version"
```

#### Migration Failed Mid-Process

**Symptoms**: Migration started but encountered error

**Solution**:
```bash
# Check git status
git status

# See partial changes
git diff

# Option 1: Complete manually
# Fix the issue, then:
# Update .sow/.version
# Commit changes

# Option 2: Rollback
git restore .
# Start over with /migrate
```

#### CLI Version Mismatch

**Symptoms**: CLI is different version than plugin

**Solution**:
```bash
# Check CLI version
sow --version

# Check plugin version
cat .claude/.plugin-version

# Download matching CLI version
# From: https://github.com/your-org/sow/releases/download/v<version>/sow-<platform>
```

#### Plugin Installation Failed

**Symptoms**: Error during plugin installation

**Solution**:
```bash
# Check Claude Code version
claude --version

# Update Claude Code if needed
# Then retry plugin installation
/plugin install sow@sow-marketplace

# Check plugin installed
/plugin
```

### Recovery Procedures

#### Corrupt Installation

**Symptoms**: Plugin installed but not working correctly

**Recovery**:
```bash
# Uninstall plugin
/plugin uninstall sow

# Restart Claude Code
exit
claude

# Reinstall plugin
/plugin install sow@sow-marketplace

# Restart again
exit
claude
```

#### Lost Version Information

**Symptoms**: `.sow/.version` missing or corrupted

**Recovery**:
```bash
# Recreate version file manually
cat > .sow/.version << EOF
sow_structure_version: 0.2.0
plugin_version: 0.2.0
last_migrated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
initialized: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF

# Commit
git add .sow/.version
git commit -m "fix: recreate version file"
```

---

## Related Documentation

- **[USER_GUIDE.md](./USER_GUIDE.md)** - Installation and daily usage
- **[COMMANDS_AND_SKILLS.md](./COMMANDS_AND_SKILLS.md)** - Commands including /migrate
- **[HOOKS_AND_INTEGRATIONS.md](./HOOKS_AND_INTEGRATIONS.md)** - SessionStart hook details
- **[PROJECT_MANAGEMENT.md](./PROJECT_MANAGEMENT.md)** - Project lifecycle
- **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - CLI command reference
- **[OVERVIEW.md](./OVERVIEW.md)** - System overview
