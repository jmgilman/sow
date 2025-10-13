---
description: Fix a bug using Test-Driven Development
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
argument-hint: "<bug-description>"
model: inherit
---

Fix the following bug using Test-Driven Development: $ARGUMENTS

## Bug-Fixing Workflow

Follow this process strictly. Test first, then fix.

### Step 1: Understand the Bug

Read task files in this order:

1. **`state.yaml`** - Check iteration number, references, feedback status
2. **`description.md`** - Read bug description, steps to reproduce, expected vs actual behavior
3. **`feedback/*.md`** - Review corrections if iteration > 1
4. **Referenced files** - Read related code, architecture docs, existing tests

Understand:
- What is the incorrect behavior?
- What should the correct behavior be?
- Under what conditions does the bug occur?
- Why wasn't this caught by existing tests?

### Step 2: Reproduce the Bug

Before writing any tests or fixes, reproduce the bug:

```bash
# Run the code that exhibits the bug
python -m app.service.process_order --user premium-123 --items item-1

# Or run existing tests that should fail (but don't)
pytest tests/test_pricing.py -v
```

Confirm:
- You can consistently reproduce the bug
- You understand the exact conditions that trigger it
- You know what the correct behavior should be

Log reproduction:
```markdown
## [2025-01-15 14:30] Reproduced bug
- Bug: Premium users not receiving discount
- Steps to reproduce:
  1. Create order with premium user
  2. Calculate total
  3. Expected: $90, Actual: $100
- Root cause hypothesis: Discount calculation not checking user tier
```

### Step 3: Write Failing Test (RED)

Write a test that demonstrates the bug. The test should fail with the current code:

```python
def test_premium_users_should_receive_discount():
    """Bug: Premium users are not receiving their 10% discount.

    Expected: Premium user with $100 order should pay $90
    Actual: Premium user pays full $100
    """
    # Arrange
    mock_repo = Mock(UserRepository)
    mock_repo.find_by_id.return_value = User(id="u123", tier="premium")
    pricing = PricingService(user_repository=mock_repo)

    # Act
    result = pricing.calculate_order_total(user_id="u123", items=[
        OrderItem(price=100, quantity=1)
    ])

    # Assert
    assert result.total == 90, f"Expected 90, got {result.total}"
    assert "discount" in result.summary.lower()
```

### Step 4: Run Test (RED)

Execute the test. It must fail, demonstrating the bug:

```bash
pytest tests/test_pricing.py::test_premium_users_should_receive_discount -v
```

Expected output:
```
FAILED - AssertionError: Expected 90, got 100
```

If the test passes, you haven't reproduced the bug correctly. Revise the test.

Log the failing test:
```markdown
## [2025-01-15 14:45] RED: Test demonstrates bug
- Wrote test: `test_premium_users_should_receive_discount`
- Test fails as expected: AssertionError: Expected 90, got 100
- Confirms bug in production code
```

### Step 5: Fix the Bug (GREEN)

Make the minimal change necessary to fix the bug. Avoid refactoring at this stage.

```python
# Before (buggy code):
def calculate_order_total(self, user_id: str, items: list[OrderItem]) -> OrderTotal:
    subtotal = sum(item.price * item.quantity for item in items)
    # BUG: Not applying discount for premium users
    return OrderTotal(total=subtotal, summary=f"Total: {subtotal}")

# After (fixed code):
def calculate_order_total(self, user_id: str, items: list[OrderItem]) -> OrderTotal:
    user = self._user_repo.find_by_id(user_id)
    subtotal = sum(item.price * item.quantity for item in items)

    # FIX: Apply discount for premium users
    discount = 0
    if user.tier == "premium":
        discount = subtotal * 0.10

    total = subtotal - discount
    summary = f"Subtotal: {subtotal}, Discount: {discount}, Total: {total}"
    return OrderTotal(total=total, summary=summary)
```

Run the failing test again. It must now pass:

```bash
pytest tests/test_pricing.py::test_premium_users_should_receive_discount -v
```

Log the fix:
```markdown
## [2025-01-15 15:00] GREEN: Bug fixed
- Modified `calculate_order_total()` to apply premium discount
- Test now passes
- Fix location: `app/services/pricing.py` line 23-28
```

### Step 6: Test for Regressions

Run the full test suite to ensure the fix didn't break anything:

```bash
pytest tests/ -v
```

All existing tests must still pass. If any tests fail:
1. Determine if the failing test was incorrect (testing wrong behavior)
2. Or if your fix introduced a regression

Log regression check:
```markdown
## [2025-01-15 15:05] Regression check
- Ran full test suite: 47 tests
- All tests pass
- No regressions detected
```

### Step 7: Add Edge Case Tests

Now that the bug is fixed, add tests for edge cases to prevent similar bugs:

```python
def test_standard_users_still_pay_full_price():
    """Ensure standard users are not affected by premium discount fix."""
    mock_repo = Mock(UserRepository)
    mock_repo.find_by_id.return_value = User(id="u456", tier="standard")
    pricing = PricingService(user_repository=mock_repo)

    result = pricing.calculate_order_total(user_id="u456", items=[
        OrderItem(price=100, quantity=1)
    ])

    assert result.total == 100  # No discount for standard users

def test_premium_discount_applies_to_multiple_items():
    """Ensure discount applies correctly to multi-item orders."""
    mock_repo = Mock(UserRepository)
    mock_repo.find_by_id.return_value = User(id="u123", tier="premium")
    pricing = PricingService(user_repository=mock_repo)

    result = pricing.calculate_order_total(user_id="u123", items=[
        OrderItem(price=50, quantity=2),   # $100 subtotal
        OrderItem(price=25, quantity=2)    # $50 subtotal
        # Total subtotal: $150, Discount: $15, Final: $135
    ])

    assert result.total == 135
```

Log edge case tests:
```markdown
## [2025-01-15 15:15] Added edge case tests
- Added test: `test_standard_users_still_pay_full_price`
- Added test: `test_premium_discount_applies_to_multiple_items`
- Both tests pass
- Coverage increased to 94%
```

### Step 8: Refactor (If Needed)

Now that the bug is fixed and tests are in place, refactor if the code is messy:

```python
class PricingService:
    PREMIUM_DISCOUNT_RATE = 0.10

    def calculate_order_total(self, user_id: str, items: list[OrderItem]) -> OrderTotal:
        user = self._user_repo.find_by_id(user_id)
        subtotal = self._calculate_subtotal(items)
        discount = self._calculate_discount(user, subtotal)
        total = subtotal - discount

        return OrderTotal(
            total=total,
            summary=self._build_summary(subtotal, discount, total)
        )

    def _calculate_discount(self, user: User, subtotal: float) -> float:
        if user.tier == "premium":
            return subtotal * self.PREMIUM_DISCOUNT_RATE
        return 0
```

Run all tests after refactoring. If any fail, revert and try again.

Log refactoring:
```markdown
## [2025-01-15 15:25] REFACTOR: Extracted discount logic
- Moved discount rate to constant
- Extracted `_calculate_discount()` method
- All tests still pass
```

## Root Cause Analysis

After fixing the bug, document why it occurred:

```markdown
## [2025-01-15 15:30] Root cause analysis
**Why did this bug occur?**
- Original implementation didn't load user data before calculating total
- Discount logic was missing entirely

**Why wasn't it caught by tests?**
- No existing test verified premium user discount behavior
- Tests only covered the happy path for standard users

**Prevention:**
- Added test for premium discount (now in test suite)
- Added tests for edge cases (multi-item orders, standard users)
- Recommendation: Review test coverage for all user tier features
```

Include root cause analysis in your log.

## Guidelines

### Minimal Fixes

Fix only what is broken. Resist the urge to refactor unrelated code:

```python
# ✅ GOOD: Targeted fix
if user.tier == "premium":
    discount = subtotal * 0.10

# ❌ BAD: Over-engineering while fixing
discount_calculator = DiscountCalculatorFactory.create(user.tier)
discount = discount_calculator.calculate(subtotal, user.membership_level,
    user.purchase_history, current_promotions)
```

### Test the Bug, Not the Fix

Your test should verify correct behavior, not implementation details:

```python
# ✅ GOOD: Tests behavior
def test_premium_users_receive_discount():
    result = service.calculate_order_total(user_id="premium", items=[...])
    assert result.total == 90

# ❌ BAD: Tests implementation
def test_premium_users_receive_discount():
    service.calculate_order_total(user_id="premium", items=[...])
    assert service._calculate_discount.called
```

### Consider Why Tests Missed the Bug

If existing tests didn't catch this bug, ask why:
- Was test coverage insufficient?
- Did tests check the wrong behavior?
- Were edge cases not tested?

Add tests to prevent recurrence.

## Logging Format

Use this format for all log entries:

```markdown
## [YYYY-MM-DD HH:MM] Reproduced bug
- Bug description
- Steps to reproduce
- Expected vs actual behavior

## [YYYY-MM-DD HH:MM] RED: Test demonstrates bug
- Wrote test: `test_name`
- Test fails as expected: <error>

## [YYYY-MM-DD HH:MM] GREEN: Bug fixed
- Modified file/function
- Test now passes
- Fix location

## [YYYY-MM-DD HH:MM] Regression check
- Ran full test suite
- X tests pass
- No regressions / Y regressions found (fixed)

## [YYYY-MM-DD HH:MM] Added edge case tests
- List of new tests
- All pass

## [YYYY-MM-DD HH:MM] REFACTOR: Description (optional)
- Refactoring details
- All tests still pass

## [YYYY-MM-DD HH:MM] Root cause analysis
- Why bug occurred
- Why tests missed it
- Prevention measures
```

## Completion Criteria

Bug fix is complete when:

1. Bug is consistently reproducible
2. Test demonstrates the bug (fails before fix)
3. Fix is implemented (test passes after fix)
4. Full test suite passes (no regressions)
5. Edge case tests added
6. Code is refactored (if needed)
7. Root cause analysis completed
8. All actions logged to `log.md`

## When to Stop and Raise Issues

Stop immediately if:

- Bug cannot be reproduced consistently
- Bug is caused by architectural design flaw (not implementation)
- Fix requires changes to requirements or design
- Multiple components need modification (might be a design issue)
- You cannot write an isolated unit test for the bug

For architectural or design issues, raise to the orchestrator. Do not attempt to redesign while bug fixing.