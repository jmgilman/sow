---
description: Implement a new feature using Test-Driven Development
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
argument-hint: "<feature-description>"
model: inherit
---

Implement the following feature using Test-Driven Development: $ARGUMENTS

## TDD Workflow

Follow this process strictly. Do not skip steps.

### Step 1: Understand Requirements

Read task files in this order:

1. **`state.yaml`** - Check iteration number, references, feedback status
2. **`description.md`** - Read complete requirements and acceptance criteria
3. **`feedback/*.md`** - Review corrections if iteration > 1
4. **Referenced files** - Read architecture docs, sinks, code examples listed in `state.yaml`

### Step 2: Plan Tests

Before writing any code, identify what behaviors need testing:

- List each requirement from `description.md`
- Identify edge cases
- Determine what inputs/outputs to test
- Note external dependencies that need mocking

Do not plan implementation. Plan tests only.

### Step 3: Write First Test (RED)

Write a single test for the first behavior:

```python
def test_premium_users_receive_10_percent_discount():
    """Premium tier users should receive 10% off their order total."""
    # Arrange
    mock_repo = Mock(UserRepository)
    mock_repo.find_by_id.return_value = User(id="u123", tier="premium")
    pricing = PricingService(user_repository=mock_repo)

    # Act
    result = pricing.calculate_order_total(user_id="u123", items=[
        OrderItem(price=100, quantity=1)
    ])

    # Assert
    assert result.total == 90  # 100 - 10% = 90
    assert "discount" in result.summary.lower()  # Loose coupling
```

**Test characteristics:**
- Tests one behavior
- Mocks external dependencies (repository port)
- Uses real domain objects (User, OrderItem)
- Asserts on outcomes, not internal calls
- Uses loose string matching for flexibility

### Step 4: Run Test (RED)

Execute the test:

```bash
pytest path/to/test_file.py::test_premium_users_receive_10_percent_discount -v
```

**The test MUST fail.** If it passes, the test is wrong. Delete and rewrite it.

Log the failure:
```markdown
## [2025-01-15 10:23] RED: Premium discount test
- Wrote test: `test_premium_users_receive_10_percent_discount`
- Test failed as expected: `AttributeError: 'PricingService' has no method 'calculate_order_total'`
```

### Step 5: Implement (GREEN)

Write the minimal code to make the test pass. Do not over-engineer.

```python
class PricingService:
    def __init__(self, user_repository: UserRepository):
        self._user_repo = user_repository

    def calculate_order_total(self, user_id: str, items: list[OrderItem]) -> OrderTotal:
        user = self._user_repo.find_by_id(user_id)
        subtotal = sum(item.price * item.quantity for item in items)

        discount = 0
        if user.tier == "premium":
            discount = subtotal * 0.10

        total = subtotal - discount
        summary = f"Subtotal: {subtotal}, Discount: {discount}, Total: {total}"

        return OrderTotal(total=total, summary=summary)
```

Run the test again. It must pass.

Log the implementation:
```markdown
## [2025-01-15 10:35] GREEN: Premium discount implementation
- Implemented `PricingService.calculate_order_total()`
- Test passes
- Coverage: 85% (need more tests for edge cases)
```

### Step 6: Refactor

Improve code quality without changing behavior:

```python
class PricingService:
    PREMIUM_DISCOUNT_RATE = 0.10

    def __init__(self, user_repository: UserRepository):
        self._user_repo = user_repository

    def calculate_order_total(self, user_id: str, items: list[OrderItem]) -> OrderTotal:
        user = self._user_repo.find_by_id(user_id)
        subtotal = self._calculate_subtotal(items)
        discount = self._calculate_discount(user, subtotal)
        total = subtotal - discount

        return OrderTotal(
            total=total,
            summary=self._build_summary(subtotal, discount, total)
        )

    def _calculate_subtotal(self, items: list[OrderItem]) -> float:
        return sum(item.price * item.quantity for item in items)

    def _calculate_discount(self, user: User, subtotal: float) -> float:
        if user.tier == "premium":
            return subtotal * self.PREMIUM_DISCOUNT_RATE
        return 0

    def _build_summary(self, subtotal: float, discount: float, total: float) -> str:
        return f"Subtotal: {subtotal}, Discount: {discount}, Total: {total}"
```

Run tests after each refactoring. If tests fail, you changed behavior—revert immediately.

Log refactoring:
```markdown
## [2025-01-15 10:42] REFACTOR: Extract methods
- Extracted `_calculate_subtotal()`, `_calculate_discount()`, `_build_summary()`
- Moved discount rate to constant
- All tests still pass
```

### Step 7: Repeat for Next Behavior

Write the next test and repeat RED-GREEN-REFACTOR:

```python
def test_standard_users_receive_no_discount():
    """Standard tier users should not receive any discount."""
    mock_repo = Mock(UserRepository)
    mock_repo.find_by_id.return_value = User(id="u456", tier="standard")
    pricing = PricingService(user_repository=mock_repo)

    result = pricing.calculate_order_total(user_id="u456", items=[
        OrderItem(price=100, quantity=1)
    ])

    assert result.total == 100  # No discount
    assert "discount: 0" in result.summary.lower() or "no discount" in result.summary.lower()
```

Continue until all requirements are implemented.

## Testing Guidelines

### Mock External Dependencies

Unit tests never touch real external systems. Mock the ports (interfaces):

```python
# ✅ CORRECT: Mock the port interface
mock_repo = Mock(UserRepository)
mock_repo.find_by_id.return_value = User(...)
service = PricingService(mock_repo)

# ❌ WRONG: Use real adapter
real_db = PostgresUserRepository(connection_string="...")
service = PricingService(real_db)  # This is an integration test, not a unit test
```

**If you cannot mock a dependency, stop immediately and raise an issue. The design is wrong.**

### Test Behavior, Not Implementation

```python
# ❌ BAD: Tests internal method calls (implementation)
def test_calculate_order_calls_calculate_discount():
    service.calculate_order_total(user_id="u123", items=[...])
    assert service._calculate_discount.called  # Testing internal structure

# ✅ GOOD: Tests outcome (behavior)
def test_premium_users_get_discount():
    result = service.calculate_order_total(user_id="u123", items=[...])
    assert result.total < 100  # Testing behavior
```

### Use Loose Assertions

Avoid brittle tests by using partial matches:

```python
# ❌ FRAGILE: Exact match
assert result.summary == "Subtotal: 100.0, Discount: 10.0, Total: 90.0"

# ✅ ROBUST: Partial match
assert "discount" in result.summary.lower()
assert "90" in result.summary or result.total == 90
```

### Test One Behavior Per Test

```python
# ❌ BAD: Tests multiple behaviors
def test_pricing():
    # Test premium discount
    result1 = service.calculate_order_total(user_id="premium", items=[...])
    assert result1.total == 90

    # Test standard pricing
    result2 = service.calculate_order_total(user_id="standard", items=[...])
    assert result2.total == 100

    # Test shipping
    result3 = service.calculate_order_total(user_id="premium", items=[...], shipping=True)
    assert result3.total == 95

# ✅ GOOD: One behavior per test
def test_premium_discount():
    result = service.calculate_order_total(user_id="premium", items=[...])
    assert result.total == 90

def test_standard_no_discount():
    result = service.calculate_order_total(user_id="standard", items=[...])
    assert result.total == 100

def test_shipping_adds_cost():
    result = service.calculate_order_total(user_id="premium", items=[...], shipping=True)
    assert result.total == 95
```

## Edge Cases to Consider

Always test edge cases:

- Empty inputs (empty lists, null values)
- Boundary values (0, negative numbers, very large numbers)
- Error conditions (user not found, invalid data)
- Concurrent edge cases (if applicable)

Example:
```python
def test_empty_order_returns_zero_total():
    result = service.calculate_order_total(user_id="u123", items=[])
    assert result.total == 0

def test_invalid_user_raises_error():
    mock_repo.find_by_id.side_effect = UserNotFoundError("u999")

    with pytest.raises(UserNotFoundError):
        service.calculate_order_total(user_id="u999", items=[...])
```

## Logging Actions

Append all actions to `log.md` in this format:

```markdown
## [YYYY-MM-DD HH:MM] RED: Test name
- Wrote test: `test_function_name`
- Test failed as expected: <error message>

## [YYYY-MM-DD HH:MM] GREEN: Implementation description
- Implemented `ClassName.method_name()`
- Test passes
- Coverage: X%

## [YYYY-MM-DD HH:MM] REFACTOR: Description of changes
- Refactored: <what changed>
- All tests still pass

## [YYYY-MM-DD HH:MM] RED: Next test name
...
```

## Completion Criteria

Feature is complete when:

1. All requirements in `description.md` have tests
2. All tests pass
3. Code coverage >90% (or all behaviors tested)
4. No external dependencies in unit tests
5. Code is refactored and clean
6. All actions logged to `log.md`

## When to Stop and Raise Issues

Stop immediately if:

- You cannot isolate a test (needs real database/API/network)
- The design doesn't provide interfaces to mock
- Requirements are ambiguous or contradictory
- Implementation requires design changes
- You discover architectural problems

Do not attempt to fix design issues. Raise them to the orchestrator.