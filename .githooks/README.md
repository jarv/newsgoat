# Git Hooks

This directory contains git hooks for the newsgoat repository.

## Setup

To enable the git hooks, run:

```bash
mise run setup-hooks
```

Or manually configure git:

```bash
git config core.hooksPath .githooks
```

## Available Hooks

### pre-commit

Runs `mise run lint` before each commit to ensure code quality. The commit will be rejected if the linter finds any issues.

To bypass the hook temporarily (not recommended):

```bash
git commit --no-verify
```

## Requirements

- [mise](https://mise.jdx.dev/) must be installed
- golangci-lint must be available (installed via mise)
