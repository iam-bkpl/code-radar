# Code Radar

Track where your git commits have been deployed across branches.

## Purpose

When you merge code into `develop`, it can flow to multiple branches:
- `qa` → `uat` → `staging` → `master` → `prod`

Code Radar helps you visualize which branches contain a specific commit.

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

```bash
# Full output with spinner
code-radar <commit-hash>

# JSON output (for scripting)
code-radar --json <commit-hash>

# Compact one-line output
code-radar --short <commit-hash>

# Interactive mode
code-radar
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help message |
| `--version` | `-v` | Show version information |
| `--json` | `-j` | Output as JSON |
| `--short` | `-s` | Compact one-line output |

## Example Output

### Full Output
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

  ┌─ DEPLOYMENT PIPELINE
  │
  ├─ DEV        develop
  ├─ QA         qa
  ├─ UAT        uat
  ├─ STAGING    staging
  └─ MASTER     master
  │

  ┌─ FLOW
  │
  DEV ──→ QA ──→ UAT ──→ STAGING ──→ MASTER
  │
```

### JSON Output
```json
{
  "hash": "a1b2c3d4e5f6",
  "message": "Add new feature",
  "author": "John Doe",
  "date": "2024-01-15 10:30:00 +0000",
  "branches": [
    { "name": "develop", "env": "DEV" },
    { "name": "qa", "env": "QA" },
    { "name": "master", "env": "MASTER" }
  ]
}
```

### Short Output
```
a1b2c3d4e5f6 Add new feature → DEV → QA → MASTER
```

## Branch Categories

Color-coded environment tags:
- `DEV` (purple) - develop, dev
- `QA` (yellow) - qa, test
- `UAT` (orange) - uat
- `STAGING` (cyan) - staging, stg
- `MASTER` (violet) - master, main
- `PROD` (red) - prod, production
- `OTHER` (gray) - everything else

## Features

- Loading spinner while scanning branches
- Color-coded environment tags
- JSON output for scripting/automation
- Compact one-line mode
- Deployment pipeline visualization with flow diagram
- Interactive or CLI argument mode
- Clean, modern terminal UI using [lipgloss](https://github.com/charmbracelet/lipgloss)
- Cross-platform support (macOS, Linux, Windows)

## License

MIT License - see [LICENSE](LICENSE) for details.
