━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

DISCOVERY PHASE DECISION

PROJECT: {{.ProjectName}}
DESCRIPTION: {{.ProjectDescription}}

Determine if discovery phase is warranted for this work.

DISCOVERY WORTHINESS RUBRIC (score 0-2 for each):

  1. Context Availability
     0 = Comprehensive docs exist
     1 = Some notes/requirements but gaps
     2 = No context available

  2. Problem Clarity
     0 = Crystal clear what needs to be done
     1 = Generally clear but some unknowns
     2 = Unclear, needs investigation

  3. Codebase Familiarity
     0 = No investigation needed (familiar area)
     1 = Some review needed
     2 = Significant exploration required

  4. Research Needs
     0 = No external research needed
     1 = Light research would help
     2 = Substantial research required

SCORING:
  0-2 points: Skip discovery (recommend strongly)
  3-5 points: Optional (ask user preference)
  6-8 points: Enable discovery (recommend strongly)

RECOMMENDED PHRASING:
  • 0-2: "Based on [existing docs/clear requirements], skip discovery
         and go to [design/implementation]?"
  • 3-5: "We could do discovery to [goal], but also reasonable to
         proceed. Your preference?"
  • 6-8: "Recommend discovery phase to [goal]. This will help us
         [benefit]. Sound good?"

NEXT ACTION:
  If discovery needed:
    sow agent project phase enable discovery --type <bug|feature|docs|refactor|general>

  If discovery not needed:
    sow agent project phase skip discovery
    (Will auto-transition to design decision)

Reference: PROJECT_LIFECYCLE.md (Discovery Rubric)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
