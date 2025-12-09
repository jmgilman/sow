# Go Style Guide

## Golden Rules

These rules are non-negotiable and must be followed at all times.

1. **Simplicity over cleverness - Write clear, straightforward code**
   - Prioritize readability over showcasing language features
   - Example:
     ```go
     // BAD: Clever one-liner
     return len(s) > 0 && s[0] != ' ' && s[len(s)-1] != ' '

     // GOOD: Clear and explicit
     if len(s) == 0 {
         return false
     }
     return s[0] != ' ' && s[len(s)-1] != ' '
     ```

2. **Make zero values useful - Design types so their zero value is valid and usable**
   - Example:
     ```go
     // BAD: Requires constructor
     type Buffer struct {
         data []byte
         size int
     }

     // GOOD: Zero value ready to use
     type Buffer struct {
         data []byte
     }
     func (b *Buffer) Write(p []byte) {
         b.data = append(b.data, p...)  // Works with nil slice
     }
     ```

3. **Accept interfaces, return concrete types**
   - Example:
     ```go
     // BAD
     func NewCache(size int) Cache

     // GOOD
     func NewCache(size int) *MemoryCache
     func ProcessData(r io.Reader) error
     ```

4. **Don't panic in library code - Return errors instead**
   - Acceptable: `init()`, `main` package, impossible conditions
   - Example:
     ```go
     // BAD
     func ParseConfig(data []byte) Config {
         var cfg Config
         if err := json.Unmarshal(data, &cfg); err != nil {
             panic(err)
         }
         return cfg
     }

     // GOOD
     func ParseConfig(data []byte) (Config, error) {
         var cfg Config
         if err := json.Unmarshal(data, &cfg); err != nil {
             return Config{}, fmt.Errorf("parse config: %w", err)
         }
         return cfg, nil
     }
     ```

5. **Avoid global mutable state**
   - Use dependency injection
   - Constants and package-level defaults are acceptable
   - Example:
     ```go
     // BAD
     var defaultClient *http.Client
     func FetchData(url string) ([]byte, error)

     // GOOD
     type Fetcher struct {
         client *http.Client
     }
     func (f *Fetcher) FetchData(url string) ([]byte, error)
     ```

6. **Explicit is better than implicit - No magic behavior**
   - Example:
     ```go
     // BAD: Hidden side effect
     func (c *Cache) Get(key string) string {
         if time.Since(c.lastCleanup) > time.Hour {
             c.cleanup()  // Surprise!
         }
         return c.data[key]
     }

     // GOOD: Explicit operations
     func (c *Cache) Get(key string) string
     func (c *Cache) Cleanup()
     ```

7. **Never re-implement functionality from the standard library**
   - Exception: Document performance/feature requirements
   - Example:
     ```go
     // BAD
     func trimWhitespace(s string) string { /* custom impl */ }

     // GOOD
     result := strings.TrimSpace(s)
     ```

## Code Organization

1. **Package structure: One package per directory, name matches directory**
   - Package names: short, lowercase, singular
   - Avoid: `util`, `common`, `helper`, underscores, mixed caps
   - Example:
     ```
     myproject/
     ├── handler/    # package handler
     ├── storage/    # package storage
     └── api/        # package api
     ```

2. **Organize by domain, not technical layer**
   - Example:
     ```
     // BAD
     myproject/
     ├── models/
     ├── controllers/
     └── repositories/

     // GOOD
     myproject/
     ├── user/
     │   ├── user.go
     │   ├── handler.go
     │   └── storage.go
     ├── product/
     └── order/
     ```

3. **Always use `cmd/` directory for binaries (even single binary)**
   - Keep `main` minimal: flag parsing, config, dependency wiring only
   - All business logic in importable packages
   - Example:
     ```
     myproject/
     ├── cmd/
     │   └── server/
     │       └── main.go
     ├── handler/
     └── storage/
     ```

4. **Use `internal/` for non-exported packages**
   - Example:
     ```
     myproject/
     ├── api/              # Exported
     └── internal/
         ├── config/       # Cannot be imported externally
         └── database/
     ```

5. **Import grouping: stdlib, external, internal**
   - Use `goimports` for automatic formatting
   - Example:
     ```go
     import (
         "context"
         "fmt"

         "github.com/gorilla/mux"

         "myproject/internal/config"
     )
     ```

6. **Keep package depth to 2-3 levels maximum**
   - Example:
     ```go
     // BAD
     import "myproject/internal/services/user/handler/http/v1"

     // GOOD
     import "myproject/internal/user"
     ```

7. **Use `doc.go` for complex package documentation**
   - Example:
     ```go
     // Package handler provides HTTP request handlers.
     //
     // Example:
     //     h := handler.New(storage)
     //     http.HandleFunc("/users", h.ListUsers)
     package handler
     ```

8. **Organize code within files in this order**
   - Imports
   - Constants
   - Type declarations (non-structs)
   - Interfaces
   - Structs
   - Struct constructors
   - Struct public methods (alphabetized)
   - Struct private methods (alphabetized)
   - Public functions (alphabetized)
   - Private functions (alphabetized)
   - Example:
     ```go
     package user

     import (
         "context"
         "errors"
     )

     const MaxNameLength = 100

     type Role string

     type Storage interface {
         Save(ctx context.Context, u *User) error
     }

     type User struct {
         id    string
         email string
     }

     func NewUser(email string) *User {
         return &User{email: email}
     }

     func (u *User) Email() string {
         return u.email
     }

     func (u *User) Validate() error {
         return u.validate()
     }

     func (u *User) validate() error {
         if u.email == "" {
             return errors.New("email required")
         }
         return nil
     }

     func FormatEmail(email string) string {
         return strings.ToLower(email)
     }

     func validateDomain(email string) bool {
         // ...
     }
     ```

## Naming Conventions

1. **Variable names: short for small scope, descriptive for large scope**
   - Example:
     ```go
     // GOOD
     for i := 0; i < len(items); i++ { }
     for _, user := range users { }

     func ProcessOrder(customerID string) error {
         var totalPrice float64
         var validationErrors []error
         // ...
     }

     // BAD
     for indexCounter := 0; indexCounter < len(items); indexCounter++ { }
     func ProcessOrder(cid string) error {
         var tp float64
         var ve []error
     }
     ```

2. **Use MixedCaps/mixedCaps, never underscores or ALL_CAPS**
   - Example:
     ```go
     // BAD
     type user_service struct{}
     const MAX_RETRY = 3

     // GOOD
     type UserService struct{}
     const MaxRetry = 3
     ```

3. **Acronyms: all uppercase or all lowercase**
   - Example:
     ```go
     // BAD
     type HttpServer struct{}
     var userId string

     // GOOD
     type HTTPServer struct{}
     var userID string
     ```

4. **Interfaces: describe behavior, often -er suffix**
   - Avoid: `Interface`, `Manager`, `Helper`, `I` prefix
   - Example:
     ```go
     // GOOD
     type Reader interface
     type Validator interface
     type Storage interface

     // BAD
     type IReader interface
     type ReaderInterface interface
     ```

5. **Avoid package name stuttering**
   - Example:
     ```go
     // BAD: package user
     type UserService struct{}

     // GOOD: package user
     type Service struct{}
     // Used as: user.Service
     ```

6. **Receiver names: short (1-2 chars), consistent**
   - Avoid: `this`, `self`, `me`
   - Example:
     ```go
     // GOOD
     func (u *User) SetName(name string)
     func (u *User) Name() string

     // BAD
     func (user *User) SetName(name string)
     func (this *User) Name() string
     ```

7. **Function names: verbs; Getters omit 'Get'; Errors use 'Err' prefix**
   - Example:
     ```go
     // Functions
     func ParseConfig() (*Config, error)
     func IsValid() bool

     // Getters/Setters
     func (u *User) Name() string          // Not GetName
     func (u *User) SetName(name string)

     // Errors
     var ErrNotFound = errors.New("not found")
     type ValidationError struct{}
     ```

## Functions & Methods

1. **Keep functions small (50-80 lines) and single-purpose**
   - Example:
     ```go
     // BAD: Does fetch, validate, email, analytics
     func ProcessUser(userID string) error { /* 100 lines */ }

     // GOOD: Split into focused functions
     func ProcessUser(userID string) error {
         user, err := fetchUser(userID)
         // ...
         if err := validateUser(user); err != nil
         if err := sendWelcomeEmail(user); err != nil
         trackUserProcessing(user.ID)
     }
     ```

2. **Use early returns to reduce nesting**
   - Example:
     ```go
     // BAD
     func Process(r *Request) error {
         if r != nil {
             if r.IsValid() {
                 if r.User != nil {
                     return process(r)
                 }
             }
         }
     }

     // GOOD
     func Process(r *Request) error {
         if r == nil {
             return errors.New("request is nil")
         }
         if !r.IsValid() {
             return errors.New("invalid")
         }
         if r.User == nil {
             return errors.New("no user")
         }
         return process(r)
     }
     ```

3. **Use pointer receivers by default for consistency**
   - Exception: Small immutable types (Point, Color)
   - Be consistent: if one method needs pointer, all methods use pointers
   - Example:
     ```go
     type UserService struct { db *sql.DB }

     func (s *UserService) GetUser(id string) (*User, error)
     func (s *UserService) UpdateUser(user *User) error
     ```

4. **Parameter order: ctx first, required before optional, errors last**
   - Example:
     ```go
     // GOOD
     func FetchUser(ctx context.Context, userID string, opts *Options) (*User, error)

     // BAD
     func FetchUser(opts *Options, userID string, ctx context.Context) (error, *User)
     ```

5. **Use functional options pattern for many optional parameters**
   - Example:
     ```go
     type ServerOption func(*Server)

     func WithTimeout(d time.Duration) ServerOption {
         return func(s *Server) { s.timeout = d }
     }

     func NewServer(addr string, opts ...ServerOption) *Server {
         s := &Server{addr: addr, timeout: 30 * time.Second}
         for _, opt := range opts {
             opt(s)
         }
         return s
     }

     // Usage
     srv := NewServer(":8080", WithTimeout(60*time.Second))
     ```

6. **Avoid named return values except for docs/short functions**
   - Never use naked returns in long functions
   - Example:
     ```go
     // GOOD: Short function
     func Split(path string) (dir, file string) {
         i := strings.LastIndex(path, "/")
         return path[:i], path[i+1:]
     }

     // BAD: Long function with naked return
     func ProcessData(input []byte) (result []byte, err error) {
         // ... 50 lines ...
         return  // Unclear what's returned
     }
     ```

7. **Group 4+ related parameters into a struct**
   - Example:
     ```go
     // BAD
     func CreateUser(name, email, phone, address, city, state, zip string) (*User, error)

     // GOOD
     type CreateUserRequest struct {
         Name    string
         Email   string
         Phone   string
         Address Address
     }
     func CreateUser(req CreateUserRequest) (*User, error)
     ```

## Error Handling

1. **Always check errors - Never ignore them**
   - Use `_` only when you're certain there's no error
   - Example:
     ```go
     // BAD
     data, _ := os.ReadFile("config.json")
     json.Unmarshal(data, &cfg)

     // GOOD
     data, err := os.ReadFile("config.json")
     if err != nil {
         return fmt.Errorf("read config: %w", err)
     }
     if err := json.Unmarshal(data, &cfg); err != nil {
         return fmt.Errorf("parse config: %w", err)
     }
     ```

2. **Wrap errors with context using `%w`**
   - Preserves error chain for `errors.Is()` and `errors.As()`
   - Add context about what operation failed
   - Example:
     ```go
     // BAD
     if err != nil {
         return err  // Lost context
     }
     if err != nil {
         return fmt.Errorf("failed: %v", err)  // Breaks error chain
     }

     // GOOD
     if err != nil {
         return fmt.Errorf("fetch user %s: %w", userID, err)
     }
     ```

3. **Use `errors.Is()` for sentinel errors, `errors.As()` for error types**
   - Don't compare error strings
   - Example:
     ```go
     // BAD
     if err.Error() == "not found" { }
     if err == io.EOF { }  // Fails if error is wrapped

     // GOOD
     if errors.Is(err, io.EOF) { }
     if errors.Is(err, ErrNotFound) { }

     var validationErr *ValidationError
     if errors.As(err, &validationErr) {
         // Handle validation error
     }
     ```

4. **Define sentinel errors at package level, custom types when needed**
   - Sentinel errors: simple cases with no additional data
   - Custom types: when you need to attach context
   - Example:
     ```go
     // Sentinel errors
     var (
         ErrNotFound     = errors.New("not found")
         ErrUnauthorized = errors.New("unauthorized")
     )

     // Custom error type with context
     type ValidationError struct {
         Field   string
         Message string
     }

     func (e *ValidationError) Error() string {
         return fmt.Sprintf("%s: %s", e.Field, e.Message)
     }

     // Usage
     if user.Email == "" {
         return &ValidationError{Field: "email", Message: "required"}
     }
     ```

5. **Handle errors at the appropriate level**
   - Log at the top level (main, handlers)
   - Add context at each level
   - Don't log-and-return (causes duplicate logs)
   - Example:
     ```go
     // BAD: Logs at every level
     func fetchUser(id string) (*User, error) {
         user, err := db.Query(id)
         if err != nil {
             log.Printf("database error: %v", err)  // Don't log here
             return nil, err
         }
         return user, nil
     }

     // GOOD: Add context, log at top level
     func fetchUser(id string) (*User, error) {
         user, err := db.Query(id)
         if err != nil {
             return nil, fmt.Errorf("query user %s: %w", id, err)
         }
         return user, nil
     }

     func Handler(w http.ResponseWriter, r *http.Request) {
         user, err := fetchUser(id)
         if err != nil {
             log.Printf("fetch user failed: %v", err)  // Log here
             http.Error(w, "internal error", 500)
             return
         }
     }
     ```

6. **Return early on errors - don't use else blocks**
   - Example:
     ```go
     // BAD
     if err != nil {
         return err
     } else {
         // continue processing
         result := process(data)
         return nil
     }

     // GOOD
     if err != nil {
         return err
     }
     result := process(data)
     return nil
     ```

7. **Use defer for cleanup, even on error paths**
   - Ensures resources are released
   - Example:
     ```go
     // GOOD
     func processFile(path string) error {
         f, err := os.Open(path)
         if err != nil {
             return err
         }
         defer f.Close()  // Always closes, even on error

         // Process file...
         if err := process(f); err != nil {
             return err  // File still closes
         }
         return nil
     }
     ```

## Interfaces & Types

1. **Keep interfaces small - prefer single-method interfaces**
   - Easier to implement and test
   - More composable and flexible
   - Example:
     ```go
     // GOOD: Small, focused interfaces
     type Reader interface {
         Read(p []byte) (n int, err error)
     }

     type Writer interface {
         Write(p []byte) (n int, err error)
     }

     type ReadWriter interface {
         Reader
         Writer
     }

     // BAD: Large interface
     type Storage interface {
         Read(key string) ([]byte, error)
         Write(key string, data []byte) error
         Delete(key string) error
         List() ([]string, error)
         Close() error
         Ping() error
         Stats() (Statistics, error)
     }
     ```

2. **Define interfaces in consumer packages, not producer packages**
   - Define interfaces where they're used, not where they're implemented
   - Keeps packages decoupled
   - Example:
     ```go
     // BAD: storage package defines interface
     package storage
     type Storage interface {
         Get(key string) ([]byte, error)
     }
     type PostgresStorage struct{}
     func (p *PostgresStorage) Get(key string) ([]byte, error)

     // GOOD: consumer defines interface
     package handler
     type storage interface {  // lowercase, internal
         Get(key string) ([]byte, error)
     }
     type Handler struct {
         storage storage
     }

     package storage
     type PostgresStorage struct{}  // Just implements it
     func (p *PostgresStorage) Get(key string) ([]byte, error)
     ```

3. **Use composition over inheritance**
   - Embed types to reuse behavior
   - Promotes flexibility and clear relationships
   - Example:
     ```go
     // GOOD: Composition via embedding
     type Logger struct {
         writer io.Writer
     }

     type HTTPLogger struct {
         Logger  // Embeds Logger
         prefix string
     }

     // GOOD: Explicit composition
     type Service struct {
         storage Storage
         cache   Cache
         logger  Logger
     }
     ```

4. **Order struct fields: groups by purpose, consider alignment**
   - Group related fields together
   - Consider memory alignment for hot-path structs
   - Example:
     ```go
     // GOOD: Grouped by purpose
     type Server struct {
         // Connection settings
         addr string
         port int

         // Timeouts
         readTimeout  time.Duration
         writeTimeout time.Duration

         // Dependencies
         storage Storage
         logger  Logger

         // Internal state
         mu      sync.Mutex
         running bool
     }

     // GOOD: Optimized for alignment (hot path)
     type Event struct {
         timestamp int64   // 8 bytes
         userID    int64   // 8 bytes
         eventType int32   // 4 bytes
         flags     int32   // 4 bytes
         name      string  // 16 bytes
     }
     ```

5. **Use pointers for large structs or when mutability is needed**
   - Small structs (< 100 bytes): can use values
   - Large structs: use pointers to avoid copies
   - Mutable structs: use pointers
   - Example:
     ```go
     // Small, immutable: use values
     type Point struct {
         X, Y int
     }
     func Translate(p Point, dx, dy int) Point

     // Large or mutable: use pointers
     type User struct {
         ID       string
         Email    string
         Profile  Profile
         Settings Settings
         // ... many fields
     }
     func UpdateUser(u *User) error

     type Counter struct {
         count int
     }
     func (c *Counter) Increment()
     ```

6. **Use struct tags for metadata, keep them consistent**
   - Common tags: `json`, `yaml`, `db`, `validate`
   - Be consistent with naming conventions
   - Example:
     ```go
     type User struct {
         ID        string    `json:"id" db:"user_id"`
         Email     string    `json:"email" db:"email" validate:"required,email"`
         CreatedAt time.Time `json:"created_at" db:"created_at"`
         UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
     }
     ```

7. **Don't export unnecessary fields - prefer getter methods when needed**
   - Unexported fields allow you to change internal representation
   - Exported fields become part of your API
   - Example:
     ```go
     // BAD: Exposes internal state
     type Cache struct {
         Data map[string]interface{}
         Mu   sync.RWMutex
     }

     // GOOD: Encapsulates internals
     type Cache struct {
         data map[string]interface{}
         mu   sync.RWMutex
     }

     func (c *Cache) Get(key string) (interface{}, bool) {
         c.mu.RLock()
         defer c.mu.RUnlock()
         val, ok := c.data[key]
         return val, ok
     }
     ```

## Comments & Documentation

1. **All exported identifiers must have doc comments**
   - Start with the identifier name
   - Use complete sentences
   - Example:
     ```go
     // GOOD
     // User represents a user account in the system.
     type User struct {
         ID    string
         Email string
     }

     // NewUser creates a new user with the given email.
     func NewUser(email string) *User {
         return &User{Email: email}
     }

     // BAD
     // user struct
     type User struct{}

     // creates a user
     func NewUser(email string) *User
     ```

2. **Package comment should describe the package purpose**
   - Place in `doc.go` or at the top of any file in the package
   - Start with "Package <name>"
   - Example:
     ```go
     // Package handler provides HTTP request handlers for the user API.
     //
     // The handlers support JSON request/response formats and include
     // built-in validation and error handling.
     package handler
     ```

4. **Use TODO comments with issue references**
   - Format: `// TODO(#issue): description`
   - Always link to an issue or ticket
   - Example:
     ```go
     // GOOD
     // TODO(#123): Add support for pagination
     // TODO(alice): Optimize this query for large datasets

     // BAD
     // TODO: fix this later
     // FIXME
     ```

5. **Use examples in doc comments for complex APIs**
   - Example:
     ```go
     // ProcessBatch processes multiple items in parallel.
     //
     // Example:
     //
     //     items := []Item{{ID: "1"}, {ID: "2"}}
     //     results, err := ProcessBatch(ctx, items, 10)
     //     if err != nil {
     //         log.Fatal(err)
     //     }
     //
     func ProcessBatch(ctx context.Context, items []Item, workers int) ([]Result, error)
     ```
