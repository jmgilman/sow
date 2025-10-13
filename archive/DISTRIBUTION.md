# Distribution & Versioning (DRAFT)

**Status**: Draft - Not Authoritative
**Last Updated**: 2025-10-12
**Purpose**: Document how `sow` is packaged, distributed, versioned, and upgraded

This document describes the distribution mechanisms, versioning strategy, and upgrade workflows for the `sow` system.

---

## Overview

`sow` is distributed as a Claude Code Plugin with an optional CLI for enhanced functionality.

**Distribution Components**:
- **Claude Code Plugin** (required) - Agents, commands, hooks, migrations
- **CLI Binary** (optional) - Fast operations, sink/repo management

**Versioning Strategy**:
- Plugin uses semantic versioning (MAJOR.MINOR.PATCH)
- Repository structure tracks version independently
- Migrations bridge version gaps

---

## Plugin Structure

### What Gets Distributed

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

### plugin.json

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
  "license": "MIT"
}
```

### .plugin-version

Simple version file for runtime access:
```
0.2.0
```

---

## Versioning Strategy

### Semantic Versioning

**Format**: `MAJOR.MINOR.PATCH`

**MAJOR** (Breaking Changes):
- Incompatible structure changes requiring migration
- Agent behavior changes that break existing workflows
- Removed commands or significant API changes
- Example: `0.9.0` → `1.0.0`

**MINOR** (New Features):
- New agents, commands, or hooks
- New phases or phase behaviors
- Enhanced functionality (backward compatible)
- Example: `0.1.0` → `0.2.0`

**PATCH** (Bug Fixes):
- Bug fixes in agent prompts
- Documentation updates
- Performance improvements
- Example: `0.1.0` → `0.1.1`

### Version Tracking

**Plugin Version**:
- Stored in `plugin.json` and `.plugin-version`
- Read by orchestrator at runtime
- Compared against repository structure version

**Repository Structure Version**:
- Stored in `.sow/.version` (committed to git)
- Tracks what structure version the repository uses
- Updated by migrations

**`.sow/.version` File**:
```yaml
sow_structure_version: 0.2.0
plugin_version: 0.2.0
last_migrated: 2025-10-12T16:30:00Z
initialized: 2025-10-12T14:00:00Z
```

### Version Compatibility

**Compatible Versions**:
- Plugin `0.2.x` works with structure `0.2.y` (patch differences OK)
- Plugin `0.2.0` prompts migration if structure is `0.1.x`

**Breaking Changes**:
- Major version bumps require migration
- Minor version bumps may require migration (if structure changes)
- Patch versions never require migration

---

## Installation Flow

### First-Time Setup

**Prerequisites**:
- Claude Code installed
- Git repository (existing or new)

**Step 1: Install Plugin**

Via marketplace:
```bash
# Add sow marketplace (one-time)
/plugin marketplace add your-org/sow-marketplace

# Install plugin
/plugin install sow@sow-marketplace
```

Via git URL:
```bash
/plugin install https://github.com/your-org/sow
```

**Step 2: Restart Claude Code**
```bash
# Exit and restart
exit
claude
```

**Step 3: SessionStart Hook**

When Claude Code starts:
```
⚠️  Not a sow repository
💡 Use /init to set up sow
```

**Step 4: Initialize Repository**

```
User: /init

Orchestrator:
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
   The CLI provides fast operations like logging and sink management.

   Download for your platform:
   - macOS: https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos
   - Linux: https://github.com/your-org/sow/releases/download/v0.2.0/sow-linux
   - Windows: https://github.com/your-org/sow/releases/download/v0.2.0/sow-windows.exe

   After download: mv ~/Downloads/sow-macos ~/.local/bin/sow && chmod +x ~/.local/bin/sow

   Test with: sow --version

[y to download now, n to skip, ? for more info]:
```

**Step 5: Ready to Work**

```
✓ sow initialized successfully!

Available commands:
  /start-project <name> - Create new project
  /continue - Resume existing project (if any)

Try: /start-project "Add authentication"
```

---

## Upgrade Flow

### Plugin Upgrade Process

**Note**: Claude Code plugin upgrade documentation is sparse. This workflow is based on available commands and reasonable assumptions.

**Step 1: Check for Updates**

Refresh marketplace metadata:
```bash
/plugin marketplace update sow-marketplace
```

List plugins to see update availability:
```bash
/plugin
# Shows: sow (installed: 0.1.0, available: 0.2.0)
```

**Step 2: Upgrade Plugin**

Reinstall to get latest version:
```bash
/plugin install sow@sow-marketplace
```

Or specify explicit version:
```bash
/plugin install sow@sow-marketplace --version 0.2.0
```

**Step 3: Restart Claude Code**

Required for plugin changes to take effect:
```bash
exit
claude
```

**Step 4: SessionStart Hook Detects Mismatch**

When Claude Code starts after upgrade:
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

**Step 5: Run Migration**

```
User: /migrate

Orchestrator:
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

---

## Migration System

### Migration Files

**Location**: `.claude/migrations/` (in plugin)

**Naming**: `<from>-to-<to>.md`

**Example**: `.claude/migrations/0.1.0-to-0.2.0.md`

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
   - Create `context/` directory
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
```

Then restart Claude Code.

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

1. **Read Versions**
   - Current: `.sow/.version` → `sow_structure_version`
   - Target: `.plugin-version` → plugin version

2. **Build Migration Chain**
   - If current `0.1.0` and target `0.3.0`:
   - Apply: `0.1.0-to-0.2.0.md`
   - Then: `0.2.0-to-0.3.0.md`

3. **Parse Migration File**
   - Read `.claude/migrations/<version-pair>.md`
   - Extract automated steps
   - Execute file operations

4. **Apply Changes**
   - Create directories
   - Convert files
   - Update state files
   - Update `.sow/.version`

5. **Commit**
   - Stage all changes
   - Commit: `chore: migrate sow structure to v<target>`

6. **Report**
   - Show what changed
   - Provide changelog link
   - Confirm completion

### Version Skipping

**Scenario**: User on `0.1.0`, latest is `0.3.0`

**Sequential Migration**:
```
User: /migrate

Orchestrator:
🔍 Multiple migrations required: 0.1.0 → 0.3.0

Migration path:
  0.1.0 → 0.2.0 (add context support)
  0.2.0 → 0.3.0 (add parallel task support)

Apply all migrations? [y/n]:

User: y

🚀 Applying 0.1.0 → 0.2.0...
✓ Complete

🚀 Applying 0.2.0 → 0.3.0...
✓ Complete

All migrations applied successfully!
Your repository is now at v0.3.0
```

---

## CLI Distribution

### Optional Enhancement

The CLI is **optional** - `sow` works fully without it. The CLI provides:
- Fast operations (`sow log`, `sow session-info`)
- Sink management (`sow sinks install/update`)
- Repository management (`sow repos add/sync`)

### Distribution Method

**GitHub Releases**:
```
https://github.com/your-org/sow/releases/
├── v0.2.0/
│   ├── sow-macos          (macOS binary)
│   ├── sow-linux          (Linux binary)
│   └── sow-windows.exe    (Windows binary)
```

**Installation** (offered during `/init`):
```bash
# macOS/Linux
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow-macos -o sow
chmod +x sow
mv sow ~/.local/bin/sow

# Verify
sow --version
# sow 0.2.0
```

**CLI Commands**:
- `sow log` - Fast logging (used by agents)
- `sow session-info` - Show repository status (SessionStart hook)
- `sow sinks install <source>` - Install sink
- `sow sinks update` - Update all sinks
- `sow sinks list` - List installed sinks
- `sow repos add <url>` - Link repository
- `sow repos sync` - Sync linked repos
- `sow validate` - Validate structure integrity
- `sow --version` - Show CLI version

**Version Alignment**:
- CLI version should match plugin version
- CLI `0.2.0` works with plugin `0.2.0`
- During `/init`, offer matching CLI version

---

## Marketplace Publishing

### Publishing to Marketplace

**Prerequisites**:
1. GitHub repository with plugin structure
2. Valid `plugin.json` with version
3. README.md with installation instructions
4. Tagged release matching version

**Steps**:

**1. Prepare Release**
```bash
# Tag version
git tag v0.2.0
git push origin v0.2.0

# Create GitHub release with tag v0.2.0
# Include CHANGELOG.md excerpt
```

**2. Create Marketplace Listing**

Create `marketplace.json` in your org's marketplace repo:
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

**3. Update Marketplace**

Push to marketplace repository:
```bash
cd sow-marketplace
# Update marketplace.json with new version
git add marketplace.json
git commit -m "Update sow to v0.2.0"
git push
```

**4. Users Can Install**

```bash
/plugin marketplace update sow-marketplace
/plugin install sow@sow-marketplace
```

### Updating Published Plugin

**For New Versions**:

1. **Update plugin.json**:
   ```json
   {
     "version": "0.3.0"
   }
   ```

2. **Update .plugin-version**:
   ```
   0.3.0
   ```

3. **Add Migration** (if structure changes):
   - Create `.claude/migrations/0.2.0-to-0.3.0.md`

4. **Update CHANGELOG.md**:
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

5. **Tag and Release**:
   ```bash
   git tag v0.3.0
   git push origin v0.3.0
   # Create GitHub release
   ```

6. **Update Marketplace**:
   - Update `marketplace.json` version to `0.3.0`
   - Push marketplace changes

---

## Version Check Implementation

### SessionStart Hook

**`.claude/hooks.json`**:
```json
{
  "SessionStart": {
    "matcher": "*",
    "command": "sow session-info"
  }
}
```

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
  exit 0
fi

echo "✓ Versions aligned (v$PLUGIN_VERSION)"
echo ""
echo "📖 Available commands:"
echo "   /start-project <name> - Create new project"
echo "   /continue - Resume existing project"
echo "   /cleanup - Delete project before merge"
```

---

## Best Practices

### For Plugin Developers

1. **Always Include Migration**
   - Even if "no changes needed", document it
   - Helps users understand what changed

2. **Test Migrations**
   - Create test repo at old version
   - Run migration
   - Verify all scenarios (with/without project, etc.)

3. **Semantic Versioning**
   - Be conservative with major bumps
   - Document breaking changes clearly

4. **Changelog Discipline**
   - Update CHANGELOG.md with every release
   - Link from migration files

5. **Version Alignment**
   - Keep CLI version aligned with plugin version
   - Test CLI + plugin together

### For Users

1. **Check for Updates Regularly**
   ```bash
   /plugin marketplace update sow-marketplace
   /plugin
   ```

2. **Read Changelogs Before Upgrading**
   - Visit release page
   - Understand what's changing

3. **Backup Before Major Upgrades**
   ```bash
   git branch backup-before-sow-upgrade
   ```

4. **Run Migration Promptly**
   - Don't work across version mismatches
   - Run `/migrate` when prompted

5. **Report Issues**
   - If migration fails, report with:
     - Old version
     - New version
     - Error message
     - Repository state

---

## Troubleshooting

### Version Mismatch Won't Clear

**Symptom**: After `/migrate`, still shows version mismatch

**Solution**:
```bash
# Check .sow/.version manually
cat .sow/.version

# Should show:
# sow_structure_version: 0.2.0
# plugin_version: 0.2.0

# If not, manually update
# Then commit
git add .sow/.version
git commit -m "fix: update sow version"
```

### Migration Failed Mid-Process

**Symptom**: Migration started but encountered error

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

### CLI Version Mismatch

**Symptom**: CLI is different version than plugin

**Solution**:
```bash
# Check CLI version
sow --version

# Check plugin version
cat .claude/.plugin-version

# Download matching CLI version
# From: https://github.com/your-org/sow/releases/download/v<version>/sow-<platform>
```

---

## Future Considerations

### Automatic Updates

**Potential Feature**: Auto-update plugins (if Claude Code adds this)

**Design**:
- Users opt-in to auto-updates
- Plugin updates automatically when available
- `/migrate` prompted on next session
- Rollback always available

### Migration Dry-Run

**Potential Feature**: Preview migrations before applying

```
User: /migrate --dry-run

Orchestrator:
🔍 Migration preview: 0.1.0 → 0.2.0

Would apply these changes:
  - Create .sow/project/context/
  - Convert state.json to state.yaml
  - Add iteration field to 3 tasks
  - Update .sow/.version

No changes will be made.
Run /migrate to apply.
```

### Rollback Command

**Potential Feature**: Easy rollback to previous version

```
User: /rollback

Orchestrator:
🔙 Rolling back to previous version...

Current: 0.2.0
Target: 0.1.0

This will:
  - Revert last migration commit
  - Restore previous structure
  - Update .sow/.version

Proceed? [y/n]:
```

---

## Related Documentation

- **`FS_STRUCTURE.md`**: Complete filesystem layout
- **`PROJECT_LIFECYCLE.md`**: Operational workflows
- **`EXECUTION.md`**: Agent behavior and execution layer
- **`BRAINSTORMING.md`**: Design decisions and exploration
