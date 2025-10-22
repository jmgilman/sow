# GitHub Issue Integration Plan

**Status**: Planning Complete
**Created**: 2025-10-20
**Purpose**: Add GitHub issue integration for project management and large work decomposition

---

## Executive Summary

This plan introduces GitHub issue integration to sow, enabling:
1. **Large work decomposition** - Break down very large features into multiple GitHub issues (projects)
2. **Project discovery** - List and claim available work via GitHub issues
3. **Branch-issue linking** - Automatic branch linking using `gh issue develop`
4. **Team coordination** - Prevent duplicate work through native GitHub features

**Core Philosophy**: Leverage GitHub's native issue and branch linking features rather than building custom state management.

---

## Problem Statement

### Current Limitations

**Works Well:**
- Small to medium tasks (single project with 5-20 tasks)
- One-off work (no project needed)

**Breaks Down:**
- Very large features requiring 50+ tasks
- Complex changes spanning multiple PRs
- Team coordination on what work exists and who's doing what

### The Gap

Currently, there's no way to:
- Decompose very large work into multiple projects
- Discover what projects are available across the team
- Prevent duplicate work (two people creating projects for the same work)
- Track cross-project dependencies

---

## Solution Overview

### High-Level Approach

**Use GitHub issues as the project registry:**
- Issues with `sow` label = sow projects
- One issue = one project = one branch = one PR
- GitHub's native branch linking via `gh issue develop`
- Orchestrator helps create and manage issues

### Key Insight

**You don't need status of ALL issues, just the ONE you're working on.**

Instead of expensive queries showing "available/claimed" for every issue:
- List command shows all `sow` labeled issues (cheap)
- Check command queries specific issue for linked branches (cheap, on-demand)
- Init command fails gracefully if issue already has linked branch

---

## Technical Approach

### GitHub CLI (`gh issue develop`)

**Discovery**: The `gh issue develop` command provides everything we need:

```bash
# Create branch linked to issue
gh issue develop 123 --checkout
# → Creates branch: 123-add-authentication
# → Links branch to issue in GitHub
# → Checks out branch locally

# Check if issue has linked branches
gh issue develop --list 123
# → Shows branches linked to issue #123
# → Empty if no branches linked

# List all sow issues
gh issue list --label sow --json number,title,url,state
# → Fast query, just label filtering
```

**Benefits:**
- ✓ Programmatic branch creation + linking
- ✓ Query linked branches
- ✓ Automatic naming convention (`<number>-<title>`)
- ✓ Visible in GitHub UI (Development section)
- ✓ No custom state management needed

---

## CLI Commands

### New Commands

#### Issue Management

```bash
sow issue list [--state open|closed|all]
# Lists all issues with 'sow' label
# Uses: gh issue list --label sow
# Output: Simple table (number, title, state, URL)

sow issue show <number>
# Shows detailed issue information
# Uses: gh issue view <number> --json title,body,labels,state
# Output: Formatted issue details

sow issue check <number>
# Checks if issue has linked branches (claimed or not)
# Uses: gh issue develop --list <number>
# Output: "Available" or "Claimed on branch <name>"

sow issue create
# Interactive issue creation (future: orchestrator-assisted)
# Uses: gh issue create --label sow --title "..." --body "..."
# Output: Created issue number
```

### Modified Commands

#### Project Initialization

```bash
# NEW: Create project from issue
sow project init --issue <number> [--branch-name <name>]

# Flow:
# 1. Fetch issue via: gh issue view <number> --json title,body,labels
# 2. Verify 'sow' label exists
# 3. Check for linked branches: gh issue develop --list <number>
# 4. If branches exist: ERROR with helpful message
# 5. If empty: Create branch via gh issue develop <number> --checkout [--name]
# 6. Initialize project state with github_issue: <number>
# 7. Continue normal project initialization

# EXISTING: Create project without issue (still supported)
sow project init
```

#### Project Status

```bash
sow project status

# If github_issue exists in state:
# - Show linked issue number and title
# - Include issue URL
# - Show issue state (open/closed)
```

---

## Workflows

### 1. Decomposing Large Work (Future)

**Orchestrator-assisted issue creation:**

```
User: "I want to add comprehensive authentication system"

Orchestrator: "This is very large work. Let me help decompose into projects."

[Discovery/Design phase or special decomposition mode]

Orchestrator: "I recommend 4 projects:
1. Add User model and database schema
2. Implement JWT token service
3. Create auth middleware and routes
4. Add frontend login/signup flows

Should I create GitHub issues for these?"

User: "Yes"

Orchestrator:
→ gh issue create --label sow --title "Add User model..." --body "..."
→ gh issue create --label sow --title "Implement JWT..." --body "..."
→ gh issue create --label sow --title "Create auth middleware..." --body "..."
→ gh issue create --label sow --title "Add frontend login..." --body "..."

Orchestrator: "Created issues #123, #124, #125, #126. Ready to start with #123?"
```

### 2. Discovering Available Work

```bash
# List all sow projects
sow issue list

OUTPUT:
NUMBER  TITLE                                STATE
#123    Add User model and database schema   open
#124    Implement JWT token service          open
#125    Create auth middleware and routes    open
#126    Add frontend login/signup flows      open

# Check if issue is claimed
sow issue check 123

OUTPUT (if available):
Issue #123: Add User model and database schema
Status: Available - no branches linked

OUTPUT (if claimed):
Issue #123: Add User model and database schema
Status: Claimed
Branch: 123-add-user-model
URL: https://github.com/org/repo/tree/123-add-user-model
```

### 3. Creating Project from Issue

```bash
# Claim and start work on issue
sow project init --issue 123

FLOW:
1. Fetches issue #123
2. Checks for 'sow' label → ✓
3. Checks for linked branches → none found
4. Creates branch: gh issue develop 123 --checkout
   → Branch: 123-add-user-model
   → Linked to issue in GitHub
   → Checked out locally
5. Creates .sow/project/state.yaml with github_issue: 123
6. Continues with truth table → phase selection → work begins

EXAMPLE ERROR (if already claimed):
Error: Issue #123 already has a linked branch
Branch: 123-add-user-model (created by @alice)
To work on this project: git checkout 123-add-user-model && sow project status
```

### 4. Completing Project Linked to Issue

```bash
# Normal project flow continues...
# When finalize phase creates PR:

sow project finalize
→ Runs tests, updates docs
→ Creates PR via: gh pr create --title "..." --body "..."
→ If github_issue exists: adds "Closes #123" to PR body
→ GitHub auto-closes issue when PR merges
→ Deletes .sow/project/ as normal
```

---

## Schema Changes

### Project State (`project_state.cue`)

```cue
project: {
    name: string
    branch: string
    description: string

    // NEW: Optional GitHub issue link
    github_issue: *null | int

    created_at: time.Time
    updated_at: time.Time
}
```

**Simple addition** - just store the issue number if linked.

---

## Implementation Phases

### Phase 1: Read-Only Integration (Foundation)

**Goal**: Query GitHub issues without modifying state

**Tasks:**
- [ ] Add `sow issue list` command
  - Use `gh issue list --label sow --json number,title,url,state`
  - Format output as table
  - Support `--state` filter
- [ ] Add `sow issue show <number>` command
  - Use `gh issue view <number> --json title,body,labels,state`
  - Display formatted issue details
- [ ] Add `sow issue check <number>` command
  - Use `gh issue develop --list <number>`
  - Parse output to determine if claimed
  - Show branch name if linked
- [ ] Add GitHub CLI detection
  - Check if `gh` is installed and authenticated
  - Provide helpful error messages if missing

**Validation:**
- Can list sow issues from test repo
- Can show issue details
- Can detect linked branches
- Graceful errors when `gh` unavailable

---

### Phase 2: Project-Issue Linking (Core)

**Goal**: Link projects to issues during creation

**Tasks:**
- [ ] Add `github_issue` field to `project_state.cue` schema
- [ ] Extend `sow project init` with `--issue <number>` flag
  - Fetch issue via `gh issue view`
  - Validate `sow` label exists
  - Check for linked branches (fail if found)
  - Create branch via `gh issue develop <number> --checkout`
  - Initialize project with `github_issue: <number>`
- [ ] Update `sow project status` to show linked issue
  - Display issue number, title, URL if linked
  - Fetch current issue state from GitHub
- [ ] Handle branch naming
  - Default: auto-generated by `gh` (`<number>-<title>`)
  - Optional: `--branch-name` flag to override

**Validation:**
- Can create project from issue
- Branch automatically linked in GitHub UI
- Duplicate creation fails gracefully
- Project status shows issue info
- Works without issue (backward compatible)

---

### Phase 3: Finalize Integration (Completion)

**Goal**: Auto-close issues when PRs merge

**Tasks:**
- [ ] Extend finalize phase PR creation
  - Check if `github_issue` exists in project state
  - Add `Closes #<number>` to PR body
  - GitHub auto-closes issue on merge
- [ ] Test PR creation workflow
  - Verify issue closes on merge
  - Verify issue remains open if PR not merged

**Validation:**
- PR body includes issue reference
- Issue closes when PR merges
- Issue stays open if PR closed without merge

---

### Phase 4: Orchestrator Integration (Enhanced UX)

**Goal**: Orchestrator awareness of issues

**Tasks:**
- [x] Extend `/sow-greet` command
  - Query open sow issues via GitHub CLI
  - Show count and suggest exploring
  - Add "Work on an existing issue" to menu options
  - Gracefully handle gh CLI unavailable
- [x] Add issue context to greet template
  - Operator mode: Shows available issues with instructions
  - Orchestrator mode: Mentions issues available for future work
  - Provides guidance on using `sow issue` commands
- [ ] (Future) Issue creation workflow
  - `/decompose` or similar command
  - Orchestrator helps break down large work
  - Creates multiple issues with descriptions

**Implementation:**
- Modified `cli/cmd/greet.go` to query GitHub issues
- Added `OpenIssues` and `GHAvailable` to `GreetContext`
- Updated `cli/cmd/templates/greet.tmpl` with issue awareness
- Conditional display: only shows if gh CLI available and issues exist
- Natural integration into existing greet flow

**Validation:**
- ✓ Greet shows issue count when issues present
- ✓ Natural flow from issue discovery → exploration → project creation
- ✓ Orchestrator aware of available work
- ✓ Graceful degradation when gh unavailable

---

### Phase 5: Advanced Features (Future)

**Goal**: Enhanced workflows and tooling

**Tasks:**
- [ ] `sow issue create` - Interactive creation
  - Prompt for title, description, labels
  - Use templates if repo has issue templates
- [ ] Orchestrator-assisted decomposition
  - Special mode for breaking down large work
  - Generate multiple issue descriptions
  - Create issues via `gh issue create`
- [ ] Dependency tracking
  - Add metadata to issue body (YAML frontmatter?)
  - Track which issues depend on others
  - Show dependency graph
- [ ] Team metrics
  - Show issues by assignee
  - Show completion rate
  - Integration with GitHub Projects

---

## Edge Cases & Error Handling

### Issue Not Found

```bash
sow project init --issue 999
ERROR: Issue #999 not found
```

### Missing `sow` Label

```bash
sow project init --issue 123
ERROR: Issue #123 does not have 'sow' label
To fix: Add 'sow' label via GitHub UI or: gh issue edit 123 --add-label sow
```

### Already Claimed

```bash
sow project init --issue 123
ERROR: Issue #123 already has a linked branch
Branch: 123-add-auth (created by @alice)
To work on this: git checkout 123-add-auth && sow project status
```

### `gh` Not Installed

```bash
sow issue list
ERROR: GitHub CLI (gh) not found
Install: https://cli.github.com/
```

### `gh` Not Authenticated

```bash
sow issue list
ERROR: GitHub CLI not authenticated
Run: gh auth login
```

### Network Unavailable

```bash
sow issue list
ERROR: Cannot reach GitHub API
Check your network connection
```

### Permissions Insufficient

```bash
sow issue create
ERROR: Insufficient permissions to create issues
Requires: write access to repository
```

### Branch Creation Failure

```bash
sow project init --issue 123
ERROR: Failed to create branch via gh issue develop
Possible causes:
- Branch name already exists
- Network error
- Insufficient permissions
```

---

## Testing Strategy

### Unit Tests

- Schema validation for `github_issue` field
- Issue number parsing and validation
- Error message formatting

### Integration Tests

- Mock `gh` CLI responses
- Test issue list parsing
- Test branch linking flow
- Test error scenarios

### E2E Tests

- Require actual GitHub repo with test issues
- Create project from issue
- Verify branch linking in GitHub
- Create PR and verify issue reference
- Clean up test data

### Manual Testing Checklist

- [ ] List issues shows correct data
- [ ] Check issue shows linked branches
- [ ] Create project from available issue succeeds
- [ ] Create project from claimed issue fails gracefully
- [ ] Project status shows issue info
- [ ] PR includes issue reference
- [ ] Issue closes when PR merges

---

## Security & Privacy Considerations

### Authentication

- Relies on `gh` CLI authentication
- No PAT storage in sow
- User manages credentials via `gh auth login`

### Repository Access

- Only works on repos user has access to
- Issue creation requires write permissions
- Read operations require read permissions

### Data Privacy

- No sensitive data stored in sow
- Issue numbers stored in project state (local)
- All data visible in GitHub (no secrets)

---

## Documentation Updates

### New Documentation

- [ ] `docs/GITHUB_INTEGRATION.md` - Complete integration guide
  - How issue integration works
  - Workflows and examples
  - Troubleshooting

### Updated Documentation

- [ ] `docs/CLI_REFERENCE.md` - Add issue commands
- [ ] `docs/USER_GUIDE.md` - Add issue workflows
- [ ] `docs/OVERVIEW.md` - Mention issue integration
- [ ] `docs/PROJECT_LIFECYCLE.md` - Add issue creation workflow
- [ ] `docs/SCHEMAS.md` - Document `github_issue` field
- [ ] `README.md` - Update features and quick start

---

## Success Criteria

### Phase 1 Success

- ✓ Can list all sow issues from GitHub
- ✓ Can check if specific issue has linked branches
- ✓ Graceful errors when `gh` unavailable

### Phase 2 Success

- ✓ Can create project from issue
- ✓ Branch automatically linked in GitHub UI
- ✓ Duplicate creation fails with helpful message
- ✓ Project status shows linked issue
- ✓ Backward compatible (projects without issues still work)

### Phase 3 Success

- ✓ PRs include `Closes #<number>` when linked
- ✓ Issues auto-close when PRs merge
- ✓ Clean project deletion still works

### Phase 4 Success

- ✓ Greet command shows issue count
- ✓ Natural flow from issue discovery to project creation
- ✓ Orchestrator suggests issues when relevant

---

## Future Enhancements

### Issue Templates

- Support GitHub issue templates
- Pre-fill project structure from template

### Milestones & Projects

- Group issues by GitHub milestones
- Integration with GitHub Projects (beta)

### Dependencies

- Track cross-issue dependencies
- Visualize dependency graph
- Warn if starting issue with unmet dependencies

### Orchestrator Decomposition

- Special mode for breaking down large work
- AI-assisted issue generation
- Dependency inference

### Team Features

- Show issues by assignee
- Team velocity metrics
- Progress tracking across issues

### Advanced Queries

```bash
sow issue list --assignee @me
sow issue list --milestone v2.0
sow issue list --project board-name
```

---

## Open Questions

### 1. Branch Naming Flexibility

**Question**: Should we allow custom branch names or always use `gh` auto-generation?

**Options**:
- A: Always auto-generate (`<number>-<title>`)
- B: Allow override via `--branch-name` flag
- C: Prompt user for confirmation/customization

**Recommendation**: Option B - Auto-generate by default, allow override for edge cases

---

### 2. Issue Linking Requirement

**Question**: Should issue linking be required or optional?

**Options**:
- A: Required - every project must link to an issue
- B: Optional - projects can exist without issues
- C: Configurable - repo setting determines requirement

**Recommendation**: Option B (initially), Option C (future) - Optional but encouraged

---

### 3. Label Management

**Question**: Who manages the `sow` label?

**Options**:
- A: Manual - users add label via GitHub
- B: Automatic - `sow issue create` adds label
- C: Validation - CLI validates and adds if missing

**Recommendation**: Option B - Automatic addition, validation on project creation

---

### 4. Multiple Branches Per Issue

**Question**: Should we support multiple branches for one issue?

**Options**:
- A: Allow - different approaches to same problem
- B: Forbid - first branch wins, others must use different issue
- C: Warn - show existing branches, let user decide

**Recommendation**: Option B initially (simpler), Option C if needed

---

## Timeline Estimate

**Phase 1**: 2-3 days
- Issue list/show/check commands
- `gh` CLI integration layer
- Basic error handling

**Phase 2**: 3-4 days
- Schema changes
- Project init with issue linking
- Branch creation via `gh issue develop`
- Comprehensive error handling

**Phase 3**: 1-2 days
- Finalize phase PR integration
- Issue auto-close testing

**Phase 4**: 2-3 days
- Orchestrator greet enhancement
- Issue suggestion flow
- UX polish

**Total**: 8-12 days for core functionality (Phases 1-3)

---

## Risks & Mitigations

### Risk: `gh` CLI Changes

**Impact**: Breaking changes to `gh issue develop` syntax
**Likelihood**: Low (stable feature)
**Mitigation**: Version detection, fallback to manual linking

### Risk: GitHub API Rate Limits

**Impact**: List commands fail under heavy usage
**Likelihood**: Low (label filtering is cheap)
**Mitigation**: Cache results, show rate limit status

### Risk: Network Dependency

**Impact**: Commands fail offline
**Likelihood**: High (by design)
**Mitigation**: Clear error messages, graceful degradation

### Risk: Permission Issues

**Impact**: Users can't create branches/issues
**Likelihood**: Medium (varies by repo)
**Mitigation**: Clear permission error messages, documentation

---

## Related Documentation

- `docs/PROJECT_LIFECYCLE.md` - Project initialization
- `docs/CLI_DESIGN.md` - CLI architecture
- `docs/ARCHITECTURE.md` - System design
- `docs/SCHEMAS.md` - State file schemas

---

## Approval & Next Steps

**Approval Required**: Architecture review, security review

**Next Steps**:
1. Review this plan with team
2. Create GitHub project for tracking
3. Begin Phase 1 implementation
4. Iterate based on feedback

---

**Document Version**: 1.0
**Last Updated**: 2025-10-20
**Status**: Ready for Implementation
