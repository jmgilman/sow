# ADR 002: Build and Distribution Strategy

**Status**: Accepted

**Date**: 2025-10-13

**Context**: Milestone 1 (CLI Foundation & Schema System)

---

## Context

The sow CLI (defined in ADR 001) requires a robust build and distribution strategy that:

1. Produces cross-platform binaries (macOS, Linux, Windows)
2. Embeds CUE schemas with guaranteed version alignment
3. Integrates with GitHub Releases for distribution
4. Provides straightforward installation methods
5. Supports version checking and upgrade workflows
6. Enables automated releases with minimal manual intervention

The CLI is a required component for using sow, and must be distributed alongside the Claude Code plugin. Users need a simple path from "install plugin" to "download CLI" to "start using sow."

## Decision

We will use **GoReleaser** for automated builds and releases, with **GitHub Releases** as the primary distribution channel.

### 1. Build Tooling: GoReleaser

**Tool**: [GoReleaser](https://goreleaser.com/)

**Rationale**:
- Industry-standard for Go project releases
- Handles cross-compilation automatically
- Generates checksums and signatures
- Creates GitHub Releases with assets
- Supports Homebrew tap generation
- Minimal configuration, maximum automation

**Configuration** (`.goreleaser.yml`):

```yaml
# .goreleaser.yml
project_name: sow

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - id: sow
    main: ./cmd/sow
    binary: sow

    # Embed version info at build time
    ldflags:
      - -s -w
      - -X sow/internal/config.Version={{.Version}}
      - -X sow/internal/config.BuildDate={{.Date}}
      - -X sow/internal/config.Commit={{.Commit}}

    # Target platforms
    goos:
      - darwin
      - linux
      - windows

    goarch:
      - amd64
      - arm64

    # Platform-specific exclusions
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - id: sow-archive
    format: tar.gz

    # Use zip for Windows
    format_overrides:
      - goos: windows
        format: zip

    # Archive naming: sow_0.2.0_darwin_amd64.tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

    files:
      - LICENSE
      - README.md
      - CHANGELOG.md

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

snapshot:
  name_template: "{{ incpatch .Version }}-snapshot"

changelog:
  sort: asc
  use: github
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
  groups:
    - title: Features
      regexp: "^.*feat[(\\w)]*:+.*$"
      order: 0
    - title: Bug Fixes
      regexp: "^.*fix[(\\w)]*:+.*$"
      order: 1
    - title: Others
      order: 999

release:
  github:
    owner: your-org
    name: sow

  # Draft release for manual review before publishing
  draft: false

  # Replace existing release if re-running
  replace: false

  # Release notes
  header: |
    ## sow {{ .Version }}

    Download the appropriate binary for your platform below.

    **Installation**:
    ```bash
    # macOS (Intel)
    curl -L https://github.com/your-org/sow/releases/download/{{ .Tag }}/sow_{{ .Version }}_darwin_amd64.tar.gz | tar xz
    sudo mv sow /usr/local/bin/

    # macOS (Apple Silicon)
    curl -L https://github.com/your-org/sow/releases/download/{{ .Tag }}/sow_{{ .Version }}_darwin_arm64.tar.gz | tar xz
    sudo mv sow /usr/local/bin/

    # Linux
    curl -L https://github.com/your-org/sow/releases/download/{{ .Tag }}/sow_{{ .Version }}_linux_amd64.tar.gz | tar xz
    sudo mv sow /usr/local/bin/

    # Windows
    # Download sow_{{ .Version }}_windows_amd64.zip and extract to PATH
    ```

    Verify installation:
    ```bash
    sow --version
    ```

  footer: |
    **Full Changelog**: https://github.com/your-org/sow/compare/{{ .PreviousTag }}...{{ .Tag }}

brews:
  - name: sow

    # GitHub repository for Homebrew tap
    tap:
      owner: your-org
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"

    # Git author for tap commits
    commit_author:
      name: sow-bot
      email: bot@example.com

    # Formula details
    homepage: "https://github.com/your-org/sow"
    description: "AI-powered system of work for software engineering"
    license: "MIT"

    # Installation
    install: |
      bin.install "sow"

    # Tests
    test: |
      system "#{bin}/sow", "--version"
```

**Why GoReleaser**:
1. **Zero Manual Work**: Tag + push = full release
2. **Cross-Compilation**: All platforms built automatically
3. **Checksums**: Security verification built-in
4. **GitHub Integration**: Creates releases with assets
5. **Homebrew Support**: Generates tap formula automatically
6. **Changelog Generation**: Extracts from git commits
7. **Industry Standard**: Well-maintained, widely-used

**Alternative Considered**: Manual Makefile

```makefile
# Rejected: Too much manual work, error-prone
VERSION := 0.2.0

build-all:
	GOOS=darwin GOARCH=amd64 go build -o sow-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o sow-darwin-arm64
	GOOS=linux GOARCH=amd64 go build -o sow-linux-amd64
	GOOS=windows GOARCH=amd64 go build -o sow-windows-amd64.exe
	# Still need to create checksums, GitHub release, upload assets...
```

**Rejected Because**:
- Manual GitHub release creation
- Manual checksum generation
- Manual asset uploads
- No changelog automation
- No Homebrew tap support
- Error-prone and tedious

---

### 2. Binary Naming Convention

**Format**: `sow_{version}_{os}_{arch}.tar.gz`

**Examples**:
- `sow_0.2.0_darwin_amd64.tar.gz` (macOS Intel)
- `sow_0.2.0_darwin_arm64.tar.gz` (macOS Apple Silicon)
- `sow_0.2.0_linux_amd64.tar.gz` (Linux)
- `sow_0.2.0_windows_amd64.zip` (Windows)

**Inside Archive**:
- `sow` (or `sow.exe` on Windows)
- `LICENSE`
- `README.md`
- `CHANGELOG.md`

**Rationale**:
- Standard naming convention (platform/tool agnostic)
- Version explicit in filename
- Easy to script downloads
- Archive includes documentation

---

### 3. Version Embedding

**At Build Time**:

```go
// internal/config/version.go
package config

// Version info injected via ldflags at build time
var (
    Version   = "dev"          // Semantic version
    BuildDate = "unknown"      // ISO 8601 timestamp
    Commit    = "none"         // Git commit hash
)
```

**Build Command** (via GoReleaser):

```bash
go build -ldflags "\
  -X sow/internal/config.Version=0.2.0 \
  -X sow/internal/config.BuildDate=2025-10-13T14:30:00Z \
  -X sow/internal/config.Commit=abc1234"
```

**Version Command Output**:

```bash
$ sow --version
sow 0.2.0
Built: 2025-10-13T14:30:00Z
Commit: abc1234
```

**Schema Version Alignment**:

The CLI version IS the schema version. CUE schemas are embedded at build time, so:

```
CLI Version 0.2.0 = Embedded Schemas v0.2.0
```

No separate schema versioning needed.

**Rationale**:
- Single source of truth (git tag)
- Version visible in binary
- Build metadata for debugging
- Schema version guaranteed to match

---

### 4. Release Process

**Fully Automated Workflow**:

```bash
# 1. Update version in code (if needed - GoReleaser uses git tags)
# No code changes needed for version bumps

# 2. Update CHANGELOG.md
cat >> CHANGELOG.md << EOF
## [0.2.0] - 2025-10-13

### Added
- Fast logging command (sow log)
- Schema validation (sow validate)
- Sink management (sow sinks)

### Fixed
- Context detection on Windows paths

### Changed
- Improved error messages
EOF

git add CHANGELOG.md
git commit -m "chore: update CHANGELOG for v0.2.0"
git push

# 3. Create and push tag
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0

# 4. GitHub Actions runs GoReleaser (triggered by tag push)
# - Builds all platforms
# - Runs tests
# - Generates checksums
# - Creates GitHub Release
# - Uploads assets
# - Updates Homebrew tap
```

**GitHub Actions Workflow** (`.github/workflows/release.yml`):

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -v ./...

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

**What Happens Automatically**:
1. Tests run (fail = no release)
2. Binaries built for all platforms
3. Archives created with docs
4. Checksums generated
5. GitHub Release created
6. Assets uploaded
7. Homebrew tap updated
8. Release notes generated from commits

**Manual Step**: Review and publish release (if draft mode enabled)

**Rationale**:
- Tag push = single trigger
- No manual binary builds
- No manual asset uploads
- No manual Homebrew updates
- Consistent, repeatable process
- Reduced human error

---

### 5. Installation Methods

#### Method 1: Direct Binary Download (Primary)

**macOS/Linux**:

```bash
# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi

VERSION="0.2.0"

# Download and install
curl -L "https://github.com/your-org/sow/releases/download/v${VERSION}/sow_${VERSION}_${OS}_${ARCH}.tar.gz" | tar xz
sudo mv sow /usr/local/bin/

# Verify
sow --version
```

**Windows** (PowerShell):

```powershell
$VERSION = "0.2.0"
$URL = "https://github.com/your-org/sow/releases/download/v$VERSION/sow_${VERSION}_windows_amd64.zip"

# Download
Invoke-WebRequest -Uri $URL -OutFile sow.zip

# Extract
Expand-Archive -Path sow.zip -DestinationPath .

# Move to PATH (adjust path as needed)
Move-Item -Path .\sow.exe -Destination $env:USERPROFILE\bin\

# Verify
sow --version
```

**Installation Script** (`install.sh`):

```bash
#!/bin/bash
# Install script for sow CLI
set -e

# Configuration
REPO="your-org/sow"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin|linux) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest version from GitHub API
echo "Fetching latest version..."
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4 | sed 's/v//')

if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version"
    exit 1
fi

echo "Installing sow v${VERSION}..."

# Download
ASSET="sow_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${ASSET}"

echo "Downloading from ${URL}..."
curl -L "$URL" | tar xz

# Install
echo "Installing to ${INSTALL_DIR}..."
sudo mv sow "${INSTALL_DIR}/"

# Verify
echo "✓ Installation complete!"
echo ""
"${INSTALL_DIR}/sow" --version
echo ""
echo "Run 'sow --help' to get started"
```

**Usage**:

```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/your-org/sow/main/install.sh | bash

# Install to custom location
curl -sSL https://raw.githubusercontent.com/your-org/sow/main/install.sh | INSTALL_DIR=$HOME/.local/bin bash

# Install specific version
curl -sSL https://raw.githubusercontent.com/your-org/sow/main/install.sh | VERSION=0.1.0 bash
```

---

#### Method 2: Homebrew (macOS/Linux)

**Installation**:

```bash
# Add tap
brew tap your-org/tap

# Install
brew install sow

# Verify
sow --version
```

**Update**:

```bash
brew update
brew upgrade sow
```

**How It Works**:
- GoReleaser automatically updates Homebrew tap on release
- Formula generated from `.goreleaser.yml` config
- Homebrew handles download, checksum verification, installation
- Updates via standard `brew upgrade`

**Formula Location**: `https://github.com/your-org/homebrew-tap/blob/main/Formula/sow.rb`

**Generated Formula** (example):

```ruby
class Sow < Formula
  desc "AI-powered system of work for software engineering"
  homepage "https://github.com/your-org/sow"
  version "0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/your-org/sow/releases/download/v0.2.0/sow_0.2.0_darwin_amd64.tar.gz"
      sha256 "abc123..."
    end
    if Hardware::CPU.arm?
      url "https://github.com/your-org/sow/releases/download/v0.2.0/sow_0.2.0_darwin_arm64.tar.gz"
      sha256 "def456..."
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/your-org/sow/releases/download/v0.2.0/sow_0.2.0_linux_amd64.tar.gz"
      sha256 "ghi789..."
    end
  end

  def install
    bin.install "sow"
  end

  test do
    system "#{bin}/sow", "--version"
  end
end
```

---

#### Method 3: Go Install (For Developers)

**Installation**:

```bash
go install github.com/your-org/sow/cmd/sow@latest

# Or specific version
go install github.com/your-org/sow/cmd/sow@v0.2.0
```

**Rationale**:
- Convenient for Go developers
- Always builds from source
- Requires Go toolchain
- Not recommended for end users

---

#### Method 4: Package Managers (Future)

**Not Implemented in Milestone 1**, but documented for future reference:

**apt (Debian/Ubuntu)**:

```bash
# Add repository
curl -s https://packagecloud.io/install/repositories/your-org/sow/script.deb.sh | sudo bash

# Install
sudo apt install sow
```

**yum/dnf (RHEL/Fedora)**:

```bash
# Add repository
curl -s https://packagecloud.io/install/repositories/your-org/sow/script.rpm.sh | sudo bash

# Install
sudo yum install sow
```

**Chocolatey (Windows)**:

```powershell
choco install sow
```

**Scoop (Windows)**:

```powershell
scoop bucket add your-org https://github.com/your-org/scoop-bucket
scoop install sow
```

**Why Defer**:
- Requires maintaining additional infrastructure
- Homebrew + direct download covers most users
- Package manager setup is time-consuming
- Can be added post-launch based on demand

---

### 6. Installation Verification

**After Installation**:

```bash
# Check version
sow --version
# Expected: sow 0.2.0

# Check installation path
which sow
# Expected: /usr/local/bin/sow (or install location)

# Run help
sow --help
# Should display command list

# Verify schemas are embedded
sow schema list
# Should list all schemas
```

**Common Issues and Solutions**:

| Issue | Solution |
|-------|----------|
| `sow: command not found` | Add install directory to PATH |
| Permission denied | Use `sudo mv` or install to user directory |
| Wrong version showing | Multiple installations, check PATH order |
| Schemas not found | Re-download binary (corruption) |

---

### 7. Version Alignment Strategy

**CLI Version = Plugin Version = Schema Version**

**Alignment Check**:

```bash
# Check CLI version
sow --version
# sow 0.2.0

# Check plugin version
cat .claude/.plugin-version
# 0.2.0

# Check repository structure version
cat .sow/.version
# sow_structure_version: 0.2.0
# plugin_version: 0.2.0
```

**Version Mismatch Detection**:

Via `sow session-info` (SessionStart hook):

```bash
$ sow session-info

📋 You are in a sow-enabled repository

⚠️  Version mismatch detected!
   Repository structure: 0.1.0
   Plugin version: 0.2.0
   CLI version: 0.2.0

💡 Run /migrate to upgrade your repository structure
```

**Upgrade Workflow**:

1. **Upgrade Plugin**:
   ```
   /plugin install sow@sow-marketplace
   # Restart Claude Code
   ```

2. **Upgrade CLI**:
   ```bash
   # Homebrew
   brew upgrade sow

   # Or direct download
   curl -sSL https://raw.githubusercontent.com/your-org/sow/main/install.sh | bash
   ```

3. **Verify Alignment**:
   ```bash
   sow --version        # Should match plugin version
   cat .claude/.plugin-version  # Should match CLI version
   ```

4. **Migrate Repository** (if needed):
   ```
   /migrate
   ```

**Rationale**:
- Single version number across all components
- Clear detection of mismatches
- Guided upgrade path
- No version drift

---

### 8. Upgrade Path Strategy

**Semantic Versioning**:

```
MAJOR.MINOR.PATCH

Examples:
0.1.0 → 0.1.1  (Patch: bug fixes, no migration)
0.1.1 → 0.2.0  (Minor: new features, may need migration)
0.9.0 → 1.0.0  (Major: breaking changes, migration required)
```

**Upgrade Scenarios**:

| From | To | CLI Upgrade | Plugin Upgrade | Migration Needed? |
|------|-----|-------------|----------------|-------------------|
| 0.1.0 | 0.1.1 | Yes | Optional | No |
| 0.1.0 | 0.2.0 | Yes | Yes | Maybe |
| 0.9.0 | 1.0.0 | Yes | Yes | Yes |

**CLI Upgrade Commands**:

```bash
# Homebrew
brew upgrade sow

# Direct download (latest)
curl -sSL https://raw.githubusercontent.com/your-org/sow/main/install.sh | bash

# Direct download (specific version)
curl -sSL https://raw.githubusercontent.com/your-org/sow/main/install.sh | VERSION=0.2.0 bash

# Go install
go install github.com/your-org/sow/cmd/sow@latest
```

**Plugin Upgrade**:

```
/plugin install sow@sow-marketplace
# Restart Claude Code
```

**Migration** (if needed):

```
/migrate
```

**Deprecation Policy**:

- CLI maintains compatibility with schemas N and N-1
- CLI version 0.3.0 supports structures from 0.2.0 and 0.3.0
- Older structures prompt migration
- No silent breaking changes

---

### 9. Checksums and Verification

**Checksums Generated** (via GoReleaser):

```
checksums.txt:
abc123...  sow_0.2.0_darwin_amd64.tar.gz
def456...  sow_0.2.0_darwin_arm64.tar.gz
ghi789...  sow_0.2.0_linux_amd64.tar.gz
jkl012...  sow_0.2.0_windows_amd64.zip
```

**Verification**:

```bash
# Download binary and checksums
curl -LO https://github.com/your-org/sow/releases/download/v0.2.0/sow_0.2.0_darwin_amd64.tar.gz
curl -LO https://github.com/your-org/sow/releases/download/v0.2.0/checksums.txt

# Verify
sha256sum --check checksums.txt --ignore-missing
# sow_0.2.0_darwin_amd64.tar.gz: OK

# Or manually
sha256sum sow_0.2.0_darwin_amd64.tar.gz
# Should match value in checksums.txt
```

**macOS/Linux Verification** (automatic in install script):

```bash
#!/bin/bash
# Download
curl -L "$URL" -o "$ASSET"

# Download checksums
curl -L "${BASE_URL}/checksums.txt" -o checksums.txt

# Verify
if sha256sum --check checksums.txt --ignore-missing; then
    echo "✓ Checksum verified"
else
    echo "✗ Checksum verification failed"
    exit 1
fi

# Extract
tar xzf "$ASSET"
```

**Rationale**:
- Prevents corrupted downloads
- Detects tampering
- Industry best practice
- Automated in install script
- Manual verification documented

---

### 10. Release Checklist

**Pre-Release**:
- [ ] All tests passing (`go test ./...`)
- [ ] CLI validated on all platforms (macOS, Linux, Windows)
- [ ] Documentation updated (README, CLI_REFERENCE.md)
- [ ] CHANGELOG.md updated with changes
- [ ] Version number decided (semantic versioning)
- [ ] Migration guide written (if breaking changes)

**Release**:
- [ ] Update CHANGELOG.md
- [ ] Commit changes: `git commit -m "chore: prepare v0.2.0"`
- [ ] Push to main: `git push`
- [ ] Create tag: `git tag -a v0.2.0 -m "Release v0.2.0"`
- [ ] Push tag: `git push origin v0.2.0`
- [ ] Verify GitHub Actions workflow runs
- [ ] Verify release created on GitHub
- [ ] Test download and installation

**Post-Release**:
- [ ] Update plugin version (`.plugin-version`)
- [ ] Test plugin + CLI integration
- [ ] Announce release (if public)
- [ ] Update documentation site (if exists)
- [ ] Monitor for issues

---

## Consequences

### Positive

1. **Automation**: Tag push = complete release (no manual steps)
2. **Consistency**: Same process every release, no variation
3. **Cross-Platform**: All platforms built automatically
4. **Verification**: Checksums prevent corrupted downloads
5. **Easy Installation**: Homebrew + install script cover most users
6. **Version Clarity**: Single version across CLI, plugin, schemas
7. **Upgrade Path**: Clear process for staying up-to-date
8. **Industry Standard**: GoReleaser is proven, well-documented
9. **Low Maintenance**: Minimal manual intervention required
10. **Flexible**: Can add package managers later if needed

### Negative

1. **GitHub Dependency**: Requires GitHub for releases (acceptable trade-off)
2. **GoReleaser Learning**: Team needs to understand GoReleaser config
3. **Token Management**: Need HOMEBREW_TAP_GITHUB_TOKEN for tap updates
4. **Initial Setup**: Requires configuring GitHub Actions once
5. **Binary Size**: Embedded schemas increase size (~5-10MB) (acceptable)

### Mitigations

1. **GitHub Dependency**: Could add alternative distribution (cloudflare, CDN) if needed
2. **GoReleaser Learning**: Comprehensive documentation in this ADR
3. **Token Management**: Document token creation in setup guide
4. **Initial Setup**: Provide step-by-step setup instructions
5. **Binary Size**: Modern systems handle 10MB binaries easily

---

## Alternatives Considered

### Alternative 1: Manual Build + GitHub Release Web UI

**Approach**:
- Build binaries manually with Makefile
- Upload to GitHub via web interface
- Manual checksum generation

**Rejected**:
- Too error-prone (human mistakes)
- Tedious and time-consuming
- No Homebrew tap automation
- Inconsistent releases
- No automatic changelog

---

### Alternative 2: Docker-Based Distribution

**Approach**:
- Package CLI in Docker image
- Distribute via Docker Hub
- Users run via `docker run`

**Rejected**:
- Requires Docker runtime
- Slower than native binary
- Poor user experience for CLI tool
- Adds unnecessary complexity
- Installation instructions more complex

---

### Alternative 3: Platform-Specific Package Managers Only

**Approach**:
- apt, yum, Chocolatey as primary distribution
- No direct binary downloads

**Rejected**:
- Requires maintaining multiple package repos
- Time-consuming to set up
- Not all users have package managers configured
- Direct download needed for fallback anyway
- Homebrew + direct download is simpler

---

## Implementation Plan

### Phase 1: GoReleaser Setup
1. Create `.goreleaser.yml` configuration
2. Test local build: `goreleaser build --snapshot --clean`
3. Verify all platforms build successfully
4. Check binary sizes and embedded schemas

### Phase 2: GitHub Actions Integration
1. Create `.github/workflows/release.yml`
2. Configure GitHub secrets (HOMEBREW_TAP_GITHUB_TOKEN)
3. Test workflow with a test tag
4. Verify release creation and asset uploads

### Phase 3: Homebrew Tap
1. Create `your-org/homebrew-tap` repository
2. Configure tap in `.goreleaser.yml`
3. Test tap update on release
4. Verify formula installs correctly

### Phase 4: Installation Scripts
1. Write `install.sh` for Unix-like systems
2. Test on macOS (Intel and Apple Silicon)
3. Test on Linux (Ubuntu, Debian, Fedora)
4. Document Windows installation (PowerShell)

### Phase 5: Documentation
1. Update DISTRIBUTION.md with new instructions
2. Update CLI_REFERENCE.md with installation methods
3. Create RELEASING.md with release checklist
4. Add troubleshooting guide for installation issues

### Phase 6: Verification
1. Create test release (v0.1.0-rc1)
2. Test all installation methods
3. Verify checksums
4. Test on all platforms
5. Gather feedback, iterate

---

## Example Commands Reference

**Local Build (Testing)**:
```bash
goreleaser build --snapshot --clean
ls -lh dist/
```

**Local Release (Testing)**:
```bash
goreleaser release --snapshot --clean --skip-publish
```

**Create Release**:
```bash
# Update CHANGELOG
git add CHANGELOG.md
git commit -m "chore: prepare v0.2.0"
git push

# Tag and push
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0

# GitHub Actions handles the rest
```

**Install (User)**:
```bash
# Homebrew
brew tap your-org/tap
brew install sow

# Install script
curl -sSL https://raw.githubusercontent.com/your-org/sow/main/install.sh | bash

# Direct download
curl -L https://github.com/your-org/sow/releases/download/v0.2.0/sow_0.2.0_darwin_amd64.tar.gz | tar xz
sudo mv sow /usr/local/bin/
```

**Verify**:
```bash
sow --version
which sow
sow --help
```

---

## References

- [GoReleaser Documentation](https://goreleaser.com/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Homebrew Tap Documentation](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Semantic Versioning](https://semver.org/)
- [CLI_REFERENCE.md](/Users/josh/code/sow/docs/CLI_REFERENCE.md)
- [DISTRIBUTION.md](/Users/josh/code/sow/docs/DISTRIBUTION.md)
- [ADR 001: Go CLI Architecture](/Users/josh/code/sow/.sow/knowledge/adrs/001-go-cli-architecture.md)

---

**Decision Made By**: architect-1

**Review Status**: Ready for implementation
