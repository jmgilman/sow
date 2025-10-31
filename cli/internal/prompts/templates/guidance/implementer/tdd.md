# Test-Driven Development (TDD) Methodology

This guidance is **mandatory** for all implementer work. TDD is non-negotiable.

## Core TDD Principles

### 1. Test First, Always

Write tests before writing implementation code. Tests define the behavior you must implement. No exceptions.

**The Red-Green-Refactor Cycle:**

1. **Red**: Write a failing test
2. **Green**: Write minimal code to make it pass
3. **Refactor**: Improve code quality without changing behavior
4. Repeat

### 2. Test Behavior, Not Implementation

Tests verify **what code does**, not **how it does it**.

Test outcomes and public APIs. Internal implementation details are irrelevant to tests.

```python
# ❌ BAD: Tests internal calls (implementation detail)
def test_calls_internal_method():
    service.process()
    assert service._internal_method.called

# ✅ GOOD: Tests outcome (observable behavior)
def test_process_returns_correct_result():
    result = service.process()
    assert result.value == expected_value
```

**Why this matters:**
- Implementation tests break when you refactor (even if behavior is unchanged)
- Behavior tests survive refactoring
- You want freedom to improve code without breaking tests

### 3. Unit Tests Stay Isolated

Unit tests never touch external systems—no databases, APIs, file systems, or network calls.

**Mock all outbound dependencies** using interfaces/ports.

**If you cannot isolate a test, stop and raise the issue.** This indicates a design problem.

## Mocking External Dependencies

Hexagonal architecture provides **ports** (interfaces) for external systems. Mock the ports, never concrete adapters.

```python
# ✅ CORRECT: Mock the port interface
mock_repo = Mock(UserRepository)  # Repository is an interface/port
mock_repo.find_by_id.return_value = User(id="123", name="Alice")
service = UserService(mock_repo)
result = service.get_user("123")
assert result.name == "Alice"

# ❌ WRONG: Use real adapter
real_db = PostgresUserRepository(connection_string)  # External dependency!
service = UserService(real_db)
result = service.get_user("123")  # This hits the real database
```

**If you cannot mock a dependency, the design is wrong.** Stop immediately and raise the issue to the orchestrator.

### Why Mock?

- **Speed**: Unit tests run in milliseconds
- **Reliability**: No flaky tests due to network/database issues
- **Isolation**: Test one unit at a time
- **Independence**: Tests don't depend on external system state

## Writing Effective Tests

### Test One Behavior Per Test

Each test should verify a single behavior. Multiple behaviors require multiple tests.

```python
# ❌ BAD: Tests multiple behaviors
def test_user_service():
    user = service.create_user("Alice")
    assert user.name == "Alice"
    updated = service.update_user(user.id, "Bob")
    assert updated.name == "Bob"
    deleted = service.delete_user(user.id)
    assert deleted is True

# ✅ GOOD: One behavior per test
def test_create_user_returns_user_with_correct_name():
    user = service.create_user("Alice")
    assert user.name == "Alice"

def test_update_user_changes_name():
    user = User(id="123", name="Alice")
    updated = service.update_user(user.id, "Bob")
    assert updated.name == "Bob"

def test_delete_user_returns_true_on_success():
    deleted = service.delete_user("123")
    assert deleted is True
```

### Use Loose Assertions

Avoid brittle tests by using partial matches instead of exact matches.

```python
# ❌ FRAGILE: Exact string match
assert message == "User premium-123 received 10% discount on item ABC"
# Breaks if: wording changes, user ID format changes, item name changes

# ✅ ROBUST: Partial matches
assert "premium" in message.lower()
assert "discount" in message
# Survives: wording changes, formatting changes
```

```javascript
// ❌ FRAGILE: Exact structure match
expect(result).toEqual({
  id: 123,
  name: "Alice",
  email: "alice@example.com",
  created_at: "2025-01-15T10:30:00Z",
  metadata: { ... }
});

// ✅ ROBUST: Match what matters
expect(result).toMatchObject({
  id: 123,
  name: "Alice"
});
expect(result.email).toContain("@example.com");
```

### Avoid Testing Private Methods Directly

Private methods are implementation details. Test them indirectly through public APIs.

```python
# ❌ BAD: Testing private method directly
def test_calculate_discount_internal():
    result = service._calculate_discount(100, "PREMIUM")
    assert result == 10

# ✅ GOOD: Test private method through public API
def test_premium_users_receive_discount():
    order = service.process_order(user_type="PREMIUM", amount=100)
    assert order.total == 90  # 10% discount applied
```

**Why:** If you refactor and rename/remove `_calculate_discount`, the public behavior test still passes. The private method test breaks.

### Test Edge Cases and Boundaries

Don't just test the happy path. Test:

- **Null/empty inputs**: What happens with `None`, `""`, `[]`?
- **Boundary values**: Minimum, maximum, just inside/outside limits
- **Error conditions**: Invalid input, missing data, permissions denied
- **Edge cases**: Zero, negative numbers, special characters

```python
# Don't just test:
def test_divide_returns_correct_result():
    assert divide(10, 2) == 5

# Also test:
def test_divide_by_zero_raises_error():
    with pytest.raises(ZeroDivisionError):
        divide(10, 0)

def test_divide_with_negative_numbers():
    assert divide(-10, 2) == -5
    assert divide(10, -2) == -5
```

## Test Coverage Goals

### Aim for >90% Behavioral Coverage

Prioritize **behavioral coverage** over **code coverage**.

**Behavioral coverage**: Every requirement in description.md has a test. Every edge case has a test.

**Code coverage**: Percentage of lines executed by tests (this is secondary).

### What to Test

**Always test:**
- Every requirement from description.md
- All error handling paths
- Boundary conditions
- Edge cases

**Don't obsess over:**
- Trivial getters/setters (if they literally just return a field)
- Auto-generated code
- Framework code you don't own

## Testing Anti-Patterns to Avoid

**❌ Testing private methods directly**
- Test behavior through public API instead

**❌ Asserting on internal state instead of behavior**
- Check what the system does, not how it stores data internally

**❌ Using exact string/structure matches**
- Use partial matches for robustness

**❌ Writing tests after implementation**
- Test first, always

**❌ Skipping tests for "simple" code**
- Simple code can still have bugs; test it

**❌ Mocking concrete classes instead of interfaces**
- Mock ports/interfaces, not adapters

**❌ Tests that depend on execution order**
- Each test should be independent

**❌ Tests that depend on external system state**
- Mock all external dependencies

**❌ One giant test that tests everything**
- One test per behavior

## Integration with Hexagonal Architecture

Your codebase likely follows **hexagonal architecture** (ports and adapters pattern).

**Ports**: Interfaces defining contracts with external systems (database, API, file system)
**Adapters**: Concrete implementations of ports

**In unit tests, mock the ports:**

```typescript
// Port (interface)
interface UserRepository {
  findById(id: string): Promise<User | null>;
  save(user: User): Promise<void>;
}

// Service depends on port
class UserService {
  constructor(private repo: UserRepository) {}

  async getUser(id: string): Promise<User> {
    const user = await this.repo.findById(id);
    if (!user) throw new Error("User not found");
    return user;
  }
}

// Test mocks the port
test("getUser returns user when found", async () => {
  const mockRepo: UserRepository = {
    findById: jest.fn().mockResolvedValue({ id: "123", name: "Alice" }),
    save: jest.fn()
  };

  const service = new UserService(mockRepo);
  const user = await service.getUser("123");

  expect(user.name).toBe("Alice");
});
```

**Never mock concrete adapters in unit tests:**
```typescript
// ❌ BAD: Using real adapter in unit test
const realAdapter = new PostgresUserRepository(config);
const service = new UserService(realAdapter);
// This is now an integration test, not a unit test
```

## When Testing Reveals Design Issues

If you cannot write isolated unit tests, this reveals a design problem:

**Problem**: Cannot mock a dependency
**Cause**: No interface/port exists for the external system
**Solution**: Stop and escalate to orchestrator (this is an architectural fix)

**Problem**: Test setup is extremely complicated
**Cause**: Too many dependencies, tight coupling
**Solution**: Stop and escalate (design needs simplification)

**Problem**: Cannot test without real database/API
**Cause**: Business logic is embedded in adapter layer
**Solution**: Stop and escalate (logic should be in domain/service layer)

Do not attempt to fix these architectural problems yourself. Raise the issue and wait for guidance.

## TDD Workflow Summary

For every feature/bug/refactor:

1. **Understand the requirement** - What behavior needs to exist?
2. **Write a failing test** - Describe the behavior in code
3. **Run the test** - Confirm it fails (red)
4. **Write minimal implementation** - Just enough to pass
5. **Run the test** - Confirm it passes (green)
6. **Refactor** - Improve code quality
7. **Run the test** - Confirm it still passes
8. **Repeat** - Until all requirements are met

Test first, always.
