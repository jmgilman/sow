# Bug Fixing Workflow

Use this guidance when **fixing defects** in existing functionality.

## Bug Fixing Process

### Step 1: Reproduce the Bug

Before fixing anything, reproduce the bug in a test:

1. **Understand the expected behavior** - What should happen?
2. **Understand the actual behavior** - What's happening instead?
3. **Identify the trigger** - What input/conditions cause the bug?

**Write a failing test that demonstrates the bug.**

This is critical because:
- Confirms you understand the bug
- Proves the bug exists
- Prevents regression (if test passes later, bug is fixed)
- Documents the bug for future reference

### Step 2: Write a Failing Test

Create a test that:
- Reproduces the bug condition
- Asserts the CORRECT behavior (not the current buggy behavior)
- Fails when run

**Example**: Bug report says "User login fails with special characters in password"

```python
# Test that demonstrates the bug
def test_login_succeeds_with_special_characters_in_password():
    user = User(username="alice", password="P@ssw0rd!")
    result = auth_service.login("alice", "P@ssw0rd!")
    assert result.success is True  # This will FAIL (bug exists)
```

Run the test. It should fail. If it passes, either:
- You didn't reproduce the bug correctly
- The bug has already been fixed
- The bug report is incorrect

### Step 3: Locate the Root Cause

Don't just fix symptoms. Find the root cause:

**Debugging techniques:**
- **Read the failing test output** - What exactly failed?
- **Trace the code path** - Follow execution from trigger to failure
- **Check assumptions** - Are there wrong assumptions in the code?
- **Review related code** - Bug might be in integration between components

**Common root causes:**
- Wrong conditional logic (`if` condition is inverted or incomplete)
- Missing validation (edge case not handled)
- Incorrect assumptions (assumed input would always be X)
- Off-by-one errors (array indices, loop conditions)
- Type coercion issues (string vs number, null vs undefined)

### Step 4: Fix the Implementation

Make the minimal change to fix the bug:

1. **Fix the root cause** - Don't just patch symptoms
2. **Make the test pass** - Run your failing test, confirm it now passes
3. **Don't add extra features** - Fix this bug only (scope discipline)

**Example**: Root cause was password validation regex didn't allow special characters

```python
# Before (buggy)
def is_valid_password(password):
    return re.match(r'^[A-Za-z0-9]+$', password)  # No special chars!

# After (fixed)
def is_valid_password(password):
    return len(password) >= 8  # Allow any characters
```

### Step 5: Verify No Regressions

Run the **full test suite**, not just your new test:

```bash
# Run all tests to ensure you didn't break anything
pytest
npm test
go test ./...
```

If existing tests fail, your fix broke something else. Adjust your fix.

### Step 6: Test Edge Cases

Your bug might have related edge cases. Test them too:

**Original bug**: Login fails with special characters
**Related edge cases to test:**
- Emoji in password
- Very long passwords
- Passwords with spaces
- Passwords with unicode characters

Add tests for any edge cases that aren't covered.

## Bug-Specific Guidance

### Root Cause Analysis Mindset

Ask "why" multiple times:

**Bug**: "Dashboard crashes when user has no orders"

- Why does it crash? → Because `orders.length` throws error
- Why does that throw an error? → Because `orders` is `null`
- Why is `orders` null? → Because API returns `null` for users with no orders
- **Root cause**: Code assumes `orders` is always an array, doesn't handle `null` case

**Fix**: Add null check:
```typescript
// Before (buggy)
const orderCount = orders.length;  // Crashes if orders is null

// After (fixed)
const orderCount = orders?.length ?? 0;  // Handles null safely
```

### Avoid Scope Creep

Bugs often reveal related issues. **Don't fix everything at once.**

**Example**: While fixing login bug, you notice:
- Password validation could be stronger
- Error messages could be more helpful
- There's duplicate validation code

**Correct approach:**
1. Fix the reported bug (special characters in password)
2. Log observations about other issues
3. Let the orchestrator decide if those should become separate tasks

**Incorrect approach:**
- "While I'm here, I'll refactor the entire auth module..."
- Now your "bug fix" is 500 lines of changed code
- Review becomes difficult
- Risk of introducing new bugs increases

### Defensive Programming

When fixing bugs, add defensive checks to prevent similar bugs:

**Example**: Bug was division by zero

```python
# Minimal fix
def calculate_average(total, count):
    if count == 0:
        return 0
    return total / count

# Defensive improvement
def calculate_average(total, count):
    if count is None or count <= 0:  # Also handles negative/null
        return 0
    return total / count
```

But don't over-engineer. Add checks for realistic scenarios, not hypothetical ones.

### Regression Test Value

The test you write for this bug is a **regression test**:
- Proves the bug is fixed
- Prevents the bug from coming back
- Documents the bug for future developers

Keep regression tests even after the bug is fixed. They're permanent guards.

## Common Bug Fixing Pitfalls

**❌ Fixing without reproducing**: Guessing at the fix without confirming the bug
- You might fix the wrong thing or miss the real issue

**❌ Fixing symptoms instead of root cause**: "I'll just add a try/catch here"
- Bug will manifest somewhere else later

**❌ Expanding scope**: "I'll refactor this while I'm fixing the bug"
- Makes review difficult, increases risk

**❌ Not testing edge cases**: Only testing the exact reported scenario
- Related bugs remain

**❌ Not running full test suite**: Only running your new test
- You might have broken something else

**❌ Skipping the failing test**: "I'll just fix it directly"
- No proof the bug is fixed, no regression protection

## Bug Completion Checklist

Before marking the task as needs_review:

- [ ] Bug reproduced with failing test
- [ ] Test fails before fix (proves bug exists)
- [ ] Test passes after fix (proves bug is fixed)
- [ ] Full test suite passes (no regressions)
- [ ] Edge cases tested
- [ ] Root cause fixed (not just symptoms)
- [ ] No scope creep (only this bug fixed)
- [ ] Modified files tracked: `sow agent task state add-file <path>`
- [ ] Actions logged to task log.md

## Example: Bug Fix Workflow

**Bug report**: "User profile update fails silently when email is invalid"

### Step 1: Reproduce

```typescript
// Write failing test
test("update profile returns error when email is invalid", async () => {
  const result = await profileService.updateEmail("user123", "notanemail");
  expect(result.success).toBe(false);
  expect(result.error).toContain("invalid email");
});

// Run test → FAILS (bug exists)
// Actual: result.success is true, but email isn't updated
```

### Step 2: Root cause

Trace code:
```typescript
// Found the bug!
async function updateEmail(userId, newEmail) {
  // Email validation happens here
  if (!isValidEmail(newEmail)) {
    console.log("Invalid email");  // ❌ Just logs, doesn't return error!
  }

  // Email gets updated anyway
  await db.updateUser(userId, { email: newEmail });
  return { success: true };
}
```

**Root cause**: Validation doesn't actually prevent the update.

### Step 3: Fix

```typescript
async function updateEmail(userId, newEmail) {
  if (!isValidEmail(newEmail)) {
    return { success: false, error: "Invalid email format" };  // ✅ Return error
  }

  await db.updateUser(userId, { email: newEmail });
  return { success: true };
}
```

### Step 4: Verify

```bash
# Run the new test → PASSES
# Run full suite → ALL PASS
```

### Step 5: Edge cases

```typescript
test("update profile handles null email", async () => {
  const result = await profileService.updateEmail("user123", null);
  expect(result.success).toBe(false);
});

test("update profile handles empty email", async () => {
  const result = await profileService.updateEmail("user123", "");
  expect(result.success).toBe(false);
});
```

Done! Bug fixed, regression test added, edge cases covered.

## When to Stop and Escalate

Stop if:
- Cannot reproduce the bug (might be environment-specific, might be fixed)
- Root cause is in architectural layer (design issue, not code bug)
- Fix requires API changes or breaking changes
- Bug is actually a feature request in disguise
- Multiple bugs are intertwined (needs decomposition)

Don't guess. Escalate and provide details about what you discovered.

## Next Steps

You've loaded the bug fixing guidance. Proceed with:

1. Reproducing the bug in a test
2. Locating the root cause
3. Fixing the implementation
4. Verifying no regressions
5. Logging your progress

Remember: Test first (reproduce the bug), fix the root cause, verify with full test suite.
