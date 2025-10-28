# sow Architecture Documentation (Arc42)

Complete Arc42 architecture documentation for the sow (System of Work) framework.

## Document Structure

| Section | Title | Description |
|---------|-------|-------------|
| [01](./01-introduction.md) | Introduction and Goals | System overview, functional requirements, quality goals, stakeholders |
| [02](./02-constraints.md) | Architecture Constraints | Technical and organizational constraints, conventions |
| [03](./03-context-scope.md) | System Context and Scope | Business and technical context, external interfaces, system boundary |
| [04](./04-solution-strategy.md) | Solution Strategy | Architectural approach, key decisions, technology choices, quality solutions |
| [05](./05-building-blocks.md) | Building Block View | System decomposition: CLI internals, project system, modes, schemas |
| [06](./06-runtime-view.md) | Runtime View | Key scenarios: project init, task execution, feedback loop, modes, resumption, refs |
| [07](./07-deployment.md) | Deployment View | Infrastructure, environments, deployment process, platform considerations |
| [08](./08-crosscutting-concepts.md) | Cross-cutting Concepts | Domain model, state management, error handling, logging, security, testing |
| [09](./09-architecture-decisions.md) | Architecture Decisions | ADR index and summaries of 12 key architectural decisions |
| [10](./10-quality-requirements.md) | Quality Requirements | Quality tree, detailed scenarios, trade-offs, priorities |
| [11](./11-risks-debt.md) | Risks and Technical Debt | Technical risks, technical debt items, repayment plan |
| [12](./12-glossary.md) | Glossary | Domain, technical, file, status, state machine, command terms, acronyms |

## Quick Navigation

### For New Developers
1. Start with [Section 1 (Introduction)](./01-introduction.md) for overview
2. Read [Section 5 (Building Blocks)](./05-building-blocks.md) for system structure
3. Review [Section 6 (Runtime View)](./06-runtime-view.md) for key workflows
4. Check [Section 12 (Glossary)](./12-glossary.md) for terminology

### For Contributors
1. Review [Section 8 (Cross-cutting Concepts)](./08-crosscutting-concepts.md) for patterns
2. Read [Section 9 (Architecture Decisions)](./09-architecture-decisions.md) for rationale
3. Check [Section 11 (Risks and Technical Debt)](./11-risks-debt.md) for known issues

### For Architects
1. Study [Section 4 (Solution Strategy)](./04-solution-strategy.md) for approach
2. Review [Section 9 (Architecture Decisions)](./09-architecture-decisions.md) for key decisions
3. Examine [Section 10 (Quality Requirements)](./10-quality-requirements.md) for quality attributes

### For Operators
1. Read [Section 7 (Deployment View)](./07-deployment.md) for infrastructure
2. Check [Section 8 (Cross-cutting Concepts)](./08-crosscutting-concepts.md) for monitoring

## Key Architectural Principles

1. **Zero-Context Resumability**: All state on disk, any agent can resume any work
2. **Filesystem as Database**: YAML/markdown state files, git versioning
3. **Human-Led Planning, AI-Led Execution**: Subservient vs. autonomous modes
4. **Multi-Agent Orchestration**: Orchestrator + specialized workers
5. **Two-Layer Architecture**: Execution (plugin) + Data (repository)
6. **Fixed 5-Phase Model**: Planning → Implementation → Review → Finalize
7. **Single Project Per Branch**: One project per git branch (enforced)

## Core Technologies

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Filesystem**: Billy (abstraction)
- **State Machine**: stateless library
- **Schemas**: CUE with Go code generation
- **Git**: go-git library
- **GitHub**: gh CLI wrapper
- **Platform**: Claude Code

## Quality Priorities

1. **Resumability** (< 2s context recovery, 100% completeness)
2. **Reliability** (zero state corruption)
3. **Performance** (< 100ms CLI operations)
4. **Usability** (human-readable state, clear errors)
5. **Maintainability** (> 80% test coverage, type safety)
6. **Portability** (macOS, Linux, Windows)

## Document Maintenance

**Update Triggers**:
- New components/services: Update Sections 5, 6, 7
- Architectural decisions: Update Section 9, affected sections
- New constraints: Update Section 2
- Quality changes: Update Section 10
- New risks/debt: Update Section 11
- New terminology: Update Section 12

**Review Schedule**: Quarterly architecture review

---

**Created**: 2025-10-27
**Version**: 1.0
**Format**: Arc42 (12 sections)
**Status**: Complete, pending review
