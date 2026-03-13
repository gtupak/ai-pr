package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"aipr/internal/ai"
	"aipr/internal/config"
	"aipr/internal/gh"
	"aipr/internal/git"
)

const defaultBaseBranch = "master"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return runCreatePR()
	}

	if args[0] == "config" {
		return runConfig(args[1:])
	}

	return fmt.Errorf("unknown command %q\n\n%s", args[0], usage())
}

func runConfig(args []string) error {
	if len(args) < 1 {
		return errors.New("missing config subcommand\n\n" + usage())
	}

	switch args[0] {
	case "base":
		return runConfigBase(args[1:])
	case "openrouter-api-key":
		return runConfigOpenRouterAPIKey(args[1:])
	default:
		return fmt.Errorf("unknown config subcommand %q\n\n%s", args[0], usage())
	}
}

func runConfigBase(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: aipr config base <branch>")
	}

	branch := strings.TrimSpace(args[0])
	if branch == "" {
		return errors.New("branch name cannot be empty")
	}

	repoRoot, err := git.RepoRoot()
	if err != nil {
		return err
	}

	cfgPath := config.Path(repoRoot)
	cfg, err := config.Load(repoRoot)
	if err != nil {
		return err
	}

	cfg.Base = branch
	if err := config.Save(repoRoot, cfg); err != nil {
		return err
	}

	fmt.Printf("saved base branch %q to %s\n", branch, cfgPath)
	return nil
}

func runConfigOpenRouterAPIKey(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: aipr config openrouter-api-key <api-key>")
	}

	apiKey := strings.TrimSpace(args[0])
	if apiKey == "" {
		return errors.New("api key cannot be empty")
	}

	cfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}
	cfg.OpenRouterAPIKey = apiKey
	if err := config.SaveGlobal(cfg); err != nil {
		return err
	}

	cfgPath, err := config.GlobalPath()
	if err != nil {
		return err
	}
	fmt.Printf("saved OpenRouter API key to %s\n", cfgPath)
	return nil
}

func runCreatePR() error {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return err
	}
	if !git.HasCommits(repoRoot) {
		fmt.Println("repository has no commits yet; create an initial commit first")
		return nil
	}

	currentBranch, err := git.CurrentBranch(repoRoot)
	if err != nil {
		return err
	}

	cfg, err := config.Load(repoRoot)
	if err != nil {
		return err
	}

	base := cfg.Base
	if strings.TrimSpace(base) == "" {
		base = defaultBaseBranch
	}

	baseRef, err := git.ResolveBaseRef(repoRoot, base)
	if err != nil {
		return err
	}

	commits, err := git.CommitsBetween(repoRoot, baseRef, "HEAD")
	if err != nil {
		return err
	}
	if len(commits) == 0 {
		fmt.Printf("no commits differ from %q; nothing to create\n", base)
		return nil
	}

	apiKey, err := resolveOpenRouterAPIKey()
	if err != nil {
		return err
	}

	title, body, err := ai.GeneratePRTitleBody(apiKey, base, currentBranch, commits)
	if err != nil {
		return fmt.Errorf("AI generation failed: %w", err)
	}

	if err := gh.CreatePR(repoRoot, gh.CreatePROptions{
		BaseBranch: base,
		HeadBranch: currentBranch,
		Title:      title,
		Body:       body,
	}); err != nil {
		return err
	}

	fmt.Printf("created PR from %q into %q\n", currentBranch, base)
	return nil
}

func usage() string {
	return `Usage:
  aipr
  aipr config base <branch>
  aipr config openrouter-api-key <api-key>

Commands:
  (no args)           Create a PR from current branch commits.
  config base <name>  Save default base branch for this repository.
  config openrouter-api-key <key>
                     Save a global OpenRouter API key.`
}

func resolveOpenRouterAPIKey() (string, error) {
	cfg, err := config.LoadGlobal()
	if err != nil {
		return "", err
	}
	if key := strings.TrimSpace(cfg.OpenRouterAPIKey); key != "" {
		return key, nil
	}

	if key := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")); key != "" {
		return key, nil
	}
	return "", fmt.Errorf("missing OpenRouter API key; set it with `aipr config openrouter-api-key <api-key>`")
}
