# 11. Risks and Technical Debt

## Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| **State Schema Evolution Breaking Changes** | Medium | High | Version field in state files, migration tooling, backward compatibility testing |
| **Claude Code Platform Changes** | Medium | High | Minimal platform coupling, abstract agent spawning, monitor Claude Code updates |
| **Concurrent State Writes Corruption** | Low | High | Atomic writes (rename), schema validation on load, git versioning as backup |
| **GitHub CLI API Changes** | Low | Medium | Wrapper abstraction, version pinning in docs, graceful error handling |
| **CUE Code Generation Breaks** | Low | Medium | Pin CUE version, test generation in CI, manual fallback possible |
| **Filesystem Differences (Windows)** | Low | Medium | Platform detection, copy instead of symlink on Windows, test on all platforms |
| **Git Repository Corruption** | Very Low | High | Rely on git's integrity, atomic operations, no direct .git manipulation |
| **Cache Poisoning (Malicious Refs)** | Low | Medium | User responsibility (trust refs), consider checksum validation, docs warning |

---

### Risk Details

#### R-1: State Schema Evolution Breaking Changes
**Description**: Changes to CUE schemas may break compatibility with existing `.sow/` directories.

**Current Mitigation**:
- Version field (`.sow/.version`) tracks structure version
- CLI detects version mismatch, can refuse to operate
- Documentation recommends committing `.sow/` to branches

**Future Mitigation**:
- Implement migration system (`sow migrate` command)
- Migrations as Go code registered by version range
- Automated migration on version detect
- Test migrations in CI with fixtures from all versions

**Likelihood Reduction**: Pin CUE schema version in docs, avoid breaking changes when possible

**Impact Reduction**: Clear error messages, migration tooling, rollback via git

---

#### R-2: Claude Code Platform Changes
**Description**: Claude Code may change Task tool API, agent spawning, or file operations.

**Current Mitigation**:
- Minimal platform coupling (agents are markdown files)
- CLI is independent (doesn't depend on Claude Code API)
- Plugin is separate from CLI (can update independently)

**Future Mitigation**:
- Monitor Claude Code release notes
- Abstract agent spawning (if platform changes)
- Maintain compatibility with older Claude Code versions
- Community testing across Claude Code versions

**Likelihood Reduction**: Claude Code is stable platform, Task tool is core feature

**Impact Reduction**: Rapid plugin updates, community communication, fallback to manual workflows

---

#### R-3: Concurrent State Writes Corruption
**Description**: Multiple CLI processes writing same state file simultaneously could corrupt state.

**Current Mitigation**:
- Atomic writes via rename (POSIX guarantees atomicity)
- Schema validation catches corruption on load
- Git versioning provides backup
- Last writer wins (acceptable for user error)

**Future Mitigation**:
- Consider advisory file locking (flock)
- Detect concurrent access, warn user
- Lock-free data structures (not applicable for files)
- Process-level mutex (only helps single machine)

**Likelihood Reduction**: User discipline (don't run multiple sow commands in parallel), rare in practice

**Impact Reduction**: State validation catches corruption, git rollback possible

---

#### R-4: GitHub CLI API Changes
**Description**: `gh` CLI may change output format, break parsing, or change commands.

**Current Mitigation**:
- Wrapper abstraction (`internal/sow/github.go`)
- Parse JSON output (more stable than text)
- Version pinning in documentation (gh 2.0+)
- Graceful error handling, clear messages

**Future Mitigation**:
- Pin `gh` minimum version, test compatibility
- Use `--json` flag where available (structured output)
- Fallback to direct API calls if `gh` breaks
- Community reporting, rapid fixes

**Likelihood Reduction**: `gh` maintains backward compatibility, JSON output is stable

**Impact Reduction**: Wrapper isolation (change in one place), clear errors, fallback options

---

#### R-5: CUE Code Generation Breaks
**Description**: `cue exp gengotypes` may change, break generation, or produce invalid Go code.

**Current Mitigation**:
- Pin CUE version in `go.mod` and CI
- Test generation in CI (build fails if broken)
- Generated types committed to repo (can inspect diffs)

**Future Mitigation**:
- Lock CUE version more strictly
- Manual fallback (hand-write Go types if needed)
- Upstream bug reports to CUE project
- Alternative: Use `cue export` + custom codegen

**Likelihood Reduction**: CUE is stable, gengotypes is mature

**Impact Reduction**: Generated code is committed (can revert), manual fallback possible

---

#### R-6: Filesystem Differences (Windows)
**Description**: Windows doesn't support symlinks reliably, NTFS differences, path separators.

**Current Mitigation**:
- Billy filesystem abstraction (handles paths)
- Platform detection for symlink vs. copy
- Test on Windows in CI (GitHub Actions)
- Clear documentation for Windows users

**Future Mitigation**:
- More comprehensive Windows testing
- Junction points instead of copies (better than copies)
- Community feedback from Windows users
- Document Windows-specific limitations

**Likelihood Reduction**: Billy abstracts most differences, tested in CI

**Impact Reduction**: Fallback to copies, clear Windows docs, community support

---

#### R-7: Git Repository Corruption
**Description**: Git repository corruption would break sow (relies on git for operations).

**Current Mitigation**:
- Never manipulate `.git/` directly (use go-git library)
- Rely on git's integrity checks
- Atomic operations (no partial writes)
- Standard git workflows only

**Future Mitigation**:
- Validate git repository health (`git fsck`)
- Detect corruption early, warn user
- Git provides recovery tools (user responsibility)

**Likelihood Reduction**: Git is extremely reliable, we don't modify `.git/`

**Impact Reduction**: User's git backups (remotes), git's own recovery tools

---

#### R-8: Cache Poisoning (Malicious Refs)
**Description**: User adds malicious git repository as ref, agents execute harmful code.

**Current Mitigation**:
- User responsibility (trust refs they add)
- Refs are just files (read-only from agent perspective)
- No automatic execution of ref code
- Documentation warns about trusting refs

**Future Mitigation**:
- Checksum validation (verify ref SHA against known-good)
- Allowlist/denylist for ref sources
- Community-maintained trusted refs list
- Sandboxing (not feasible with current architecture)

**Likelihood Reduction**: User controls refs, most refs are public style guides

**Impact Reduction**: User responsibility, no automatic execution, docs warning

---

## Technical Debt

| Item | Impact | Effort | Priority | Introduced |
|------|--------|--------|----------|------------|
| **No Migration System** | High | Medium | P1 | Initial design |
| **Limited Windows Testing** | Medium | Medium | P2 | Initial implementation |
| **No Advisory Locking (Concurrent Writes)** | Medium | Low | P3 | Initial design |
| **GitHub CLI Dependency** | Medium | High | P4 | Initial design |
| **No Comprehensive Benchmarks** | Low | Low | P5 | Initial implementation |
| **Hardcoded Prompt Templates** | Low | Medium | P6 | Initial implementation |
| **No Plugin Version Pinning** | Medium | Low | P3 | Initial design |

---

### Debt Details

#### D-1: No Migration System
**Description**: State schema changes require manual migration or break existing repositories.

**Impact**:
- Schema changes are breaking (can't evolve easily)
- Users stuck on old CLI versions
- Painful upgrade path for major changes

**Effort**: Medium
- Design migration system (registry, versioning)
- Implement migrations as Go functions
- Test with fixtures from all versions
- Document migration process

**Priority**: P1 (blocks schema evolution)

**Repayment Plan**:
- Phase 1 (Q1): Design migration system
- Phase 2 (Q1): Implement migration registry
- Phase 3 (Q2): Backfill migrations for existing versions
- Phase 4 (Q2): Test and document

**Alternatives**:
- Accept breaking changes (not viable long-term)
- Manual migration scripts (user burden)
- Version detection without migration (refuse to run)

---

#### D-2: Limited Windows Testing
**Description**: Primarily tested on macOS/Linux, Windows may have issues.

**Impact**:
- Windows users may encounter bugs
- Symlink fallback (copies) not thoroughly tested
- Path separator issues possible

**Effort**: Medium
- Set up Windows testing environment
- Add Windows-specific tests
- Test symlink fallback thoroughly
- Document Windows-specific behavior

**Priority**: P2 (affects user segment)

**Repayment Plan**:
- Phase 1 (Q1): Add Windows CI job (GitHub Actions)
- Phase 2 (Q1): Comprehensive Windows testing
- Phase 3 (Q2): Fix discovered issues
- Phase 4 (Q2): Document Windows-specific behavior

**Alternatives**:
- Community-driven testing (slower, less reliable)
- Drop Windows support (not desirable)
- Accept bugs (poor UX)

---

#### D-3: No Advisory Locking (Concurrent Writes)
**Description**: Multiple CLI processes can write to same state file simultaneously.

**Impact**:
- Race conditions possible (rare in practice)
- Last writer wins (acceptable but not ideal)
- No detection of concurrent access

**Effort**: Low
- Implement file locking (flock on Unix, LockFileEx on Windows)
- Detect lock contention, retry or error
- Test concurrent access scenarios

**Priority**: P3 (low probability, mitigated by atomic writes)

**Repayment Plan**:
- Phase 1 (Q2): Research cross-platform locking
- Phase 2 (Q2): Implement locking wrapper
- Phase 3 (Q2): Test concurrent scenarios
- Phase 4 (Q3): Deploy with feature flag

**Alternatives**:
- Accept current behavior (reasonable for user error)
- Warn users in docs (don't run concurrent commands)
- Process-level mutex (only helps single machine)

---

#### D-4: GitHub CLI Dependency
**Description**: Hard dependency on `gh` CLI limits flexibility.

**Impact**:
- Users must install `gh` separately
- Breaking changes in `gh` affect sow
- Cannot customize GitHub integration easily

**Effort**: High
- Implement direct GitHub API client
- Handle OAuth, token management
- Maintain parity with `gh` features
- Test authentication flows

**Priority**: P4 (low priority, current approach works well)

**Repayment Plan**:
- Phase 1 (Future): Evaluate if worth the effort
- Phase 2 (Future): Implement API client (if needed)
- Phase 3 (Future): Dual support (gh + API)
- Phase 4 (Future): Deprecate gh dependency (if desired)

**Alternatives**:
- Keep `gh` dependency (simplest, works well)
- Optional API client (fallback if `gh` unavailable)
- Accept gh as requirement (reasonable for target audience)

---

#### D-5: No Comprehensive Benchmarks
**Description**: Performance not systematically measured.

**Impact**:
- Regressions may go unnoticed
- No baseline for optimization
- Unclear performance characteristics

**Effort**: Low
- Write benchmark tests (Go benchmarks)
- Measure CLI startup, state load, logging
- Run in CI, track over time
- Document performance baselines

**Priority**: P5 (low impact, nice-to-have)

**Repayment Plan**:
- Phase 1 (Q2): Write benchmark suite
- Phase 2 (Q2): Integrate into CI (track trends)
- Phase 3 (Q3): Optimize hotspots if needed
- Phase 4 (Q3): Document performance characteristics

**Alternatives**:
- Manual performance testing (less systematic)
- Accept current performance (sufficient for now)
- Optimize reactively (when users report issues)

---

#### D-6: Hardcoded Prompt Templates
**Description**: Prompt templates embedded in CLI binary, not easily customizable.

**Impact**:
- Users cannot customize prompts
- Changes require CLI rebuild
- Experimentation difficult

**Effort**: Medium
- Load templates from filesystem (`.sow/prompts/`)
- Fallback to embedded (default)
- Document template customization
- Validate custom templates

**Priority**: P6 (low priority, advanced use case)

**Repayment Plan**:
- Phase 1 (Future): Design template override system
- Phase 2 (Future): Implement filesystem loading
- Phase 3 (Future): Document customization
- Phase 4 (Future): Community template sharing

**Alternatives**:
- Keep embedded templates (simplest)
- Fork CLI for customization (developer burden)
- External template repo (complexity)

---

#### D-7: No Plugin Version Pinning
**Description**: CLI doesn't enforce compatible plugin version.

**Impact**:
- Plugin and CLI may drift out of sync
- State structure changes may break
- Unclear which versions are compatible

**Effort**: Low
- Add version compatibility check
- Plugin declares compatible CLI versions
- CLI checks on startup, warns if mismatch
- Document compatibility matrix

**Priority**: P3 (medium priority, affects stability)

**Repayment Plan**:
- Phase 1 (Q1): Define compatibility versioning (semver)
- Phase 2 (Q1): Implement version check
- Phase 3 (Q1): Document compatibility
- Phase 4 (Q2): Test across version combinations

**Alternatives**:
- Accept drift (causes user confusion)
- Manual version tracking (docs burden)
- Lock CLI and plugin versions together (less flexible)

---

## Debt Repayment Prioritization

**Immediate (Q1 2025)**:
1. D-1: Migration system (blocks evolution)
2. D-7: Plugin version pinning (stability)
3. D-2: Windows testing (user coverage)

**Short-term (Q2 2025)**:
4. D-3: Advisory locking (safety)
5. D-5: Benchmarks (measurement)

**Long-term (Q3+ 2025)**:
6. D-6: Template customization (advanced feature)
7. D-4: GitHub API client (only if needed)

---

## Risk Monitoring

**Quarterly Risk Review**:
- Review probability and impact assessments
- Update mitigation strategies
- Prioritize new risks
- Track debt repayment progress

**Community Feedback**:
- GitHub issues for bug reports
- Feature requests inform debt prioritization
- User surveys (quarterly)
- Community contributions toward debt reduction
