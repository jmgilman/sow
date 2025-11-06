---

## Guidance: Summarizing Findings

You are in the **Summarizing** state. All research topics are resolved. Now synthesize findings into comprehensive summary document(s).

### Creating Summaries

Summary documents synthesize insights from all research topics.

**Single summary** (recommended for focused exploration):
- Create one comprehensive document covering all findings
- Organize by theme or chronologically
- Include key insights, decisions, recommendations

**Multiple summaries** (for complex explorations):
- Create separate documents for distinct themes
- Each summary covers a subset of topics
- Maintain cross-references between summaries

### Summary Structure

Recommended sections:

```markdown
# Exploration: <Project Name>

## Context
Brief description of what was explored and why

## Key Findings
Main insights discovered during research

## Research Topics
### <Topic Name>
Summary of findings for this topic

[Repeat for each completed topic]

## Recommendations
Actionable next steps or decisions based on findings

## References
Links to:
- Task directories with detailed findings
- External resources discovered
- Related documentation
```

### Workflow

1. **Review all research outputs**:
   - Read task logs: `.sow/project/phases/exploration/tasks/*/log.md`
   - Review any artifacts created during research
   - Identify patterns and insights

2. **Write summary document(s)**:
   - Save to: `.sow/project/phases/exploration/outputs/`
   - Use descriptive filenames (e.g., `findings.md`, `authentication-research.md`)

3. **Register summaries**:
   ```bash
   sow output add --type summary --path "phases/exploration/outputs/findings.md"
   ```

4. **Present to user**:
   - User reviews summary quality
   - User approves when satisfied:
     ```bash
     sow output set --index <N> approved true
     ```

5. **Advance when approved**:
   ```bash
   sow project advance  # Guard: at least one summary approved
   ```

### Quality Guidelines

**Synthesis over listing**: Don't just list what you found - synthesize insights and draw conclusions.

**Context matters**: Explain why findings are significant and how they relate to project goals.

**Actionable**: Include clear recommendations or next steps based on findings.

**Self-contained**: Reader should understand the exploration without reading all task logs.

**Preserve depth**: Link to task directories for readers who want detailed findings.

### Approval Process

User approval means:
- Summary is comprehensive and accurate
- Insights are clearly communicated
- Ready to become permanent knowledge artifact

If summary needs revision:
- User provides feedback
- Update document based on feedback
- Re-register if path changed, or user re-approves same artifact
