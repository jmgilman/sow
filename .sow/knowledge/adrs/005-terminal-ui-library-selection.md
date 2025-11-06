# ADR: Terminal UI Library Selection for Interactive Project Wizard

**Status**: Proposed
**Date**: 2025-01-06
**Task**: 020
**Deciders**: Architecture Team
**Related**: [Wizard UX Flow](../010/wizard-ux-flow.md), [Wizard Technical Implementation](../010/wizard-technical-implementation.md), [huh Library Verification](../../context/huh-library-verification.md)

## Context

The `sow` project wizard requires an interactive terminal UI to replace the current flag-based interface. The wizard needs to support:

1. **Select prompts** - Choosing from lists (project types, issues, projects)
2. **Input fields** - Text entry with real-time validation and preview
3. **Text areas** - Multi-line input for prompts and descriptions
4. **External editor integration** - Opening $EDITOR for longer text input
5. **Inline error display** - Showing validation errors next to fields
6. **Dynamic forms** - Conditional fields based on user choices
7. **Professional UX** - Clean, intuitive interface matching modern CLI tools

### Critical Requirement: External Editor Support

The ability to open $EDITOR (Ctrl+E) for multi-line input is **non-negotiable**. Users need to compose longer prompts and descriptions using their preferred editor (vim, emacs, VSCode, etc.). This is essential for:
- Writing detailed initial prompts for Claude
- Composing continuation prompts with context
- Providing multi-line project descriptions

Without this feature, the user experience would be significantly degraded for power users.

### Target User Experience

Users should feel like they're using a modern, well-crafted CLI tool similar to:
- `gh` (GitHub CLI) - interactive issue/PR creation
- `npm init` - interactive project setup
- `aws configure` - interactive configuration
- Modern installers and setup wizards

## Decision

**We will use `charmbracelet/huh` as the terminal UI library for the interactive wizard.**

Repository: https://github.com/charmbracelet/huh
License: MIT
Version: Latest stable (v0.x as of writing)

## Rationale

### Why charmbracelet/huh?

#### 1. External Editor Support ✅ Critical

**huh is the ONLY library among the alternatives that supports external editor integration.**

- Built-in via `WithEditor(true)` on text fields
- Keybinding: Ctrl+E (standard across Charm tools)
- Seamless integration with $EDITOR environment variable
- Works with vim, emacs, VSCode, nano, any editor

**Evidence**: [huh Library Verification](../../context/huh-library-verification.md) confirms this works perfectly.

#### 2. Complete Feature Coverage

All required features are fully supported:
- ✅ Select prompts (single and multi-selection)
- ✅ Input validation with inline error display
- ✅ Real-time field updates (via `DescriptionFunc`)
- ✅ Text areas with character limits
- ✅ Dynamic forms (conditional fields via `TitleFunc`, `OptionsFunc`)
- ✅ Built-in help text and keybinding hints

**Evidence**: [huh Library Verification](../../context/huh-library-verification.md) provides code examples for every feature.

#### 3. Clean, Intuitive API

```go
form := huh.NewForm(
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("What would you like to do?").
            Options(
                huh.NewOption("Create new project", "create"),
                huh.NewOption("Continue existing project", "continue"),
            ).
            Value(&action),
    ),
)
```

The API is:
- Type-safe (generic options)
- Declarative (describe what, not how)
- Composable (groups and fields combine naturally)
- Self-documenting (clear method names)

#### 4. Production Ready

- **Active maintenance**: Part of Charm ecosystem, regular updates
- **Well-documented**: Comprehensive README, examples directory, GoDoc
- **Battle-tested**: Used in production by Charm and community projects
- **Stable API**: Semantic versioning, backward compatibility maintained

#### 5. Accessibility Support

- Built-in accessible mode via `WithAccessible(true)`
- Screen-reader friendly text prompts
- Keyboard-only navigation
- No additional effort required to support

#### 6. Part of Proven Ecosystem

huh is built on Bubble Tea, Charm's TUI framework:
- **Bubble Tea**: 26k+ GitHub stars, proven architecture
- **Lip Gloss**: Styling (if we need custom themes later)
- **Spinner**: Loading indicators (available when needed)
- **Log**: Structured logging integration

Using huh gives us access to the entire Charm ecosystem if we need to extend functionality.

#### 7. Excellent Examples and Community

- 20+ examples in repository covering every use case
- Active community on Charm Discord
- Good Stack Overflow presence
- Clear issue tracker with responsive maintainers

### Why Not Alternatives?

#### Alternative 1: promptui

**Repository**: https://github.com/manifoldco/promptui
**Stars**: ~6k

**Pros**:
- Simple API
- Lightweight
- Good for basic prompts

**Cons**:
- ❌ **No external editor support** (deal-breaker)
- ❌ No text area widget
- ❌ Limited form composition
- ❌ Stale maintenance (last major update 2020)
- ❌ No dynamic forms

**Verdict**: Missing critical feature (editor) and insufficient for complex forms.

#### Alternative 2: survey

**Repository**: https://github.com/AlecAivazis/survey
**Stars**: ~4k

**Pros**:
- Rich set of prompt types
- Good validation support

**Cons**:
- ❌ **No external editor support** (deal-breaker)
- ❌ Limited real-time updates
- ❌ Less active maintenance
- ❌ API feels dated compared to modern Go practices

**Verdict**: Missing critical feature and less polished API.

#### Alternative 3: Bubble Tea (directly)

**Repository**: https://github.com/charmbracelet/bubbletea
**Stars**: 26k+

**Pros**:
- Maximum flexibility
- Full control over behavior
- Same ecosystem as huh

**Cons**:
- ❌ **Massive implementation overhead** - MVU architecture requires significant boilerplate
- ❌ Need to build form primitives from scratch
- ❌ Complex state management for multi-screen flows
- ❌ Longer development time (weeks vs days)
- ❌ More surface area for bugs

**Verdict**: Overkill for our needs. huh is built on Bubble Tea and provides exactly what we need without the complexity.

#### Alternative 4: tview

**Repository**: https://github.com/rivo/tview
**Stars**: ~10k

**Pros**:
- Full TUI framework
- Rich widget set
- Good for complex layouts

**Cons**:
- ❌ **Designed for persistent TUIs, not wizards** (wrong paradigm)
- ❌ Complex layout management
- ❌ Overkill for sequential forms
- ❌ Steeper learning curve
- ❌ No built-in form wizard patterns

**Verdict**: Wrong tool for the job. Designed for persistent TUI apps (like htop), not interactive wizards.

### Decision Matrix

| Feature | huh | promptui | survey | Bubble Tea | tview |
|---------|-----|----------|--------|------------|-------|
| External Editor | ✅ **YES** | ❌ No | ❌ No | ⚠️ DIY | ⚠️ DIY |
| Select Prompts | ✅ | ✅ | ✅ | ⚠️ DIY | ✅ |
| Input Validation | ✅ | ✅ | ✅ | ⚠️ DIY | ✅ |
| Text Areas | ✅ | ❌ | ⚠️ Limited | ⚠️ DIY | ✅ |
| Real-time Updates | ✅ | ❌ | ⚠️ Limited | ✅ | ✅ |
| Dynamic Forms | ✅ | ❌ | ❌ | ⚠️ DIY | ⚠️ Complex |
| API Simplicity | ✅ Clean | ✅ Simple | ⚠️ OK | ❌ Complex | ❌ Complex |
| Maintenance | ✅ Active | ⚠️ Stale | ⚠️ Slow | ✅ Active | ✅ Active |
| Development Time | ✅ Days | ✅ Days | ✅ Days | ❌ Weeks | ❌ Weeks |
| **Suitability** | ✅ **Perfect** | ❌ Insufficient | ❌ Insufficient | ❌ Overkill | ❌ Wrong paradigm |

## Consequences

### Positive

1. **Fast development**: Clean API enables rapid implementation of all wizard screens
2. **Excellent UX**: Professional, polished interface matches user expectations
3. **External editor support**: Power users can compose prompts in their editor of choice
4. **Maintainability**: Well-documented code, clear patterns, active upstream
5. **Future-proof**: Part of growing ecosystem with long-term support
6. **Accessibility**: Built-in support for screen readers and keyboard-only usage
7. **Low risk**: Proven library with strong community adoption

### Negative

1. **Learning curve**: Team needs to learn huh API and patterns
   - **Mitigation**: Excellent examples and documentation available
   - **Mitigation**: Similar API to other Charm tools if team already familiar

2. **Dependency**: Adding external dependency to project
   - **Mitigation**: MIT license, no legal concerns
   - **Mitigation**: Small dependency footprint (~10 transitive deps, all Charm ecosystem)
   - **Mitigation**: Active maintenance reduces risk of abandonment

3. **Version lock-in**: Need to track huh versions and updates
   - **Mitigation**: Semantic versioning maintained
   - **Mitigation**: Go modules make version management straightforward

4. **Limited customization**: Can't easily modify core behavior
   - **Mitigation**: Library provides sufficient configuration options
   - **Mitigation**: Can drop to Bubble Tea if truly custom behavior needed (escape hatch)

### Risks

**Low Risk**: External editor bug ([GitHub issue #686](https://github.com/charmbracelet/huh/issues/686))
- When using Ctrl+E, may overwrite all fields in current group
- **Our design is safe**: Each text area is in its own group, so bug won't affect us
- **Workaround available**: If encountered, split into multiple forms
- **Status**: Non-blocking, documented, has workaround

**Very Low Risk**: Breaking API changes in future versions
- **Mitigation**: Pin to known-good version
- **Mitigation**: Test before upgrading
- **Mitigation**: Active maintainers responsive to migration concerns

## Implementation Notes

### Getting Started

```go
import (
    "github.com/charmbracelet/huh"
)

// Minimal example
func main() {
    var name string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Enter project name:").
                Value(&name),
        ),
    )

    if err := form.Run(); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Project name: %s\n", name)
}
```

### Key Patterns

**Validation**:
```go
huh.NewInput().
    Validate(func(s string) error {
        if strings.TrimSpace(s) == "" {
            return fmt.Errorf("cannot be empty")
        }
        return nil
    })
```

**Real-time Preview**:
```go
huh.NewNote().
    DescriptionFunc(func() string {
        return fmt.Sprintf("Preview: %s", computeValue(input))
    }, &input) // Bind to input field
```

**External Editor**:
```go
huh.NewText().
    WithEditor(true) // Enables Ctrl+E
```

**Dynamic Options**:
```go
huh.NewSelect[string]().
    OptionsFunc(func() []huh.Option[string] {
        // Compute options dynamically
        return buildOptions()
    }, &dependency) // Recompute when dependency changes
```

### Testing Approach

1. **Unit tests**: Validation functions, helper functions
2. **Integration tests**: Mock form inputs, test state transitions
3. **Manual tests**: Actual terminal testing for UX validation

### Installation

```bash
go get github.com/charmbracelet/huh@latest
```

No additional setup or configuration required.

## Alternatives Reconsidered

If huh proves insufficient (unlikely based on verification), fallback options:

1. **Bubble Tea directly**: Full control, but significant dev time investment
2. **Hybrid approach**: Use huh for most screens, Bubble Tea for complex ones
3. **Custom forms**: Build minimal form library on top of termbox-go or similar

**Confidence**: High that huh will meet all needs without fallback required.

## References

- **huh Repository**: https://github.com/charmbracelet/huh
- **huh Examples**: https://github.com/charmbracelet/huh/tree/main/examples
- **huh Documentation**: https://pkg.go.dev/github.com/charmbracelet/huh
- **Bubble Tea**: https://github.com/charmbracelet/bubbletea
- **Charm Ecosystem**: https://charm.sh/
- **[huh Library Verification](../../context/huh-library-verification.md)**: Complete capability verification with code examples
- **[Wizard UX Flow](../010/wizard-ux-flow.md)**: User experience design
- **[Wizard Technical Implementation](../010/wizard-technical-implementation.md)**: Implementation architecture

## Decision History

**Proposed**: 2025-01-06
**Status**: Pending review and approval
**Deciders**: Architecture Team

---

## Appendix: Feature Verification Summary

This ADR is supported by comprehensive feature verification documented in [huh Library Verification](../../context/huh-library-verification.md).

**Verification Results**:
- ✅ All required features confirmed working
- ✅ Code examples provided for each feature
- ✅ Edge cases documented
- ✅ Known issues identified and assessed (non-blocking)
- ✅ Alternative approaches documented where applicable

**Confidence Level**: Very High

The verification report provides detailed evidence that huh can implement the complete wizard design without compromises.
