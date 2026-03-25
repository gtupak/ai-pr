package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

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
		return runCreatePR(createOptions{})
	}

	if args[0] == "config" {
		return runConfig(args[1:])
	}

	createOpts, err := parseCreateOptions(args)
	if err != nil {
		return err
	}
	return runCreatePR(createOpts)
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
	case "model":
		return runConfigModel(args[1:])
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

	if err := config.SetRepoBaseBranch(repoRoot, branch); err != nil {
		return err
	}
	cfgPath, err := config.GlobalPath()
	if err != nil {
		return err
	}

	fmt.Printf("saved base branch %q for repo %q to %s\n", branch, repoRoot, cfgPath)
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

func runConfigModel(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: aipr config model <openrouter-model>")
	}

	model := strings.TrimSpace(args[0])
	if model == "" {
		return errors.New("model cannot be empty")
	}

	cfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}
	cfg.OpenRouterModel = model
	if err := config.SaveGlobal(cfg); err != nil {
		return err
	}

	cfgPath, err := config.GlobalPath()
	if err != nil {
		return err
	}
	fmt.Printf("saved OpenRouter model %q to %s\n", model, cfgPath)
	return nil
}

type createOptions struct {
	HeadOwner string
}

func parseCreateOptions(args []string) (createOptions, error) {
	fs := flag.NewFlagSet("aipr", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var opts createOptions
	fs.StringVar(&opts.HeadOwner, "head-owner", "", "GitHub owner for fork head ref (uses owner:current-branch)")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, usage())
	}

	if err := fs.Parse(args); err != nil {
		return createOptions{}, err
	}
	if len(fs.Args()) > 0 {
		return createOptions{}, fmt.Errorf("unexpected arguments: %s\n\n%s", strings.Join(fs.Args(), " "), usage())
	}

	opts.HeadOwner = strings.TrimSpace(opts.HeadOwner)
	return opts, nil
}

func runCreatePR(opts createOptions) error {
	step("Booting PR rocket...")

	repoRoot, err := withLoaderValue("Finding git repository root", func() (string, error) {
		return git.RepoRoot()
	})
	if err != nil {
		return err
	}
	ok(fmt.Sprintf("Repo located: %s", repoRoot))

	step("Checking repository has commits")
	if !git.HasCommits(repoRoot) {
		fmt.Println("repository has no commits yet; create an initial commit first")
		return nil
	}
	ok("Commit history found")

	currentBranch, err := withLoaderValue("Reading current branch", func() (string, error) {
		return git.CurrentBranch(repoRoot)
	})
	if err != nil {
		return err
	}
	ok(fmt.Sprintf("Current branch: %s", currentBranch))
	headRef := currentBranch
	if opts.HeadOwner != "" {
		headRef = fmt.Sprintf("%s:%s", opts.HeadOwner, currentBranch)
		ok(fmt.Sprintf("Using fork head ref: %s", headRef))
	}

	base, err := withLoaderValue("Resolving base branch config for this repo", func() (string, error) {
		return config.GetRepoBaseBranch(repoRoot)
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(base) == "" {
		base = defaultBaseBranch
		ok(fmt.Sprintf("No custom base configured; using default %q", base))
	} else {
		ok(fmt.Sprintf("Configured base branch: %s", base))
	}

	baseRef, err := withLoaderValue(fmt.Sprintf("Verifying base branch reference for %q", base), func() (string, error) {
		return git.ResolveBaseRef(repoRoot, base)
	})
	if err != nil {
		return err
	}
	ok(fmt.Sprintf("Base branch reference ready: %s", baseRef))

	commits, err := withLoaderValue("Collecting commits that differ from base", func() ([]git.Commit, error) {
		return git.CommitsBetween(repoRoot, baseRef, "HEAD")
	})
	if err != nil {
		return err
	}
	if len(commits) == 0 {
		fmt.Printf("no commits differ from %q; nothing to create\n", base)
		return nil
	}
	ok(fmt.Sprintf("Found %d commit(s) to include", len(commits)))

	apiKey, err := withLoaderValue("Loading OpenRouter API key", func() (string, error) {
		return resolveOpenRouterAPIKey()
	})
	if err != nil {
		return err
	}
	ok("API key loaded")

	model, err := withLoaderValue("Loading OpenRouter model", func() (string, error) {
		return resolveOpenRouterModel()
	})
	if err != nil {
		return err
	}
	ok(fmt.Sprintf("Using model: %s", model))

	type prDraft struct {
		title string
		body  string
	}
	draft, err := withLoaderValue("Asking AI to craft PR title and description", func() (prDraft, error) {
		title, body, err := ai.GeneratePRTitleBody(apiKey, model, base, currentBranch, commits)
		if err != nil {
			return prDraft{}, err
		}
		return prDraft{title: title, body: body}, nil
	})
	if err != nil {
		return fmt.Errorf("AI generation failed: %w", err)
	}
	ok("AI drafted PR content")

	err = withLoader("Launching gh to open the PR", func() error {
		return gh.CreatePR(repoRoot, gh.CreatePROptions{
			BaseBranch: base,
			HeadBranch: headRef,
			Title:      draft.title,
			Body:       draft.body,
		})
	})
	if err != nil {
		return err
	}

	ok(fmt.Sprintf("Created PR from %q into %q", currentBranch, base))
	fmt.Println("[aipr] Mission complete. Time for celebratory coffee.")
	return nil
}

func usage() string {
	return `Usage:
  aipr
  aipr --head-owner <owner>
  aipr config base <branch>
  aipr config openrouter-api-key <api-key>
  aipr config model <openrouter-model>

Commands:
  (no args)                 Create a PR from current branch commits.
  --head-owner <owner>      Use owner:current-branch as gh head ref (fork workflow).
  config base <name>  Save default base branch for this repository.
  config openrouter-api-key <key>
                     Save a global OpenRouter API key.
  config model <name> Save a global OpenRouter model.`
}

func resolveOpenRouterAPIKey() (string, error) {
	cfg, err := config.LoadGlobal()
	if err != nil {
		return "", err
	}
	if key := strings.TrimSpace(cfg.OpenRouterAPIKey); key != "" {
		return key, nil
	}
	return "", fmt.Errorf("missing OpenRouter API key; set it with `aipr config openrouter-api-key <api-key>`")
}

func resolveOpenRouterModel() (string, error) {
	cfg, err := config.LoadGlobal()
	if err != nil {
		return "", err
	}
	if model := strings.TrimSpace(cfg.OpenRouterModel); model != "" {
		return model, nil
	}
	return ai.DefaultModel(), nil
}

func step(message string) {
	fmt.Printf("[aipr] %s...\n", message)
}

func ok(message string) {
	fmt.Printf("[aipr] %s\n", message)
}

func withLoader(message string, fn func() error) error {
	stop := startLoader(message)
	err := fn()
	stop(err)
	return err
}

func withLoaderValue[T any](message string, fn func() (T, error)) (T, error) {
	stop := startLoader(message)
	value, err := fn()
	stop(err)
	return value, err
}

func startLoader(message string) func(error) {
	done := make(chan struct{})
	go func() {
		frames := []rune{'-', '\\', '|', '/'}
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		idx := 0

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "\r[aipr] %s... %c", message, frames[idx%len(frames)])
				idx++
			}
		}
	}()

	return func(err error) {
		close(done)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\r[aipr] %s... failed\n", message)
			return
		}
		fmt.Fprintf(os.Stderr, "\r[aipr] %s... done\n", message)
	}
}
