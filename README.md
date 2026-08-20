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

### Via GoReleaser

Download the latest release from [GitHub Releases](https://github.com/iam-bkpl/code-radar/releases).

### Build from Source

```bash
go build -o code-radar
```

## Usage

### With CLI argument:
```bash
code-radar <commit-hash>
```

### Interactive mode:
```bash
code-radar
# Then enter the commit hash when prompted
```

## Example Output

```
  📡  CODE RADAR  

  ┌─ COMMIT DETAILS
  │
  │ Hash:     a1b2c3d4e5f6
  │ Message:  Add new feature
  │ Author:   John Doe
  │ Date:     2024-01-15 10:30:00 +0000
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
```

## Branch Categories

The tool automatically categorizes branches with color coding:
- `DEV` (purple) - develop, dev
- `QA` (yellow) - qa, test
- `UAT` (orange) - uat
- `STAGING` (cyan) - staging, stg
- `MASTER` (violet) - master, main
- `PROD` (red) - prod, production
- `OTHER` (gray) - everything else

## Features

- Color-coded environment tags
- Interactive or CLI argument mode
- Deployment pipeline visualization
- Clean, modern terminal UI using [lipgloss](https://github.com/charmbracelet/lipgloss)
- Cross-platform support (macOS, Linux, Windows)
- Auto-update via Homebrew

## Development

### Prerequisites

- Go 1.21+
- GoReleaser (for releases)

### Building

```bash
# Build locally
go build -o code-radar

# Build for all platforms
goreleaser build --clean
```

### Releasing

1. Create a GitHub token with `repo` scope
2. Set the token: `export HOMEBREW_TAP_GITHUB_TOKEN=your_token`
3. Create a new release tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. Run GoReleaser:
   ```bash
   goreleaser release --clean
   ```

## License

MIT License - see [LICENSE](LICENSE) for details.
