# Go Testing Best Practices Guide

## Golden Rules

These rules are non-negotiable and must be followed at all times.

1. **Don't test implementation details, test behavior**
   - Focus on testing the public API and observable behavior
   - Avoid testing private functions directly
   - Don't rely on specific internal state
   - Example:
     ```go
     // BAD: Testing internal state
     if cache.internalMap["key"] != expected {
         t.Error("internal map mismatch")
     }

     // GOOD: Testing behavior
     got := cache.Get("key")
     if got != expected {
         t.Errorf("Get(key) = %v, want %v", got, expected)
     }
     ```

2. **Unit tests must NEVER have external dependencies**
   - No network calls (HTTP, gRPC, database connections, etc.)
   - No real filesystem access (except temporary filesystem when required)
   - Use mocks or in-memory implementations for external dependencies
   - Tests with external dependencies are integration tests, not unit tests
   - Exception: `os.CreateTemp()`, `t.TempDir()` and similar temporary filesystem operations are acceptable
   - Example:
     ```go
     // BAD: External dependency in unit test
     func TestFetchUser(t *testing.T) {
         resp, err := http.Get("https://api.example.com/user/123")
         // ...
     }

     // GOOD: Mock the external dependency
     func TestFetchUser(t *testing.T) {
         client := &mockHTTPClient{
             getFunc: func(url string) (*Response, error) {
                 return &Response{Status: 200, Body: `{"id": 123}`}, nil
             },
         }
         svc := NewService(client)
         // ...
     }

     // ACCEPTABLE: Temporary filesystem
     func TestWriteConfig(t *testing.T) {
         dir := t.TempDir()
         configPath := filepath.Join(dir, "config.json")
         // test logic with temporary directory
     }
     ```

3. **Don't test constants or functions with no logic**
   - Don't write tests for constant values
   - Don't test trivial constructors that only assign parameters to struct fields
   - Don't test functions with no logic (simple getters, setters with no validation)
   - Focus testing efforts on code with actual logic, branching, or transformations
   - Example:
     ```go
     // BAD: Testing constants
     func TestMaxRetries(t *testing.T) {
         if MaxRetries != 3 {
             t.Errorf("MaxRetries = %d, want 3", MaxRetries)
         }
     }

     // BAD: Testing trivial constructor
     func TestNewUser(t *testing.T) {
         user := NewUser("John", 30)
         if user.Name != "John" {
             t.Errorf("Name = %q, want %q", user.Name, "John")
         }
         if user.Age != 30 {
             t.Errorf("Age = %d, want %d", user.Age, 30)
         }
     }

     // Where NewUser is just:
     // func NewUser(name string, age int) User {
     //     return User{Name: name, Age: age}
     // }

     // GOOD: Test functions with actual logic
     func TestValidateUser(t *testing.T) {
         user := User{Name: "", Age: -1}
         err := user.Validate()
         if err == nil {
             t.Error("expected validation error for invalid user")
         }
     }
     ```

4. **Strive for 80% behavioral coverage, not line coverage**
   - Focus on testing 80% of behaviors and use cases, not 80% of lines executed
   - Line coverage metrics can be misleading - executing a line doesn't mean the behavior is tested
   - A single test can execute many lines but only test one behavior
   - Multiple tests may be needed to cover all behaviors in a single function
   - Example:
     ```go
     // This function has multiple behaviors to test:
     func ProcessOrder(order *Order) error {
         if order == nil {
             return ErrNilOrder
         }
         if order.Total < 0 {
             return ErrNegativeTotal
         }
         if order.Items == nil || len(order.Items) == 0 {
             return ErrNoItems
         }
         // process order
         return nil
     }

     // GOOD: Test each behavior separately
     // - Behavior: nil order returns error
     // - Behavior: negative total returns error
     // - Behavior: empty items returns error
     // - Behavior: valid order processes successfully
     // This is 4 behaviors, and we should test all 4 (100% behavioral coverage)

     // BAD: One test that executes all lines but only tests success case
     // This would give high line coverage but poor behavioral coverage
     func TestProcessOrder(t *testing.T) {
         order := &Order{Total: 100, Items: []Item{{ID: 1}}}
         err := ProcessOrder(order)
         if err != nil {
             t.Errorf("unexpected error: %v", err)
         }
     }
     ```

## Test Organization

1. **Test files must be named with `_test.go` suffix, matching the source file name**
   - Pattern: For `foo.go`, create `foo_test.go`
   - Place test files in the same directory as the code being tested
   - When splitting tests across multiple files, maintain the prefix: `foo_test.go`, `foo_integration_test.go`, `foo_benchmark_test.go`
   - Examples: `handler.go` → `handler_test.go`; `user.go` → `user_test.go`, `user_validation_test.go`

2. **Always use the same package name as the code being tested**
   - Use `package mypackage`, not `package mypackage_test`
   - This allows tests to access both exported and unexported functions/types
   - Example: If testing `package handler`, use `package handler` in test files

3. **Name test functions with `Test` prefix followed by the function/feature being tested**
   - Pattern: `func TestFunctionName(t *testing.T)`
   - For package-level functions: `func TestValidateEmail(t *testing.T)`
   - For methods: `func TestUser_Login(t *testing.T)` or `func TestUser_Validate(t *testing.T)`
     - Use the pattern `TestTypeName_MethodName`
   - For specific scenarios: `func TestUser_Login_InvalidPassword(t *testing.T)`

4. **Place mocks appropriately based on reusability**
   - Single-use mocks: Keep inline within the test file where they're used
   - Large single-use mocks: Extract to `mypkg_mocks_test.go` in the same directory
   - Reusable mocks: Place in a `mocks` subpackage within the current package
     - Directory structure: `mypkg/mocks/`
     - File naming: Match the source file being mocked (e.g., `mypkg/client.go` → `mypkg/mocks/client.go`)
     - Example: `handler/mocks/storage.go` for mocks of `handler/storage.go`

5. **Place shared test helpers in a `testutils` package at the module root**
   - Single-use test helpers: Keep inline within the test file or in `mypkg_test_helpers.go`
   - Shared test helpers used across multiple packages: Place in `testutils` package at module root
     - Directory structure: `<module-root>/testutils/`
     - Example: `mymodule/testutils/assertions.go`, `mymodule/testutils/fixtures.go`
   - Mark helper functions with `t.Helper()` to ensure correct line numbers in error reports

6. **Group related tests in the same file**
   - All tests for `User.Validate()` should be in the same test file
   - Split into multiple test files only when a single file becomes unwieldy (>500-1000 lines)

## Test Structure

1. **Use table-driven tests for multiple test cases**
   - Define test cases as a slice of structs
   - Each test case should have a `name` field for identification
   - Example structure:
     ```go
     func TestAdd(t *testing.T) {
         tests := []struct {
             name string
             a, b int
             want int
         }{
             {name: "positive numbers", a: 2, b: 3, want: 5},
             {name: "negative numbers", a: -1, b: -1, want: -2},
             {name: "zero", a: 0, b: 0, want: 0},
         }
         for _, tt := range tests {
             t.Run(tt.name, func(t *testing.T) {
                 got := Add(tt.a, tt.b)
                 if got != tt.want {
                     t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
                 }
             })
         }
     }
     ```

2. **Always use subtests with `t.Run()` for table-driven tests**
   - Allows running individual test cases: `go test -run TestAdd/positive`
   - Provides clear test output showing which case failed
   - Enables parallel execution of test cases when appropriate

3. **For complex table-driven tests, extract validation and setup logic**
   - Add a `validate` function field to the test struct for complex validation instead of inline assertions
   - If `validate` takes more than 3 arguments, wrap return data in an inline type
   - Use a `setup()` function for conditional setup instead of adding many fields to the test struct
     - Exception: Simple tests with uniform setup or only a few toggle fields can skip `setup()`
   - Example:
     ```go
     func TestComplexOperation(t *testing.T) {
         type result struct {
             data   []byte
             status int
             meta   map[string]string
         }

         tests := []struct {
             name      string
             input     string
             setupOpts []SetupOption // instead of: useCache bool, enableRetry bool, timeout int, etc.
             want      result
             validate  func(t *testing.T, got result)
         }{
             {
                 name:      "case1",
                 input:     "...",
                 setupOpts: []SetupOption{WithCache(), WithRetry()},
                 want:      result{...},
                 validate: func(t *testing.T, got result) {
                     if got.status != tt.want.status {
                         t.Errorf("status = %d, want %d", got.status, tt.want.status)
                     }
                     // additional complex validation logic
                 },
             },
         }

         for _, tt := range tests {
             t.Run(tt.name, func(t *testing.T) {
                 svc := setup(t, tt.setupOpts...)
                 got := svc.Execute(tt.input)
                 tt.validate(t, got)
             })
         }
     }
     ```

4. **Use `t.Run()` to separate multiple test cases, even in non-table-driven tests**
   - When a test function covers multiple scenarios, use subtests for each
   - Single test case scenarios can omit `t.Run()`
   - Example:
     ```go
     func TestUserService(t *testing.T) {
         t.Run("create user", func(t *testing.T) {
             // test user creation
         })

         t.Run("update user", func(t *testing.T) {
             // test user update
         })

         t.Run("delete user", func(t *testing.T) {
             // test user deletion
         })
     }
     ```

4. **Follow the Arrange-Act-Assert pattern**
   - Arrange: Set up test data and preconditions
   - Act: Execute the code being tested
   - Assert: Verify the results
   - Use blank lines to visually separate these sections
   - Example:
     ```go
     func TestUserValidate(t *testing.T) {
         user := User{Name: "John", Email: "invalid"}
         err := user.Validate()
         if err == nil {
             t.Error("expected validation error, got nil")
         }
     }
     ```

4. **Keep test cases independent and isolated**
   - Each test should set up its own data
   - Avoid shared mutable state between tests
   - Clean up resources using `t.Cleanup()` or defer
   - Tests should pass when run individually or as part of the suite

## Test Helpers & Utilities

1. **Use `testify/assert` and `testify/require` for all assertions**
   - Import `github.com/stretchr/testify/assert` for assertions that allow the test to continue
   - Import `github.com/stretchr/testify/require` for assertions where the test must stop on failure
   - `assert` maps to `t.Error()` behavior (test continues)
   - `require` maps to `t.Fatal()` behavior (test stops immediately)
   - Example:
     ```go
     import (
         "testing"
         "github.com/stretchr/testify/assert"
         "github.com/stretchr/testify/require"
     )

     func TestUserService(t *testing.T) {
         // Use require for setup/preconditions that must succeed
         db, err := setupTestDB()
         require.NoError(t, err, "failed to setup test database")
         require.NotNil(t, db)

         // Use assert for multiple assertions that should all be checked
         user := GetUser(db, "123")
         assert.Equal(t, "John", user.Name)
         assert.Equal(t, 30, user.Age)
         assert.True(t, user.IsActive)
     }
     ```

2. **Always use `t.Helper()` in test helper functions**
   - Mark helper functions with `t.Helper()` as the first line
   - Ensures error messages point to the calling test, not the helper
   - Example:
     ```go
     func assertUser(t *testing.T, got, want User) {
         t.Helper()
         if got.Name != want.Name {
             t.Errorf("Name = %q, want %q", got.Name, want.Name)
         }
     }
     ```

3. **Use `t.Cleanup()` for resource cleanup**
   - Prefer `t.Cleanup()` over defer for test cleanup
   - Cleanup functions run in LIFO order after test completion
   - Works correctly with subtests and parallel tests
   - Example:
     ```go
     func TestWithTempFile(t *testing.T) {
         f, err := os.CreateTemp("", "test")
         if err != nil {
             t.Fatal(err)
         }
         t.Cleanup(func() {
             os.Remove(f.Name())
         })
         // test code
     }
     ```

4. **Create test fixtures in a consistent location**
   - Place test data files in a `testdata` directory
   - Go tools ignore `testdata` directories by convention
   - Structure: `package/testdata/fixture.json`
   - Example:
     ```go
     func loadFixture(t *testing.T, name string) []byte {
         t.Helper()
         data, err := os.ReadFile(filepath.Join("testdata", name))
         if err != nil {
             t.Fatalf("failed to load fixture %s: %v", name, err)
         }
         return data
     }
     ```

5. **Use golden files for complex output validation**
   - Store expected output in `testdata/*.golden` files
   - Compare actual output against golden files
   - Update golden files with `-update` flag pattern
   - Example:
     ```go
     var update = flag.Bool("update", false, "update golden files")

     func TestRender(t *testing.T) {
         got := Render(input)
         golden := filepath.Join("testdata", "output.golden")

         if *update {
             os.WriteFile(golden, got, 0644)
         }

         want, _ := os.ReadFile(golden)
         if !bytes.Equal(got, want) {
             t.Errorf("output mismatch")
         }
     }
     ```

## Error Handling

1. **Use `t.Fatal()` when the test cannot continue after a failure**
   - Use `t.Fatal()` or `t.Fatalf()` for setup failures or preconditions
   - Stops test execution immediately
   - Example: Failed to create test resources, nil pointer checks
   - Example:
     ```go
     func TestProcess(t *testing.T) {
         f, err := os.Open("input.txt")
         if err != nil {
             t.Fatalf("failed to open file: %v", err)
         }
         // test continues only if file opened successfully
     }
     ```

2. **Use `t.Error()` when the test can continue after a failure**
   - Use `t.Error()` or `t.Errorf()` for assertion failures
   - Allows multiple assertions to be checked in a single test run
   - Example:
     ```go
     func TestUser(t *testing.T) {
         user := GetUser()
         if user.Name != "John" {
             t.Errorf("Name = %q, want %q", user.Name, "John")
         }
         if user.Age != 30 {
             t.Errorf("Age = %d, want %d", user.Age, 30)
         }
         // both errors will be reported
     }
     ```

3. **Test error values using `errors.Is()` and error types using `errors.As()`**
   - Use `errors.Is()` to check if an error matches a sentinel error
   - Use `errors.As()` to check if an error is of a specific type
   - Don't compare error strings directly
   - Example:
     ```go
     func TestValidate(t *testing.T) {
         err := Validate(input)

         // Check sentinel error
         if !errors.Is(err, ErrInvalidInput) {
             t.Errorf("expected ErrInvalidInput, got %v", err)
         }

         // Check error type
         var validationErr *ValidationError
         if !errors.As(err, &validationErr) {
             t.Errorf("expected ValidationError, got %T", err)
         }
     }
     ```

4. **When testing for expected errors, verify both presence and type**
   - Don't just check `err != nil`, verify it's the correct error
   - Include error details in failure messages
   - Example:
     ```go
     tests := []struct {
         name    string
         input   string
         wantErr error
     }{
         {name: "invalid email", input: "bad", wantErr: ErrInvalidEmail},
         {name: "valid email", input: "test@example.com", wantErr: nil},
     }

     for _, tt := range tests {
         t.Run(tt.name, func(t *testing.T) {
             err := ValidateEmail(tt.input)
             if !errors.Is(err, tt.wantErr) {
                 t.Errorf("got error %v, want %v", err, tt.wantErr)
             }
         })
     }
     ```

## Mocking & Dependencies

1. **All public interfaces must have mocks generated via `moq`**
   - Generate mocks for all exported interfaces
   - Place generated mocks in the `mocks` subpackage (see Test Organization rule 4)
   - Use `go generate` with a generation comment
   - Prefer `go run` to avoid requiring `moq` to be installed locally
   - Pin `moq` to a recent version
   - Example:
     ```go
     // In storage.go
     package handler

     //go:generate go run github.com/matryer/moq@latest -out mocks/storage.go . Storage

     type Storage interface {
         Get(id string) (*Item, error)
         Save(item *Item) error
     }
     ```

2. **Use `moq` for consumed interfaces, except for simple cases**
   - Generate mocks for external interfaces you depend on
   - Exception: If testing only 1-2 methods, inline mocks are acceptable
   - Follow the same `go generate` pattern as public interfaces
   - Example of acceptable inline mock:
     ```go
     type mockStorage struct {
         getFunc func(id string) (*Item, error)
     }

     func (m *mockStorage) Get(id string) (*Item, error) {
         return m.getFunc(id)
     }
     ```

3. **NEVER mock internal logic or implementation details**
   - Mocks should ONLY be used to control data returned from external dependencies
   - Do NOT check that mocks were called with specific parameters
   - Do NOT count the number of times a mock was called
   - Do NOT verify the order of mock calls
   - Invalid example (DO NOT DO THIS):
     ```go
     // BAD: Testing implementation details
     mock.AssertCalled(t, "Save", specificItem)
     mock.AssertNumberOfCalls(t, "Get", 3)
     ```

4. **Use mocks only to provide test data, not to verify behavior**
   - Valid use: Return test data or errors from external dependencies
   - Invalid use: Asserting on how the mock was used
   - Test behavior through observable outputs, not mock interactions
   - Example:
     ```go
     func TestUserService_Create(t *testing.T) {
         // GOOD: Mock returns test data
         storage := &mockStorage{
             saveFunc: func(item *Item) error {
                 return nil // or return an error to test error handling
             },
         }

         svc := NewUserService(storage)
         err := svc.Create(user)

         // GOOD: Assert on the actual result
         if err != nil {
             t.Errorf("unexpected error: %v", err)
         }

         // BAD: Don't assert on mock internals
         // if !storage.saveCalled { ... }
     }
     ```

## Concurrency

1. **Always run tests with the race detector enabled**
   - Use `go test -race` to detect data races
   - Run race detector in CI/CD pipelines
   - Example:
     ```bash
     go test -race ./...
     ```

2. **Use `t.Parallel()` for independent tests that can run concurrently**
   - Mark tests as parallel when they don't share mutable state
   - Speeds up test execution
   - Place `t.Parallel()` at the start of the test function
   - Example:
     ```go
     func TestUserValidation(t *testing.T) {
         t.Parallel() // This test can run in parallel with others

         user := User{Name: "John"}
         err := user.Validate()
         // assertions
     }
     ```

3. **Use synchronization primitives to test goroutines safely**
   - Use `sync.WaitGroup` to wait for goroutines to complete
   - Use channels to coordinate and receive results
   - Don't rely on `time.Sleep()` for synchronization
   - Example:
     ```go
     func TestConcurrentWrites(t *testing.T) {
         var wg sync.WaitGroup
         cache := NewCache()

         for i := 0; i < 10; i++ {
             wg.Add(1)
             go func(val int) {
                 defer wg.Done()
                 cache.Set(fmt.Sprintf("key%d", val), val)
             }(i)
         }

         wg.Wait()
         // assertions
     }
     ```

4. **Test timeout scenarios using contexts with deadlines**
   - Use `context.WithTimeout()` or `context.WithDeadline()` for timeout tests
   - Verify that operations respect context cancellation
   - Example:
     ```go
     func TestOperation_Timeout(t *testing.T) {
         ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
         defer cancel()

         err := LongRunningOperation(ctx)

         if !errors.Is(err, context.DeadlineExceeded) {
             t.Errorf("expected context.DeadlineExceeded, got %v", err)
         }
     }
     ```

5. **Use channels with timeouts to prevent test hangs**
   - When testing goroutines that send results via channels, always use timeouts
   - Prevents tests from hanging indefinitely if goroutines fail
   - Example:
     ```go
     func TestAsyncOperation(t *testing.T) {
         result := make(chan string, 1)

         go func() {
             result <- PerformOperation()
         }()

         select {
         case got := <-result:
             if got != "expected" {
                 t.Errorf("got %q, want %q", got, "expected")
             }
         case <-time.After(5 * time.Second):
             t.Fatal("test timed out waiting for result")
         }
     }
     ```

## Integration Testing

1. **Integration tests must always be behind a build flag**
   - Use build tags to separate integration tests from unit tests
   - Prevents integration tests from running by default with `go test ./...`
   - Run integration tests explicitly with `go test -tags=integration ./...`
   - Example:
     ```go
     //go:build integration
     // +build integration

     package mypackage

     import "testing"

     func TestDatabaseIntegration(t *testing.T) {
         // integration test code
     }
     ```

2. **Integration tests must NEVER rely on services running outside the user's machine**
   - All external dependencies must be managed within the test
   - Don't assume databases, APIs, or other services are already running
   - Don't rely on shared test environments or infrastructure
   - Tests should be fully self-contained and reproducible

3. **Use `testcontainers` for spinning up real service instances**
   - Prefer `github.com/testcontainers/testcontainers-go` for most external services
   - Provides real instances of databases, message queues, caches, etc.
   - Tests run against actual service implementations, not mocks
   - Example:
     ```go
     //go:build integration
     // +build integration

     package mypackage

     import (
         "context"
         "testing"
         "github.com/testcontainers/testcontainers-go"
         "github.com/testcontainers/testcontainers-go/wait"
     )

     func TestWithPostgres(t *testing.T) {
         ctx := context.Background()

         req := testcontainers.ContainerRequest{
             Image:        "postgres:15",
             ExposedPorts: []string{"5432/tcp"},
             Env: map[string]string{
                 "POSTGRES_PASSWORD": "test",
                 "POSTGRES_DB":       "testdb",
             },
             WaitingFor: wait.ForLog("database system is ready to accept connections"),
         }

         container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
             ContainerRequest: req,
             Started:          true,
         })
         require.NoError(t, err)
         t.Cleanup(func() {
             container.Terminate(ctx)
         })

         // Get connection details and run tests
         host, _ := container.Host(ctx)
         port, _ := container.MappedPort(ctx, "5432")
         // test logic
     }
     ```

4. **For services incompatible with `testcontainers`, use docker-compose**
   - Create `docker-compose.test.yml` files to define the test environment
   - Use environment variables for configuration (ports, passwords, hostnames)
   - Document the docker-compose setup in test file comments or README
   - Example docker-compose.test.yml:
     ```yaml
     version: '3.8'
     services:
       redis:
         image: redis:7
         ports:
           - "${REDIS_PORT:-6379}:6379"
       postgres:
         image: postgres:15
         environment:
           POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-testpass}
           POSTGRES_DB: ${POSTGRES_DB:-testdb}
         ports:
           - "${POSTGRES_PORT:-5432}:5432"
     ```
   - Example test:
     ```go
     //go:build integration
     // +build integration

     func TestWithDockerCompose(t *testing.T) {
         // Read config from environment variables
         redisAddr := fmt.Sprintf("localhost:%s", getEnv("REDIS_PORT", "6379"))
         pgHost := getEnv("POSTGRES_HOST", "localhost")
         pgPort := getEnv("POSTGRES_PORT", "5432")

         // test logic using these connection details
     }
     ```

5. **In-process HTTP test servers are NOT valid integration tests**
   - Using `httptest.Server` to mock external APIs is NOT an integration test
   - In-process mocks only test your ability to emulate a service, not the actual service behavior
   - Integration tests must use real service instances (via testcontainers or docker-compose)
   - Example of what NOT to do:
     ```go
     // BAD: This is NOT an integration test, it's a unit test with a mock
     func TestAPIIntegration(t *testing.T) {
         server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
             w.WriteHeader(200)
             w.Write([]byte(`{"status": "ok"}`))
         }))
         defer server.Close()

         // This only tests that you can parse your own mock response
         // It does NOT test integration with the real API
     }
     ```

## Common Pitfalls

1. **Don't use `panic()` in tests**
   - Use `t.Fatal()` or `t.Fatalf()` instead of `panic()`
   - `panic()` stops the entire test suite, not just the current test
   - Example:
     ```go
     // BAD
     if err != nil {
         panic(err)
     }

     // GOOD
     if err != nil {
         t.Fatalf("unexpected error: %v", err)
     }
     ```

2. **Avoid global state and shared mutable variables**
   - Global variables can cause flaky tests and race conditions
   - Tests can interfere with each other when modifying shared state
   - Each test should create its own isolated data
   - Example:
     ```go
     // BAD: Global shared state
     var testCache = make(map[string]string)

     func TestA(t *testing.T) {
         testCache["key"] = "value"
         // test logic
     }

     // GOOD: Isolated state per test
     func TestA(t *testing.T) {
         cache := make(map[string]string)
         cache["key"] = "value"
         // test logic
     }
     ```

3. **Don't ignore test cleanup**
   - Always clean up resources (files, network connections, databases)
   - Use `t.Cleanup()` or `defer` to ensure cleanup happens
   - Leaked resources can cause subsequent tests to fail
   - Example:
     ```go
     func TestWithDatabase(t *testing.T) {
         db := setupTestDB(t)
         t.Cleanup(func() {
             db.Close()
         })
         // test logic
     }
     ```

4. **Don't rely on test execution order**
   - Tests should be independent and runnable in any order
   - Don't assume one test runs before another
   - Package-level `init()` functions should not depend on test order
   - Each test should set up its own preconditions

5. **Avoid using `time.Sleep()` for synchronization**
   - `time.Sleep()` makes tests slow and flaky
   - Use proper synchronization primitives instead
   - Example:
     ```go
     // BAD: Unreliable timing
     go doWork()
     time.Sleep(100 * time.Millisecond)
     // hope work is done

     // GOOD: Explicit synchronization
     done := make(chan bool)
     go func() {
         doWork()
         done <- true
     }()
     <-done
     ```

6. **Don't skip error checks in test code**
   - Check all errors, even in test setup
   - Use `t.Fatal()` if setup errors prevent the test from continuing
   - Example:
     ```go
     // BAD
     data, _ := os.ReadFile("testdata/input.json")

     // GOOD
     data, err := os.ReadFile("testdata/input.json")
     if err != nil {
         t.Fatalf("failed to read test file: %v", err)
     }
     ```

7. **Don't use overly complex table-driven tests**
   - If test cases require extensive setup or have different logic, consider separate test functions
   - Table-driven tests work best when test cases follow the same pattern
   - Break up tests when the table becomes hard to read or maintain
