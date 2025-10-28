# 1. Introduction and Goals

## Overview

**sow** (System of Work) is an AI-powered framework for structured software development that orchestrates multiple specialized AI agents through a fixed 5-phase project lifecycle. The system coordinates human-led planning with AI-autonomous execution to build software features systematically and with complete resumability.

The framework addresses the challenge of managing AI agents in software development by providing:
- **Clear phase boundaries** that define when humans lead and when AI operates autonomously
- **Zero-context resumability** through comprehensive filesystem-based state management
- **Multi-agent orchestration** with specialized workers coordinated by an orchestrator
- **Integrated workflow** from exploration through design, implementation, review, and finalization

Primary users include software developers and teams seeking to augment their workflow with AI assistance while maintaining control over planning and architecture decisions.

## Key Functional Requirements

- **Project Lifecycle Management**: 5-phase workflow (Planning → Implementation → Review → Finalize) with state machine-driven transitions
- **Multi-Agent Orchestration**: Orchestrator agent coordinates specialized worker agents (implementer, architect, etc.) via Task tool spawning
- **Zero-Context Resumability**: Complete project and task state persisted to filesystem, enabling any agent to resume any work
- **Operating Modes**: Support exploration (research), design (documentation), breakdown (issue creation), and project (structured implementation) modes
- **External Knowledge Integration**: Reference system for style guides, conventions, and code examples via local caching
- **GitHub Integration**: Issue management, branch linking, and pull request creation via gh CLI
- **Schema-Driven Validation**: CUE schemas generate Go types for type-safe state management

## Quality Goals

| Priority | Quality Attribute         | Specific Target                                                             | Rationale                                                                           |
| -------- | ------------------------- | --------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| 1        | **Resumability**          | 100% of context recoverable from filesystem within 2 seconds                | Core value proposition - enables session independence and multi-developer workflows |
| 2        | **State Consistency**     | Zero state corruption incidents through atomic writes and schema validation | Prevents context loss that would break resumability                                 |
| 3        | **Agent Coordination**    | < 5 second orchestrator spawn-to-execution for worker agents                | Maintains developer flow during task delegation                                     |
| 4        | **CLI Performance**       | < 100ms for non-git operations (logging, state reads)                       | Frequent operations must be instantaneous                                           |
| 5        | **Documentation Quality** | All state files human-readable markdown/YAML with clear structure           | Enables human debugging and transparency                                            |

## Stakeholders

| Role                      | Expectations                                                            | Concerns                                          |
| ------------------------- | ----------------------------------------------------------------------- | ------------------------------------------------- |
| **Software Developers**   | Accelerate feature development without losing control over architecture | Transparency, resumability, git integration       |
| **Engineering Managers**  | Track progress through filesystem state, review PR quality              | Audit trail, quality gates (review phase)         |
| **Plugin Developers**     | Extend sow with custom agents or commands                               | Clean interfaces, documented extension points     |
| **CLI Maintainers**       | Type-safe code generation, comprehensive test coverage                  | Schema validation, backward compatibility         |
| **Documentation Writers** | Understand architecture for onboarding and troubleshooting              | Clear architectural documentation (this document) |
