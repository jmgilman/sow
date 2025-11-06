# Exploration Project Type

**Project**: {{.Name}}
**Type**: exploration
**Branch**: {{.Branch}}
{{if .Description}}**Description**: {{.Description}}
{{end}}

This project follows the **exploration workflow**: Active → Summarizing → Finalizing.

Exploration projects are designed for open-ended research, investigation, and knowledge gathering with minimal structure.

---

## Phase Overview

### 1. Exploration Phase (Active → Summarizing)

The exploration phase has two states:

#### Active State

**Your role**: Facilitate research and investigation

**Workflow**:
1. **Identify research topics**:
   - Create tasks as research topics emerge
   - Each task represents an area to investigate
   ```bash
   sow task add "Research authentication patterns" --id 010
   ```

2. **Investigate topics**:
   - Work directly on research (no spawning required for simple exploration)
   - For complex investigation, can spawn agents
   - Document findings in `.sow/project/phases/exploration/tasks/{id}/`

3. **Resolve topics**:
   - Mark completed when findings are documented
   - Mark abandoned if topic proves unfruitful
   ```bash
   sow task complete 010
   # or
   sow task abandon 010 --reason "Not relevant to current scope"
   ```

4. **Advance when all topics resolved**:
   ```bash
   sow project advance  # Guard: all tasks completed or abandoned
   ```

#### Summarizing State

**Your role**: Synthesize findings into summary documents

**Workflow**:
1. **Review all research findings**:
   - Read all completed task logs and outputs
   - Identify key insights and patterns

2. **Create summary document(s)**:
   - Write to `.sow/project/phases/exploration/outputs/`
   - Can create single summary or multiple thematic summaries
   - Markdown format recommended

3. **Register summaries**:
   ```bash
   sow output add --type summary --path "phases/exploration/outputs/findings.md"
   ```

4. **User reviews and approves**:
   ```bash
   sow output set --index 0 approved true
   ```

5. **Advance when summaries approved**:
   ```bash
   sow project advance  # Guard: at least one summary approved
   ```

---

### 2. Finalization Phase

**State**: Finalizing

**Your role**: Move artifacts to permanent location and create PR

**Workflow**:
1. **Create finalization tasks**:
   ```bash
   sow task add "Move summaries to knowledge base" --id move-artifacts --phase finalization
   sow task add "Create pull request" --id create-pr --phase finalization
   ```

2. **Move artifacts**:
   - Copy approved summaries to `.sow/knowledge/explorations/`
   - Preserve project context and metadata

3. **Create PR**:
   - Draft PR body summarizing exploration
   - Include links to findings documents
   - Use `gh pr create` to submit

4. **Complete tasks and advance**:
   ```bash
   sow task complete move-artifacts
   sow task complete create-pr
   sow project advance  # Guard: all finalization tasks complete
   ```

---

## Key Characteristics

### Minimal Structure
- No planning phase - topics emerge organically
- No formal review - user approves summaries directly
- Flexible task management

### Research-Focused
- Tasks represent topics, not implementation work
- Emphasis on documentation and knowledge capture
- Abandoned topics are acceptable

### Summary-Driven
- At least one summary document required
- Can create multiple summaries for different themes
- Summaries become permanent knowledge artifacts

---

## State Transition Logic

```
NoProject
  → (project_init) → Active

Active
  → (all_tasks_resolved, guard: all tasks completed/abandoned)
  → Summarizing

Summarizing
  → (summaries_approved, guard: ≥1 summary approved)
  → Finalizing

Finalizing
  → (finalization_complete, guard: all tasks completed)
  → Completed
```

---

## Critical Notes

### Flexible Research Process
Unlike standard projects, exploration projects are intentionally unstructured. Add topics as needed, abandon freely, iterate naturally.

### Direct Work vs Agent Delegation
For lightweight research, work directly. For complex investigation or code exploration, spawn agents.

### Summary Quality Matters
Summaries become permanent knowledge. Take time to synthesize insights, not just list findings.

### User Approval Required For
- Summary artifacts (must be approved before advancing)
- Finalization tasks (user confirms artifacts moved correctly)

---

## Your Current State

The tactical guidance for your current state follows below (if provided).
