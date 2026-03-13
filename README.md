# aipr

`aipr` is a small Go CLI that creates a GitHub PR from your current branch using commits that differ from a configured base branch, and uses OpenRouter to generate the PR title/body.

## Features

- Per-repo base branch mapping stored globally in `~/.aipr/config.json`
- Default base branch is `master`
- AI-generated PR title and body via OpenRouter
- PR creation via GitHub CLI (`gh`)

## Requirements

- `git`
- `gh` (authenticated: `gh auth login`)
- Go 1.22+
- Run `aipr config openrouter-api-key <your-api-key>`

## Install globally

From this repository:

```bash
go install .
```

Make sure your Go bin directory is in `PATH` (typically `$(go env GOPATH)/bin`).

## Usage

Set the base branch for the current repo (stored globally by repo path):

```bash
aipr config base develop
```

Set the global OpenRouter API key:

```bash
aipr config openrouter-api-key <your-api-key>
```

Set the global OpenRouter model:

```bash
aipr config model qwen/qwen3.5-flash-02-23
```

Create a PR from the current branch:

```bash
aipr
```

## Behavior

When you run `aipr`:

1. It resolves the current git repo root.
2. It looks up the repo path in `~/.aipr/config.json` for `base`; falls back to `master`.
3. It reads the global OpenRouter key from `~/.aipr/config.json`.
4. It finds commits in `<base>..HEAD`.
5. It resolves the OpenRouter model (global config, then default `qwen/qwen3.5-flash-02-23`).
6. It sends commit history to OpenRouter to generate a PR title/body.
7. If AI generation fails, it exits with an error and does not create a PR.
8. It runs (only when AI generation succeeds):

```bash
gh pr create --base <base> --head <current-branch> --title <generated-title> --body <generated-body>
```
