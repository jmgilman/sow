---

## Guidance: Summarizing Findings

You are in the **Summarizing** state. All research topics are resolved. Now synthesize the **user-approved findings** into comprehensive summary document(s).

**Remember**: You're drafting summaries based on investigations the user already reviewed and approved. Your job is to synthesize insights, not make new research decisions.

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

1. **Review all user-approved research outputs**:
   - Read task logs: `.sow/project/phases/exploration/tasks/*/log.md`
   - Review artifacts created during research
   - These represent findings the user already reviewed and approved
   - Identify patterns and insights across topics

2. **Write summary document(s)**:
   - Save to: `.sow/project/phases/exploration/outputs/`
   - Use descriptive filenames (e.g., `findings.md`, `authentication-research.md`)

3. **Register summaries**:
   ```bash
   sow output add --type summary --path "phases/exploration/outputs/findings.md"
   ```

4. **Present draft to user for review**:
   - Show the summary you've created
   - User reviews summary quality and completeness
   - User either:
     - **Approves**: `sow output set --index <N> approved true`
     - **Requests revisions**: Provides feedback on what to change

5. **Iterate based on feedback**:
   - If user requests changes, revise summary accordingly
   - Re-present updated summary
   - Continue until user approves

6. **Advance when approved**:
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

**You draft, user approves**:
- Draft summaries based on user-approved research
- Present draft to user
- User reviews for comprehensiveness, accuracy, clarity
- Revise based on feedback until user approves

User approval means:
- Summary is comprehensive and accurate
- Insights are clearly communicated
- Ready to become permanent knowledge artifact

**Do not**:
- Advance to Finalizing without user approval
- Make assumptions about whether summary is sufficient
- Skip user review
