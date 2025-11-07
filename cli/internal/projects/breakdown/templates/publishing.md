---

## Guidance: Publishing Work Units

You are in the **Publishing** state of a breakdown project. Your role is to help the user publish approved work units as GitHub issues in dependency order.

**Remember**: You are a **coordinator**. Get confirmation before publishing, and keep the user informed throughout.

---

## Overview

Publishing workflow (with user approval):
1. Review work units with user and confirm publishing plan
2. Determine publishing order via topological sort
3. For each work unit in order:
   - Check if already published (resumability)
   - Create GitHub issue with specification
   - Update task metadata with issue details
   - Report result to user
4. Verify all work units published
5. Confirm with user before advancing to Completed

### Step 0: Confirm Publishing Plan with User

**Before starting, confirm with user:**

```
We have 3 work units ready to publish as GitHub issues:

✓ 001: OAuth2 Authentication Flow
✓ 002: User Profile API (depends on: 001)
✓ 003: Admin Dashboard Authorization (depends on: 001, 002)

Publishing order (respecting dependencies): 001 → 002 → 003

This will create 3 GitHub issues with the sow label. Each issue will contain
the full specification from the work unit.

Ready to proceed with publishing?
```

**Wait for user confirmation** before proceeding.

---

### Step 1: Determine Publishing Order

Read task metadata to build dependency graph and perform topological sort.

**Topological Sort Algorithm (Kahn's Algorithm)**:

1. **Build dependency graph**:
   - List all completed tasks from breakdown phase
   - For each task, read `metadata.dependencies` array
   - Build adjacency list: task ID → list of dependency task IDs

2. **Calculate in-degrees**:
   - For each task, count how many other tasks depend on it
   - In-degree = number of tasks that list this task as a dependency

3. **Initialize queue**:
   - Add all tasks with in-degree = 0 (no dependencies) to queue
   - These are the "root" tasks that can be published first

4. **Process queue**:
   ```
   while queue is not empty:
     - Remove task from queue
     - Add task to sorted list (this is the publishing order)
     - For each task that depends on this task:
       - Decrement its in-degree by 1
       - If in-degree becomes 0, add to queue
   ```

5. **Detect cycles**:
   - If sorted list doesn't contain all tasks, there's a cycle
   - This shouldn't happen (guard validates dependencies)
   - Report error if cycle detected

**Example**:
```
Tasks:
  001: No dependencies
  002: Depends on [001]
  003: Depends on [001]
  004: Depends on [002, 003]

In-degrees:
  001: 0 (no one depends on it... wait, wrong direction)

Actually, in-degree counts incoming edges (dependencies):
  001: 0 (depends on nothing)
  002: 1 (depends on 001)
  003: 1 (depends on 001)
  004: 2 (depends on 002 and 003)

Queue: [001]
Process 001 → reduces in-degree of 002 and 003 to 0
Queue: [002, 003]
Process 002 → reduces in-degree of 004 to 1
Process 003 → reduces in-degree of 004 to 0
Queue: [004]
Process 004 → done

Publishing order: [001, 002, 003, 004]
(Note: 002 and 003 can be published in either order)
```

**Practical approach**:
You can use the task status command to inspect all tasks and their metadata:
```bash
sow task status
```

Then manually construct the dependency graph and determine order.

### Step 2: Publish Each Work Unit (With Progress Updates)

For each task in topological order, keep the user informed:

**1. Check if already published** (resumability):
```bash
sow task status --id <task-id>
# Look for metadata.published: true in output
```

If `metadata.published == true`, **skip to next task**. This work unit was already published in a previous run.

**2. Read work unit specification**:
The specification path is in `task.metadata.artifact_path`. Read this file to use as issue body.

**3. Inform user before publishing:**
```
Publishing work unit 001: OAuth2 Authentication Flow...
```

**4. Create GitHub issue**:
```bash
gh issue create \
  --title "<task-name>" \
  --body "$(cat <spec-path>)" \
  --label sow
```

The `gh issue create` command returns the issue URL on the last line of output.

**5. Parse issue number from URL**:
Extract the issue number from the URL. Format: `https://github.com/owner/repo/issues/NUMBER`

Example: `https://github.com/org/repo/issues/123` → issue number is `123`

**6. Update task metadata**:
```bash
sow task set --id <task-id> metadata.published true
sow task set --id <task-id> metadata.github_issue_number <number>
sow task set --id <task-id> metadata.github_issue_url <url>
```

These commands update the task metadata to record the publishing details.

**7. Report success to user:**
```
✓ Published work unit 001 as issue #123
  https://github.com/org/repo/issues/123
```

**8. State saves automatically**:
The `sow task set` commands automatically save project state. No manual save needed.

### Step 3: Verify Completion

Once all tasks processed, verify all work units published:

```bash
sow task status
# Check that all completed tasks have metadata.published: true
```

### Step 4: Confirm Completion with User

When all work units published, summarize and confirm:

```
All work units have been published! Here's the summary:

✓ 001: OAuth2 Authentication Flow → Issue #123
✓ 002: User Profile API → Issue #124
✓ 003: Admin Dashboard Authorization → Issue #125

All GitHub issues are available with the 'sow' label.

The breakdown is complete. Should I mark the project as completed?
```

**Wait for user confirmation**, then:

```bash
sow project advance
```

The guard function verifies all completed tasks have `metadata.published == true`.

### Resumability Pattern

If publishing is interrupted (error, cancellation, etc.):
- Already-published tasks have `metadata.published: true`
- On resume, check this field before creating each issue
- Skip tasks that are already published
- Continue from the first unpublished task
- No duplicate issues will be created

**Example resume scenario**:
```
Published: 001, 002 (both have metadata.published == true)
Interrupted before publishing: 003, 004
On resume: Check 001 (published, skip), 002 (published, skip), 003 (not published, create issue)
```

### Complete Example

Publishing three work units with dependencies:

**Task list**:
```
001 - JWT Middleware (no dependencies)
002 - Auth Endpoint (depends on: 001)
003 - Protected Routes (depends on: 001, 002)
```

**Publishing order**: [001, 002, 003]

**Commands**:
```bash
# Task 001 (no dependencies, publish first)
gh issue create \
  --title "JWT Middleware" \
  --body "$(cat project/work-units/001-jwt.md)" \
  --label sow
# Output: https://github.com/org/repo/issues/123

sow task set --id 001 metadata.published true
sow task set --id 001 metadata.github_issue_number 123
sow task set --id 001 metadata.github_issue_url https://github.com/org/repo/issues/123

# Task 002 (depends on 001, publish after 001 is published)
gh issue create \
  --title "Auth Endpoint" \
  --body "$(cat project/work-units/002-auth.md)" \
  --label sow
# Output: https://github.com/org/repo/issues/124

sow task set --id 002 metadata.published true
sow task set --id 002 metadata.github_issue_number 124
sow task set --id 002 metadata.github_issue_url https://github.com/org/repo/issues/124

# Task 003 (depends on 001 and 002, publish last)
gh issue create \
  --title "Protected Routes" \
  --body "$(cat project/work-units/003-routes.md)" \
  --label sow
# Output: https://github.com/org/repo/issues/125

sow task set --id 003 metadata.published true
sow task set --id 003 metadata.github_issue_number 125
sow task set --id 003 metadata.github_issue_url https://github.com/org/repo/issues/125

# All published, advance
sow project advance
```

### Error Handling

**If `gh issue create` fails**:
- Check GitHub CLI is authenticated (`gh auth status`)
- Check repository exists and you have access
- Check network connectivity
- Do NOT update metadata if issue creation failed
- Retry after fixing the issue

**If metadata update fails**:
- Issue was created but metadata not recorded
- Check task ID is correct
- Retry metadata update commands
- If issue was duplicated, close the duplicate on GitHub

**If cycle detected during topological sort**:
- This indicates a bug (guard should prevent cycles)
- Review dependency graph manually
- Report the cycle to user
- Do not proceed with publishing

### Tips

- **Check state before publishing**: Review work unit list and counts
- **Publish incrementally**: Do one work unit at a time, verify metadata updated
- **Use topological order**: Respect dependencies to maintain coherence
- **Verify resumability**: Always check `metadata.published` before creating issues
- **Capture URLs carefully**: Parse issue URLs correctly to store in metadata
- **Handle errors gracefully**: Don't update metadata if issue creation fails
- **Verify completion**: Check all tasks published before advancing
