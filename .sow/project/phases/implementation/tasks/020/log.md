# Task 020 Log

Worker actions will be logged here.

## Action: Started Task 020: Replace Project Command with New/Continue Subcommands
**Timestamp**: 2025-11-03
**Reasoning**: Read task description, existing project.go (633 lines), SDK state types, and field path parser from Task 010. Understand requirements: delete old unified command, create 4 subcommands (new, continue, set, delete) using SDK state layer.

**Files reviewed**:
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/project.go` - existing implementation to migrate
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/internal/sdks/project/state/loader.go` - SDK Load/Save API
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/internal/cmdutil/fieldpath.go` - field path parser for set command

**Key findings**:
- Old project.go handles branch scenarios, worktree management, issue linking, Claude Code launching
- SDK provides Load(ctx) and project.Save() pattern
- Need to implement Create() function in SDK state layer (doesn't exist yet)
- Field path parser supports direct fields and metadata paths

**Next step**: Write integration test first (TDD requirement)

## Action: Wrote integration test for project lifecycle commands
**Timestamp**: 2025-11-03
**Reasoning**: Following TDD methodology - write test before implementation. Test covers all 4 commands: new, continue, set, delete.

**File created**:
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/testdata/script/unified_commands/project/project_lifecycle.txtar`

**Test scenarios**:
1. Create new project with `sow project new`
2. Verify state file created with correct fields
3. Set direct field (description) with `sow project set`
4. Set metadata fields with `sow project set metadata.key value`
5. Delete project with `sow project delete`
6. Error handling: commands fail when no project exists

**Next step**: Implement Create() function in SDK state layer

## Action: Implemented all command substructure and SDK integration
**Timestamp**: 2025-11-03
**Reasoning**: Completed full TDD cycle - test passed after implementing all commands.

**Files created**:
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/project/project.go` - Root command
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/project/new.go` - new subcommand
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/project/continue.go` - continue subcommand
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/project/set.go` - set subcommand
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/project/delete.go` - delete subcommand
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/helpers.go` - Shared launchClaudeCode helper
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/testdata/script/unified_project_lifecycle.txtar` - Integration test

**Files modified**:
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/internal/sdks/project/state/loader.go` - Added Create() function and registered standard project type
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/root.go` - Updated to import and use new project package

**Files deleted**:
- `/Users/josh/code/sow/.sow/worktrees/50-unified-cli-command-structure/cli/cmd/project.go` - Old unified command

**Implementation notes**:
1. SDK Create() function takes branch explicitly to avoid context issues
2. Standard project type registered in init() function
3. Project-level metadata not supported in current schema (github_issue handling omitted)
4. All commands work from worktree context
5. Field path parser integrated for set command

**Test result**: PASS
