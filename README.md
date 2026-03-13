# aipr

`aipr` is a small Go CLI that creates a GitHub PR from your current branch using commits that differ from a configured base branch.

## Features

- Repo-local base branch config in `.aipr.json`
- Default base branch is `master`
- Commit-based PR title and body generation
- PR creation via GitHub CLI (`gh`)

## Requirements

- `git`
- `gh` (authenticated: `gh auth login`)
- Go 1.22+

## Install globally

From this repository:

```bash
go install .
```

Make sure your Go bin directory is in `PATH` (typically `$(go env GOPATH)/bin`).

## Usage

Set the repo-level base branch:

```bash
aipr config base develop
```

Create a PR from the current branch:

```bash
aipr
```

## Behavior

When you run `aipr`:

1. It resolves the current git repo root.
2. It reads `.aipr.json` for `base`; falls back to `master`.
3. It finds commits in `<base>..HEAD`.
4. It builds a PR title/body from commit subjects.
5. It runs:

```bash
gh pr create --base <base> --head <current-branch> --title <generated-title> --body <generated-body>
```
