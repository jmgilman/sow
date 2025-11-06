---

## Guidance: Active Research

You are in the **Active** state of an exploration project. Your focus is identifying research topics and investigating them.

### Creating Research Topics

Add topics as they emerge during exploration:

```bash
sow task add "Topic name" --id <3-digit-id>
```

Examples:
- "Evaluate authentication libraries"
- "Research deployment options"
- "Investigate performance bottlenecks"
- "Understand user workflow patterns"

### Investigating Topics

For each research topic:

1. **Start investigation**:
   ```bash
   sow task start <id>
   ```

2. **Document findings**:
   - Write findings to task directory: `.sow/project/phases/exploration/tasks/<id>/`
   - Use markdown for notes
   - Include code samples, links, observations

3. **Resolve topic**:
   - Complete when you have sufficient findings:
     ```bash
     sow task complete <id>
     ```
   - Abandon if topic proves irrelevant:
     ```bash
     sow task abandon <id> --reason "Out of scope"
     ```

### Working Style

**Direct work**: For lightweight research, work directly without spawning agents.

**Agent delegation**: For complex investigation, spawn agents:
```
Use Task tool to spawn investigator
Provide topic context and research questions
Review findings when agent completes
```

### Advancement Criteria

You can advance to Summarizing when:
- All research topics are resolved (completed or abandoned)
- At least one topic is completed (not all abandoned)

Ready to advance? Run:
```bash
sow project advance
```

### Tips

- Don't over-plan: Add topics as you discover them
- Document as you go: Capture insights immediately
- Abandon freely: Not all topics pan out
- Stay focused: Keep topic scope narrow and manageable
