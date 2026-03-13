package gh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CreatePROptions struct {
	BaseBranch string
	HeadBranch string
	Title      string
	Body       string
}

func CreatePR(repoRoot string, opts CreatePROptions) error {
	if strings.TrimSpace(opts.BaseBranch) == "" {
		return fmt.Errorf("base branch is required")
	}
	if strings.TrimSpace(opts.HeadBranch) == "" {
		return fmt.Errorf("head branch is required")
	}
	if strings.TrimSpace(opts.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI `gh` is required but was not found in PATH")
	}

	args := []string{
		"pr", "create",
		"--base", opts.BaseBranch,
		"--head", opts.HeadBranch,
		"--title", opts.Title,
		"--body", opts.Body,
	}

	cmd := exec.Command("gh", args...)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh pr create failed: %w (ensure `gh auth status` is healthy)", err)
	}
	return nil
}
