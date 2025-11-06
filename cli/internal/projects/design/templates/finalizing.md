---

## Guidance: Finalizing

You are in the **Finalizing** state of a design project. All documents are approved. Your focus is moving artifacts to permanent locations and completing the project.

### Finalization Tasks

When entering the Finalizing state, three tasks are automatically created:

1. **Move approved documents to targets**
2. **Create PR with design artifacts**
3. **Delete .sow/project/ directory**

### Moving Documents

Move approved design documents to their permanent locations:

**For knowledge base documents**:
```bash
# Copy to .sow/knowledge/
cp project/architecture.md .sow/knowledge/designs/architecture-name.md
cp project/adr-001.md .sow/knowledge/adrs/001-decision-title.md
```

**For repository documentation**:
```bash
# Copy to docs/ or appropriate location
cp project/implementation-guide.md docs/guides/feature-name.md
```

**Complete the task when done**:
```bash
sow task set --id move-docs status completed
```

### Creating Pull Request

Draft a PR summarizing the design work:

1. **Review all approved documents**
2. **Create PR body** with:
   - Design summary
   - Links to design documents
   - Key decisions made
   - Next steps for implementation

3. **Use gh CLI**:
   ```bash
   gh pr create --title "Design: <project-name>" --body "$(cat <<'EOF'
   ## Design Summary
   <1-3 paragraphs summarizing the design>

   ## Documents
   - Architecture: [link to doc]
   - ADR 001: [link to doc]
   - Implementation Guide: [link to doc]

   ## Key Decisions
   - Decision 1
   - Decision 2

   ## Next Steps
   - [ ] Review and approve design
   - [ ] Create implementation project
   - [ ] Begin development

   Generated with Claude Code
   EOF
   )"
   ```

4. **Complete the task**:
   ```bash
   sow task set --id create-pr status completed
   ```

### Cleanup

Delete the project workspace:

```bash
rm -rf .sow/project/
sow task set --id cleanup status completed
```

**Important**: Only delete after documents are moved and PR is created.

### Advancement Criteria

You can advance to Completed when:
- All three finalization tasks are completed
- No tasks remain in pending or in_progress status

Ready to complete the design project? Run:
```bash
sow project advance
```

### Tips

- **Verify targets**: Confirm document destinations before moving
- **Preserve metadata**: Include document type and creation date in headers
- **Comprehensive PR**: Write detailed PR description for reviewers
- **Clean branch**: Ensure only permanent artifacts remain after cleanup
- **Track decisions**: Link to ADRs from architecture documents
