# Readme that Delivers: A Compact Guide for LLMs

## Objectives

* Explain **what the project is, who it’s for, and how to use it—fast**. ([GitHub Docs][1])
* Keep the README a **high-signal landing page** that links to deeper docs (tutorials, how-tos, reference, explanations). ([docs.divio.com][2])

## Structure (in order)

1. **Overview** (1–3 sentences: what/why/for whom). ([GitHub Docs][1])
2. **Quick Start** (copy-paste install + “hello world” usage).
3. **Usage** (the 2–4 tasks most users do first).
4. **Configuration** (only the essential options with defaults).
5. **Troubleshooting** (top 3–5 issues → fixes).
6. **Links** to full docs (tutorials/how-tos/reference/explanations). ([docs.divio.com][2])
7. **Contributing / Security / License** (short blurbs; link to policy files). ([GitHub Docs][3])
8. **Changelog** (link; follow Keep a Changelog + mention SemVer). ([keepachangelog.com][4])

## Style Rules

* **Write for skimmers**: descriptive headings, short paragraphs, front-load the point.
* **Use active voice** and imperative verbs (“Run…”, “Add…”). ([Google for Developers][5])
* **Make commands runnable**; state prerequisites right before commands.
* **Link out** instead of bloating the README (place depth in docs). ([docs.divio.com][2])
* **Badges**: keep to essentials (CI, release, license) only.

## Positive / Negative Reinforcement

**Do (Positive):**

* **Lead with value**: “CLI to diff Kubernetes manifests with policy checks.” ([GitHub Docs][1])
* **Provide a 60-second proof**: one install + one command that works.
* **Separate concerns**: tutorial vs how-to vs reference vs explanation. ([docs.divio.com][2])
* **Use a human-readable changelog** and note SemVer. ([keepachangelog.com][4])

**Avoid (Negative):**

* ❌ Wall-of-text intros or marketing copy before the Quick Start.
* ❌ Mixing tutorial prose with API reference in the README (link instead). ([docs.divio.com][2])
* ❌ Long release notes inside the README (use CHANGELOG). ([keepachangelog.com][4])
* ❌ Passive voice and vague steps (“It can be installed by…”). ([Google for Developers][5])

---

## Example README (Modern, Minimal)

````markdown
# FluxDiff

CLI to diff Kubernetes manifests and flag policy violations before deploys.

## Quick start

```bash
# Prerequisites: kubectl v1.30+, Go 1.22+ (for source builds)
curl -L https://github.com/acme/fluxdiff/releases/latest/download/fluxdiff_$(uname -s)_$(uname -m) -o /usr/local/bin/fluxdiff
chmod +x /usr/local/bin/fluxdiff

# Hello world: diff two folders and print a summary
fluxdiff diff ./manifests/dev ./manifests/prod
````

## Usage

Common tasks:

```bash
# 1) Show only breaking changes (delete/replace)
fluxdiff diff a b --severity=breaking

# 2) Enforce policy (blocks noncompliant changes)
fluxdiff diff a b --policy ./policy.rego --fail-on=violation

# 3) Output formats for CI (table, json)
fluxdiff diff a b --output=json > report.json
```

More examples → **docs/how-to**.

## Configuration

| Option       | Default | Description                    |
| ------------ | ------- | ------------------------------ |
| `--policy`   | none    | Rego policy file(s) to enforce |
| `--severity` | all     | `all` | `breaking` | `safety`  |
| `--output`   | table   | `table` | `json`               |

See full reference → **docs/reference**.

## Troubleshooting

* `command not found`: Ensure the binary is executable and on `$PATH`.
* “kubeconfig not found”: Set `KUBECONFIG` or run `kubectl config view`.
* CI fails with `exit 2`: A violation was found; re-run with `--output=json` to inspect.

More fixes → **docs/faq**.

## Contributing

We welcome issues and PRs. See **CONTRIBUTING.md** for setup, style, and review process.

### Security

Please report vulnerabilities via **SECURITY.md** (private disclosure). We publish fixes and credits in the **CHANGELOG**.

## License

Apache-2.0. See **LICENSE**.

## Links

* Tutorials: **docs/tutorials**
* How-tos: **docs/how-to**
* Reference: **docs/reference**
* Explanations: **docs/explanations**
* Changelog: **CHANGELOG.md** (SemVer)

```
