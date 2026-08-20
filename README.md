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
╔══════════════════════════════════════════════════════════════╗
║                    CODE COMMIT TRACKER                       ║
╠══════════════════════════════════════════════════════════════╣
║ Commit:  a1b2c3d4e5f6...                                    ║
║ Message: Add new feature                                     ║
║ Author:  John Doe                                            ║
║ Date:    2024-01-15 10:30:00 +0000                           ║
╠══════════════════════════════════════════════════════════════╣
║ Found in 5 branch(es):                                       ║
╠══════════════════════════════════════════════════════════════╣
║  [DEV] develop                                               ║
║  [QA] qa                                                     ║
║  [UAT] uat                                                   ║
║  [STAGING] staging                                           ║
║  [MASTER] master                                             ║
╚══════════════════════════════════════════════════════════════╝

Deployment Summary:
-------------------
  [DEV] develop
  [QA] qa
  [UAT] uat
  [STAGING] staging
  [MASTER] master
```

## Branch Categories

The tool automatically categorizes branches:
- `[DEV]` - develop, dev
- `[QA]` - qa, test
- `[UAT]` - uat
- `[STAGING]` - staging, stg
- `[MASTER]` - master, main
- `[PROD]` - prod, production
- `[OTHER]` - everything else
