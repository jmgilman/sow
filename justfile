# Default recipe - show available commands
default:
    @just --list

# Run all tests with race detection and coverage
test:
    @echo "Running tests..."
    cd cli && go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
    @echo "✅ Tests passed"

# Run golangci-lint
lint:
    @echo "Running golangci-lint..."
    cd cli && golangci-lint run --config=.golangci.yml
    @echo "✅ Linting passed"

# Build the sow binary for current platform
build:
    @echo "Building sow binary..."
    cd cli && CGO_ENABLED=0 go build -v -trimpath -ldflags="-s -w" -o ../sow .
    @echo "✅ Binary built: ./sow"

# Validate CUE schemas and check generated types
cue-validate:
    @echo "Validating CUE schemas..."
    cd cli/schemas && cue vet ./...
    @echo "✅ CUE schemas validated"
    @echo "Checking generated types are up-to-date..."
    cd cli && go generate ./schemas
    @if ! git diff --quiet schemas/cue_types_gen.go; then \
        echo "❌ Generated Go types are out of date!"; \
        echo "Run 'just cue-generate' to update them."; \
        exit 1; \
    fi
    @echo "✅ Generated types are up-to-date"

# Generate Go types from CUE schemas
cue-generate:
    @echo "Generating Go types from CUE schemas..."
    cd cli && go generate ./schemas
    @echo "✅ Types generated"

# Run all CI checks locally (matches GitHub Actions)
ci: test lint cue-validate build
    @echo ""
    @echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    @echo "✅ All CI checks passed!"
    @echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test GoReleaser locally with snapshot (no publish)
release-local VERSION:
    @echo "Testing GoReleaser with version {{VERSION}}..."
    @echo "This will create a snapshot build in dist/"
    goreleaser release --snapshot --clean --skip=publish
    @echo "✅ Snapshot build complete - check dist/ directory"

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -f sow
    rm -rf cli/coverage.out
    rm -rf dist/
    @echo "✅ Cleaned"

# Install the sow binary to $GOPATH/bin
install:
    @echo "Installing sow to $GOPATH/bin..."
    cd cli && go install .
    @echo "✅ Installed"

# Format all Go code
fmt:
    @echo "Formatting Go code..."
    cd cli && go fmt ./...
    @echo "✅ Formatted"

# Tidy go.mod
tidy:
    @echo "Tidying go.mod..."
    cd cli && go mod tidy
    @echo "✅ Tidied"

# Show coverage report in browser
coverage: test
    @echo "Opening coverage report..."
    cd cli && go tool cover -html=coverage.out
