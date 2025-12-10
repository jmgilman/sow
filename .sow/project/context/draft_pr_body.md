# libs/project module

## Intent

Create a new `libs/project` Go module consolidating `cli/internal/sdks/state/` and `cli/internal/sdks/project/` with a storage backend abstraction. This introduces a `Backend` interface for pluggable persistence, decouples project state from `sow.Context`, and provides a `MemoryBackend` for improved testability.

## Status

**Draft** - Implementation in progress

## Progress

- [x] Planning phase
- [ ] Implementation phase
- [ ] Review phase
- [ ] Final checks

---

_This PR body will be updated with full details before marking ready for review._

Generated with [sow](https://github.com/jmgilman/sow)
