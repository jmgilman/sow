---

## Guidance: Finalizing Exploration

You are in the **Finalizing** state. Summaries are approved. Complete finalization tasks to make findings permanent and create a PR.

### Finalization Tasks

Create tasks for finalization steps:

```bash
sow task add "Move summaries to knowledge base" --id move-artifacts --phase finalization
sow task add "Create pull request" --id create-pr --phase finalization
```

### Moving Artifacts to Knowledge Base

Summary documents should become permanent knowledge artifacts.

**Destination**: `.sow/knowledge/explorations/{project-name}/`

**Steps**:
1. Create directory:
   ```bash
   mkdir -p .sow/knowledge/explorations/{project-name}
   ```

2. Copy approved summaries:
   ```bash
   cp .sow/project/phases/exploration/outputs/*.md .sow/knowledge/explorations/{project-name}/
   ```

3. Add metadata file (optional but recommended):
   ```bash
   cat > .sow/knowledge/explorations/{project-name}/metadata.yaml << EOF
   exploration: {project-name}
   date: $(date +%Y-%m-%d)
   branch: {branch-name}
   topics: {count}
   summaries:
     - findings.md
   EOF
   ```

4. Mark task complete:
   ```bash
   sow task complete move-artifacts
   ```

### Creating Pull Request

Exploration PRs capture research context and make findings accessible.

**PR Body Structure**:

```markdown
# Exploration: <Project Name>

## Overview
Brief description of what was explored and why

## Key Findings
3-5 bullet points summarizing main insights

## Artifacts Created
- Link to summary documents in `.sow/knowledge/explorations/`
- Mention any other significant outputs

## Recommendations
Next steps or decisions informed by this exploration

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

**Create PR**:
```bash
gh pr create --title "Exploration: <project name>" --body "$(cat pr-body.md)"
```

**Store PR URL** (optional):
```bash
# If you want to track PR URL in project metadata
sow metadata set pr_url "https://github.com/..."
```

**Mark task complete**:
```bash
sow task complete create-pr
```

### Advancement

Once all finalization tasks are complete:

```bash
sow project advance  # Guard: all finalization tasks completed
```

This transitions to **Completed** state and marks the exploration as finished.

### Cleanup Note

Unlike standard projects, exploration projects typically:
- Keep `.sow/project/` directory (contains full research context)
- Merge PR to preserve findings
- Delete project directory only after merge completes

The project state remains until explicitly cleaned up or after PR merge.

### Tips

**Knowledge organization**: Use descriptive directory names within `.sow/knowledge/explorations/` for easy discovery later.

**PR clarity**: Focus PR body on insights and recommendations, not detailed findings (those are in summary docs).

**Preserve context**: Keep links between PR, summaries, and research task directories for traceability.
