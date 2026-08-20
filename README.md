# Code Commit Tracker

Track where your git commits have been deployed across branches.

## Purpose

When you merge code into `develop`, it can flow to multiple branches:
- `qa` → `uat` → `staging` → `master` → `prod`

This tool helps you visualize which branches contain a specific commit.

## Installation

```bash
go build -o code-commit
```

## Usage

### With CLI argument:
```bash
./code-commit <commit-hash>
```

### Interactive mode:
```bash
./code-commit
# Then enter the commit hash when prompted
```

## Example Output

```
  📦  CODE COMMIT TRACKER  

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
