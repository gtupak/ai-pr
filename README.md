# aipr

`aipr` is a small Go CLI that creates a GitHub PR from your current branch using commits that differ from a configured base branch, and uses OpenRouter to generate the PR title/body.

## Features

- Repo-local base branch config in `.aipr.json`
- Default base branch is `master`
- AI-generated PR title and body via OpenRouter
- PR creation via GitHub CLI (`gh`)

## Requirements

- `git`
- `gh` (authenticated: `gh auth login`)
- Go 1.22+

Optional:
- `OPENROUTER_API_KEY` environment variable (fallback if global config is not set)
- `AIPR_OPENROUTER_MODEL` to override the default model (`openai/gpt-4o-mini`)

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

Set the global OpenRouter API key:

```bash
aipr config openrouter-api-key <your-api-key>
```

Create a PR from the current branch:

```bash
aipr
```

## Behavior

When you run `aipr`:

1. It resolves the current git repo root.
2. It reads `.aipr.json` for `base`; falls back to `master`.
3. It reads the global OpenRouter key from `~/.aipr/config.json` (or `OPENROUTER_API_KEY` fallback).
4. It finds commits in `<base>..HEAD`.
5. It sends commit history to OpenRouter to generate a PR title/body.
6. If AI generation fails, it exits with an error and does not create a PR.
7. It runs (only when AI generation succeeds):

```bash
gh pr create --base <base> --head <current-branch> --title <generated-title> --body <generated-body>
```
