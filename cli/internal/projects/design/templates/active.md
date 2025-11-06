---

## Guidance: Active Design

You are in the **Active** state of a design project. Your focus is planning, drafting, and reviewing design documents.

### Planning Documents

Create tasks for each document you need to produce:

```bash
sow task add "Document title" --id <3-digit-id>
```

Examples:
- "Authentication Architecture Design"
- "API Design Specification"
- "Database Schema ADR"
- "Implementation Guide for Feature X"

**Important**: Create at least one document task before adding artifacts. Each artifact must link to a task.

### Drafting Documents

For each document:

1. **Start drafting**:
   ```bash
   sow task start <id>
   ```

2. **Write the document**:
   - Create document in project workspace (e.g., `project/architecture.md`)
   - Use markdown format
   - Include diagrams, code samples, and detailed specifications
   - Reference input sources when relevant

3. **Register artifact**:
   ```bash
   sow output add --type <doc-type> --path "project/<filename>.md"
   ```

   Common document types:
   - `architecture`: System architecture documentation
   - `design`: Design specifications
   - `adr`: Architecture Decision Records
   - `guide`: Implementation guides

4. **Link artifact to task**:
   ```bash
   sow task set <id> artifact_path "project/<filename>.md"
   sow task set <id> document_type "<doc-type>"
   ```

### Review Workflow

When document draft is ready:

1. **Request review**:
   ```bash
   sow task set <id> status needs_review
   ```

2. **User reviews document**:
   - User reads the document
   - User provides feedback if changes needed
   - User signals approval when satisfied

3. **Iterate if needed**:
   - If changes needed, return to `in_progress`:
     ```bash
     sow task set <id> status in_progress
     ```
   - Make revisions
   - Request review again when ready

4. **Complete when approved**:
   ```bash
   sow task set --id <id> status completed
   ```

   **Note**: Completing the task automatically approves the linked artifact. No separate approval step needed.

### Abandoning Documents

If a document is no longer needed:

```bash
sow task abandon <id> --reason "No longer required"
```

Abandoned documents don't count against advancement readiness.

### Working with Inputs

If your design phase has input artifacts (e.g., exploration findings):
- Reference them in your documents
- Build on insights from previous work
- Link to input sources for traceability

### Advancement Criteria

You can advance to Finalizing when:
- All document tasks are resolved (completed or abandoned)
- At least one document is completed (not all abandoned)

Ready to advance? Run:
```bash
sow project advance
```

The guard will check that all documents are properly resolved before allowing advancement.

### Tips

- **Plan iteratively**: Start with core documents, add more as needed
- **Review early**: Request review on partial drafts to catch issues early
- **Link everything**: Always link artifacts to tasks for proper tracking
- **Use metadata**: Store document type and target location in task metadata
- **Iterate freely**: The review workflow supports multiple revision cycles
- **Reference inputs**: Build on existing exploration work or other designs
