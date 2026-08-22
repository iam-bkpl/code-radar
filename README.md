# Code Radar

Track where your git commits have been deployed across branches.

## Purpose

When you merge code into `develop`, it can flow to multiple branches:
- `qa` → `uat` → `staging` → `master` → `prod`

Code Radar helps you visualize which branches contain a specific commit across your remote repositories.

## Installation

### Via Homebrew (Recommended)

```bash
brew tap iam-bkpl/tap
brew install code-radar
```

### Build from Source

```bash
go build -o code-radar
```

## Usage

### CLI

```bash
# Scan a commit
code-radar <commit-hash>

# Compact output
code-radar --short <commit-hash>

# JSON output (for scripting)
code-radar --json <commit-hash>

# Interactive mode
code-radar

# Start web UI
code-radar --web

# Generate config file
code-radar --init
```

### Web UI

```bash
code-radar --web
# Open http://localhost:9876 in your browser
```

Features:
- Scan commits via browser
- Edit environment configurations visually
- Color-coded branch tags

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help message |
| `--version` | `-v` | Show version |
| `--json` | `-j` | Output as JSON |
| `--short` | `-s` | Compact one-line output |
| `--web` | | Start web UI on port 9876 |
| `--init` | | Generate `.code-radar.yaml` config |

## Configuration

Create a `.code-radar.yaml` in your project root (or `~/.config/code-radar/config.yaml` globally):

```bash
code-radar --init
```

Default config:

```yaml
environments:
  - name: DEV
    pattern: [develop, dev]
    color: "#6272A4"
  - name: QA
    pattern: [qa, test, release]
    color: "#F1FA8C"
  - name: UAT
    pattern: [uat]
    color: "#FFB86C"
  - name: STAGING
    pattern: [staging, stg]
    color: "#8BE9FD"
  - name: MASTER
    pattern: [master, main]
    color: "#BD93F9"
  - name: PROD
    pattern: [prod, production]
    color: "#FF5555"
```

Customize to match your branch naming:

```yaml
environments:
  - name: FEATURE
    pattern: [feature, feat]
    color: "#50FA7B"
  - name: RELEASE
    pattern: [release, rel]
    color: "#FFB86C"
```

**Matching**: Branch names are matched case-insensitively using exact match or contains. First match wins.

## Example Output

```
  📡  CODE RADAR  

  ┌─ COMMIT DETAILS
  │
  Hash:        a1b2c3d4e5f6
  Message:     Add new feature
  Author:      John Doe
  Date:        2024-01-15 10:30:00 +0000
  │

  ┌─ BRANCHES (5 found)
  │
  ├─ → DEV develop
  ├─ → QA qa
  ├─ → UAT uat
  ├─ → STAGING staging
  └─ → MASTER master
  │
```

## Features

- Scans remote branches only (not local)
- Auto-fetches before scanning
- Configurable environments with custom patterns and colors
- Web UI with config editor
- JSON output for scripting
- Loading spinner in terminal
- Cross-platform (macOS, Linux, Windows)

## License

MIT License - see [LICENSE](LICENSE) for details.
