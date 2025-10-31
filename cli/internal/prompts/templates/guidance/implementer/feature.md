# Feature Implementation Workflow

Use this guidance when implementing **new functionality** that doesn't exist yet in the codebase.

## Feature Implementation Process

### Phase 1: Understand the Requirement

Before writing any code, ensure you understand:

1. **What** needs to be built (functional requirements)
2. **Why** it's needed (context, use case)
3. **How** it should behave (acceptance criteria)
4. **Where** it fits in the existing architecture

Read description.md carefully. If anything is unclear, stop and escalate.

### Phase 2: Break Down into Testable Units

Decompose the feature into small, independently testable units:

**Example**: "Implement JWT authentication middleware"

Break down:
1. Token parsing from Authorization header
2. Token signature validation
3. Token expiration checking
4. User context extraction from token
5. Request rejection on invalid token
6. Request continuation on valid token

Each unit becomes a test (or set of tests).

### Phase 3: Identify Integration Points

Determine where your feature integrates with existing code:

- **Inbound**: How does the feature get invoked? (HTTP endpoint, function call, event listener)
- **Outbound**: What external systems does it interact with? (database, API, cache)
- **Data structures**: What domain models does it use or create?

**Important**: For outbound dependencies, verify that **ports (interfaces) exist** for mocking. If not, stop and escalate.

### Phase 4: Test-First Implementation

For each testable unit:

1. **Write the test first**
   - Describe the behavior you want
   - Use mocks for all external dependencies
   - Test one behavior per test

2. **Watch it fail**
   - Run the test, confirm it fails (red)
   - If it passes without implementation, your test is wrong

3. **Write minimal implementation**
   - Just enough code to make the test pass
   - Don't over-engineer or add extra features

4. **Watch it pass**
   - Run the test, confirm it passes (green)

5. **Refactor if needed**
   - Improve code quality without changing behavior
   - Run tests again to ensure they still pass

6. **Repeat** for the next unit

### Phase 5: Integration Verification

After implementing all units:

1. **Run the full test suite** - Ensure you didn't break existing functionality
2. **Verify your feature works end-to-end** (if possible in your test environment)
3. **Check code coverage** - Aim for >90% of your new code

## Feature-Specific Guidance

### Creating New Files vs Modifying Existing

**Create new files when:**
- Adding a new domain concept or service
- Implementing a new layer (new adapter, new port)
- Tests for new code

**Modify existing files when:**
- Extending existing functionality
- Adding to an existing service or handler
- Integrating with existing code

**Rule of thumb**: Follow the existing code organization patterns. If authentication code lives in `pkg/auth/`, your new auth feature should too.

### Working with Existing Code

Your feature likely integrates with existing code:

1. **Read the existing code** - Understand how it works
2. **Follow existing patterns** - Match the style and structure
3. **Respect abstractions** - Use ports/interfaces, don't bypass them
4. **Don't refactor while implementing** - That's a separate task

If existing code is problematic (no testability, tight coupling), stop and escalate. Don't fix architecture issues during feature implementation.

### Incremental Implementation Approach

Build features incrementally, not all at once:

**✅ Good (incremental):**
1. Implement token parsing → test → commit
2. Implement signature validation → test → commit
3. Implement expiration checking → test → commit
4. (Continue for each unit)

**❌ Bad (big bang):**
1. Write all the code for the entire feature
2. Try to make all tests pass at once
3. Spend hours debugging

Incremental commits give you:
- Easier debugging (smaller change sets)
- Clearer progress
- Ability to roll back partially if needed

### Documentation

For new features, update documentation:

1. **Code comments** - Explain non-obvious decisions
2. **README** - If feature affects usage, update README
3. **API docs** - If adding public APIs, document them

Don't write extensive design docs (that's the architect's job). Just document **usage** and **non-obvious implementation choices**.

## Common Feature Implementation Pitfalls

**❌ Over-engineering**: Adding flexibility beyond requirements
- Build what's specified, not what might be needed later (YAGNI)

**❌ Skipping tests**: "I'll add tests later"
- Tests never get added, or you discover your code isn't testable

**❌ Big changes without testing**: Implementing entire feature before running tests
- When tests fail, you have no idea what's wrong

**❌ Modifying requirements**: "I think it should work this way instead"
- Implement what's specified; propose changes separately

**❌ Refactoring existing code as part of feature work**
- Mixing refactoring with new features makes review difficult

## Feature Completion Checklist

Before marking the task as needs_review:

- [ ] All requirements from description.md are implemented
- [ ] All new code has tests
- [ ] All tests pass (including existing tests)
- [ ] Code coverage >90% for new code
- [ ] No real database/API calls in unit tests (all mocked)
- [ ] Code follows existing patterns and conventions
- [ ] Modified files tracked: `sow agent task state add-file <path>`
- [ ] Actions logged to task log.md
- [ ] No "TODO" or "FIXME" comments left in code

## Example: JWT Middleware Implementation

**Requirement**: "Implement JWT authentication middleware that validates tokens and sets user context"

**Test-first approach:**

```typescript
// Test 1: Parsing
test("extracts token from Authorization header", () => {
  const req = { headers: { authorization: "Bearer abc123" } };
  const token = extractToken(req);
  expect(token).toBe("abc123");
});

// Implementation 1: Just enough to pass
function extractToken(req): string {
  const auth = req.headers.authorization;
  return auth.split(" ")[1];
}

// Test 2: Validation
test("validates token signature", async () => {
  const mockVerifier = { verify: jest.fn().mockResolvedValue(true) };
  const result = await validateToken("valid-token", mockVerifier);
  expect(result).toBe(true);
});

// Implementation 2: Just enough to pass
async function validateToken(token, verifier): Promise<boolean> {
  return await verifier.verify(token);
}

// Continue for all units...
```

Notice:
- Each test is small and focused
- Implementation is minimal
- External dependencies (token verifier) are mocked
- Tests define behavior before implementation exists

## When to Stop and Escalate

Stop if:
- Requirements are ambiguous (which scenarios should be handled?)
- No interface exists for external dependency (can't mock)
- Existing code structure prevents proper integration
- Feature requires architectural changes

Don't guess or make assumptions. Escalate and wait for clarification.

## Next Steps

You've loaded the feature implementation guidance. Proceed with:

1. Breaking down your feature into testable units
2. Writing tests first
3. Implementing incrementally
4. Logging your progress

Remember: Test first, always. One unit at a time. Mock external dependencies.
