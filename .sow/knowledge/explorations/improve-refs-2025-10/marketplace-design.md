# Marketplace Design (Homebrew-Style)

## Summary

Ref discovery via git-based marketplaces, modeled after Homebrew's tap system. A marketplace is any git repository containing `.sow/marketplace.yml`. No central infrastructure required. Versions discovered dynamically via registry API.

## Design Principles

**Git-native**: Marketplaces are just git repos with a manifest file.

**Decentralized**: Anyone can publish a marketplace (official, team, community).

**Version-agnostic**: Marketplace lists refs, not versions. Versions queried from registry on-demand.

**Familiar UX**: Homebrew-style `user/repo` shorthand for GitHub, full URLs for other hosts.

## Marketplace Manifest Schema

```yaml
# .sow/marketplace.yml
schema_version: "1.0.0"

marketplace:
  name: "Official Sow Refs"
  description: "Curated collection of community refs"
  homepage: "https://github.com/sow-project/refs"

refs:
  go-standards:
    registry: ghcr.io/sow-project/go-standards
    title: "Go Team Standards Template"
    description: "Template for Go coding standards and conventions"
    tags:
      - golang
      - conventions
      - guidelines
    classifications:
      - type: guidelines
        description: "Coding standards and best practices for Go"
      - type: code-templates
        description: "Template files for team standards documentation"

  python-style:
    registry: ghcr.io/sow-project/python-style
    title: "Python Style Guide Template"
    description: "PEP 8 based style guide with team customization examples"
    tags:
      - python
      - style
      - pep8
    classifications:
      - type: guidelines
        description: "Python coding conventions based on PEP 8"
```

## Field Definitions

**Marketplace metadata:**
- `schema_version`: Manifest format version ("1.0.0")
- `marketplace.name`: Human-readable marketplace name
- `marketplace.description`: What this marketplace provides
- `marketplace.homepage`: Link to docs/website (optional)

**Ref entries (map of name → metadata):**
- Map key: Ref name (e.g., `go-standards`)
  - Used for CLI references and symlink name
  - Must be unique within marketplace
  - Pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
- `registry`: OCI reference WITHOUT tag (e.g., `ghcr.io/org/repo`)
- `title`: Human-readable display name
- `description`: One-sentence summary
- `tags`: Search keywords (array)
- `classifications`: Content types (array of type+description objects)

**What's NOT included:**
- Version/tags (queried from registry dynamically)
- Full `.sow-ref.yaml` contents (pulled from image)

## Discovery Workflow

### Adding a Marketplace

```bash
# GitHub shorthand (default: github.com)
sow refs marketplace add sow-project/refs

# Explicit GitHub URL
sow refs marketplace add https://github.com/sow-project/refs

# Other git hosts
sow refs marketplace add https://gitlab.com/myorg/sow-refs

# With custom name
sow refs marketplace add myorg/refs --name team-refs
```

**Internally:**
1. Clone git repo to `~/.config/sow/marketplaces/<name>/`
2. Validate `.sow/marketplace.yml` exists and is valid
3. Add to `~/.config/sow/marketplaces/config.yaml`

### Searching Refs

```bash
# Search across all marketplaces
sow refs search golang

# Search specific marketplace
sow refs search golang --marketplace official

# Filter by classification
sow refs search --classification guidelines

# Show all refs in marketplace
sow refs browse --marketplace team-refs
```

**Internally:**
1. Read marketplace.yml from each configured marketplace
2. Filter by search terms (title, description, tags, classifications)
3. Display matching refs with metadata

**Output:**
```
Found 3 refs matching "golang":

go-standards (sow-project/refs)
  Registry: ghcr.io/sow-project/go-standards
  Description: Template for Go coding standards and conventions
  Tags: golang, conventions, guidelines
  Classifications: guidelines, code-templates
  Add with: sow refs add go-standards

  (Note: Name collision with myteam/refs/go-standards)
  Use: sow refs add sow-project/refs/go-standards

go-examples (community/refs)
  Registry: ghcr.io/awesome-go/examples
  Description: Runnable Go examples for common patterns
  Tags: golang, examples, patterns
  Classifications: code-examples, tutorial
  Add with: sow refs add go-examples
```

### Adding Refs

**Three resolution strategies:**

```bash
# 1. By name (searches marketplaces)
sow refs add go-standards

# If unique across all marketplaces:
# → Resolves to marketplace entry
# → Shows available versions
# → User selects version (or latest)
# → Installs to .sow/refs/go-standards

# 2. By marketplace/name (explicit marketplace)
sow refs add sow-project/refs/go-standards

# If name collision detected across marketplaces:
# → Prompts: "Multiple refs found for 'go-standards':
#            - sow-project/refs/go-standards
#            - myteam/refs/go-standards
#            Use marketplace/name syntax to disambiguate"

# 3. By full OCI reference (exact)
sow refs add ghcr.io/sow-project/go-standards:v3.0.0

# Bypasses marketplace lookup entirely
# → Direct registry pull
# → Uses .sow-ref.yaml from image for metadata
```

**Resolution process:**

```bash
# User runs: sow refs add go-standards

# Step 1: Search marketplaces for name "go-standards"
# Found in: sow-project/refs, myteam/refs

# Step 2: Detect collision
# Prompt user to disambiguate or show options:
# "Found 2 refs named 'go-standards':
#  1. sow-project/refs/go-standards (Official Sow Refs)
#     Go Team Standards Template
#  2. myteam/refs/go-standards (Platform Team Refs)
#     Team-specific Go standards
#
#  Select [1-2] or use marketplace/name syntax:"

# User selects or runs: sow refs add sow-project/refs/go-standards

# Step 3: Query available versions
# Registry: ghcr.io/sow-project/go-standards
# Versions: v3.0.0, v2.1.0, v2.0.0, latest

# Prompt: "Select version [default: latest]:"

# Step 4: Install
# Pull ghcr.io/sow-project/go-standards:v3.0.0
# Extract .sow-ref.yaml
# Create symlink .sow/refs/go-standards → cache
# Add to .sow/refs/index.json
```

**Internally:**
1. Parse input (name vs marketplace/name vs full URI)
2. If name: Search all enabled marketplaces
3. If collision: Prompt for disambiguation
4. If marketplace/name: Look up in specific marketplace
5. If full URI: Skip marketplace, query registry directly
6. Query registry tags API: `GET /v2/<name>/tags/list`
7. Display versions, user selects
8. Pull image, extract `.sow-ref.yaml`, install

## Default Marketplace

Sow ships with default marketplace configured:

```yaml
# ~/.config/sow/marketplaces/config.yaml (initial state)
default_marketplace: official

marketplaces:
  - name: official
    source: https://github.com/sow-project/refs
    enabled: true
```

Users can disable, add, or configure defaults.

## Marketplace Management

```bash
# List configured marketplaces
sow refs marketplace list

# Update marketplace (git pull)
sow refs marketplace update official

# Update all marketplaces
sow refs marketplace update --all

# Remove marketplace
sow refs marketplace remove team-refs

# Enable/disable marketplace
sow refs marketplace disable community
sow refs marketplace enable community
```

## Publishing a Marketplace

### Create Marketplace Repo

```bash
mkdir my-sow-refs
cd my-sow-refs

# Create manifest
cat > .sow/marketplace.yml <<EOF
schema_version: "1.0.0"

marketplace:
  name: "My Team Refs"
  description: "Internal team refs for platform engineering"

refs:
  go-standards:
    registry: ghcr.io/myorg/go-standards
    title: "Go Standards"
    description: "Team Go coding standards"
    tags: [golang, conventions]
    classifications:
      - type: guidelines
        description: "Go coding standards"
EOF

# Commit and push
git init
git add .sow/marketplace.yml
git commit -m "Initial marketplace"
git remote add origin https://github.com/myorg/sow-refs
git push -u origin main
```

### Share with Team

```bash
# Team members add marketplace
sow refs marketplace add myorg/sow-refs

# Search and install refs
sow refs search golang
sow refs add go-standards  # or myorg/sow-refs/go-standards if collision
# → Prompts for version
# → Installs to .sow/refs/go-standards
```

## Auto-Generated Marketplace (CI/CD)

For teams publishing many refs, automate marketplace updates:

```yaml
# .github/workflows/update-marketplace.yml
name: Update Marketplace

on:
  workflow_dispatch:
  repository_dispatch:
    types: [ref-published]

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Query published refs
        run: |
          # List all refs in ghcr.io/myorg
          # For each ref, pull manifest to get title/description/tags
          # Update .sow/marketplace.yml

      - name: Commit and push
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add .sow/marketplace.yml
          git commit -m "Update marketplace catalog"
          git push
```

**Trigger on ref publish:**
```yaml
# In ref publishing workflow
- name: Notify marketplace
  run: |
    curl -X POST \
      -H "Authorization: token ${{ secrets.PAT }}" \
      https://api.github.com/repos/myorg/sow-refs/dispatches \
      -d '{"event_type":"ref-published"}'
```

## Validation

Sow validates marketplace.yml schema:

```bash
# Validate marketplace manifest
sow refs marketplace validate .sow/marketplace.yml
```

**Validation rules:**
- `schema_version` required, must be valid semver
- `marketplace.name` and `marketplace.description` required
- `refs` must be a map with at least one entry
- Ref names (map keys) must match pattern: `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
- Each ref must have: `registry`, `title`, `description`, `tags`, `classifications`
- `registry` must NOT include tag (e.g., `ghcr.io/org/repo`, not `ghcr.io/org/repo:v1.0.0`)
- `classifications[].type` must be valid classification type
- `tags` must be non-empty array

## Marketplace Discovery

How do users find marketplaces?

**Official marketplace**: Documented in sow docs, configured by default

**Community marketplaces**:
- Listed in sow docs
- Shared via GitHub topics (e.g., `sow-marketplace` topic)
- Word of mouth, team wikis

**Awesome list**: Community-maintained `awesome-sow` repo listing marketplaces

## Trust and Security

**Marketplace trust:**
- Marketplaces are just metadata (registry URLs, descriptions)
- No code execution from marketplace itself
- Users should trust marketplace source (official, team, vetted community)

**Ref trust:**
- Pulling refs from OCI registry is where trust matters
- Use private registries for team refs (authenticated)
- Verify public ref sources before adding
- OCI image signing (future enhancement)

## Comparison to Homebrew

| Homebrew | Sow Refs |
|----------|----------|
| `brew tap` | `sow refs marketplace add` |
| `homebrew/core` (default tap) | `sow-project/refs` (default marketplace) |
| Formula files (Ruby) | Marketplace manifest (YAML) |
| Formula contains install logic | Manifest contains metadata only |
| Versions in formula | Versions from registry tags API |
| Bottles (pre-built binaries) | OCI images (pre-packaged refs) |

## Example: Team Adoption Flow

```bash
# Platform team creates marketplace
mkdir team-sow-refs
cd team-sow-refs

cat > .sow/marketplace.yml <<EOF
schema_version: "1.0.0"
marketplace:
  name: "Platform Team Refs"
  description: "Internal refs for platform engineering"
refs:
  go-standards:
    registry: ghcr.io/acme/go-standards
    title: "Go Standards"
    description: "ACME Go coding standards"
    tags: [golang, conventions]
    classifications:
      - type: guidelines
        description: "Go coding standards"
EOF

git init && git add . && git commit -m "Init"
git remote add origin https://github.com/acme/team-sow-refs
git push -u origin main

# Developer onboarding
sow refs marketplace add acme/team-sow-refs
sow refs search golang
sow refs add go-standards
# → Prompts for version (v2.1.0, v2.0.0, latest)
# → Selects v2.1.0
# → Ref appears in .sow/refs/go-standards
# → LLM agents can now use team standards
```

## Open Questions

1. **Marketplace updates**: Auto-update marketplaces periodically, or manual `sow refs marketplace update`?
2. **Conflict resolution**: What if two marketplaces list same registry with different metadata?
3. **Marketplace caching**: How long to cache marketplace.yml before re-fetching?
4. **Version recommendations**: Should marketplace suggest "recommended" version (e.g., `latest`, `stable`)?
5. **Deprecation**: How to mark refs as deprecated in marketplace?
