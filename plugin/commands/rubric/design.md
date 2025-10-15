# /rubric:design - Design Phase Assessment

**Purpose**: Determine if design phase is warranted

**Score 0-2 for each**:
1. **Scope**: 0=1-3 tasks, 1=4-9 tasks, 2=10+ tasks
2. **Architecture**: 0=no changes, 1=minor adjustments, 2=significant changes/new components
3. **Integration**: 0=self-contained, 1=1-2 integrations, 2=3+ integrations/external systems
4. **Decisions**: 0=straightforward, 1=1-2 decisions worth documenting, 2=multiple decisions/ADRs needed

**Bug Fix Penalty**: If work type is bug fix, subtract 3 points (minimum 0)

**Result**:
- **0-2**: Skip design (recommend strongly)
- **3-5**: Optional (ask user preference)
- **6-8**: Enable design (recommend strongly)

**Phrasing**:
- 0-2: "Based on [small scope/straightforward nature], skip formal design docs and go to implementation?"
- 3-5: "Medium-sized project. Create design documents or proceed with [existing context]? Your preference?"
- 6-8: "Given [large scope/complexity], recommend formal design docs (ADRs, design docs) before implementation. Sound good?"
