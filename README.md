# OpenRouter AI PRs

`aipr` creates a GitHub pull request from your current branch using AI-generated title/body based on commits that differ from your configured base branch.

## What it does

- Generates PR title and description with OpenRouter
- Uses `gh pr create` to open the PR
- Stores config globally in `~/.aipr/config.json`
- Lets you set a different base branch per repo path

## Requirements

- `git`
- `gh` (authenticated with `gh auth login`)
- Go 1.22+

## Install

From this repo:

```bash
go install .
```

Make sure your Go bin path is in `PATH` (usually `$(go env GOPATH)/bin`).

## Quick start

1) Set your OpenRouter API key:

```bash
aipr config openrouter-api-key <your-api-key>
```

2) (Optional) Set the model globally:

```bash
aipr config model qwen/qwen3.5-flash-02-23
```

3) Set base branch for the current repo:

```bash
aipr config base develop
```

4) Run it:

```bash
aipr
```

## Commands

```bash
aipr
aipr config base <branch>
aipr config openrouter-api-key <api-key>
aipr config model <openrouter-model>
```

## Example

```bash
# On feature branch: feat/better-readme
aipr config base master
aipr config model qwen/qwen3.5-flash-02-23
aipr
```

Expected flow:

- `aipr` finds commits in `master..feat/better-readme`
- sends commit context to OpenRouter
- receives generated PR title/body
- runs `gh pr create --base master --head feat/better-readme ...`

## Notes

- If no commits differ from base, no PR is created.
- If AI generation fails, `aipr` exits with an error (no fallback).
- Default base branch is `master` if none is configured for the repo.
