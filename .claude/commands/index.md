---
name: index
description: "Generate or update the repository index"
allowed-tools: Bash(find:*)
---

# Index Repository

Scans the repository structure and generates/updates `.sow/index.md` with a minimal overview of the repository's logical components.

## Repository Structure

!`find . -type f -o -type d | grep -v -E '(\.git/|node_modules/|\.pyc$|__pycache__|\.DS_Store)' | head -500`

## Your Task

Analyze the repository structure above and generate a **minimal, scannable index** at `.sow/index.md`.

### Critical Guidelines

**KEEP IT MINIMAL:**
- This is a quick reference map, not comprehensive documentation
- Focus on WHERE things are, not WHAT they contain
- No file counts, no exhaustive listings, no git status
- Target length: 50-100 lines total
- If in doubt, leave it out

**IDENTIFY LOGICAL COMPONENTS:**

Look for and document the locations of:
- **Services/APIs** - Backend services, microservices, API implementations
- **Software libraries** - Reusable code libraries, shared packages
- **Documentation** - Architecture docs, guides, ADRs, design documents
- **Binary projects** - CLI tools, executables, applications
- **Configuration** - Important config files or directories
- **Knowledge bases** - Style guides, conventions, standards, `.sow/knowledge/`, `.sow/sinks/`
- **Testing infrastructure** - Test directories, test utilities
- **Build/deployment** - Build scripts, deployment configs, CI/CD

**Don't force-fit these categories.** Every repository is different. Identify what's actually present and logically group similar components.

### Index Structure

Use this template structure:

```markdown
# Repository Index

**Generated**: <timestamp>

---

## Quick Start

**New to this repository?**
- `README.md` - Project overview
- <Point to main documentation directory if exists>

**Starting a project?**
- <Point to roadmap, planning docs, or project guides if they exist>
- Check `.sow/knowledge/` for repository-specific context

---

## Knowledge

<Identify and list knowledge sources: docs/, .sow/knowledge/, .sow/sinks/, etc.>
<For each source, one-line description of what's there>
<If empty or non-existent, note it>

---

## Key Directories

<Create a minimal tree showing logical components - 2-3 levels deep max>
<Include brief inline comments explaining purpose>
<Mark empty directories as (empty)>

---

## Common Tasks

<Provide 3-5 brief task-oriented pointers>
<Format: "Working on X? Read Y" or "Before starting X, check Y">
<Focus on most common workflows for this specific repository>
```

### Examples of Good Component Identification

**Example 1 - Monorepo with services:**
```
repo/
├── services/
│   ├── auth-api/      # Authentication service
│   ├── user-api/      # User management service
│   └── billing-api/   # Billing service
├── packages/          # Shared libraries
└── docs/              # Architecture documentation
```

**Example 2 - Library project:**
```
repo/
├── src/               # Library source code
├── tests/             # Test suites
├── docs/              # API documentation
└── examples/          # Usage examples
```

**Example 3 - CLI tool:**
```
repo/
├── cmd/               # CLI entry points
├── internal/          # Internal packages
├── docs/              # User guides
└── .sow/knowledge/    # Architecture decisions
```

### What NOT to Include

❌ File counts ("15 files in docs/")
❌ Git status ("3 modified files")
❌ Exhaustive file listings
❌ Verbose descriptions
❌ Obvious information (everyone knows what node_modules is)
❌ Temporary or generated directories

### Final Steps

1. **Generate the index** following the template
2. **Display the complete index** to the user
3. **Ask for approval and/or feedback**: "Does this index look good? Any changes needed?"
4. **If approved**: Write to `.sow/index.md` (overwrite if exists)
5. **If changes requested**: Revise based on feedback and repeat steps 2-4
6. **Confirm completion** with a brief summary of what components you identified

**Remember**: Minimal and scannable. When in doubt, leave it out.
