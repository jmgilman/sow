# Task Index: Project SDK Builder & Configuration

## Overview

This project implements the fluent builder API, options pattern, registry, OnAdvance configuration, BuildMachine with guard closure binding, and two-tier validation for the Project SDK.

**Foundation:** Unit-002 (state types & persistence) is complete. This unit builds on top of that foundation.

**Design Reference:** `.sow/knowledge/designs/project-sdk-implementation.md`

---

## Task List

| ID  | Task Name | Estimated Time | Dependencies |
|-----|-----------|----------------|--------------|
| 010 | Core Configuration Types and Options Pattern | 1.5h | None |
| 020 | Builder API Implementation | 2h | 010 |
| 030 | Registry Implementation | 1h | 020 |
| 040 | BuildMachine with Closure Binding | 2h | 030 |
| 050 | OnAdvance Configuration and Project.Advance() | 2h | 040 |
| 060 | Two-Tier Validation Implementation | 2.5h | 020 |
| 070 | Integration Test - Complete Project Type Configuration | 1.5h | 010-060 |

**Total Estimated Time:** 12.5 hours

---

## Task Files

- `010-core-config-types-options.md` - Configuration structures and options pattern
- `020-builder-api.md` - Fluent builder API implementation
- `030-registry.md` - Global registry for project types
- `040-build-machine-closure-binding.md` - State machine with bound guards
- `050-onadvance-and-advance.md` - Event determination and generic advance
- `060-two-tier-validation.md` - Artifact type and metadata validation
- `070-integration-test.md` - Complete workflow integration test

---

## Testing Approach

All tasks follow **Test-Driven Development (TDD)** with **behavior-only testing**:
- Write tests first
- Test observable behavior, not internal implementation details
- No testing of private methods or internal structures
- Focus on public API contracts and outcomes
