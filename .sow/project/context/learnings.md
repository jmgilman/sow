# Project Learnings: milestone-1-foundation

This file captures insights, discoveries, and lessons learned while building Milestone 1.

---

## 2025-10-12

### Plugin Structure Clarification

**Discovery**: The architecture docs initially referred to `.claude/` as the execution layer,
but the actual Claude Code plugin structure requires:
- Source files in `plugin/` directory (marketplace repo)
- Installation copies `plugin/` contents to `.claude/` (user repos)

**Impact**: Updated all architecture docs to clarify this distinction:
- FILE_STRUCTURE.md - Added "How It Gets There" section
- ARCHITECTURE.md - Added development note
- ROADMAP.md - Clarified Milestone 1 deliverables
- DISTRIBUTION.md - Rewrote Package Structure section

**Lesson**: When dogfooding, always validate assumptions against actual platform
requirements early.

---

## Bootstrap Command Enhancement

**Action**: Expanded `.claude/commands/bootstrap.md` to be more comprehensive.

**Rationale**: Future agents need clear guidance on manual project management during
bootstrap phase. Original version was too brief.

**Changes**:
- Added detailed responsibilities section (7 areas)
- Included workflow example
- Listed important constraints
- Clarified tools available
- Added final reminders

**Outcome**: Future agents will have much clearer understanding of their role.

---
