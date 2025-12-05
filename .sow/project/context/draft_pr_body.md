# Executor System

## Intent

Implement an executor system that enables sow to spawn AI CLI tools (Claude Code, Cursor) with agent prompts and manage resumable sessions. This allows workers to be invoked programmatically and supports bidirectional communication through session resumption for paused workflows and review iterations.

## Status

**Draft** - Implementation in progress

## Progress

- [x] Planning phase
- [ ] Implementation phase
- [ ] Review phase
- [ ] Final checks

## Scope

- Executor interface (Spawn, Resume, SupportsResumption)
- ExecutorRegistry for managing executor instances
- ClaudeExecutor implementation (invokes `claude` CLI)
- CursorExecutor implementation (invokes `cursor-agent` CLI)
- Task schema update: add `session_id` field and `paused` status

---

_This PR body will be updated with full details before marking ready for review._

Generated with [sow](https://github.com/jmgilman/sow)
