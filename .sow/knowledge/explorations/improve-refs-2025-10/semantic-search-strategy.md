# Semantic Search Strategy for Refs

## Summary

Two-phase approach for ref search: Start with SQLite FTS5 keyword search (zero dependencies), then add optional Ollama-based semantic search with graceful fallback. Uses chromem-go pure-Go vector DB, LangChainGo chunking, and official Ollama SDK. Implementation: ~1 week Phase 1, ~1 week Phase 2.

## Problem Statement

Enable users to search across installed refs to find relevant knowledge and code examples. Must work offline, respect privacy, and scale to 100+ refs without requiring external services.

## Two-Phase Approach

### Phase 1: Keyword Search (MVP)

**Technology**: SQLite FTS5 (Full-Text Search)

**Capabilities**:
- Keyword matching with boolean operators
- Ranking by relevance
- Snippet extraction with context
- Fast (millisecond queries)
- Zero dependencies beyond SQLite

**User Experience**:
```bash
sow refs add go-standards
sow refs search "error handling"

# Returns matches for documents containing "error" AND "handling"
✓ Found 3 results in 2 refs:

go-standards/error-handling.md
  ...proper error handling requires wrapping errors with context...

api-patterns/resilience.md
  ...database error handling should include retries...
```

**Implementation**:
```sql
-- Schema
CREATE VIRTUAL TABLE refs_fts USING fts5(
  ref_id,
  file_path,
  content,
  tokenize='porter unicode61'
);

-- Index on ref add
INSERT INTO refs_fts (ref_id, file_path, content)
SELECT 'go-standards', 'error-handling.md', 'content here...';

-- Search
SELECT
  ref_id,
  file_path,
  snippet(refs_fts, 2, '<mark>', '</mark>', '...', 64) as snippet
FROM refs_fts
WHERE content MATCH 'error AND handling'
ORDER BY rank
LIMIT 10;
```

**Pros**:
- Works immediately (no setup)
- Fast indexing (~seconds per ref)
- Fast search (<10ms)
- No external dependencies
- Always available (offline)
- Small index size (~10% of content)

**Cons**:
- Keyword-based only (no semantic understanding)
- "error handling" won't find "exception management"
- Must match exact terms (or synonyms you explicitly add)
- No "find similar" capability

**When Sufficient**:
- Technical documentation (consistent terminology)
- Users know specific keywords
- Ref count <100
- Terms are standardized (e.g., "error handling" always called that)

**Storage**: Single SQLite database at `~/.cache/sow/refs-index.db`

**Indexing Trigger**: On `sow refs add` or `sow refs update`

**Effort**: 2-3 days implementation

---

### Phase 2: Semantic Search with Fallback

**Technology**:
- chromem-go (pure-Go vector DB)
- LangChainGo textsplitter (markdown-aware chunking)
- Official Ollama SDK (embedding generation)
- FTS5 (fallback when Ollama unavailable)

**Capabilities**:
- Natural language queries
- Semantic similarity ("error handling" matches "exception management")
- Cross-ref discovery (find related content across refs)
- "Find similar" functionality
- Automatic fallback to FTS5 when Ollama unavailable

**User Experience**:

**First Time Setup**:
```bash
sow refs search "how should I handle database errors?"

⚠ Semantic search requires Ollama for better results.
  Install Ollama? [Y/n]

# User approves
→ Opening installation guide...
→ Run: curl -fsSL https://ollama.com/install.sh | sh

# After user installs Ollama
sow refs search "how should I handle database errors?"

⚠ Downloading embedding model (200MB, one-time)...
→ ollama pull nomic-embed-text
✓ Model ready

⚠ Indexing refs for semantic search (first time, ~2 min)...
[████████░░] 80% (3/5 refs)
✓ Indexing complete

✓ Found 5 results across 3 refs:

go-standards/error-handling.md (similarity: 0.92)
  ...when handling database errors, wrap with context and retry...

api-patterns/resilience.md (similarity: 0.87)
  ...database connection errors should trigger exponential backoff...

python-guide/exceptions.md (similarity: 0.79)
  ...exception handling for database operations requires...
```

**Ollama Not Available**:
```bash
sow refs search "error handling patterns"

⚠ Ollama not available, using keyword search
  Tip: Install Ollama for better semantic search

✓ Found 3 results (keyword matching):

go-standards/error-handling.md
  ...error handling patterns in Go...
```

**Architecture**:
```
Query: "how do I handle errors?"
    ↓
[Check Ollama] → Available?
    ↓                    ↓
   Yes                  No
    ↓                    ↓
[Semantic Search]   [FTS5 Search]
    ↓                    ↓
[chromem-go]        [SQLite FTS5]
    ↓                    ↓
Results            Results
```

**Implementation Flow**:

```go
type RefSearcher struct {
    fts5DB      *sql.DB           // Always present
    vectorDB    *chromem.DB       // Optional (if Ollama available)
    ollamaReady bool
}

func (s *RefSearcher) Search(query string, limit int) ([]Result, error) {
    // Try semantic search first
    if s.ollamaReady {
        results, err := s.semanticSearch(query, limit)
        if err == nil {
            return results, nil
        }
        // Ollama failed, fall back
        log.Warn("Semantic search failed, falling back to keyword search")
    }

    // Fallback to FTS5
    return s.keywordSearch(query, limit)
}

func (s *RefSearcher) semanticSearch(query string, limit int) ([]Result, error) {
    collection := s.vectorDB.GetCollection("sow-refs")

    // chromem-go automatically:
    // 1. Embeds query via Ollama
    // 2. Searches vectors
    // 3. Returns ranked results
    results, err := collection.Query(
        context.Background(),
        query,
        limit,
        nil, // no where filter
        nil, // no where document filter
    )

    return convertResults(results), err
}

func (s *RefSearcher) keywordSearch(query string, limit int) ([]Result, error) {
    rows, err := s.fts5DB.Query(`
        SELECT ref_id, file_path, snippet(refs_fts, 2, '<mark>', '</mark>', '...', 64)
        FROM refs_fts
        WHERE content MATCH ?
        ORDER BY rank
        LIMIT ?
    `, query, limit)

    return scanResults(rows), err
}
```

**Indexing Process**:

```go
func IndexRef(refPath, refID string) error {
    // Always index for FTS5 (fast, guaranteed)
    if err := indexFTS5(refPath, refID); err != nil {
        return err
    }

    // If Ollama available, also index for semantic search
    if ollamaAvailable() {
        if err := indexSemantic(refPath, refID); err != nil {
            log.Warn("Semantic indexing failed: %v", err)
            // Not fatal - FTS5 still works
        }
    }

    return nil
}

func indexSemantic(refPath, refID string) error {
    // 1. Read markdown files
    files := readMarkdownFiles(refPath)

    // 2. Chunk with LangChainGo
    splitter := textsplitter.NewMarkdownTextSplitter(
        textsplitter.WithChunkSize(512),
        textsplitter.WithChunkOverlap(50),
    )

    var allChunks []chromem.Document
    for _, file := range files {
        chunks, _ := splitter.SplitText(file.Content)

        for i, chunk := range chunks {
            allChunks = append(allChunks, chromem.Document{
                ID:      fmt.Sprintf("%s-%s-%d", refID, file.Path, i),
                Content: chunk,
                Metadata: map[string]string{
                    "ref_id":    refID,
                    "file_path": file.Path,
                },
            })
        }
    }

    // 3. Add to chromem-go (embeds via Ollama, stores vectors)
    collection := db.GetCollection("sow-refs")
    err := collection.AddDocuments(context.Background(), allChunks)

    // 4. Save to disk
    return db.ExportToFile("~/.cache/sow/embeddings.gob")
}
```

**Pros**:
- True semantic understanding
- Natural language queries work
- Finds related content even with different terminology
- Graceful degradation (always has FTS5 fallback)
- Privacy-first (local embeddings, no cloud APIs)
- Pure Go (chromem-go has zero dependencies)

**Cons**:
- Requires Ollama installation (one-time user action)
- First-time indexing delay (~1-2 min per 10MB ref)
- Larger index size (~2-3x content size for embeddings)
- 200MB embedding model download
- Re-indexing needed if ref updates

**Storage**:
- FTS5 index: `~/.cache/sow/refs-index.db` (~10% of content)
- Vector index: `~/.cache/sow/embeddings.gob` (~2-3x content size)

**Total overhead example** (10 refs, 50MB total content):
- FTS5 index: ~5MB
- Vector index: ~100-150MB
- Total: ~105-155MB additional storage

**Effort**: 5-7 days implementation

---

## Technology Stack

### Why These Libraries?

**chromem-go** (Vector Database):
- ✅ Pure Go, zero dependencies (no CGO)
- ✅ Built-in Ollama integration (`NewEmbeddingFuncOllama`)
- ✅ In-memory with export/import (fast + persistent)
- ✅ Simple API (similar to ChromaDB)
- ✅ Scales to 100k+ vectors (sufficient for sow)
- ✅ Recently published (Jan 2025, actively maintained)

**Alternative considered**: sqlite-vec
- ❌ Requires CGO or WASM runtime
- ✅ Persistent by default
- Decision: chromem-go preferred for simpler deployment

**LangChainGo textsplitter** (Chunking):
- ✅ Markdown-aware (preserves structure)
- ✅ Battle-tested (100+ projects using it)
- ✅ Configurable chunk size/overlap
- ✅ Heading hierarchy preservation
- ✅ Published Feb 2025

**Alternative considered**: Custom chunking
- Could do simple splitting (~100 lines of code)
- Decision: LangChainGo gives better quality for minimal cost

**Official Ollama SDK** (Embeddings):
- Not directly used (chromem-go wraps it)
- chromem-go's `NewEmbeddingFuncOllama` handles API calls
- Package: `github.com/ollama/ollama/api` (published Oct 2025)

**SQLite FTS5** (Keyword Search):
- ✅ Built into SQLite (no extra dependencies)
- ✅ Fast indexing and search
- ✅ Porter stemming (handles plurals, verb forms)
- ✅ Snippet extraction with highlighting
- ✅ Proven technology

---

## User Automation

**What Sow Automates**:

| Task | Phase 1 (FTS5) | Phase 2 (Semantic) |
|------|----------------|-------------------|
| Detect search capability | ✅ Auto (always available) | ✅ Auto (check Ollama) |
| Install dependencies | ✅ None needed | ⚠️ Guide user to install Ollama |
| Download models | ✅ None needed | ✅ Auto (`ollama pull`) |
| Index refs | ✅ Auto (on add/update) | ✅ Auto (on add/update) |
| Search | ✅ Auto | ✅ Auto (with fallback) |
| Persist index | ✅ Auto (SQLite) | ✅ Auto (gob export) |

**User Responsibilities**:

**Phase 1**: None (works out of box)

**Phase 2**:
1. Install Ollama (one-time, ~1 minute)
   - Guided by sow with instructions
   - Or automated: `curl -fsSL https://ollama.com/install.sh | sh`
2. Wait for first-time indexing (~1-2 min per ref)
3. Keep Ollama running (daemon, starts on boot)

---

## Implementation Roadmap

### Phase 1: FTS5 Keyword Search (Week 1)

**Day 1-2**: Schema and indexing
- Create SQLite database with FTS5 virtual table
- Implement ref indexing on add/update
- Read markdown files, extract content
- Insert into FTS5 table

**Day 3**: Search implementation
- Query FTS5 with boolean operators
- Extract snippets with highlighting
- Format results for CLI output

**Day 4**: CLI integration
- `sow refs search <query>` command
- Display results with file paths and snippets
- Pagination for many results

**Day 5**: Testing and polish
- Test with various query patterns
- Performance testing (100+ refs)
- Error handling

**Deliverable**: Working keyword search with zero dependencies

---

### Phase 2: Semantic Search (Week 2)

**Day 1-2**: Ollama detection and setup
- Check if Ollama installed (`which ollama` or API ping)
- Guide user through installation if missing
- Auto-pull embedding model (`ollama pull nomic-embed-text`)
- Test Ollama connectivity

**Day 3-4**: Semantic indexing
- Integrate LangChainGo textsplitter
- Chunk markdown files (512 tokens, 50 overlap)
- Create chromem-go collection with Ollama embedding function
- Add chunked documents (chromem-go embeds automatically)
- Export to disk (`~/.cache/sow/embeddings.gob`)

**Day 5**: Semantic search
- Load chromem-go database from disk
- Query collection (chromem-go embeds query automatically)
- Rank results by similarity
- Format output with similarity scores

**Day 6**: Fallback logic
- Detect Ollama availability before search
- Fall back to FTS5 if Ollama unavailable
- User messaging (explain fallback, suggest installation)

**Day 7**: Testing and optimization
- Test with/without Ollama
- Performance testing (indexing time, search latency)
- Memory profiling (chromem-go in-memory usage)
- Error handling (Ollama crashes, network issues)

**Deliverable**: Semantic search with graceful FTS5 fallback

---

## Configuration

**User config** (`~/.config/sow/config.yaml`):

```yaml
refs:
  search:
    # Search mode: auto (semantic if available, else keyword), keyword (force FTS5), semantic (error if Ollama unavailable)
    mode: auto

    # Semantic search settings
    semantic:
      enabled: true  # User can disable to force FTS5
      ollama_endpoint: http://localhost:11434
      embedding_model: nomic-embed-text

    # Indexing settings
    indexing:
      chunk_size: 512
      chunk_overlap: 50
      auto_index: true  # Index on ref add/update

  # Storage paths
  cache_dir: ~/.cache/sow
```

**CLI flags**:

```bash
# Force keyword search (skip semantic)
sow refs search "query" --keyword

# Force semantic search (error if unavailable)
sow refs search "query" --semantic

# Rebuild indexes
sow refs reindex --all
sow refs reindex go-standards

# Check indexing status
sow refs status
# Output:
# FTS5 index: 10 refs, 5,234 documents
# Semantic index: 8 refs, 2,145 chunks (Ollama required for 2 refs)
```

---

## Migration Path

**Existing refs** (git-based):
- Phase 1: Index existing refs automatically on first search
- Phase 2: Prompt user to install Ollama, then reindex
- Backward compatible (FTS5 always works)

**OCI refs** (future):
- Same indexing process
- Extract from OCI image, index content
- No changes to search implementation

---

## Performance Characteristics

### Phase 1 (FTS5)

**Indexing**:
- 10MB ref: ~1-2 seconds
- 100 refs (500MB): ~2-3 minutes
- Linear scaling

**Search**:
- Query latency: <10ms
- Scales to millions of documents
- No memory overhead (SQLite on disk)

**Storage**:
- ~10% of content size
- Example: 500MB refs → 50MB index

### Phase 2 (Semantic)

**Indexing**:
- 10MB ref: ~30-60 seconds (embedding generation)
  - 512-token chunks: ~20 chunks per MB
  - 200 chunks × 0.1s per embedding = ~20s
  - Plus chunking overhead: ~10s
- 100 refs (500MB): ~30-50 minutes first time
- Bottleneck: Ollama embedding generation (CPU-bound)

**Search**:
- Query latency: ~100-200ms
  - Embed query: ~100ms
  - Vector search (chromem-go): ~10-50ms (for 100k vectors)
  - Format results: ~10ms
- Memory: ~2-3x content size loaded in RAM
  - Example: 500MB refs → 1-1.5GB RAM for vectors

**Storage**:
- FTS5: ~50MB (10% of 500MB)
- Vectors: ~1-1.5GB (2-3x of 500MB)
- Total: ~1.55GB for 500MB of refs

**Scalability limits** (chromem-go exhaustive search):
- <100k vectors: Excellent (<50ms)
- 100k-1M vectors: Good (50-200ms)
- >1M vectors: Consider alternatives (HNSW, external DB)

**For sow's expected scale**:
- 100 refs × 500 chunks/ref = 50k chunks
- Well within chromem-go's sweet spot

---

## Open Questions

1. **Incremental indexing**: When ref updates, re-index entire ref or detect changed files only?
   - Phase 1: Full re-index (simple, safe)
   - Phase 2 optimization: Delta detection

2. **Index versioning**: How to handle:
   - Embedding model updates (nomic-embed-text v1.5 → v2.0)
   - Schema changes in chromem-go format
   - Migration strategy for users

3. **Background indexing**: Index during `sow refs add` (blocking) or spawn background process?
   - Phase 1: Blocking (FTS5 is fast, <5s)
   - Phase 2: Consider background for semantic (30-60s per ref)

4. **Multi-ref queries**: Search across specific refs only (`sow refs search "query" --refs go-standards,api-patterns`) or always search all?

5. **Result ranking**: In semantic search, how to balance:
   - Similarity score
   - Recency (newer refs ranked higher?)
   - Ref priority (user-starred refs?)

6. **Ollama model choice**: Default to `nomic-embed-text` (768-dim, 100M params) or offer alternatives?
   - `jina-embeddings-v2-small`: 512-dim, 33M (faster, smaller)
   - `mxbai-embed-large`: 1024-dim, 335M (more accurate, slower)

7. **Embedding cache**: If user removes and re-adds same ref, reuse embeddings?
   - Track by ref source URL + version/commit?
   - Cache at `~/.cache/sow/embeddings-cache/{ref-id}-{version}.gob`

8. **Search scope**: Search just indexed content or also:
   - Ref metadata (title, description, tags)
   - File names
   - Code comments only vs full code

---

## Success Metrics

**Phase 1 (FTS5)**:
- ✅ Works offline with zero setup
- ✅ Search latency <50ms for 100 refs
- ✅ Indexing completes in <5 minutes for 100 refs
- ✅ Index size <15% of content size

**Phase 2 (Semantic)**:
- ✅ Finds relevant results for natural language queries
- ✅ Graceful fallback to FTS5 when Ollama unavailable
- ✅ First-time indexing completes in <1 hour for 100 refs
- ✅ Subsequent searches <500ms
- ✅ Clear user guidance for Ollama installation
- ✅ Works entirely offline after setup

---

## Alternative Approaches Considered

### Precomputed Embeddings in OCI Images

**Idea**: Publisher generates embeddings, ships in OCI image alongside content

**Analysis**:
- ✅ Instant search after download (no indexing wait)
- ✅ Consistent results (same embeddings for all users)
- ❌ User still needs Ollama to embed queries (no dependency reduction)
- ❌ 10-30% larger OCI images permanently
- ❌ Model lock-in (can't change without republishing)
- **Decision**: Not worth trade-offs

### Cloud Embedding APIs

**Idea**: Use Voyage AI / OpenAI APIs instead of local Ollama

**Analysis**:
- ✅ No local installation required
- ✅ Always available (no Ollama daemon)
- ❌ Ref content sent to third-party (privacy concern)
- ❌ Requires internet connection
- ❌ Costs money (small but ongoing)
- ❌ Not acceptable for enterprise/sensitive content
- **Decision**: Local-first approach better aligns with sow philosophy

### FTS5 Only (No Semantic)

**Idea**: Ship Phase 1, skip Phase 2 entirely

**Analysis**:
- ✅ Simplest implementation
- ✅ Zero dependencies
- ✅ Fast and reliable
- ❌ Limited to keyword matching
- ❌ Misses semantic relationships
- ⚠️ May be sufficient for technical docs with standardized terminology
- **Decision**: Implement Phase 1, evaluate user demand before Phase 2

---

## Recommendation

**Implement Phase 1 immediately** (FTS5 keyword search):
- Low effort (2-3 days)
- High value (search works immediately)
- Zero friction (no user setup)
- Validates search UX

**Evaluate Phase 2 based on**:
- User requests for semantic search
- Ref count growth (semantic more valuable at scale)
- Feedback on FTS5 limitations

**If proceeding to Phase 2**:
- Use recommended stack (chromem-go + LangChainGo + Ollama SDK)
- 5-7 days implementation
- Clear user communication about Ollama requirement
- Maintain FTS5 fallback permanently (graceful degradation)

**Estimated total effort**: 1 week (Phase 1) + 1 week (Phase 2) = 2 weeks
