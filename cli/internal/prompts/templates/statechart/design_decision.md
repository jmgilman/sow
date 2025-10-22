━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DESIGN PHASE DECISION

PROJECT: {{.ProjectName}}
DESCRIPTION: {{.ProjectDescription}}
{{if .HasDiscovery}}DISCOVERY: Completed ({{.DiscoveryArtifactCount}} artifacts){{end}}

Determine if design phase is warranted for this work.

DESIGN WORTHINESS RUBRIC (score 0-2 for each):

  1. Scope Size
     0 = 1-3 tasks expected
     1 = 4-9 tasks expected
     2 = 10+ tasks expected

  2. Architectural Impact
     0 = No architectural changes
     1 = Minor adjustments to existing patterns
     2 = Significant changes/new components

  3. Integration Complexity
     0 = Self-contained work
     1 = 1-2 integration points
     2 = 3+ integrations or external systems

  4. Design Decisions
     0 = Straightforward implementation
     1 = 1-2 decisions worth documenting
     2 = Multiple decisions, ADRs needed

BUG FIX PENALTY: If this is a bug fix, subtract 3 points (minimum 0)

SCORING:
  0-2 points: Skip design (recommend strongly)
  3-5 points: Optional (ask user preference)
  6-8 points: Enable design (recommend strongly)

RECOMMENDED PHRASING:
  • 0-2: "Based on [small scope/straightforward nature], skip formal
         design docs and go to implementation?"
  • 3-5: "Medium-sized project. Create design documents or proceed
         with [existing context]? Your preference?"
  • 6-8: "Given [large scope/complexity], recommend formal design docs
         (ADRs, design docs) before implementation. Sound good?"

NEXT ACTION:
  If design needed:
    sow agent project phase enable design

  If design not needed:
    sow agent project phase skip design
    (Will auto-transition to implementation)

Reference: PROJECT_LIFECYCLE.md (Design Rubric)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
