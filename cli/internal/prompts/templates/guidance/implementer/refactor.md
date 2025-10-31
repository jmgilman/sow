# Refactoring Workflow

Use this guidance when **improving code structure or quality** without changing external behavior.

## What is Refactoring?

**Refactoring**: Restructuring existing code without changing its observable behavior.

**Key principle**: Tests pass before refactoring, tests pass after refactoring. Same tests.

**Not refactoring:**
- Fixing bugs (behavior changes)
- Adding features (behavior changes)
- Performance optimization without behavior preservation
- Rewriting code from scratch

## Refactoring Process

### Step 0: Prerequisites (Critical!)

Before any refactoring, ensure:

1. **Comprehensive test coverage exists**
   - Tests for the code you're refactoring must already exist
   - Tests must pass before you start
   - If no tests exist, STOP and escalate (need to write tests first)

2. **Tests verify behavior, not implementation**
   - Tests should test public API and outcomes
   - Tests should NOT test internal structure
   - If tests are implementation-coupled, they'll break during refactoring

**Without good tests, refactoring is dangerous.** You can't prove behavior is preserved.

### Step 1: Understand Current Behavior

Before changing anything:

1. **Run the test suite** - Everything should pass
2. **Read the code** - Understand what it does
3. **Identify the problem** - What needs improvement?
   - Poor naming
   - Duplicate code
   - Complex logic
   - Tight coupling
   - Long functions/methods
   - Poor structure

### Step 2: Plan Small Steps

Break refactoring into tiny, incremental changes:

**❌ Bad (big bang):**
- "Refactor entire authentication module"
- Change everything at once
- Tests might break, hard to debug

**✅ Good (incremental):**
1. Extract method from long function
2. Run tests → Pass
3. Rename variables for clarity
4. Run tests → Pass
5. Move method to appropriate class
6. Run tests → Pass
7. Remove duplication
8. Run tests → Pass

Each step is reversible. If tests fail, you know exactly what broke them.

### Step 3: Make One Change at a Time

**The refactoring cycle:**

1. **Make a small change** (extract, rename, move, etc.)
2. **Run tests** - Must pass
3. **Commit** (optional, but helpful for rollback)
4. **Repeat**

If tests fail:
- Your refactoring changed behavior (undo and try differently)
- OR your tests were testing implementation details (fix the tests)

### Step 4: Common Refactoring Patterns

#### Extract Method

**When**: Function is too long or does multiple things

```python
# Before
def process_order(order):
    # Validate order (15 lines)
    # Calculate total (10 lines)
    # Apply discount (8 lines)
    # Save to database (5 lines)

# After
def process_order(order):
    validate_order(order)
    total = calculate_total(order)
    total = apply_discount(total, order.user)
    save_order(order)

def validate_order(order):
    # 15 lines moved here

def calculate_total(order):
    # 10 lines moved here
    return total
```

Run tests after EACH extraction.

#### Rename

**When**: Names are unclear or misleading

```typescript
// Before
function calc(u, p) {  // What is this?
  return u * p * 0.8;
}

// After
function calculateDiscountedPrice(unitPrice, quantity) {
  const DISCOUNT_RATE = 0.8;
  return unitPrice * quantity * DISCOUNT_RATE;
}
```

#### Remove Duplication

**When**: Same code appears in multiple places

```python
# Before
def process_payment_card(card):
    logger.log("Processing card payment")
    validate_card(card)
    result = charge_card(card)
    logger.log("Card payment processed")
    return result

def process_payment_paypal(paypal):
    logger.log("Processing PayPal payment")
    validate_paypal(paypal)
    result = charge_paypal(paypal)
    logger.log("PayPal payment processed")
    return result

# After
def process_payment(payment_method):
    logger.log(f"Processing {payment_method.type} payment")
    validate(payment_method)
    result = charge(payment_method)
    logger.log(f"{payment_method.type} payment processed")
    return result
```

#### Simplify Conditional Logic

**When**: Complex nested conditions

```javascript
// Before
if (user.role === 'admin' || user.role === 'moderator') {
  if (user.verified && user.active) {
    if (resource.visibility === 'private') {
      return true;
    }
  }
}
return false;

// After
function canAccessPrivateResource(user, resource) {
  const isStaff = user.role === 'admin' || user.role === 'moderator';
  const isAuthorized = user.verified && user.active;
  const isPrivate = resource.visibility === 'private';

  return isStaff && isAuthorized && isPrivate;
}
```

## Red-Green-Refactor Cycle

Refactoring is the third step in TDD:

1. **Red**: Write failing test
2. **Green**: Write minimal code to pass
3. **Refactor**: Improve code quality ← You are here
4. Confirm tests still pass
5. Repeat

After making tests pass, you're allowed to improve the code. Just keep the tests passing.

## Refactoring vs. Rewriting

**Refactoring**: Incremental improvements while preserving behavior
**Rewriting**: Throwing away old code and starting over

**Prefer refactoring over rewriting** because:
- Refactoring is lower risk (tests verify each step)
- Refactoring preserves working behavior
- Rewriting often introduces new bugs
- Rewriting takes longer

**Only rewrite if:**
- Code is truly unsalvageable
- Tests don't exist and can't be written (architectural issue)
- Architecture needs fundamental change

If you think a rewrite is needed, escalate to orchestrator. This is a design decision.

## When to Stop Refactoring

Refactoring can go on forever. Know when to stop:

**Stop when:**
- Code is "good enough" (doesn't have to be perfect)
- Improvements are getting smaller and smaller (diminishing returns)
- You've addressed the specific issue mentioned in description.md
- Further changes would require API changes (breaking changes)

**Don't refactor just because you can**. Refactor when it provides clear value:
- Makes code easier to understand
- Makes code easier to test
- Removes duplication
- Simplifies complex logic

## Common Refactoring Pitfalls

**❌ Refactoring without tests**: Can't prove behavior is preserved
- Write tests first, then refactor

**❌ Changing behavior during refactoring**: "While I'm here, I'll fix this bug too"
- Refactoring changes structure, not behavior
- Fix bugs separately

**❌ Big-bang refactoring**: Changing everything at once
- Incremental changes are safer and easier to debug

**❌ Refactoring for its own sake**: Making code "perfect"
- Refactor when there's a specific problem to solve

**❌ Mixing refactoring with feature work**: "I'll implement the feature AND refactor"
- Separate concerns: either refactor OR implement, not both

**❌ Breaking tests and "fixing" them**: Tests fail, so you change the tests
- If tests fail, you changed behavior (undo refactoring)
- Only fix tests if they were testing implementation details

## Refactoring Checklist

Before marking the task as needs_review:

- [ ] Test suite passes BEFORE refactoring
- [ ] Test suite passes AFTER refactoring
- [ ] Same tests pass (didn't need to change tests)
- [ ] No behavior changes (feature still works the same)
- [ ] Code is more readable/maintainable than before
- [ ] Addressed the specific improvement mentioned in description.md
- [ ] No unnecessary changes (stayed focused on the goal)
- [ ] Modified files tracked: `sow agent task state add-file <path>`
- [ ] Actions logged to task log.md

## Example: Refactoring Session

**Task**: "Refactor UserService.authenticate() - too complex and hard to test"

### Before Refactoring

```typescript
class UserService {
  async authenticate(username: string, password: string) {
    // 50 lines of complex logic mixing:
    // - Database queries
    // - Password hashing
    // - Session creation
    // - Logging
    // - Error handling
  }
}
```

**Problems:**
- Too long (50 lines)
- Does too many things
- Hard to test (mixes database calls with business logic)

### Refactoring Steps

**Step 1**: Extract password validation
```typescript
async authenticate(username, password) {
  const isValid = await this.validatePassword(username, password);
  // ... rest of the 45 lines
}

private async validatePassword(username, password) {
  // 5 lines extracted
}
```
Run tests → Pass ✓

**Step 2**: Extract user lookup
```typescript
async authenticate(username, password) {
  const user = await this.findUser(username);
  const isValid = await this.validatePassword(user, password);
  // ... rest of the 40 lines
}

private async findUser(username) {
  // 5 lines extracted
}
```
Run tests → Pass ✓

**Step 3**: Extract session creation
```typescript
async authenticate(username, password) {
  const user = await this.findUser(username);
  const isValid = await this.validatePassword(user, password);

  if (!isValid) {
    throw new AuthenticationError();
  }

  const session = await this.createSession(user);
  this.logAuthSuccess(user);
  return session;
}

private async createSession(user) {
  // 10 lines extracted
}
```
Run tests → Pass ✓

**Step 4**: Extract logging
```typescript
private logAuthSuccess(user) {
  // 3 lines extracted
}
```
Run tests → Pass ✓

### After Refactoring

```typescript
class UserService {
  async authenticate(username: string, password: string) {
    const user = await this.findUser(username);
    const isValid = await this.validatePassword(user, password);

    if (!isValid) {
      throw new AuthenticationError();
    }

    const session = await this.createSession(user);
    this.logAuthSuccess(user);
    return session;
  }

  // Private methods extracted (each testable independently)
  private async findUser(username) { ... }
  private async validatePassword(user, password) { ... }
  private async createSession(user) { ... }
  private logAuthSuccess(user) { ... }
}
```

**Result:**
- Main method now 10 lines (was 50)
- Each concern separated
- Easier to test (can test each method independently if needed)
- Same behavior (all tests still pass)

## When to Stop and Escalate

Stop if:
- No tests exist for the code you're refactoring
- Tests are tightly coupled to implementation (need to be rewritten)
- Refactoring requires API changes (breaking changes)
- Refactoring reveals architectural problems
- Code is so bad it needs rewriting, not refactoring

Escalate with clear description of the problem and why refactoring isn't sufficient.

## Next Steps

You've loaded the refactoring guidance. Proceed with:

1. Verifying comprehensive test coverage exists
2. Running tests to confirm current behavior
3. Making small, incremental improvements
4. Running tests after each change
5. Logging your progress

Remember: Small steps, run tests after each change, preserve behavior.
