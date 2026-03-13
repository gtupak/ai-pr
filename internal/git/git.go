package git

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	recordSep = "\x1e"
	fieldSep  = "\x1f"
)

type Commit struct {
	Subject string
	Body    string
}

func RepoRoot() (string, error) {
	out, err := runGit("", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("unable to find git repo root: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func CurrentBranch(repoRoot string) (string, error) {
	out, err := runGit(repoRoot, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("resolve current branch: %w", err)
	}
	branch := strings.TrimSpace(out)
	if branch == "" || branch == "HEAD" {
		return "", fmt.Errorf("detached HEAD is not supported")
	}
	return branch, nil
}

func HasCommits(repoRoot string) bool {
	return verifyRef(repoRoot, "HEAD") == nil
}

func ResolveBaseRef(repoRoot, base string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		return "", fmt.Errorf("base branch cannot be empty")
	}

	if err := verifyRef(repoRoot, "refs/heads/"+base); err == nil {
		return base, nil
	}
	if err := verifyRef(repoRoot, "refs/remotes/origin/"+base); err == nil {
		return "origin/" + base, nil
	}
	return "", fmt.Errorf("base branch %q not found locally or in origin", base)
}

func CommitsBetween(repoRoot, from, to string) ([]Commit, error) {
	rangeRef := fmt.Sprintf("%s..%s", from, to)
	out, err := runGit(
		repoRoot,
		"log",
		"--reverse",
		"--pretty=format:%s"+fieldSep+"%b"+recordSep,
		rangeRef,
	)
	if err != nil {
		return nil, fmt.Errorf("collect commits for %s: %w", rangeRef, err)
	}

	rawRecords := strings.Split(out, recordSep)
	commits := make([]Commit, 0, len(rawRecords))
	for _, rec := range rawRecords {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		parts := strings.SplitN(rec, fieldSep, 2)

		subject := strings.TrimSpace(parts[0])
		body := ""
		if len(parts) > 1 {
			body = strings.TrimSpace(parts[1])
		}
		if subject == "" {
			continue
		}
		commits = append(commits, Commit{
			Subject: subject,
			Body:    body,
		})
	}

	return commits, nil
}

func verifyRef(repoRoot, ref string) error {
	_, err := runGit(repoRoot, "rev-parse", "--verify", "--quiet", ref)
	return err
}

func runGit(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if repoRoot != "" {
		cmd.Dir = repoRoot
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return string(out), nil
}
