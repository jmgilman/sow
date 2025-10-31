# MCP Strategy for Sow

## Summary

Ship sow with 2 MCP servers (Context7 for library docs, Exa for code/web search) to complement static refs with dynamic, public data. Deploy via plugin-declared configuration (`.mcp.json`) for zero-friction setup. LLMs prefer Context7 first, use Exa for follow-up research, and always use refs for team-specific knowledge.

## What is MCP?

**Model Context Protocol** (MCP) is an open standard introduced by Anthropic (November 2024) and adopted by OpenAI (March 2025) for connecting AI systems to external data sources and tools.

**Architecture**:
```
MCP Client (Claude Code)
    ↓
MCP Servers (tools/resources)
    ↓
External Systems (APIs, databases, services)
```

**Server Capabilities**:
1. **Tools** - Functions LLM can invoke (e.g., `search_github`, `fetch_api_docs`)
2. **Resources** - Data the LLM can read (e.g., file contents, API responses)
3. **Prompts** - Pre-configured prompt templates

**Transport Types**:
- **Remote HTTP** - Cloud-hosted servers (recommended, no local dependencies)
- **Local stdio** - Local processes (requires npx/Node.js)
- **Remote SSE** - Server-sent events (deprecated)

## MCP vs Refs: Complementary Systems

### When to Use MCP

**Dynamic, queryable, public data**:
- Public library/framework documentation (React, Django, Express)
- Live services (GitHub issues, PRs - if available remotely)
- Real-time web content
- Large datasets already indexed by provider

**Characteristics**:
- Data changes frequently
- Query-based access (search, filter, paginate)
- Can't cache everything locally
- Maintained by third parties

**Examples**:
- Context7: Fetch current React docs on-demand
- Exa: Search GitHub repos for code examples
- Exa: Crawl specific URLs for content

### When to Use Refs

**Static, versioned, team-specific knowledge**:
- Internal coding standards
- Team-specific style guides
- Custom code templates and examples
- Architecture decisions (ADRs)
- Company-specific runbooks

**Characteristics**:
- Content relatively stable
- Version-controlled
- Organization-specific
- Small enough to cache locally
- Not available via public MCP servers

**Examples**:
- Team Go conventions (not in public docs)
- Internal API patterns and examples
- Company best practices
- Project templates

### Decision Heuristic

**Use MCP when**:
- Public + frequently updated + already indexed → MCP

**Use Refs when**:
- Team-specific OR stable versions OR not in MCP ecosystem → Refs

**Overlap Examples**:

| Content Type | MCP Approach | Refs Approach | Recommendation |
|--------------|--------------|---------------|----------------|
| Public API docs (React) | Context7 MCP | Package as ref | **MCP** - Always current |
| Team coding standards | - | OCI ref | **Refs** - Org-specific |
| Public code examples | Exa code search | Git/OCI ref | **Depends** - MCP for discovery, refs for curated |
| Internal runbooks | - | OCI ref | **Refs** - Private |
| Third-party tutorials | Exa web search | OCI ref | **MCP** if updated frequently |

---

## Selected MCP Servers

### 1. Context7 (Upstash) - PRIMARY

**Purpose**: Up-to-date library and framework documentation

**What it provides**:
- Current documentation for public packages/frameworks
- Version-specific code examples
- Intelligent project ranking
- Customizable token limits

**Tools**:
- `resolve-library-id` - Resolve library name to Context7 ID
- `get-library-docs` - Fetch documentation for library

**Why selected**:
- ✅ Perfect complement to refs (public docs vs team docs)
- ✅ Remote HTTP (no local dependencies)
- ✅ Free tier with optional API key
- ✅ Maintained by Upstash (reputable)
- ✅ Direct use case: "How do I use React hooks?"

**Configuration**:
```json
{
  "context7": {
    "type": "remote",
    "url": "https://context7.com/mcp",
    "description": "Up-to-date library and framework documentation (primary)",
    "env": {
      "CONTEXT7_API_KEY": "${CONTEXT7_API_KEY}"  // Optional
    }
  }
}
```

**Setup**:
- Works immediately with free tier (rate-limited)
- Optional API key from context7.com/dashboard for higher limits
- Free tier sufficient for most usage

**Usage priority**: PRIMARY - always try first for library/framework documentation

---

### 2. Exa (Exa Labs) - SECONDARY

**Purpose**: Web and code search for follow-up research

**What it provides**:
- Search code snippets and documentation from open source
- Real-time web search with content extraction
- URL content extraction
- Business intelligence (company research)
- Deep research capabilities

**Tools** (primary ones):
- `get_code_context_exa` - Search code from GitHub repos, find examples
- `web_search_exa` - Broader web search beyond library docs
- `crawling` - Extract content from specific URLs

**Why selected**:
- ✅ Complements Context7 (broader search when Context7 doesn't have it)
- ✅ Code search from real GitHub repositories
- ✅ Remote HTTP (no local dependencies)
- ✅ Free tier ($10 credits) with optional API key
- ✅ Active development by Exa Labs

**Configuration**:
```json
{
  "exa": {
    "type": "remote",
    "url": "https://mcp.exa.ai/mcp",
    "description": "Web and code search for follow-up research (secondary)",
    "env": {
      "EXA_API_KEY": "${EXA_API_KEY}"  // Optional
    }
  }
}
```

**Setup**:
- Works immediately with free tier ($10 credits, rate-limited)
- Optional API key from dashboard.exa.ai/api-keys for predictable usage
- Pricing: Per 1k requests/pages (see exa.ai/pricing)
- Free tier sufficient for moderate usage

**Usage priority**: SECONDARY - use only when Context7 doesn't have the answer or for broader research

---

## Deployment Strategy

### Plugin-Declared MCP Servers

**Approach**: Declare MCP servers in sow plugin's `.mcp.json` file

**Plugin structure**:
```
sow-plugin/
├── .mcp.json              # MCP server declarations
├── plugin.json            # Plugin metadata
└── README.md              # Setup documentation
```

**Complete `.mcp.json`**:
```json
{
  "context7": {
    "type": "remote",
    "url": "https://context7.com/mcp",
    "description": "Up-to-date library and framework documentation (primary)",
    "env": {
      "CONTEXT7_API_KEY": "${CONTEXT7_API_KEY}"
    }
  },
  "exa": {
    "type": "remote",
    "url": "https://mcp.exa.ai/mcp",
    "description": "Web and code search for follow-up research (secondary)",
    "env": {
      "EXA_API_KEY": "${EXA_API_KEY}"
    }
  }
}
```

**How it works**:
1. User enables sow plugin in Claude Code
2. Claude Code reads `.mcp.json` from plugin
3. Auto-configures MCP servers (Context7 + Exa)
4. Servers immediately available to LLM
5. Environment variables substituted from user's environment

**Why this approach**:
- ✅ Simple (just declare servers in file)
- ✅ Claude Code handles all configuration
- ✅ No sow CLI implementation needed
- ✅ Both remote HTTP (no local dependencies)
- ✅ Free tiers work out of box
- ✅ Optional API keys (user choice)
- ✅ Standard MCP pattern

**Alternatives considered**:
- ❌ **Sow as proxy MCP server** - More complex, unnecessary for 2 servers
- ❌ **Local stdio servers** - Requires Node.js/npx dependency
- ❌ **MCPB bundles** - Not supported in Claude Code

**Decision**: Plugin-declared is simplest and works perfectly for our needs

---

## Usage Guidance

### LLM Decision-Making

**System prompt additions for Claude Code**:

```markdown
# Knowledge Sources

You have access to multiple knowledge sources with clear priorities:

## MCP Servers (for dynamic, public data)

### context7 (PRIMARY)
Use for current documentation of public libraries and frameworks.

**When to use**:
- User asks about public library/framework (React, Django, Express, etc.)
- Need current API documentation
- Checking usage of third-party package

**Always try context7 FIRST for library/framework questions.**

Examples:
- "How do I use React hooks?" → query context7
- "What's the syntax for Express middleware?" → query context7
- "How to configure Django authentication?" → query context7

### exa (SECONDARY)
Use for follow-up research when context7 doesn't have the answer.

**When to use**:
- Finding real-world code examples from GitHub repos (get_code_context_exa)
- Broader web search beyond library docs (web_search_exa)
- Extracting content from specific URLs (crawling)

**Only use exa when context7 doesn't provide sufficient information.**

Examples:
- "Find production examples of React hooks in large apps" → use exa get_code_context_exa
- "Search for error handling patterns in Go microservices" → use exa get_code_context_exa
- "Fetch content from https://example.com/blog" → use exa crawling

## Refs (for team-specific, static knowledge)

Read from `.sow/refs/` for team standards and internal documentation.

**When to use**:
- User asks about team-specific standards or conventions
- Internal coding patterns
- Company-specific practices
- Project architecture

**Always use refs for team/org-specific knowledge.**

Examples:
- "What are our Go error handling conventions?" → read .sow/refs/go-standards/
- "Show me our API design patterns" → read .sow/refs/api-patterns/
- "What's our React component structure?" → read .sow/refs/react-guide/

## Decision Guide

**Priority order**:
1. **Team-specific** → refs (.sow/refs/)
2. **Public library docs** → context7 MCP (primary)
3. **Code examples / broader search** → exa MCP (secondary)
4. **Project architecture** → knowledge (.sow/knowledge/)

**Usage pattern**:
1. Library/framework questions → Try context7 first
2. If context7 doesn't have it → Try exa for code examples or web search
3. Team-specific questions → Always use refs
```

### User Documentation

**In sow plugin README**:

```markdown
## MCP Servers

Sow ships with 2 MCP servers for enhanced AI capabilities:

### Context7 (Upstash) - Primary Documentation Source
Provides up-to-date library and framework documentation for public packages.

**Works immediately** with free tier (rate-limited)

**Optional Setup**: Get API key at https://context7.com/dashboard
- Set `CONTEXT7_API_KEY` environment variable for higher rate limits
- Free tier is sufficient for most usage

### Exa (Exa Labs) - Web & Code Search
Search code examples, documentation, and web content for follow-up research.

**Works immediately** with free tier ($10 credits, rate-limited)

**Optional Setup**: Get API key at https://dashboard.exa.ai/api-keys
- Set `EXA_API_KEY` environment variable for predictable usage
- Pricing: Per 1k requests/pages (see https://exa.ai/pricing)
- Free tier sufficient for moderate usage

### How LLMs Use These Servers

**Context7 (Primary)**: Claude tries Context7 first for all library/framework documentation questions.

**Exa (Secondary)**: Claude uses Exa when:
- Context7 doesn't have the answer
- You need real-world code examples from GitHub
- You need broader web research beyond library docs
- You want to crawl specific URLs

**Refs (Team Knowledge)**: Claude always uses refs for team-specific knowledge like coding standards, internal patterns, and company practices.

### Setting API Keys (Optional)

**macOS/Linux**:
```bash
# Add to ~/.bashrc or ~/.zshrc
export CONTEXT7_API_KEY="your-key-here"
export EXA_API_KEY="your-key-here"
```

**Windows**:
```powershell
# Add to PowerShell profile
$env:CONTEXT7_API_KEY="your-key-here"
$env:EXA_API_KEY="your-key-here"
```

Both servers work without API keys (free tiers). API keys provide higher rate limits and more predictable usage.
```

---

## Integration with Refs System

### Clear Separation of Concerns

**MCP Servers** → Dynamic, public data:
- **Context7**: Library/framework documentation (always current)
- **Exa**: Code examples from GitHub, web research (on-demand)

**Refs System** → Static, versioned, team knowledge:
- Team coding standards
- Internal API patterns
- Company-specific guides
- Architecture decisions
- Runbooks and procedures

### Example Workflows

**Workflow 1: Public Library Question**
```
User: "How do I use React hooks?"

LLM Decision:
1. Public library question → Try Context7
2. Context7 has React documentation → Use it
3. Return answer from Context7

Result: ✅ Answer from Context7 React docs
```

**Workflow 2: Real-World Code Examples**
```
User: "Find production examples of React hooks in large apps"

LLM Decision:
1. Code examples from GitHub → Use Exa
2. Exa get_code_context_exa searches GitHub repos
3. Return code examples

Result: ✅ GitHub code examples from Exa
```

**Workflow 3: Team-Specific Standards**
```
User: "What are our React component conventions?"

LLM Decision:
1. Team-specific → Always use refs
2. Read .sow/refs/react-guide/
3. Return team conventions

Result: ✅ Team conventions from refs
```

**Workflow 4: Hybrid Question**
```
User: "How do we implement JWT authentication in Express?"

LLM Decision:
1. Unclear if team-specific or general
2. Try Context7 first (general Express JWT docs)
3. Also check refs for team-specific auth patterns
4. Combine both sources

Result: ✅ General JWT docs (Context7) + team auth patterns (refs)
```

---

## Risks and Mitigations

### Risk 1: API Rate Limits

**Issue**:
- Context7: Rate-limited without API key
- Exa: $10 free credits, then requires paid plan

**Mitigation**:
- Document API key setup clearly (optional but recommended)
- Handle rate limit errors gracefully with clear messages
- Both free tiers sufficient for moderate usage
- Fall back to suggesting documentation URLs or refs
- Emphasize Context7 as primary to preserve Exa credits

### Risk 2: Free Tier Exhaustion

**Issue**:
- Exa $10 credits will eventually run out with heavy usage
- Users must add API key or accept limits

**Mitigation**:
- Clear messaging when credits low ("$2 remaining")
- Document how to add API key in plugin README
- System prompt emphasizes Context7 primary, Exa secondary
- Most users won't exhaust credits with normal usage

### Risk 3: MCP Server Failures

**Issue**:
- Network issues
- Service outages
- Rate limit errors

**Mitigation**:
- Don't block sow functionality on MCP availability
- Log errors, continue gracefully
- Provide clear user feedback
- Fall back to suggesting refs or documentation URLs
- LLM can still function with just refs

### Risk 4: Overuse of Exa

**Issue**:
- LLM might use Exa instead of Context7
- Exhausts credits faster than needed

**Mitigation**:
- System prompt clearly states Context7 is PRIMARY
- Explicit instructions to try Context7 first
- Monitor usage patterns post-launch
- Refine prompts if Exa overused

---

## Success Criteria

### Phase 1 (Initial Shipping)

- ✅ Context7 and Exa declared in plugin `.mcp.json`
- ✅ Both remote HTTP (no local dependencies)
- ✅ Servers auto-configure when plugin enabled
- ✅ Documentation covers optional API key setup
- ✅ Graceful degradation when rate limits hit
- ✅ Clear separation from refs in system prompts
- ✅ Context7 primary, Exa secondary usage pattern enforced

### Phase 2 (Post-Launch Validation)

- ✅ User feedback on MCP server usefulness
- ✅ Usage analytics (which servers used, how often)
- ✅ Monitor Exa credit exhaustion rates
- ✅ Evaluate additional servers based on requests
- ✅ Consider remote HTTP alternatives for other servers

---

## Future Considerations

### Additional MCP Servers

**Evaluate based on**:
- User requests and feedback
- Identified gaps in coverage
- New remote HTTP servers becoming available
- Clear differentiation from existing servers

**Candidates for future evaluation**:
- GitHub (if remote HTTP version becomes available)
- Memory (if integration with refs is clarified)
- Sequential Thinking (if user demand emerges)
- Database servers (SQLite, PostgreSQL - as optional, not default)

**Decision timeline**: Quarterly review based on usage patterns and feedback

### Proxy Architecture

**When to consider**:
- Shipping 5+ MCP servers
- Need aggregation or routing logic
- Want to add caching layer
- Custom server composition required

**Not needed now**: 2 servers work fine with plugin-declared approach

### Custom User Servers

**Question**: Should users add their own MCP servers?

**Answer**: Yes, but via Claude Code's native MCP configuration, not sow.

**Approach**:
- Document how users can add more servers to Claude Code
- Sow doesn't manage user-added servers
- Keep sow plugin's `.mcp.json` minimal (just Context7 + Exa)

---

## Implementation Checklist

**Phase 1: Plugin Configuration**
- [ ] Create `.mcp.json` in sow plugin root
- [ ] Add Context7 configuration (remote HTTP)
- [ ] Add Exa configuration (remote HTTP)
- [ ] Test with Claude Code plugin system

**Phase 2: Documentation**
- [ ] Add MCP section to plugin README
- [ ] Document optional API key setup
- [ ] Explain Context7 (primary) vs Exa (secondary) strategy
- [ ] Provide example workflows

**Phase 3: System Prompts**
- [ ] Add MCP usage guidance to agent prompts
- [ ] Emphasize Context7 first, Exa for follow-up
- [ ] Document decision flow (public → context7, team → refs)
- [ ] Test with real queries

**Phase 4: Testing**
- [ ] Test Context7 with/without API key
- [ ] Test Exa with/without API key
- [ ] Test rate limit handling
- [ ] Test error messages and fallbacks
- [ ] Verify refs still work when MCP unavailable

**Phase 5: Launch & Monitor**
- [ ] Ship plugin with MCP servers
- [ ] Gather user feedback
- [ ] Monitor usage patterns
- [ ] Track Exa credit exhaustion rates
- [ ] Refine system prompts based on actual usage

---

## Recommendation Summary

**Ship with 2 MCP servers**:

1. **Context7** (Upstash) - Library/framework docs (PRIMARY)
2. **Exa** (Exa Labs) - Code/web search (SECONDARY)

**Deployment**: Plugin-declared in `.mcp.json`, auto-configured by Claude Code

**Zero local dependencies**: Both remote HTTP, no Node.js/npx required

**Free tiers work**: Both have generous free tiers, API keys optional

**Clear strategy**: Context7 first, Exa for follow-up, refs for team knowledge

**Timeline**: Can ship immediately with plugin, no sow CLI changes needed

**Next steps**:
1. Create `.mcp.json` in sow plugin
2. Document setup in plugin README
3. Add system prompt guidance
4. Test integration with Claude Code
5. Monitor usage and gather feedback
